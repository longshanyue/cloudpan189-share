package jobs

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/xxcheng123/cloudpan189-interface/client"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ScanFileJob struct {
	db      *gorm.DB
	running bool
	mu      sync.Mutex
	client  client.Client
	logger  *zap.Logger
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewScanFileJob(db *gorm.DB, logger *zap.Logger) Job {
	return &ScanFileJob{
		db:     db,
		logger: logger.With(zap.String("job", "scan_file")),
		client: client.New(),
	}
}

func (s *ScanFileJob) Start(ctx context.Context) error {
	s.ctx, s.cancel = context.WithCancel(ctx)

	if !s.mu.TryLock() {
		return ErrJobRunning
	}

	defer s.mu.Unlock()

	if s.running {
		return ErrJobRunning
	}

	s.running = true

	gopool.Go(func() {
		for {
			if !s.doJob(ctx) {
				break
			}
		}

		s.logger.Info("scan file job stopped")
	})

	return nil
}

func (s *ScanFileJob) doJob(ctx context.Context) bool {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("scan file job panic",
				zap.Any("panic", r),
				zap.String("stack", string(debug.Stack())))
		}
	}()

	select {
	case <-s.ctx.Done():
		s.logger.Info("scan file job stopped")

		return false
	case msg := <-shared.ScanJobRead():
		s.logger.Info("scan file job received message", zap.Any("msg", msg))
		// 先检查文件还在不在
		var count int64
		s.db.WithContext(ctx).Model(new(models.VirtualFile)).Where("id = ?", msg.Msg.ID).Count(&count)
		if count == 0 {
			s.logger.Error("file not found", zap.Int64("file_id", msg.Msg.ID))

			return true
		}

		switch msg.Type {
		case shared.ScanJobTypeRefresh:
			if err := s.handleFile(s.ctx, msg.Msg); err != nil {
				s.logger.Error("handle file error", zap.Error(err))
			}
		case shared.ScanJobTypeDeepRefresh:
			if err := s.deepHandleFile(s.ctx, msg.Msg); err != nil {
				s.logger.Error("handle file error", zap.Error(err))
			}
		case shared.ScanJobTypeDel:
			if err := s.recursiveDelete(s.ctx, []*models.VirtualFile{msg.Msg}); err != nil {
				s.logger.Error("delete file error", zap.Error(err))
			}
		}
	case <-time.After(time.Minute * 10):
		if !shared.Setting.EnableTopFileAutoRefresh {
			return true
		}
		s.logger.Info("scan top file job started")
		if err := s.scanTop(s.ctx); err != nil {
			s.logger.Error("scan file error", zap.Error(err))
		}
	}

	return true
}

func (s *ScanFileJob) Stop() {
	if s.running {
		s.cancel()

		s.running = false
	}
}

