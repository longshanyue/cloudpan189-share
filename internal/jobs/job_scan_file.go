package jobs

import (
	"context"
	"runtime/debug"
	"sync"
	"time"

	"github.com/xxcheng123/cloudpan189-share/internal/bus"

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

		s.logger.Info("文件扫描任务已停止")
	})

	return nil
}

func (s *ScanFileJob) doJob(ctx context.Context) bool {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("文件扫描任务发生异常",
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
		s.logger.Info("文件扫描任务已停止")

		return false
	case <-time.After(refreshMinutes * time.Minute):
		if !shared.Setting.EnableTopFileAutoRefresh {
			return true
		}

		status := bus.Status()

		if (status.RunningCount + status.QueueLength) > 0 {
			s.logger.Info("有正在待处理或者处理中的任务，跳过本次定时扫描")

			return true
		}

		s.logger.Info("开始定时扫描文件")
		if err := bus.PublishVirtualFileScanTop(ctx); err != nil {
			s.logger.Error("定时扫描文件失败", zap.Error(err))
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
