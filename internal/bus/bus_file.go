package bus

import (
	"context"
	errors2 "errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type virtualFileWalkFunc func(ctx context.Context, file *models.VirtualFile, childrenFiles []*models.VirtualFile) (nextWalkFiles []*models.VirtualFile)

func (w *busWorker) scanVirtualFile(ctx context.Context, rootId int64, deep bool) error {
	var scanErrors []error
	var mu sync.Mutex

	fss := &FileScanStat{
		FileId:       rootId,
		ScannedCount: 0,
		WaitCount:    0,
	}

	w.fileScanStat.Store(rootId, fss)
	defer w.fileScanStat.Delete(rootId)

	err := w.walkVirtualFile(ctx, rootId, func(ctx context.Context, file *models.VirtualFile, oldFiles []*models.VirtualFile) (nextWalkFiles []*models.VirtualFile) {
		var (
			newFiles = make([]*models.VirtualFile, 0)
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
			mu.Lock()
			scanErrors = append(scanErrors, fmt.Errorf("获取文件列表失败 [%s]: %w", file.Name, err))
			mu.Unlock()
			return nil
		}

		fss.WaitCount += int64(len(newFiles))
		defer func() {
			fss.ScannedCount += int64(len(newFiles))
		}()

		// 数据准备

		// 创建映射表，用于快速查找
		var (
			newFileMap = make(map[string]*models.VirtualFile)
			oldFileMap = make(map[string]*models.VirtualFile)
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
			filesToCreate []*models.VirtualFile
			// 待删除的文件
			filesToDelete []*models.VirtualFile
			// 找出需要更新的文件
			filesToUpdateMap = map[int64]map[string]any{}
			// 需要深度扫描的文件
			filesToDeep []*models.VirtualFile
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

			count, err = w.batchCreateVirtualFile(ctx, file.ID, filesToCreate)
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
			if err = w.updateVirtualFile(ctx, id, item); err != nil {
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
			if err = w.deleteVirtualFile(ctx, item.ID); err != nil {
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

		// 收集当前文件处理过程中的错误
		if len(errs) > 0 {
			mu.Lock()
			scanErrors = append(scanErrors, errs...)
			mu.Unlock()
		}

		nextWalkFiles = append(filesToCreate, filesToDeep...)

		return nextWalkFiles
	})

	// 返回收集到的错误
	if err != nil {
		return err
	}

	if len(scanErrors) > 0 {
		return fmt.Errorf("扫描过程中发生 %d 个错误: %v", len(scanErrors), scanErrors)
	}

	return nil
}

func (w *busWorker) walkVirtualFile(ctx context.Context, rootId int64, walkFunc virtualFileWalkFunc) error {
	db := w.db.WithContext(ctx)

	file := &models.VirtualFile{}

	if rootId == 0 {
		file = &models.VirtualFile{
			Name:       "root",
			IsFolder:   1,
			IsTop:      1,
			OsType:     models.OsTypeFolder,
			ParentId:   0,
			ModifyDate: time.Now().Format(time.DateTime),
			CreateDate: time.Now().Format(time.DateTime),
		}
	} else {
		if err := db.Where("id", rootId).First(file).Error; err != nil {
			return err
		}
	}

	children := make([]*models.VirtualFile, 0)

	// 如果是文件夹类型，递归处理
	if file.IsFolder == 1 {
		if err := db.Where("parent_id", file.ID).Find(&children).Error; err != nil {
			return err
		}
	}

	w.logger.Debug("开始处理文件", zap.String("file_name", file.Name))

	// 判断是否还要继续
	if nextFiles := walkFunc(ctx, file, children); len(nextFiles) > 0 {
		// 获取线程数配置
		threadCount := shared.Setting.JobThreadCount
		if threadCount <= 0 {
			threadCount = 1
		}

		// 如果只有一个线程或者文件数量很少，使用串行处理
		if threadCount == 1 || len(nextFiles) <= 1 {
			for _, nextFile := range nextFiles {
				if err := w.walkVirtualFile(ctx, nextFile.ID, walkFunc); err != nil {
					return err
				}
			}
		} else {
			// 使用多线程并发处理
			var wg sync.WaitGroup
			errorChan := make(chan error, len(nextFiles))
			semaphore := make(chan struct{}, threadCount)

			for _, nextFile := range nextFiles {
				wg.Add(1)
				go func(file *models.VirtualFile) {
					defer wg.Done()

					// 获取信号量
					semaphore <- struct{}{}
					defer func() { <-semaphore }()

					if err := w.walkVirtualFile(ctx, file.ID, walkFunc); err != nil {
						errorChan <- err
					}
				}(nextFile)
			}

			// 等待所有goroutine完成
			wg.Wait()
			close(errorChan)

			// 检查是否有错误
			for err := range errorChan {
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (w *busWorker) batchCreateVirtualFile(ctx context.Context, parentId int64, files []*models.VirtualFile) (int64, error) {
	w.logger.Debug("批量创建文件", zap.Int64("parent_id", parentId), zap.Int("file_count", len(files)))

	// 检查 pid
	if parentId <= 0 {
		return 0, errors.New("parent_id is invalid")
	}

	for _, file := range files {
		file.ParentId = parentId
	}

	result := w.withLock(ctx, func(db *gorm.DB) *gorm.DB {
		return db.CreateInBatches(files, 1000)
	})

	//hook
	for _, file := range files {
		_ = w.createVirtualFileHook(ctx, file)
	}

	return result.RowsAffected, result.Error
}

func (w *busWorker) deleteVirtualFile(ctx context.Context, id int64) error {
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
		children := make([]*models.VirtualFile, 0)
		if err := w.getDB(ctx).Where("parent_id = ?", id).Find(&children).Error; err != nil {
			return fmt.Errorf("获取子节点失败: %w", err)
		}

		// 获取线程数配置
		threadCount := shared.Setting.JobThreadCount
		if threadCount <= 0 {
			threadCount = 1
		}

		// 如果只有一个线程或者子文件很少，使用串行处理
		if threadCount == 1 || len(children) <= 1 {
			var errs []error
			for _, child := range children {
				if err := w.deleteVirtualFile(ctx, child.ID); err != nil {
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
		} else {
			// 使用多线程并发删除
			var wg sync.WaitGroup
			errorChan := make(chan error, len(children))
			semaphore := make(chan struct{}, threadCount)

			for _, child := range children {
				wg.Add(1)
				go func(childFile *models.VirtualFile) {
					defer wg.Done()

					// 获取信号量
					semaphore <- struct{}{}
					defer func() { <-semaphore }()

					if err := w.deleteVirtualFile(ctx, childFile.ID); err != nil {
						w.logger.Error("删除子文件失败",
							zap.Int64("parent_id", id),
							zap.Int64("child_id", childFile.ID),
							zap.Error(err))
						errorChan <- err
					}
				}(child)
			}

			// 等待所有goroutine完成
			wg.Wait()
			close(errorChan)

			// 收集所有错误
			var errs []error
			for err := range errorChan {
				if err != nil {
					errs = append(errs, err)
				}
			}

			if len(errs) > 0 {
				return errors2.Join(errs...)
			}
		}
	}

	// hook
	_ = w.deleteVirtualFileHook(ctx, file.ID)

	return w.withLock(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Where("id", id).Delete(&models.VirtualFile{})
	}).Error
}

func (w *busWorker) updateVirtualFile(ctx context.Context, id int64, mp map[string]any) error {
	w.logger.Debug("更新文件", zap.Int64("file_id", id), zap.Any("data", mp))

	return w.withLock(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.VirtualFile{}).Where("id", id).Updates(mp)
	}).Error
}

func (w *busWorker) deleteVirtualFileHook(ctx context.Context, fileId int64) error {
	w.logger.Debug("删除文件", zap.Int64("file_id", fileId))

	if fileId == 0 {
		return errors.New("file.ID is invalid")
	}

	var errs []error

	if shared.LinkFileAutoDelete {
		if err := w.delMediaFile(ctx, fileId); err != nil {
			errs = append(errs, err)
		}
	}

	return errors2.Join(errs...)
}

func (w *busWorker) createVirtualFileHook(ctx context.Context, file *models.VirtualFile) error {
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

		if err = w.addStrmMediaFile(ctx, file.ID, filePath); err != nil {
			errs = append(errs, err)
		}
	}
strmOver:

	return errors2.Join(errs...)
}

func (w *busWorker) scanTopVirtualFiles(ctx context.Context) error {
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

		_ = w.scanVirtualFile(ctx, f.ID, false)
	}

	if len(errs) > 0 {
		return errors2.Join(errs...)
	}

	return nil
}

func (w *busWorker) buildMediaFile(ctx context.Context, fileId int64) (int64, error) {
	var count int64

	if err := w.walkVirtualFile(ctx, fileId, func(ctx context.Context, file *models.VirtualFile, childrenFiles []*models.VirtualFile) (nextWalkFiles []*models.VirtualFile) {
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

		if err = w.addStrmMediaFile(ctx, file.ID, filePath); err == nil {
			atomic.AddInt64(&count, 1)
		}

		return nextWalkFiles
	}); err != nil {
		return 0, err
	}

	return atomic.LoadInt64(&count), nil
}

// calFilePath 计算文件的路径
func (w *busWorker) calFilePath(ctx context.Context, id int64) (string, error) {
	return w.calFilePathWithCache(ctx, id, make(map[int64]*models.VirtualFile))
}

// calFilePathWithCache 使用缓存优化的路径计算方法
func (w *busWorker) calFilePathWithCache(ctx context.Context, id int64, cache map[int64]*models.VirtualFile) (string, error) {
	if id == 0 {
		return "/", nil
	}

	// 检查缓存
	file, exists := cache[id]
	if !exists {
		// 批量查询当前文件及其所有父级文件
		files, err := w.batchQueryParentFiles(ctx, id)
		if err != nil {
			return "", err
		}

		// 将查询结果加入缓存
		for _, f := range files {
			cache[f.ID] = f
		}

		file, exists = cache[id]
		if !exists {
			return "", FileNotFound
		}
	}

	parentPath, err := w.calFilePathWithCache(ctx, file.ParentId, cache)
	if err != nil {
		return "", err
	}

	return path.Join(parentPath, file.Name), nil
}

// batchQueryParentFiles 批量查询文件及其所有父级文件
func (w *busWorker) batchQueryParentFiles(ctx context.Context, id int64) ([]*models.VirtualFile, error) {
	var files []*models.VirtualFile
	var currentId = id
	var ids []int64

	// 收集所有需要查询的ID
	for currentId != 0 {
		ids = append(ids, currentId)

		// 查询当前文件的父ID
		var parentId int64
		if err := w.getDB(ctx).Model(&models.VirtualFile{}).Select("parent_id").Where("id = ?", currentId).Scan(&parentId).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}
			return nil, fmt.Errorf("查询父ID失败 id=%d: %w", currentId, err)
		}

		currentId = parentId
	}

	// 批量查询所有文件信息
	if err := w.getDB(ctx).Where("id IN ?", ids).Find(&files).Error; err != nil {
		return nil, fmt.Errorf("批量查询文件信息失败: %w", err)
	}

	return files, nil
}
