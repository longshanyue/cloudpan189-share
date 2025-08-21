package bus

import (
	"context"
	"errors"
	"fmt"
	"github.com/xxcheng123/cloudpan189-share/configs"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/messagebus"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"sync"
)

type mediaBusWorker struct {
	db     *gorm.DB
	logger *zap.Logger

	lock sync.Mutex
}

var (
	mediaBusWorkerInstance *mediaBusWorker
)

func (w *mediaBusWorker) Register() error {
	if !w.lock.TryLock() {
		return ErrBusRegistered
	}

	mediaBus := messagebus.New(messagebus.Config{
		WorkerCount: 8,
		BufferSize:  1024,
	}, w.logger.With(zap.String("bus", "media_bus")))

	mediaBus.Subscribe(TopicMediaAddStrmFile, func(ctx context.Context, data interface{}) {
		if req, ok := data.(TopicMediaAddStrmFileRequest); ok {
			log, _ := newLog(w.db, fmt.Sprintf("添加 strm 文件:%s", req.Path))

			if err := w.addStrm(ctx, req.FileID, req.Path); err != nil {
				w.logger.Error("添加 strm 文件失败", zap.Error(err))

				_ = log.End(fmt.Sprintf("添加 strm 文件失败: %s", err.Error()))
			} else {
				_ = log.End("添加成功")
			}
		}
	})

	mediaBus.Subscribe(TopicMediaDeleteLinkFile, func(ctx context.Context, data interface{}) {
		if req, ok := data.(TopicMediaDeleteLinkFileRequest); ok {
			log, _ := newLog(w.db, fmt.Sprintf("删除文件id:%d关联的媒体文件", req.FileID))
			if err := w.delMedia(ctx, req.FileID); err != nil {
				w.logger.Error("删除文件关联的媒体文件失败", zap.Error(err))

				_ = log.End(fmt.Sprintf("删除文件关联的媒体文件失败: %s", err.Error()))
			} else {
				_ = log.End("删除成功")
			}
		}
	})

	mediaBus.Subscribe(TopicMediaClearEmptyDir, func(ctx context.Context, data interface{}) {
		if _, ok := data.(TopicMediaClearEmptyDirRequest); ok {
			log, _ := newLog(w.db, "清理空文件夹")

			if count, err := w.clearEmptyDirs(ctx); err != nil {
				w.logger.Error("清理空文件夹失败", zap.Error(err))

				_ = log.End(fmt.Sprintf("清理空文件夹失败: %s", err.Error()))
			} else {
				_ = log.End(fmt.Sprintf("清理完成，删除了 %d 个空文件夹", count))
			}
		}
	})

	mediaBus.Subscribe(TopicMediaClearAllMedia, func(ctx context.Context, data interface{}) {
		if req, ok := data.(TopicMediaClearAllMediaRequest); ok {
			log, _ := newLog(w.db, "清理媒体文件")

			if count, err := w.clearAllMedia(ctx, req.MediaTypes...); err != nil {
				w.logger.Error("清理媒体文件失败", zap.Error(err))

				_ = log.End(fmt.Sprintf("清理媒体文件失败: %s", err.Error()))
			} else {
				_ = log.End(fmt.Sprintf("清理完成，删除了 %d 个媒体文件", count))
			}
		}
	})

	shared.MediaBus = mediaBus

	return nil
}

func (w *mediaBusWorker) addStrm(ctx context.Context, fileId int64, path string) (err error) {
	name := filepath.Base(path)

	content := generateDownloadURLWithNeverExpire(fileId)

	if err = w.write(path, []byte(content)); err != nil {
		return err
	}

	hash := utils.MD5([]byte(content))

	file := &models.MediaFile{
		FID:       fileId,
		Name:      name,
		Path:      path,
		Hash:      hash,
		Size:      int64(len(content)),
		MediaType: models.MediaTypeStrm,
	}

	return w.getDB(ctx).Create(file).Error
}

func (w *mediaBusWorker) delMedia(ctx context.Context, fileId int64, mediaTypes ...models.MediaType) error {
	files := make([]*models.MediaFile, 0)

	query := w.getDB(ctx).Where("fid = ?", fileId)
	if len(mediaTypes) > 0 {
		query = query.Where("media_type in ?", mediaTypes)
	}

	if err := query.Find(&files).Error; err != nil {
		return err
	}

	var errs []error

	for _, file := range files {
		if err := w.getDB(ctx).Delete(file).Error; err != nil {
			errs = append(errs, err)
		}

		_ = os.Remove(configs.GetConfig().MediaJoinPath(file.Path))
	}

	return errors.Join(errs...)
}

