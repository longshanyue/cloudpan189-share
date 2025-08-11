package jobs

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/xxcheng123/cloudpan189-interface/client"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

	refreshMinutes := time.Duration(shared.Setting.AutoRefreshMinutes)
	if refreshMinutes == 0 {
		refreshMinutes = 10
	}

	select {
	case <-s.ctx.Done():
		s.logger.Info("scan file job stopped")

		return false
	case msg := <-shared.ScanJobRead():
		_ = shared.RunningStat(msg.Msg.ID)
		defer shared.FinishStat(msg.Msg.ID)
		s.logger.Info("scan file job received message", zap.Any("msg", msg))
		if msg.Msg.ID != 0 {
			// 先检查文件还在不在
			var count int64
			s.db.WithContext(ctx).Model(new(models.VirtualFile)).Where("id = ?", msg.Msg.ID).Count(&count)
			if count == 0 {
				s.logger.Error("file not found", zap.Int64("file_id", msg.Msg.ID))

				return true
			}
		}

		switch msg.Type {
		case shared.ScanJobTypeRefresh:
			if err := newScanWorker(s).start(ctx, msg.Msg, false); err != nil {
				s.logger.Error("handle file error", zap.Error(err))
			}
		case shared.ScanJobTypeDeepRefresh:
			if err := newScanWorker(s).start(ctx, msg.Msg, true); err != nil {
				s.logger.Error("handle file error", zap.Error(err))
			}
		case shared.ScanJobTypeDel:
			if err := s.recursiveDelete(ctx, []*models.VirtualFile{msg.Msg}); err != nil {
				s.logger.Error("delete file error", zap.Error(err))
			}
		case shared.ScanJobRebuildStrm:
			if err := s.buildAllStrm(ctx); err != nil {
				s.logger.Error("rebuild stream error", zap.Error(err))
			}
		case shared.ScanJobClearStream:
			if err := s.clearAllStrm(ctx); err != nil {
				s.logger.Error("clear stream error", zap.Error(err))
			}
		}
	case <-time.After(refreshMinutes * time.Minute):
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
		_ = shared.RunningStat(f.ID)
		if err := newScanWorker(s).start(ctx, f, false); err != nil {
			errs = append(errs, err)
		}
		shared.FinishStat(f.ID)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func newScanWorker(job *ScanFileJob) *scanWorker {
	return &scanWorker{
		job:    job,
		logger: job.logger,
		db:     job.db,
		thread: shared.Setting.JobThreadCount,
	}
}

type scanWorker struct {
	job    *ScanFileJob
	logger *zap.Logger
	db     *gorm.DB
	id     int64
	// 多线程
	thread int
}

func (w *scanWorker) start(ctx context.Context, f *models.VirtualFile, deep bool) (err error) {
	w.id = f.ID

	return w.execute(ctx, f, deep)
}

func (w *scanWorker) execute(ctx context.Context, f *models.VirtualFile, deep bool) (err error) {
	var (
		scannedFiles = make([]*models.VirtualFile, 0)
	)

	switch f.OsType {
	case models.OsTypeSubscribe:
		scannedFiles, err = w.job.getSubscribeUserFiles(ctx, f)
	case models.OsTypeSubscribeShare:
		scannedFiles, err = w.job.getSubscribeShareFiles(ctx, f)
	case models.OsTypeShare:
		scannedFiles, err = w.job.getShareFiles(ctx, f)
	default:
		return errors.New("unsupported os type")
	}

	_ = shared.UpdateJobProgress(w.id, 0, int64(len(scannedFiles)))
	defer func() {
		_ = shared.UpdateJobProgress(w.id, int64(len(scannedFiles)), 0)
	}()

	if err != nil {
		return fmt.Errorf("failed to scan files for %s(%d): %w", f.Name, f.ID, err)
	}

	// 数据缓存文件
	var dbFiles = make([]*models.VirtualFile, 0)
	if err = w.db.WithContext(ctx).Where("parent_id = ?", f.ID).Find(&dbFiles).Error; err != nil {
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
				w.logger.Debug("file needs update - rev changed",
					zap.String("parent", f.Name),
					zap.String("file_name", name),
					zap.String("old_rev", dbFile.Rev),
					zap.String("new_rev", scannedFile.Rev))
				// Rev不同，需要更新
				dbFile.Name = scannedFile.Name
				dbFile.Rev = scannedFile.Rev
				dbFile.Size = scannedFile.Size
				dbFile.Hash = strings.ToLower(scannedFile.Hash)
				dbFile.CreateDate = scannedFile.CreateDate
				dbFile.ModifyDate = scannedFile.ModifyDate
				dbFile.UpdatedAt = time.Now()
				filesToUpdate = append(filesToUpdate, dbFile)
			} else if deep {
				filesToUpdate = append(filesToUpdate, dbFile)
			}
		} else {
			w.logger.Debug("new file found - not in database",
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
		if _, exists := scannedFileMap[name]; !exists && dbFile.IsTop != 1 {
			w.logger.Debug("file to be deleted - not in remote",
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
		if err = w.db.WithContext(ctx).CreateInBatches(filesToCreate, 100).Error; err != nil {
			errs = append(errs, fmt.Errorf("failed to create files: %w", err))
		}

		if shared.StrmFileEnable {
			strmFilesToCreate := make([]*models.VirtualFile, 0)
			for _, file := range filesToCreate {
				if strmFile, ok := getStrm(file); ok {
					strmFilesToCreate = append(strmFilesToCreate, strmFile)
				}
			}

			if len(strmFilesToCreate) > 0 {
				if err = w.db.WithContext(ctx).CreateInBatches(strmFilesToCreate, 100).Error; err != nil {
					errs = append(errs, fmt.Errorf("failed to create strm files: %w", err))
				}
			}
		}
	}

	// 批量处理更新文件
	for _, file := range filesToUpdate {
		if err = w.db.WithContext(ctx).Save(file).Error; err != nil {
			errs = append(errs, fmt.Errorf("failed to update file %s: %w", file.Name, err))
		}
	}

	// 递归删除文件（包括子文件和子文件夹）
	if len(filesToDelete) > 0 {
		if err = w.job.recursiveDelete(ctx, filesToDelete); err != nil {
			errs = append(errs, fmt.Errorf("failed to delete files: %w", err))
		}
	}

	// 递归处理子文件夹
	allFiles := append(filesToCreate, filesToUpdate...)
	folders := make([]*models.VirtualFile, 0)
	for _, file := range allFiles {
		if file.IsFolder == 1 { // 如果是文件夹
			folders = append(folders, file)
		}
	}

	if len(folders) > 0 {
		if err = w.processSubfoldersParallel(ctx, folders, deep); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// processSubfoldersParallel 并行处理子文件夹
func (w *scanWorker) processSubfoldersParallel(ctx context.Context, folders []*models.VirtualFile, deep bool) error {
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
	folderChan := make(chan *models.VirtualFile, len(folders))
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

	// 删除关联的文件
	if err := s.db.WithContext(ctx).Where("link_id IN ?", fileIDs).Delete(&models.VirtualFile{}).Error; err != nil {
		errs = append(errs, fmt.Errorf("failed to delete files: %w", err))
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

func (s *ScanFileJob) buildStrm(ctx context.Context, f *models.VirtualFile) error {
	if f.IsFolder != 1 {
		return nil
	}

	subFiles := make([]*models.VirtualFile, 0)

	if err := s.db.WithContext(ctx).Where("parent_id", f.ID).Find(&subFiles).Error; err != nil {
		return err
	}

	strmFilesToCreate := make([]*models.VirtualFile, 0)

	var errs []error

	for _, subFile := range subFiles {
		if strmFile, ok := getStrm(subFile); ok {
			strmFilesToCreate = append(strmFilesToCreate, strmFile)
		}

		if subFile.IsFolder == 1 {
			if err := s.buildStrm(ctx, subFile); err != nil {
				s.logger.Error("failed to build strm for %s", zap.Error(err), zap.String("file_name", subFile.Name))

				errs = append(errs, fmt.Errorf("failed to build strm for %s: %w", subFile.Name, err))
			}
		}
	}

	if len(strmFilesToCreate) > 0 {
		if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "parent_id"}, {Name: "name"}},
			DoNothing: true, // 忽略重复数据
		}).CreateInBatches(strmFilesToCreate, 100).Error; err != nil {
			errs = append(errs, fmt.Errorf("failed to create strm files: %w", err))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (s *ScanFileJob) buildAllStrm(ctx context.Context) error {
	s.logger.Info("start build all strm")

	if err := s.clearAllStrm(ctx); err != nil {
		return err
	}

	// 读取所有顶层文件
	var topFiles = make([]*models.VirtualFile, 0)
	if err := s.db.WithContext(ctx).Where("is_top = 1").Find(&topFiles).Error; err != nil {
		return err
	}

	var errs = make([]error, 0)

	for _, f := range topFiles {
		_ = shared.RunningStat(f.ID)
		if err := s.buildStrm(ctx, f); err != nil {
			errs = append(errs, err)
		}
		shared.FinishStat(f.ID)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (s *ScanFileJob) clearAllStrm(ctx context.Context) error {
	if err := s.db.WithContext(ctx).Where("os_type = ?", models.OsTypeStrmFile).Delete(&models.VirtualFile{}).Error; err != nil {
		s.logger.Error("failed to delete strm files", zap.Error(err))

		return err
	}

	return nil
}

func getStrm(f *models.VirtualFile) (*models.VirtualFile, bool) {
	if f.IsFolder == 1 {
		return nil, false
	}

	extName := strings.TrimPrefix(filepath.Ext(f.Name), ".")
	if len(shared.StrmSupportFileExtList) > 0 && lo.IndexOf(shared.StrmSupportFileExtList, extName) == -1 {
		return nil, false
	}

	now := time.Now()

	// 计算 size
	size := len(fmt.Sprintf("%s/api/file_download?id=%d&random=%s&sign=12345678123456781234567812345678&timestamp=-1", shared.Setting.BaseURL, f.ID, uuid.NewString()))

	strmFile := &models.VirtualFile{
		ParentId:   f.ParentId,
		LinkId:     f.ID,
		Name:       strings.TrimSuffix(f.Name, filepath.Ext(f.Name)) + ".strm",
		IsTop:      f.IsTop,
		Size:       int64(size),
		IsFolder:   f.IsFolder,
		Hash:       "-",
		CreateDate: now.Format(time.DateTime),
		ModifyDate: now.Format(time.DateTime),
		OsType:     models.OsTypeStrmFile,
		Addition:   make(datatypes.JSONMap),
		Rev:        now.Format("20060102150405"),
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	return strmFile, true
}
