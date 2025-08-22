package bus

import (
	"context"
	"github.com/bradenaw/juniper/xsync"
	"sync"

	"github.com/pkg/errors"
	"github.com/xxcheng123/cloudpan189-interface/client"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/eventbus"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	FileNotFound         = errors.New("file not found")
	ErrRequestDataFormat = errors.New("request data format error")
)

var singletonBusWork *busWorker

type busWorker struct {
	db     *gorm.DB
	logger *zap.Logger
	bus    eventbus.EventBus
	client client.Client

	dbLock sync.Mutex

	fileScanStat xsync.Map[int64, *FileScanStat]
}

type FileScanStat struct {
	FileId       int64 `json:"fileId"`
	WaitCount    int64 `json:"waitCount"`
	ScannedCount int64 `json:"scannedCount"`
}

// withLock  默认使用 sqlite，性能差
func (w *busWorker) withLock(ctx context.Context, fn func(db *gorm.DB) *gorm.DB) *gorm.DB {
	w.dbLock.Lock()
	defer w.dbLock.Unlock()

	return fn(w.db.WithContext(ctx))
}
func (w *busWorker) getDB(ctx context.Context) *gorm.DB {
	return w.db.WithContext(ctx)
}

func (w *busWorker) doSubscribe() {
	w.doSubscribeTopicFileRefreshFile()
	w.doSubscribeTopicFileDeleteFile()
	w.doSubscribeTopicScanTop()
	w.doSubscribeTopicRebuildMediaFile()

	w.doSubscribeTopicMediaClearEmptyDir()
	w.doSubscribeTopicMediaClearAllMedia()
	w.doSubscribeTopicAddStrmFile()
	w.doSubscribeTopicDeleteLinkVirtualFile()
}

func Status() eventbus.BusStats {
	a := singletonBusWork.bus.GetPendingTasks()
	b := singletonBusWork.bus.GetRunningTasks()
	_, _ = a, b

	return singletonBusWork.bus.GetStats()
}

type DetailInfo struct {
	RunningTasks []eventbus.TaskInfo `json:"runningTasks"`
	PendingTasks []eventbus.TaskInfo `json:"pendingTasks"`
	Stats        eventbus.BusStats   `json:"stats"`
}

func Detail() DetailInfo {
	return DetailInfo{
		RunningTasks: singletonBusWork.bus.GetRunningTasks(),
		PendingTasks: singletonBusWork.bus.GetPendingTasks(),
		Stats:        singletonBusWork.bus.GetStats(),
	}
}

func FindScanFileStat(fileId int64) *FileScanStat {
	v, ok := singletonBusWork.fileScanStat.Load(fileId)
	if !ok {
		return nil
	}

	return v
}
