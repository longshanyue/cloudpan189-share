package bus

import (
	"context"
	errors2 "errors"
	"fmt"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/xxcheng123/cloudpan189-interface/client"
	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/messagebus"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type fileBusWorker struct {
	db     *gorm.DB
	logger *zap.Logger
	client client.Client

	lock sync.Mutex
}

var (
	fileBusWorkerInstance *fileBusWorker

	allowAutoDelOsTypes = []string{
		models.OsTypeFolder,
		models.OsTypeFile,
		models.OsTypeSubscribe,
		models.OsTypeSubscribeShare,
		models.OsTypeShare,
	}
)

func (w *fileBusWorker) Register() error {
	if !w.lock.TryLock() {
		return ErrBusRegistered
	}

	fileBus := messagebus.New(messagebus.DefaultConfig(), w.logger.With(zap.String("bus", "file_bus")))

	fileBus.Subscribe(TopicFileRefreshFile, func(ctx context.Context, data interface{}) {
		if req, ok := data.(TopicFileRefreshRequest); ok {
			log, _ := newLog(w.db, fmt.Sprintf("扫描顶层文件ID:%d及子文件", req.FID))

			if err := w.scanFile(ctx, req.FID, req.Deep); err != nil {
				w.logger.Error("扫描文件失败", zap.Error(err))

				_ = log.End(fmt.Sprintf("扫描文件失败: %s", err.Error()))
			} else {
				_ = log.End("扫描成功")
			}
		}
	})

	fileBus.Subscribe(TopicFileDeleteFile, func(ctx context.Context, data interface{}) {
		if req, ok := data.(TopicFileDeleteRequest); ok {
			log, _ := newLog(w.db, fmt.Sprintf("删除文件ID:%d", req.FID))

			if err := w.delete(ctx, req.FID); err != nil {
				w.logger.Error("删除文件失败", zap.Error(err))

				_ = log.End(fmt.Sprintf("删除文件失败: %s", err.Error()))
			} else {
				_ = log.End("删除成功")
			}
		}
	})

	fileBus.Subscribe(TopicFileScanTop, func(ctx context.Context, data interface{}) {
		log, _ := newLog(w.db, "扫描所有文件")

		err := w.scanTop(ctx)
		if err != nil {
			w.logger.Error("扫描所有文件失败", zap.Error(err))

			_ = log.End(fmt.Sprintf("扫描所有文件失败: %s", err.Error()))
		} else {
			_ = log.End("扫描成功")
		}
	})

	fileBus.Subscribe(TopicFileRebuildMediaFile, func(ctx context.Context, data interface{}) {
		if req, ok := data.(TopicFileRebuildMediaFileRequest); ok {
			log, _ := newLog(w.db, "（重建）媒体文件")

			mediaReq := TopicMediaClearAllMediaRequest{}
			if len(req.MediaTypes) > 0 {
				mediaReq.MediaTypes = req.MediaTypes
			}

			if err := shared.MediaBus.PublishSync(ctx, TopicMediaClearAllMedia, mediaReq); err != nil {
				w.logger.Error("（重建）媒体文件", zap.Error(err))

				_ = log.End(fmt.Sprintf("删除媒体文件执行失败: %s", err.Error()))
			}

			if count, err := w.buildMediaFile(ctx, 0); err != nil {
				w.logger.Error("（重建）媒体文件", zap.Error(err))

				_ = log.End(fmt.Sprintf("执行失败: %s", err.Error()))
			} else {
				_ = log.End(fmt.Sprintf("重建完成，处理了 %d 个文件", count))
			}
		}
	})

	shared.FileBus = fileBus

	return nil
}

type File = *models.VirtualFile

type walkFunc func(ctx context.Context, file File, childrenFiles []File) (nextWalkFiles []File)

