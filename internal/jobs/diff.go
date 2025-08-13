package jobs

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/samber/lo"
	"github.com/xxcheng123/cloudpan189-interface/client"
	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/fs"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	allowAutoDelOsTypes = []string{
		models.OsTypeFolder,
		models.OsTypeFile,
		models.OsTypeSubscribe,
		models.OsTypeSubscribeShare,
		models.OsTypeShare,
	}
)

type diffWorker struct {
	logger  *zap.Logger
	db      *gorm.DB
	startId int64 // 扫描开始时的文件ID，用于断点续传
	thread  int   // 线程数
	client  client.Client
	fs      fs.FS
}

func newDiffWorker(logger *zap.Logger, db *gorm.DB, startId int64) *diffWorker {
	thread := shared.Setting.JobThreadCount
	if thread <= 0 {
		thread = 1
	}

	return &diffWorker{
		logger:  logger,
		db:      db,
		startId: startId,
		thread:  thread,
		client:  client.New(),
		fs:      fs.NewFS(db, logger.With(zap.String("owner", "diffWorker"))),
	}
}

func (w *diffWorker) execute(ctx context.Context, file File, deep bool) (err error) {
	if file == nil {
		return nil
	}

	var (
		newFiles = make([]File, 0)
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

	_ = shared.UpdateJobProgress(w.startId, 0, int64(len(newFiles)))
	defer func() {
		_ = shared.UpdateJobProgress(w.startId, int64(len(newFiles)), 0)
	}()

	if err != nil {
		return fmt.Errorf("failed to scan files for %s(%d): %w", file.Name, file.ID, err)
	}

	oldFiles := make([]File, 0)
	if err = w.db.WithContext(ctx).Where("parent_id = ?", file.ID).Find(&oldFiles).Error; err != nil {
		return fmt.Errorf("failed to query db files: %w", err)
	}

	// 创建映射表，用于快速查找
	newFileMap := make(map[string]File)
	for _, item := range newFiles {
		key := item.Name
		newFileMap[key] = item
	}

	oldFileMap := make(map[string]File)
	for _, item := range oldFiles {
		key := item.Name
		oldFileMap[key] = item
	}

	// 找出需要新增的文件
	var filesToCreate []File
	// 找出需要更新的文件
	var filesToUpdateMap = map[int64]map[string]any{}
	// 找出需要删除的文件
	var filesToDelete []File
	// 文件没有更新，但是需要深度扫描
	var filesToDeep []File

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

	var errs = make([]error, 0)

	// 新增文件
	if len(filesToCreate) > 0 {
		count, err := w.fs.BatchCreate(ctx, file.ID, filesToCreate...)
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
		if err = w.fs.Update(ctx, id, item); err != nil {
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
		if err = w.fs.Delete(ctx, item.ID); err != nil {
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

	folders := filesToDeep
	for _, item := range filesToCreate {
		if item.IsFolder == 1 {
			folders = append(folders, item)
		}
	}

	if len(folders) > 0 {
		if err = w.processSubfoldersParallel(ctx, folders, deep); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// processSubfoldersParallel 并行处理子文件夹
func (w *diffWorker) processSubfoldersParallel(ctx context.Context, folders []File, deep bool) error {
	if len(folders) == 0 {
		return nil
	}

	// 限制并发数量，避免创建过多goroutine
	maxWorkers := w.thread
	if maxWorkers <= 0 {
		maxWorkers = 3 // 默认3个并发
	}

	// 如果文件夹数量少于线程数，使用文件夹数量
	if len(folders) < maxWorkers {
		maxWorkers = len(folders)
	}

	// 创建工作池
	folderChan := make(chan File, len(folders))
	errorChan := make(chan error, len(folders))

	var wg sync.WaitGroup

	// 启动worker goroutines
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)

		workerID := i
		gopool.Go(func() {
			defer wg.Done()
			w.logger.Debug("subfolder worker started", zap.Int("worker_id", workerID))

			for folder := range folderChan {
				select {
				case <-ctx.Done():
					errorChan <- ctx.Err()

					return
				default:
					w.logger.Debug("processing subfolder",
						zap.Int("worker_id", workerID),
						zap.String("folder_name", folder.Name),
						zap.Int64("folder_id", folder.ID))

					if err := w.execute(ctx, folder, deep); err != nil {
						w.logger.Error("failed to process subfolder",
							zap.Int("worker_id", workerID),
							zap.String("folder_name", folder.Name),
							zap.Error(err))
						errorChan <- fmt.Errorf("failed to handle subfolder %s: %w", folder.Name, err)
					} else {
						w.logger.Debug("subfolder processed successfully",
							zap.Int("worker_id", workerID),
							zap.String("folder_name", folder.Name))
					}
				}
			}
		})
	}

	// 发送任务到channel
	go func() {
		defer close(folderChan)
		for _, folder := range folders {
			select {
			case <-ctx.Done():
				return
			case folderChan <- folder:
			}
		}
	}()

	// 等待所有worker完成
	go func() {
		wg.Wait()
		close(errorChan)
	}()

	// 收集错误
	var errs []error
	for err := range errorChan {
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (w *diffWorker) getSubscribeUserFiles(ctx context.Context, f File) ([]File, error) {
	_userId, ok := f.Addition[consts.FileAdditionKeySubscribeUser]
	if !ok {
		return nil, errors.New("no subscribe_user")
	}

	userId := utils.String(_userId)

	var (
		pageNum  int64 = 1
		pageSize int64 = 200
		files          = make([]File, 0)
	)

	resp, err := w.client.GetUpResourceShare(ctx, userId, pageNum, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get first page: %w", err)
	}

	if resp.Data != nil {
		for _, v := range resp.Data.FileList {
			files = append(files, &models.VirtualFile{
				ParentId:   f.ID,
				Name:       v.Name,
				IsTop:      0,
				Size:       v.Size,
				IsFolder:   int8(v.Folder),
				Hash:       strings.ToLower(v.Md5),
				CreateDate: v.CreateDate,
				ModifyDate: v.LastOpTime,
				OsType:     models.OsTypeSubscribeShare,
				Addition: map[string]any{
					consts.FileAdditionKeySubscribeUser: userId,
					consts.FileAdditionKeyShareId:       v.ShareId,
					consts.FileAdditionKeyFileId:        v.Id,
				},
				Rev: v.Rev,
			})
		}
	}

	if resp.Data != nil && int64(len(files)) < resp.Data.Count {
		var (
			mu       sync.Mutex
			wg       sync.WaitGroup
			errs     []error
			allFiles [][]File
		)

		totalPages := (resp.Data.Count + pageSize - 1) / pageSize
		allFiles = make([][]File, totalPages-1)

		for i := int64(2); i <= totalPages; i++ {
			wg.Add(1)
			go func(pageNum int64, index int) {
				defer wg.Done()

				subResp, subErr := w.client.GetUpResourceShare(ctx, userId, pageNum, pageSize)
				if subErr != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("failed to get page %d: %w", pageNum, subErr))
					mu.Unlock()
					return
				}

				if subResp.Data != nil {
					var pageFiles []File
					for _, v := range subResp.Data.FileList {
						pageFiles = append(pageFiles, &models.VirtualFile{
							ParentId:   f.ID,
							Name:       v.Name,
							IsTop:      0,
							Size:       v.Size,
							IsFolder:   int8(v.Folder),
							Hash:       strings.ToLower(v.Md5),
							CreateDate: v.CreateDate,
							ModifyDate: v.LastOpTime,
							OsType:     models.OsTypeSubscribeShare,
							Addition: map[string]any{
								consts.FileAdditionKeySubscribeUser: userId,
								consts.FileAdditionKeyShareId:       v.ShareId,
								consts.FileAdditionKeyFileId:        v.Id,
							},
							Rev: v.Rev,
						})
					}
					allFiles[index] = pageFiles
				}
			}(i, int(i-2))
		}

		wg.Wait()

		if len(errs) > 0 {
			return nil, errors.Join(errs...)
		}

		for _, pageFiles := range allFiles {
			files = append(files, pageFiles...)
		}
	}

	return files, nil
}

func (w *diffWorker) getSubscribeShareFiles(ctx context.Context, f File) ([]File, error) {
	_userId, ok := f.Addition[consts.FileAdditionKeySubscribeUser]
	if !ok {
		return nil, errors.New("no subscribe_user")
	}

	_shareId, ok := f.Addition[consts.FileAdditionKeyShareId]
	if !ok {
		return nil, errors.New("no share_id")
	}

	_fileId, ok := f.Addition[consts.FileAdditionKeyFileId]
	if !ok {
		return nil, errors.New("no file_id")
	}

	var (
		userId     = utils.String(_userId)
		shareId, _ = utils.Int64(_shareId)
		fileId     = utils.String(_fileId)
		pageNum    = 1
		pageSize   = 200
		files      = make([]File, 0)
	)

	resp, err := w.client.ListShareDir(ctx, shareId, client.String(fileId), func(req *client.ListShareFileRequest) {
		req.PageNum = pageNum
		req.PageSize = pageSize
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get first page: %w", err)
	}

	for _, v := range resp.FileListAO.FolderList {
		files = append(files, &models.VirtualFile{
			ParentId:   f.ID,
			Name:       v.Name,
			IsTop:      0,
			Size:       0,
			IsFolder:   1,
			Hash:       "",
			CreateDate: v.CreateDate,
			ModifyDate: v.LastOpTime,
			OsType:     models.OsTypeSubscribeShare,
			Addition: map[string]any{
				consts.FileAdditionKeySubscribeUser: userId,
				consts.FileAdditionKeyShareId:       shareId,
				consts.FileAdditionKeyFileId:        v.Id,
			},
			Rev: v.Rev,
		})
	}

	for _, v := range resp.FileListAO.FileList {
		files = append(files, &models.VirtualFile{
			ParentId:   f.ID,
			Name:       v.Name,
			IsTop:      0,
			Size:       v.Size,
			IsFolder:   0,
			Hash:       strings.ToLower(v.Md5),
			CreateDate: v.CreateDate,
			ModifyDate: v.LastOpTime,
			OsType:     models.OsTypeFile,
			Addition: map[string]any{
				consts.FileAdditionKeySubscribeUser: userId,
				consts.FileAdditionKeyShareId:       shareId,
				consts.FileAdditionKeyFileId:        v.Id,
			},
			Rev: v.Rev,
		})
	}

	if int64(len(files)) < resp.FileListAO.Count {
		var (
			mu       sync.Mutex
			wg       sync.WaitGroup
			errs     []error
			allFiles [][]File
		)

		totalPages := (resp.FileListAO.Count + int64(pageSize) - 1) / int64(pageSize)
		allFiles = make([][]File, totalPages-1)

		for i := int64(2); i <= totalPages; i++ {
			wg.Add(1)
			go func(pageNum int64, index int) {
				defer wg.Done()

				subResp, subErr := w.client.ListShareDir(ctx, shareId, client.String(fileId), func(req *client.ListShareFileRequest) {
					req.PageNum = int(pageNum)
					req.PageSize = pageSize
				})
				if subErr != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("failed to get page %d: %w", pageNum, subErr))
					mu.Unlock()
					return
				}

				var pageFiles []File
				for _, v := range subResp.FileListAO.FolderList {
					pageFiles = append(pageFiles, &models.VirtualFile{
						ParentId:   f.ID,
						Name:       v.Name,
						IsTop:      0,
						Size:       0,
						IsFolder:   1,
						Hash:       "",
						CreateDate: v.CreateDate,
						ModifyDate: v.LastOpTime,
						OsType:     models.OsTypeSubscribeShare,
						Addition: map[string]any{
							consts.FileAdditionKeySubscribeUser: userId,
							consts.FileAdditionKeyShareId:       shareId,
							consts.FileAdditionKeyFileId:        v.Id,
						},
						Rev: v.Rev,
					})
				}

				for _, v := range subResp.FileListAO.FileList {
					pageFiles = append(pageFiles, &models.VirtualFile{
						ParentId:   f.ID,
						Name:       v.Name,
						IsTop:      0,
						Size:       v.Size,
						IsFolder:   0,
						Hash:       strings.ToLower(v.Md5),
						CreateDate: v.CreateDate,
						ModifyDate: v.LastOpTime,
						OsType:     models.OsTypeFile,
						Addition: map[string]any{
							consts.FileAdditionKeySubscribeUser: userId,
							consts.FileAdditionKeyShareId:       shareId,
							consts.FileAdditionKeyFileId:        v.Id,
						},
						Rev: v.Rev,
					})
				}
				allFiles[index] = pageFiles
			}(i, int(i-2))
		}

		wg.Wait()

		if len(errs) > 0 {
			return nil, errors.Join(errs...)
		}

		for _, pageFiles := range allFiles {
			files = append(files, pageFiles...)
		}
	}

	return files, nil
}

func (w *diffWorker) getShareFiles(ctx context.Context, f File) ([]File, error) {
	var vv, ok = f.Addition[consts.FileAdditionKeyShareId]
	if !ok {
		return nil, errors.New("no share_id")
	}

	shareId, _ := utils.Int64(vv)

	vv, ok = f.Addition[consts.FileAdditionKeyFileId]
	if !ok {
		return nil, errors.New("no file_id")
	}

	fileId := utils.String(vv)

	vv, ok = f.Addition[consts.FileAdditionKeyShareMode]
	if !ok {
		return nil, errors.New("no share_mode")
	}

	shareMode, _ := utils.Int(vv)

	vv, ok = f.Addition[consts.FileAdditionKeyAccessCode]
	if !ok {
		return nil, errors.New("no access_code")
	}

	accessCode := utils.String(vv)

	vv, ok = f.Addition[consts.FileAdditionKeyIsFolder]
	if !ok {
		return nil, errors.New("no is_folder")
	}

	var (
		pageNum  = 1
		pageSize = 200
		files    = make([]File, 0)
		addMpFn  = func(mp map[string]any) map[string]any {
			mp[consts.FileAdditionKeyShareId] = shareId
			mp[consts.FileAdditionKeyShareMode] = shareMode
			mp[consts.FileAdditionKeyAccessCode] = accessCode

			return mp
		}
	)

	resp, err := w.client.ListShareDir(ctx, shareId, client.String(fileId), func(req *client.ListShareFileRequest) {
		req.PageNum = pageNum
		req.PageSize = pageSize
		req.IsFolder, _ = utils.Bool(vv)
		req.AccessCode = accessCode
		req.ShareMode = shareMode
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get first page: %w", err)
	}

	for _, v := range resp.FileListAO.FolderList {
		files = append(files, &models.VirtualFile{
			ParentId:   f.ID,
			Name:       v.Name,
			IsTop:      0,
			Size:       0,
			IsFolder:   1,
			Hash:       "",
			CreateDate: v.CreateDate,
			ModifyDate: v.LastOpTime,
			OsType:     models.OsTypeShare,
			Addition: addMpFn(map[string]any{
				consts.FileAdditionKeyFileId:   v.Id,
				consts.FileAdditionKeyIsFolder: true,
			}),
			Rev: v.Rev,
		})
	}

	for _, v := range resp.FileListAO.FileList {
		files = append(files, &models.VirtualFile{
			ParentId:   f.ID,
			Name:       v.Name,
			IsTop:      0,
			Size:       v.Size,
			IsFolder:   0,
			Hash:       strings.ToLower(v.Md5),
			CreateDate: v.CreateDate,
			ModifyDate: v.LastOpTime,
			OsType:     models.OsTypeFile,
			Addition: addMpFn(map[string]any{
				consts.FileAdditionKeyFileId:   v.Id,
				consts.FileAdditionKeyIsFolder: false,
			}),
			Rev: v.Rev,
		})
	}

	if int64(len(files)) < resp.FileListAO.Count {
		var (
			mu       sync.Mutex
			wg       sync.WaitGroup
			errs     []error
			allFiles [][]File
		)

		totalPages := (resp.FileListAO.Count + int64(pageSize) - 1) / int64(pageSize)
		allFiles = make([][]File, totalPages-1)

		for i := int64(2); i <= totalPages; i++ {
			wg.Add(1)
			go func(pageNum int64, index int) {
				defer wg.Done()

				subResp, subErr := w.client.ListShareDir(ctx, shareId, client.String(fileId), func(req *client.ListShareFileRequest) {
					req.PageNum = int(pageNum)
					req.PageSize = pageSize
				})
				if subErr != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("failed to get page %d: %w", pageNum, subErr))
					mu.Unlock()
					return
				}

				var pageFiles []File
				for _, v := range subResp.FileListAO.FolderList {
					pageFiles = append(pageFiles, &models.VirtualFile{
						ParentId:   f.ID,
						Name:       v.Name,
						IsTop:      0,
						Size:       0,
						IsFolder:   1,
						Hash:       "",
						CreateDate: v.CreateDate,
						ModifyDate: v.LastOpTime,
						OsType:     models.OsTypeShare,
						Addition: addMpFn(map[string]any{
							consts.FileAdditionKeyFileId:   v.Id,
							consts.FileAdditionKeyIsFolder: true,
						}),
						Rev: v.Rev,
					})
				}

				for _, v := range subResp.FileListAO.FileList {
					pageFiles = append(pageFiles, &models.VirtualFile{
						ParentId:   f.ID,
						Name:       v.Name,
						IsTop:      0,
						Size:       v.Size,
						IsFolder:   0,
						Hash:       strings.ToLower(v.Md5),
						CreateDate: v.CreateDate,
						ModifyDate: v.LastOpTime,
						OsType:     models.OsTypeFile,
						Addition: addMpFn(map[string]any{
							consts.FileAdditionKeyFileId:   v.Id,
							consts.FileAdditionKeyIsFolder: false,
						}),
						Rev: v.Rev,
					})
				}
				allFiles[index] = pageFiles
			}(i, int(i-2))
		}

		wg.Wait()

		if len(errs) > 0 {
			return nil, errors.Join(errs...)
		}

		for _, pageFiles := range allFiles {
			files = append(files, pageFiles...)
		}
	}

	return files, nil
}
