package bus

import (
	"github.com/pkg/errors"
	"github.com/xxcheng123/cloudpan189-interface/client"
	"github.com/xxcheng123/cloudpan189-share/configs"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"sync"
	"time"
)

type Worker interface {
	Register() error
	Close()
}

var (
	ErrBusRegistered = errors.New("bus registered")
	FileNotFound     = errors.New("file not found")
)

func Init() {
	fileBusWorkerInstance = &fileBusWorker{
		logger: configs.Logger().With(zap.String("worker", "bus_worker")),
		client: client.New(),
		db:     configs.DB(),
		lock:   sync.Mutex{},
	}

	if err := fileBusWorkerInstance.Register(); err != nil {
		panic(err)
	}

	mediaBusWorkerInstance = &mediaBusWorker{
		logger: configs.Logger().With(zap.String("worker", "media_bus_worker")),

		db:   configs.DB(),
		lock: sync.Mutex{},
	}

	if err := mediaBusWorkerInstance.Register(); err != nil {
		panic(err)
	}
}

func Close() {
	fileBusWorkerInstance.Close()

	mediaBusWorkerInstance.Close()
}

type busLog struct {
	id        int64
	startTime time.Time
	db        *gorm.DB
}

func (log *busLog) End(result string) error {
	if log == nil {
		return errors.New("log is nil")
	}

	return log.db.Model(&models.SystemLog{}).Where("id = ?", log.id).Updates(map[string]any{
		"result": result,
		"cost":   time.Now().Sub(log.startTime).Milliseconds(),
	}).Error
}

func newLog(db *gorm.DB, title string) (*busLog, error) {
	now := time.Now()

	log := &models.SystemLog{
		Title:  title,
		Type:   "bus",
		Cost:   0,
		Result: "创建成功",
	}

	if err := db.Create(log).Error; err != nil {
		return nil, err
	}

	return &busLog{
		id:        log.ID,
		startTime: now,
		db:        db,
	}, nil
}
