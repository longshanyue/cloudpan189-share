package jobs

import (
	"context"
	"github.com/xxcheng123/cloudpan189-share/internal/bus"
	"runtime/debug"
	"sync"
	"time"

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
	case <-time.After(refreshMinutes * time.Minute):
		if !shared.Setting.EnableTopFileAutoRefresh {
			return true
		}

		//case <-time.After(time.Second * 5):
		stats := shared.FileBus.GetRuntimeStats()
		if (stats.QueueLength + stats.ProcessingWorkers) > 0 {
			s.logger.Info("正在待处理或者处理中的任务，跳过本次定时扫描")

			return true
		}

		s.logger.Info("scan top file job started")
		if err := shared.FileBus.Publish(context.Background(), bus.TopicFileScanTop, nil); err != nil {
			s.logger.Error("failed to publish scan top job", zap.Error(err))
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
