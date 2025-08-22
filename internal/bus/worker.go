package bus

import (
	"context"
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