func (w *fileBusWorker) scanFile(ctx context.Context, rootId int64, deep bool) error {
	return w.walk(ctx, rootId, func(ctx context.Context, file File, oldFiles []File) (nextWalkFiles []File) {
		var (
			newFiles = make([]File, 0)
			err      error
		)

		switch file.OsType {
		case models.OsTypeSubscribe:
			newFiles, err = w.getSubscribeUserFiles(ctx, file)
		case models.OsTypeSubscribeShare:
			newFiles, err = w.getSubscribeShareFiles(ctx, file)
		case models.OsTypeShare:
			newFiles, err = w.getShareFiles(ctx, file)
		default:
			return nil
		}

		if err != nil {
			w.logger.Error("获取文件列表失败", zap.Error(err))

			return nil
		}

		// 数据准备

		// 创建映射表，用于快速查找
		var (
			newFileMap = make(map[string]File)
			oldFileMap = make(map[string]File)
		)

		for _, item := range newFiles {
			key := item.Name
			newFileMap[key] = item
		}

		for _, item := range oldFiles {
			key := item.Name
			oldFileMap[key] = item
		}

		var (
			// 新增的文件
			filesToCreate []File
			// 待删除的文件
			filesToDelete []File
			// 找出需要更新的文件
			filesToUpdateMap = map[int64]map[string]any{}
			// 需要深度扫描的文件
			filesToDeep []File
		)

		// 遍历扫描到的文件，找出新增和更新的文件
		for name, newFile := range newFileMap {
			if oldFile, exists := oldFileMap[name]; exists {
				// 文件存在，检查是否需要更新（通过Rev比较）
				if oldFile.Rev != newFile.Rev {
					w.logger.Debug("文件存在差异 - rev changed",
						zap.String("parent", file.Name),
						zap.String("file_name", name),
						zap.String("old_rev", oldFile.Rev),
						zap.String("new_rev", newFile.Rev))

					mp := map[string]any{
						"name":        newFile.Name,
						"rev":         newFile.Rev,
						"size":        newFile.Size,
						"hash":        strings.ToLower(newFile.Hash),
						"modify_date": newFile.ModifyDate,
					}

					filesToUpdateMap[oldFile.ID] = mp
				} else if oldFile.IsFolder == 1 && deep {
					filesToDeep = append(filesToDeep, oldFile)
				}
			} else {
				w.logger.Debug("发现新文件",
					zap.String("parent", file.Name),
					zap.String("file_name", name),
					zap.String("rev", newFile.Rev))
				// 文件不存在，需要新增
				newFile.ParentId = file.ID
				filesToCreate = append(filesToCreate, newFile)
			}
		}

		// 遍历数据库中的文件，找出需要删除的文件
		for name, dbFile := range oldFileMap {
			if _, exists := newFileMap[name]; !exists &&
				dbFile.IsTop != 1 &&
				// 允许自动删除的文件类型
				lo.IndexOf(allowAutoDelOsTypes, dbFile.OsType) > -1 {
				w.logger.Debug("文件不存在 - 删除",
					zap.String("parent", file.Name),
					zap.String("file_name", name),
					zap.Int64("file_id", dbFile.ID),
					zap.String("rev", dbFile.Rev))
				// 扫描结果中不存在该文件，需要删除

				filesToDelete = append(filesToDelete, dbFile)
			}
		}

		// 开始执行
		var errs = make([]error, 0)

		// 新增文件
		if len(filesToCreate) > 0 {
			var count int64

			count, err = w.batchCreate(ctx, file.ID, filesToCreate)
			if err != nil {
				w.logger.Error("批量创建子文件失败",
					zap.Error(err),
					zap.Int64("count", count),
					zap.Int64("file_id", file.ID),
					zap.String("file_name", file.Name))

				errs = append(errs, fmt.Errorf("批量创建子文件失败: %w", err))
			} else {
				w.logger.Info("批量创建子文件成功",
					zap.Int64("count", count),
					zap.Int64("file_id", file.ID),
					zap.String("file_name", file.Name))
			}
		}

		// 更新文件
		for id, item := range filesToUpdateMap {
			if err = w.update(ctx, id, item); err != nil {
				w.logger.Error("更新文件失败",
					zap.Error(err),
					zap.String("file_name", item["name"].(string)),
					zap.Int64("file_id", id))

				errs = append(errs, fmt.Errorf("更新文件失败: %w", err))
			} else {
				w.logger.Debug("更新文件成功",
					zap.String("file_name", item["name"].(string)),
				)
			}
		}

		for _, item := range filesToDelete {
			if err = w.delete(ctx, item.ID); err != nil {
				w.logger.Error("删除文件失败",
					zap.Error(err),
					zap.String("file_name", item.Name),
					zap.Int64("file_id", item.ID))

				errs = append(errs, fmt.Errorf("删除文件失败: %w", err))
			} else {
				w.logger.Debug("删除文件成功",
					zap.String("file_name", item.Name),
				)
			}
		}

		nextWalkFiles = append(filesToCreate, filesToDeep...)

		return nextWalkFiles
	})
}