func (s *ScanFileJob) scanTop(ctx context.Context) error {
	// 读取所有顶层文件
	var topFiles = make([]*models.VirtualFile, 0)
	if err := s.db.WithContext(ctx).Where("is_top = 1").Find(&topFiles).Error; err != nil {
		return err
	}

	var errs = make([]error, 0)

	for _, f := range topFiles {
		if err := s.handleFile(ctx, f); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (s *ScanFileJob) handleFile(ctx context.Context, f *models.VirtualFile) error {
	var scannedFiles = make([]*models.VirtualFile, 0)
	var err error

	switch f.OsType {
	case models.OsTypeSubscribe:
		scannedFiles, err = s.getSubscribeUserFiles(ctx, f)
	case models.OsTypeSubscribeShare:
		scannedFiles, err = s.getSubscribeShareFiles(ctx, f)
	case models.OsTypeShare:
		scannedFiles, err = s.getShareFiles(ctx, f)
	default:
		return errors.New("unsupported os type")
	}

	if err != nil {
		return fmt.Errorf("failed to scan files for %s: %w", f.Name, err)
	}

	var dbFiles = make([]*models.VirtualFile, 0)
	if err = s.db.WithContext(ctx).Where("parent_id = ?", f.ID).Find(&dbFiles).Error; err != nil {
		return fmt.Errorf("failed to query db files: %w", err)
	}

	// 创建映射表，用于快速查找
	scannedFileMap := make(map[string]*models.VirtualFile)
	for _, file := range scannedFiles {
		key := file.Name
		scannedFileMap[key] = file
	}

	dbFileMap := make(map[string]*models.VirtualFile)
	for _, file := range dbFiles {
		key := file.Name
		dbFileMap[key] = file
	}

	// 找出需要新增的文件
	var filesToCreate []*models.VirtualFile
	// 找出需要更新的文件
	var filesToUpdate []*models.VirtualFile
	// 找出需要删除的文件
	var filesToDelete []*models.VirtualFile

	// 遍历扫描到的文件，找出新增和更新的文件
	for name, scannedFile := range scannedFileMap {
		if dbFile, exists := dbFileMap[name]; exists {
			// 文件存在，检查是否需要更新（通过Rev比较）
			if dbFile.Rev != scannedFile.Rev {
				s.logger.Debug("file needs update - rev changed",
					zap.String("parent", f.Name),
					zap.String("file_name", name),
					zap.String("old_rev", dbFile.Rev),
					zap.String("new_rev", scannedFile.Rev))
				// Rev不同，需要更新
				scannedFile.ID = dbFile.ID // 保持原有ID
				scannedFile.ParentId = f.ID
				filesToUpdate = append(filesToUpdate, scannedFile)
			}
		} else {
			s.logger.Debug("new file found - not in database",
				zap.String("parent", f.Name),
				zap.String("file_name", name),
				zap.String("rev", scannedFile.Rev))
			// 文件不存在，需要新增
			scannedFile.ParentId = f.ID
			filesToCreate = append(filesToCreate, scannedFile)
		}
	}

	// 遍历数据库中的文件，找出需要删除的文件
	for name, dbFile := range dbFileMap {
		if _, exists := scannedFileMap[name]; !exists {
			s.logger.Debug("file to be deleted - not in remote",
				zap.String("parent", f.Name),
				zap.String("file_name", name),
				zap.Int64("file_id", dbFile.ID),
				zap.String("rev", dbFile.Rev))
			// 扫描结果中不存在该文件，需要删除
			filesToDelete = append(filesToDelete, dbFile)
		}
	}

	var errs = make([]error, 0)

	// 批量处理新增文件
	if len(filesToCreate) > 0 {
		if err = s.db.WithContext(ctx).CreateInBatches(filesToCreate, 100).Error; err != nil {
			errs = append(errs, fmt.Errorf("failed to create files: %w", err))
		}
	}

	// 批量处理更新文件
	for _, file := range filesToUpdate {
		if err = s.db.WithContext(ctx).Save(file).Error; err != nil {
			errs = append(errs, fmt.Errorf("failed to update file %s: %w", file.Name, err))
		}
	}

	// 递归删除文件（包括子文件和子文件夹）
	if len(filesToDelete) > 0 {
		if err = s.recursiveDelete(ctx, filesToDelete); err != nil {
			errs = append(errs, fmt.Errorf("failed to delete files: %w", err))
		}
	}

	// 递归处理子文件夹
	allFiles := append(filesToCreate, filesToUpdate...)
	for _, file := range allFiles {
		if file.IsFolder == 1 { // 如果是文件夹
			if err = s.handleFile(ctx, file); err != nil {
				errs = append(errs, fmt.Errorf("failed to handle subfolder %s: %w", file.Name, err))
			}
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// recursiveDelete 递归删除文件及其所有子文件和子文件夹
func (s *ScanFileJob) recursiveDelete(ctx context.Context, files []*models.VirtualFile) error {
	if len(files) == 0 {
		return nil
	}

	var fileIDs []int64
	var folderIDs []int64
	var errs = make([]error, 0)

	// 分离文件和文件夹ID
	for _, file := range files {
		fileIDs = append(fileIDs, file.ID)
		if file.IsFolder == 1 {
			folderIDs = append(folderIDs, file.ID)
		}
	}

	// 如果有文件夹，需要先递归删除其子文件
	if len(folderIDs) > 0 {
		var childFiles []*models.VirtualFile
		if err := s.db.WithContext(ctx).Where("parent_id IN ?", folderIDs).Find(&childFiles).Error; err != nil {
			errs = append(errs, fmt.Errorf("failed to find child files: %w", err))
		} else {
			// 递归删除子文件
			if len(childFiles) > 0 {
				if err := s.recursiveDelete(ctx, childFiles); err != nil {
					errs = append(errs, err)
				}
			}
		}
	}

	// 删除当前层级的所有文件
	if err := s.db.WithContext(ctx).Where("id IN ?", fileIDs).Delete(&models.VirtualFile{}).Error; err != nil {
		errs = append(errs, fmt.Errorf("failed to delete files: %w", err))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// recursiveDeleteSimple 简化版递归删除
func (s *ScanFileJob) recursiveDeleteSimple(ctx context.Context, parentIDs []int64) error {
	if len(parentIDs) == 0 {
		return nil
	}

	var errs = make([]error, 0)

	// 查找所有子文件
	var childFiles []*models.VirtualFile
	if err := s.db.WithContext(ctx).Where("parent_id IN ?", parentIDs).Find(&childFiles).Error; err != nil {
		errs = append(errs, fmt.Errorf("failed to find child files: %w", err))
	} else {
		// 如果有子文件，先递归删除
		if len(childFiles) > 0 {
			var childIDs []int64
			for _, child := range childFiles {
				childIDs = append(childIDs, child.ID)
			}
			if err := s.recursiveDeleteSimple(ctx, childIDs); err != nil {
				errs = append(errs, err)
			}
		}
	}

	// 删除当前层级的文件
	if err := s.db.WithContext(ctx).Where("id IN ?", parentIDs).Delete(&models.VirtualFile{}).Error; err != nil {
		errs = append(errs, fmt.Errorf("failed to delete files: %w", err))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (s *ScanFileJob) getSubscribeUserFiles(ctx context.Context, f *models.VirtualFile) ([]*models.VirtualFile, error) {
	_userId, ok := f.Addition["subscribe_user"]
	if !ok {
		return nil, errors.New("no subscribe_user")
	}

	userId := utils.String(_userId)

	var (
		pageNum  int64 = 1
		pageSize int64 = 200
		files          = make([]*models.VirtualFile, 0)
	)

	resp, err := s.client.GetUpResourceShare(ctx, userId, pageNum, pageSize)
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
					"subscribe_user": userId,
					"share_id":       v.ShareId,
					"file_id":        v.Id,
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
			allFiles [][]*models.VirtualFile
		)

		totalPages := (resp.Data.Count + pageSize - 1) / pageSize
		allFiles = make([][]*models.VirtualFile, totalPages-1)

		for i := int64(2); i <= totalPages; i++ {
			wg.Add(1)
			go func(pageNum int64, index int) {
				defer wg.Done()

				subResp, subErr := s.client.GetUpResourceShare(ctx, userId, pageNum, pageSize)
				if subErr != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("failed to get page %d: %w", pageNum, subErr))
					mu.Unlock()
					return
				}

				if subResp.Data != nil {
					var pageFiles []*models.VirtualFile
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
								"subscribe_user": userId,
								"share_id":       v.ShareId,
								"file_id":        v.Id,
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

func (s *ScanFileJob) getSubscribeShareFiles(ctx context.Context, f *models.VirtualFile) ([]*models.VirtualFile, error) {
	_userId, ok := f.Addition["subscribe_user"]
	if !ok {
		return nil, errors.New("no subscribe_user")
	}

	_shareId, ok := f.Addition["share_id"]
	if !ok {
		return nil, errors.New("no share_id")
	}

	_fileId, ok := f.Addition["file_id"]
	if !ok {
		return nil, errors.New("no file_id")
	}

	var (
		userId     = utils.String(_userId)
		shareId, _ = utils.Int64(_shareId)
		fileId     = utils.String(_fileId)
		pageNum    = 1
		pageSize   = 200
		files      = make([]*models.VirtualFile, 0)
	)

	resp, err := s.client.ListShareDir(ctx, shareId, client.String(fileId), func(req *client.ListShareFileRequest) {
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
				"subscribe_user": userId,
				"share_id":       shareId,
				"file_id":        v.Id,
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
				"subscribe_user": userId,
				"share_id":       shareId,
				"file_id":        v.Id,
			},
			Rev: v.Rev,
		})
	}

	if int64(len(files)) < resp.FileListAO.Count {
		var (
			mu       sync.Mutex
			wg       sync.WaitGroup
			errs     []error
			allFiles [][]*models.VirtualFile
		)

		totalPages := (resp.FileListAO.Count + int64(pageSize) - 1) / int64(pageSize)
		allFiles = make([][]*models.VirtualFile, totalPages-1)

		for i := int64(2); i <= totalPages; i++ {
			wg.Add(1)
			go func(pageNum int64, index int) {
				defer wg.Done()

				subResp, subErr := s.client.ListShareDir(ctx, shareId, client.String(fileId), func(req *client.ListShareFileRequest) {
					req.PageNum = int(pageNum)
					req.PageSize = pageSize
				})
				if subErr != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("failed to get page %d: %w", pageNum, subErr))
					mu.Unlock()
					return
				}

				var pageFiles []*models.VirtualFile
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
							"subscribe_user": userId,
							"share_id":       shareId,
							"file_id":        v.Id,
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
							"subscribe_user": userId,
							"share_id":       shareId,
							"file_id":        v.Id,
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

func (s *ScanFileJob) getShareFiles(ctx context.Context, f *models.VirtualFile) ([]*models.VirtualFile, error) {
	// need fileId、shareId、isFolder、shareMode、accessCode、
	var vv, ok = f.Addition["share_id"]
	if !ok {
		return nil, errors.New("no share_id")
	}

	shareId, _ := utils.Int64(vv)

	vv, ok = f.Addition["file_id"]
	if !ok {
		return nil, errors.New("no file_id")
	}

	fileId := utils.String(vv)

	vv, ok = f.Addition["share_mode"]
	if !ok {
		return nil, errors.New("no share_mode")
	}

	shareMode, _ := utils.Int(vv)

	vv, ok = f.Addition["access_code"]
	if !ok {
		return nil, errors.New("no access_code")
	}

	accessCode := utils.String(vv)

	vv, ok = f.Addition["is_folder"]
	if !ok {
		return nil, errors.New("no is_folder")
	}

	var (
		pageNum  = 1
		pageSize = 200
		files    = make([]*models.VirtualFile, 0)
		addMpFn  = func(mp map[string]any) map[string]any {
			mp["share_id"] = shareId
			mp["share_mode"] = shareMode
			mp["access_code"] = accessCode

			return mp
		}
	)

	resp, err := s.client.ListShareDir(ctx, shareId, client.String(fileId), func(req *client.ListShareFileRequest) {
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
				"file_id":   v.Id,
				"is_folder": true,
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
				"file_id":   v.Id,
				"is_folder": false,
			}),
			Rev: v.Rev,
		})
	}

	if int64(len(files)) < resp.FileListAO.Count {
		var (
			mu       sync.Mutex
			wg       sync.WaitGroup
			errs     []error
			allFiles [][]*models.VirtualFile
		)

		totalPages := (resp.FileListAO.Count + int64(pageSize) - 1) / int64(pageSize)
		allFiles = make([][]*models.VirtualFile, totalPages-1)

		for i := int64(2); i <= totalPages; i++ {
			wg.Add(1)
			go func(pageNum int64, index int) {
				defer wg.Done()

				subResp, subErr := s.client.ListShareDir(ctx, shareId, client.String(fileId), func(req *client.ListShareFileRequest) {
					req.PageNum = int(pageNum)
					req.PageSize = pageSize
				})
				if subErr != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("failed to get page %d: %w", pageNum, subErr))
					mu.Unlock()
					return
				}

				var pageFiles []*models.VirtualFile
				for _, v := range subResp.FileListAO.FolderList {
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
							"file_id":   v.Id,
							"is_folder": true,
						}),
						Rev: v.Rev,
					})
				}

				for _, v := range subResp.FileListAO.FileList {
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
							"file_id":   v.Id,
							"is_folder": false,
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

// deepHandleFile 深度处理文件，强制刷新所有文件夹
func (s *ScanFileJob) deepHandleFile(ctx context.Context, f *models.VirtualFile) error {
	var scannedFiles = make([]*models.VirtualFile, 0)
	var err error

	switch f.OsType {
	case models.OsTypeSubscribe:
		scannedFiles, err = s.getSubscribeUserFiles(ctx, f)
	case models.OsTypeSubscribeShare:
		scannedFiles, err = s.getSubscribeShareFiles(ctx, f)
	case models.OsTypeShare:
		scannedFiles, err = s.getShareFiles(ctx, f)
	default:
		return errors.New("unsupported os type")
	}

	if err != nil {
		return fmt.Errorf("failed to scan files for %s: %w", f.Name, err)
	}

	var dbFiles = make([]*models.VirtualFile, 0)
	if err = s.db.WithContext(ctx).Where("parent_id = ?", f.ID).Find(&dbFiles).Error; err != nil {
		return fmt.Errorf("failed to query db files: %w", err)
	}

	// 创建映射表，用于快速查找
	scannedFileMap := make(map[string]*models.VirtualFile)
	for _, file := range scannedFiles {
		key := file.Name
		scannedFileMap[key] = file
	}

	dbFileMap := make(map[string]*models.VirtualFile)
	for _, file := range dbFiles {
		key := file.Name
		dbFileMap[key] = file
	}

	// 找出需要新增的文件
	var filesToCreate []*models.VirtualFile
	// 找出需要更新的文件
	var filesToUpdate []*models.VirtualFile
	// 找出需要删除的文件
	var filesToDelete []*models.VirtualFile
	// 需要递归处理的文件夹
	var foldersToProcess []*models.VirtualFile

	// 遍历扫描到的文件，找出新增和更新的文件
	for name, scannedFile := range scannedFileMap {
		if dbFile, exists := dbFileMap[name]; exists {
			// 文件存在，检查是否需要更新
			needUpdate := false

			if dbFile.Rev != scannedFile.Rev {
				s.logger.Debug("file needs update - rev changed",
					zap.String("parent", f.Name),
					zap.String("file_name", name),
					zap.String("old_rev", dbFile.Rev),
					zap.String("new_rev", scannedFile.Rev))
				needUpdate = true
			} else if s.needUpdateFile(dbFile, scannedFile) {
				// 即使rev相同也检查其他字段是否需要更新
				s.logger.Debug("file needs update - other fields changed",
					zap.String("parent", f.Name),
					zap.String("file_name", name),
					zap.String("rev", scannedFile.Rev))
				needUpdate = true
			}

			if needUpdate {
				// 需要更新
				scannedFile.ID = dbFile.ID // 保持原有ID
				scannedFile.ParentId = f.ID
				filesToUpdate = append(filesToUpdate, scannedFile)

				// 如果是文件夹，添加到递归处理列表
				if scannedFile.IsFolder == 1 {
					foldersToProcess = append(foldersToProcess, scannedFile)
				}
			} else {
				// 不需要更新，但如果是文件夹仍需要递归处理（这是关键区别）
				if dbFile.IsFolder == 1 {
					s.logger.Debug("folder unchanged but will still scan children",
						zap.String("parent", f.Name),
						zap.String("folder_name", name),
						zap.String("rev", dbFile.Rev))
					foldersToProcess = append(foldersToProcess, dbFile)
				}
			}
		} else {
			s.logger.Debug("new file found - not in database",
				zap.String("parent", f.Name),
				zap.String("file_name", name),
				zap.String("rev", scannedFile.Rev))
			// 文件不存在，需要新增
			scannedFile.ParentId = f.ID
			filesToCreate = append(filesToCreate, scannedFile)

			// 如果是新增的文件夹，也需要递归处理
			if scannedFile.IsFolder == 1 {
				foldersToProcess = append(foldersToProcess, scannedFile)
			}
		}
	}

	// 遍历数据库中的文件，找出需要删除的文件
	for name, dbFile := range dbFileMap {
		if _, exists := scannedFileMap[name]; !exists {
			s.logger.Debug("file to be deleted - not in remote",
				zap.String("parent", f.Name),
				zap.String("file_name", name),
				zap.Int64("file_id", dbFile.ID),
				zap.String("rev", dbFile.Rev))
			// 扫描结果中不存在该文件，需要删除
			filesToDelete = append(filesToDelete, dbFile)
		}
	}

	var errs = make([]error, 0)

	// 批量处理新增文件
	if len(filesToCreate) > 0 {
		if err = s.db.WithContext(ctx).CreateInBatches(filesToCreate, 100).Error; err != nil {
			errs = append(errs, fmt.Errorf("failed to create files: %w", err))
		}
	}

	// 批量处理更新文件
	for _, file := range filesToUpdate {
		if err = s.db.WithContext(ctx).Save(file).Error; err != nil {
			errs = append(errs, fmt.Errorf("failed to update file %s: %w", file.Name, err))
		}
	}

	// 递归删除文件（包括子文件和子文件夹）
	if len(filesToDelete) > 0 {
		if err = s.recursiveDelete(ctx, filesToDelete); err != nil {
			errs = append(errs, fmt.Errorf("failed to delete files: %w", err))
		}
	}

	// 递归处理子文件夹（关键：总是处理所有文件夹）
	for _, folder := range foldersToProcess {
		if err = s.deepHandleFile(ctx, folder); err != nil {
			errs = append(errs, fmt.Errorf("failed to handle subfolder %s: %w", folder.Name, err))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// needUpdateFile 检查文件是否需要更新（除了rev之外的其他字段）
func (s *ScanFileJob) needUpdateFile(dbFile, scannedFile *models.VirtualFile) bool {
	// 检查文件大小
	if dbFile.Size != scannedFile.Size {
		return true
	}

	// 检查哈希值
	if dbFile.Hash != scannedFile.Hash {
		return true
	}

	// 检查修改时间
	if dbFile.ModifyDate != scannedFile.ModifyDate {
		return true
	}

	// 检查创建时间
	if dbFile.CreateDate != scannedFile.CreateDate {
		return true
	}

	// 检查是否为文件夹的标识
	if dbFile.IsFolder != scannedFile.IsFolder {
		return true
	}

	// 可以根据需要添加更多字段的比较
	// 比如检查 Addition 字段中的某些关键信息
	if s.needUpdateAddition(dbFile.Addition, scannedFile.Addition) {
		return true
	}

	return false
}

// needUpdateAddition 检查Addition字段是否需要更新
func (s *ScanFileJob) needUpdateAddition(dbAddition, scannedAddition map[string]any) bool {
	// 检查关键字段
	keyFields := []string{"share_id", "file_id"}

	for _, key := range keyFields {
		dbValue, dbExists := dbAddition[key]
		scannedValue, scannedExists := scannedAddition[key]

		if dbExists != scannedExists {
			return true
		}

		if dbExists && dbValue != scannedValue {
			return true
		}
	}

	return false
}
