package bus

import (
	"sync"

	"github.com/xxcheng123/cloudpan189-interface/client"
	"github.com/xxcheng123/cloudpan189-share/configs"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/eventbus"
	"go.uber.org/zap"
)

var onceLoad sync.Once

func Init() {
	onceLoad.Do(func() {
		logger := configs.Logger().With(zap.String("bus", "singleton_bus"))

		singletonBusWork = &busWorker{
			db:     configs.DB(),
			logger: logger,
			bus: eventbus.NewWithConfig(&eventbus.Config{
				BufferSize:     8,
				MaxConcurrency: 1,
			}),
			client: client.New(),
		}

		singletonBusWork.doSubscribe()
	})
}