func (w *fileBusWorker) walk(ctx context.Context, rootId int64, walkFunc walkFunc) error {
	db := w.db.WithContext(ctx)

	file := &models.VirtualFile{}

	if rootId == 0 {
		file = &models.VirtualFile{
			Name:       "root",
			IsFolder:   1,
			IsTop:      1,
			OsType:     models.OsTypeFolder,
			ParentId:   0,
			ModifyDate: time.Now().Format("2006-01-02 15:04:05"),
			CreateDate: time.Now().Format("2006-01-02 15:04:05"),
		}
	} else {
		if err := db.Where("id", rootId).First(file).Error; err != nil {
			return err
		}
	}

	children := make([]File, 0)

	// 如果是文件夹类型，递归处理
	if file.IsFolder == 1 {
		if err := db.Where("parent_id", file.ID).Find(&children).Error; err != nil {
			return err
		}
	}

	w.logger.Debug("开始处理文件", zap.String("file_name", file.Name))

	// 判断是否还要继续
	if nextFiles := walkFunc(ctx, file, children); len(nextFiles) > 0 {
		for _, nextFile := range nextFiles {
			if err := w.walk(ctx, nextFile.ID, walkFunc); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *fileBusWorker) Close() {
	if !w.lock.TryLock() {
		w.lock.Unlock()
	}
}

func (w *fileBusWorker) batchCreate(ctx context.Context, parentId int64, files []File) (int64, error) {
	w.logger.Debug("批量创建文件", zap.Int64("parent_id", parentId), zap.Int("file_count", len(files)))

	// 检查 pid
	if parentId <= 0 {
		return 0, errors.New("parent_id is invalid")
	}

	for _, file := range files {
		file.ParentId = parentId
	}

	result := w.db.WithContext(ctx).CreateInBatches(files, 1000)

	// hook
	for _, file := range files {
		_ = w.createHook(ctx, file)
	}

	return result.RowsAffected, result.Error
}

func (w *fileBusWorker) createHook(ctx context.Context, file File) error {
	if file.ID == 0 {
		return errors.New("file.ID is invalid")
	}

	var errs []error

	{
		if !shared.StrmFileEnable {
			goto strmOver
		}

		extName := strings.TrimPrefix(filepath.Ext(file.Name), ".")

		if len(shared.StrmSupportFileExtList) > 0 && lo.IndexOf(shared.StrmSupportFileExtList, extName) == -1 {
			goto strmOver
		}

		filePath, err := w.calFilePath(ctx, file.ID)
		if err != nil {
			errs = append(errs, err)
		}

		filePath = path.Join(path.Dir(filePath), strings.TrimSuffix(file.Name, filepath.Ext(file.Name))+".strm")

		_ = shared.MediaBus.Publish(ctx, TopicMediaAddStrmFile, TopicMediaAddStrmFileRequest{
			FileID: file.ID,
			Path:   filePath,
		})
	}
strmOver:

	return errors2.Join(errs...)
}

func (w *fileBusWorker) update(ctx context.Context, id int64, mp map[string]any) error {
	w.logger.Debug("更新文件", zap.Int64("file_id", id), zap.Any("data", mp))

	return w.db.WithContext(ctx).Model(&models.VirtualFile{}).Where("id", id).Updates(mp).Error
}

func (w *fileBusWorker) delete(ctx context.Context, id int64) error {
	w.logger.Debug("删除文件", zap.Int64("file_id", id))

	file := new(models.VirtualFile)
	if err := w.getDB(ctx).Where("id = ?", id).First(file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.logger.Warn("文件不存在，跳过删除", zap.Int64("id", id))

			return nil
		}

		return fmt.Errorf("获取文件信息失败 id=%d: %w", id, err)
	}

	if file.IsFolder == 1 {
		children := make([]File, 0)
		if err := w.getDB(ctx).Where("parent_id = ?", id).Find(&children).Error; err != nil {
			return fmt.Errorf("获取子节点失败: %w", err)
		}

		var errs []error

		for _, child := range children {
			if err := w.delete(ctx, child.ID); err != nil {
				w.logger.Error("删除子文件失败",
					zap.Int64("parent_id", id),
					zap.Int64("child_id", child.ID),
					zap.Error(err))

				errs = append(errs, err)
			}
		}

		if len(errs) > 0 {
			return errors2.Join(errs...)
		}
	}

	// hook
	_ = w.deleteHook(ctx, file.ID)

	return w.getDB(ctx).Where("id", id).Delete(&models.VirtualFile{}).Error
}

func (w *fileBusWorker) deleteHook(ctx context.Context, fileId int64) error {
	if fileId == 0 {
		return errors.New("file.ID is invalid")
	}

	var errs []error

	if shared.LinkFileAutoDelete {
		err := shared.MediaBus.Publish(ctx, TopicMediaDeleteLinkFile, TopicMediaDeleteLinkFileRequest{
			FileID: fileId,
		})

		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors2.Join(errs...)
}

func (w *fileBusWorker) scanTop(ctx context.Context) error {
	// 读取所有顶层文件
	var topFiles = make([]*models.VirtualFile, 0)
	if err := w.getDB(ctx).Where("is_top = 1").Find(&topFiles).Error; err != nil {
		return err
	}

	var errs = make([]error, 0)

	for _, f := range topFiles {
		// 检查是否设置了禁用自动扫描队列标志（仅is_top=1时生效）
		if f.IsTop == 1 && f.Addition != nil {
			if disableAutoScanValue, exists := f.Addition[consts.FileAdditionKeyDisableAutoScan]; exists {
				if disableAutoScan, err := utils.Bool(disableAutoScanValue); err == nil && disableAutoScan {
					w.logger.Info("跳过自动扫描，文件已设置禁用自动扫描队列标志",
						zap.Int64("file_id", f.ID),
						zap.String("file_name", f.Name))

					continue
				}
			}
		}

		_ = w.scanFile(ctx, f.ID, false)
	}

	if len(errs) > 0 {
		return errors2.Join(errs...)
	}

	return nil
}

func (w *fileBusWorker) getDB(ctx context.Context) *gorm.DB {
	return w.db.WithContext(ctx).Model(&models.VirtualFile{})
}

// CalFilePath 计算文件的路径
func (w *fileBusWorker) calFilePath(ctx context.Context, id int64) (string, error) {
	if id == 0 {
		return "/", nil
	}

	file := new(models.VirtualFile)
	if err := w.getDB(ctx).Where("id = ?", id).First(file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", FileNotFound
		}

		return "", fmt.Errorf("获取文件信息失败 id=%d: %w", id, err)
	}

	parentPath, err := w.calFilePath(ctx, file.ParentId)
	if err != nil {
		return "", err
	}

	return path.Join(parentPath, file.Name), nil
}

func (w *fileBusWorker) buildMediaFile(ctx context.Context, fileId int64) (int64, error) {
	var count int64

	if err := w.walk(ctx, fileId, func(ctx context.Context, file File, childrenFiles []File) (nextWalkFiles []File) {
		if file.IsFolder == 1 {
			return childrenFiles
		}

		extName := strings.TrimPrefix(filepath.Ext(file.Name), ".")

		if len(shared.StrmSupportFileExtList) > 0 && lo.IndexOf(shared.StrmSupportFileExtList, extName) == -1 {
			return nil
		}

		filePath, err := w.calFilePath(ctx, file.ID)
		if err != nil {
			return nil
		}

		filePath = path.Join(path.Dir(filePath), strings.TrimSuffix(file.Name, filepath.Ext(file.Name))+".strm")

		err = shared.MediaBus.Publish(ctx, TopicMediaAddStrmFile, TopicMediaAddStrmFileRequest{
			FileID: file.ID,
			Path:   filePath,
		})
		if err == nil {
			atomic.AddInt64(&count, 1)
		}

		return nextWalkFiles
	}); err != nil {
		return 0, err
	}

	return int64(int(count)), nil
}
