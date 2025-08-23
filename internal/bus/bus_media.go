package bus

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xxcheng123/cloudpan189-share/configs"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (w *busWorker) addStrmMediaFile(ctx context.Context, fileId int64, path string) (err error) {
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

	return w.withLock(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Create(file)
	}).Error
}

func (w *busWorker) delMediaFile(ctx context.Context, fileId int64, mediaTypes ...models.MediaType) error {
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
		if err := w.withLock(ctx, func(db *gorm.DB) *gorm.DB {
			return db.Delete(file)
		}).Error; err != nil {
			errs = append(errs, err)
		}

		_ = os.Remove(configs.GetConfig().MediaJoinPath(file.Path))
	}

	return errors.Join(errs...)
}

func (w *busWorker) clearEmptyDirs(ctx context.Context) (int64, error) {
	mediaDir := configs.GetConfig().MediaDir

	// 检查媒体目录是否存在
	if _, err := os.Stat(mediaDir); os.IsNotExist(err) {
		w.logger.Info("媒体目录不存在，跳过清理", zap.String("media_dir", mediaDir))
		return 0, nil
	} else if err != nil {
		return 0, fmt.Errorf("检查媒体目录失败: %v", err)
	}

	var deletedCount int64

	err := w.clearEmptyDirsRecursive(mediaDir, &deletedCount)
	if err != nil {
		return deletedCount, err
	}

	w.logger.Info("清理空文件夹完成",
		zap.String("media_dir", mediaDir),
		zap.Int64("deleted_count", deletedCount),
	)

	return deletedCount, nil
}

func (w *busWorker) clearEmptyDirsRecursive(dirPath string, deletedCount *int64) error {
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

func (w *busWorker) clearAllMediaFiles(ctx context.Context, mediaTypes ...models.MediaType) (count int64, err error) {
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

		if err = w.withLock(ctx, func(db *gorm.DB) *gorm.DB {
			return db.Where("id IN (?)", fileIds).Delete(new(models.MediaFile))
		}).Error; err != nil {
			return count, err
		}
	}
}

func (w *busWorker) write(path string, content []byte) error {
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
