package main

import (
	"context"

	"github.com/xxcheng123/cloudpan189-share/configs"
	"github.com/xxcheng123/cloudpan189-share/internal/jobs"
	"github.com/xxcheng123/cloudpan189-share/internal/router"
	"go.uber.org/zap"
)

func main() {
	configs.Init()

	defer configs.Logger().Sync()

	scanJob := jobs.NewScanFileJob(configs.DB(), configs.Logger())
	if err := scanJob.Start(context.Background()); err != nil {
		panic(err)
	}

	defer scanJob.Stop()

	if err := router.StartHTTPServer(); err != nil {
		configs.Logger().Error("start http server error", zap.Error(err))
	}
}