func (w *mediaBusWorker) write(path string, content []byte) error {
	fullPath := configs.GetConfig().MediaJoinPath(path)

	if _, err := os.Stat(fullPath); err == nil {
		// 文件已存在，直接返回
		return os.ErrExist
	} else if !os.IsNotExist(err) {
		// 其他错误（权限问题等）
		return fmt.Errorf("检查文件状态失败 %s: %v", fullPath, err)
	}

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败 %s: %v", dir, err)
	}

	// 写入文件
	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		return fmt.Errorf("写入文件失败 %s: %v", fullPath, err)
	}

	return nil
}

func (w *mediaBusWorker) Close() {
	if !w.lock.TryLock() {
		w.lock.Unlock()
	}
}

func (w *mediaBusWorker) clearEmptyDirs(ctx context.Context) (int, error) {
	mediaDir := configs.GetConfig().MediaDir

	// 检查媒体目录是否存在
	if _, err := os.Stat(mediaDir); os.IsNotExist(err) {
		w.logger.Info("媒体目录不存在，跳过清理", zap.String("media_dir", mediaDir))
		return 0, nil
	} else if err != nil {
		return 0, fmt.Errorf("检查媒体目录失败: %v", err)
	}

	deletedCount := 0

	err := w.clearEmptyDirsRecursive(mediaDir, &deletedCount)
	if err != nil {
		return deletedCount, err
	}

	w.logger.Info("清理空文件夹完成",
		zap.String("media_dir", mediaDir),
		zap.Int("deleted_count", deletedCount),
	)

	return deletedCount, nil
}

func (w *mediaBusWorker) clearEmptyDirsRecursive(dirPath string, deletedCount *int) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("读取目录失败 %s: %v", dirPath, err)
	}

	// 先递归处理子目录
	for _, entry := range entries {
		if entry.IsDir() {
			subDirPath := filepath.Join(dirPath, entry.Name())
			if err := w.clearEmptyDirsRecursive(subDirPath, deletedCount); err != nil {
				w.logger.Warn("清理子目录失败",
					zap.String("dir", subDirPath),
					zap.Error(err))
			}
		}
	}

	// 再次检查当前目录是否为空（子目录可能已被删除）
	entries, err = os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("重新读取目录失败 %s: %v", dirPath, err)
	}

	// 如果目录为空且不是根媒体目录，则删除
	if len(entries) == 0 && dirPath != configs.GetConfig().MediaDir {
		if err := os.Remove(dirPath); err != nil {
			w.logger.Warn("删除空目录失败",
				zap.String("dir", dirPath),
				zap.Error(err))
			return fmt.Errorf("删除空目录失败 %s: %v", dirPath, err)
		}

		w.logger.Debug("删除空目录", zap.String("dir", dirPath))
		*deletedCount++
	}

	return nil
}

func (w *mediaBusWorker) clearAllMedia(ctx context.Context, mediaTypes ...models.MediaType) (count int64, err error) {
	defer func() {
		if err != nil {
			w.logger.Error("清理所有媒体文件失败", zap.Error(err))
		} else {
			w.logger.Info("清理所有媒体文件成功", zap.Int64("count", count))

			// 清理所有空目录
			_, _ = w.clearEmptyDirs(ctx)
		}
	}()

	buildQuery := func() *gorm.DB {
		query := w.getDB(ctx).Limit(1000)

		if len(mediaTypes) > 0 {
			query = query.Where("media_type in ?", mediaTypes)
		}

		return query
	}

	for {
		files := make([]*models.MediaFile, 0)
		if err = buildQuery().Find(&files).Error; err != nil {
			return count, err
		}

		if len(files) == 0 {
			return count, nil
		}

		for _, file := range files {
			_ = os.Remove(configs.GetConfig().MediaJoinPath(file.Path))
		}

		fileIds := make([]int64, 0)
		for _, file := range files {
			fileIds = append(fileIds, file.ID)
		}

		if err = w.getDB(ctx).Where("id IN (?)", fileIds).Delete(new(models.MediaFile)).Error; err != nil {
			return count, err
		}
	}
}

func (w *mediaBusWorker) getDB(ctx context.Context) *gorm.DB {
	return w.db.WithContext(ctx).Model(&models.MediaFile{})
}
