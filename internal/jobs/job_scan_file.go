package jobs

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/xxcheng123/cloudpan189-share/internal/fs"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/xxcheng123/cloudpan189-interface/client"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type File = *models.VirtualFile

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
		s.logger.Info("开始执行扫描文件任务", zap.Any("msg", msg))
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
			if err := newDiffWorker(s.logger, s.db, msg.Msg.ID).execute(ctx, msg.Msg, false); err != nil {
				s.logger.Error("扫描文件任务执行失败", zap.Error(err))
			}
		case shared.ScanJobTypeDeepRefresh:
			if err := newDiffWorker(s.logger, s.db, msg.Msg.ID).execute(ctx, msg.Msg, true); err != nil {
				s.logger.Error("深度扫描文件任务执行失败", zap.Error(err))
			}
		case shared.ScanJobTypeDel:
			if err := fs.NewFS(s.db, s.logger).Delete(ctx, msg.Msg.ID); err != nil {
				s.logger.Error("删除文件任务执行失败", zap.Error(err))
			}
		case shared.ScanJobRebuildStrm:
			if err := s.buildAllStrm(ctx); err != nil {
				s.logger.Error("rebuild stream error", zap.Error(err))
			}
		case shared.ScanJobClearStrm:
			if count, err := fs.NewFS(s.db, s.logger).ClearOsType(ctx, models.OsTypeStrmFile); err != nil {
				s.logger.Error("删除strm文件任务执行失败", zap.Error(err))
			} else {
				s.logger.Info("删除strm文件成功", zap.Int64("count", count))
			}
		case shared.ScanJobClearRealFile:
			if count, err := fs.NewFS(s.db, s.logger).ClearOsType(ctx, models.OsTypeRealFile); err != nil {
				s.logger.Error("删除真存储文件任务执行失败", zap.Error(err))
			} else {
				s.logger.Info("删除真存储文件成功", zap.Int64("count", count))
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
	s.mu.Lock()
	defer s.mu.Unlock()
	
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
		if err := newDiffWorker(s.logger, s.db, f.ID).execute(ctx, f, false); err != nil {
			errs = append(errs, err)
		}
		shared.FinishStat(f.ID)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (s *ScanFileJob) buildStrm(ctx context.Context, f *models.VirtualFile) error {
	if f.IsFolder != 1 {
		return nil
	}

	subFiles := make([]*models.VirtualFile, 0)

	if err := s.db.WithContext(ctx).Where("parent_id = ?", f.ID).Find(&subFiles).Error; err != nil {
		return err
	}

	strmFilesToCreate := make([]*models.VirtualFile, 0)

	var errs []error

	for _, subFile := range subFiles {
		if strmFile, ok := fs.GetStrm(subFile); ok {
			strmFilesToCreate = append(strmFilesToCreate, strmFile)
		}

		if subFile.IsFolder == 1 {
			if err := s.buildStrm(ctx, subFile); err != nil {
				s.logger.Error("failed to build strm", zap.Error(err), zap.String("file_name", subFile.Name))

				errs = append(errs, fmt.Errorf("failed to build strm for %s: %w", subFile.Name, err))
			}
		}
	}

	if len(strmFilesToCreate) > 0 {
		count, err := fs.NewFS(s.db, s.logger).BatchCreate(ctx, f.ID, strmFilesToCreate...)
		if err != nil {
			s.logger.Error("failed to create strm files", zap.Error(err))

			errs = append(errs, fmt.Errorf("failed to create strm files: %w", err))
		} else {
			s.logger.Info("创建strm文件成功", zap.Int64("count", count))
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
