package bus

import (
	"context"

	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/eventbus"
	"go.uber.org/zap"
)

var (
	allowAutoDelOsTypes = []string{
		models.OsTypeFolder,
		models.OsTypeFile,
		models.OsTypeSubscribe,
		models.OsTypeSubscribeShare,
		models.OsTypeShare,
	}
)

func (w *busWorker) doSubscribeTopicFileRefreshFile() eventbus.Subscription {
	return w.bus.Subscribe(TopicFileRefreshFile, func(ctx context.Context, data interface{}) error {
		req, ok := data.(TopicFileRefreshFileRequest)
		if !ok {
			return ErrRequestDataFormat
		}

		var (
			fileId = req.FileId
			deep   = req.Deep
		)

		return w.scanVirtualFile(ctx, fileId, deep)
	})
}

func (w *busWorker) doSubscribeTopicFileDeleteFile() eventbus.Subscription {
	return w.bus.Subscribe(TopicFileDeleteFile, func(ctx context.Context, data interface{}) error {
		req, ok := data.(TopicFileDeleteRequest)
		if !ok {
			return ErrRequestDataFormat
		}

		return w.deleteVirtualFile(ctx, req.FileId)
	})
}

func (w *busWorker) doSubscribeTopicScanTop() eventbus.Subscription {
	return w.bus.Subscribe(TopicFileScanTop, func(ctx context.Context, data interface{}) error {
		return w.scanTopVirtualFiles(ctx)
	})
}

func (w *busWorker) doSubscribeTopicRebuildMediaFile() eventbus.Subscription {
	return w.bus.Subscribe(TopicFileRebuildMediaFile, func(ctx context.Context, data interface{}) error {
		req, ok := data.(TopicFileRebuildMediaFileRequest)
		if !ok {
			return ErrRequestDataFormat
		}

		mediaReq := TopicMediaClearAllMediaRequest{}
		if len(req.MediaTypes) > 0 {
			mediaReq.MediaTypes = req.MediaTypes
		}

		_, err := w.clearAllMediaFiles(ctx, mediaReq.MediaTypes...)
		if err != nil {
			w.logger.Error("删除旧媒体文件失败", zap.Error(err))
		}

		count, err := w.buildMediaFile(ctx, 0)
		if err != nil {
			w.logger.Error("重建媒体文件失败", zap.Error(err))

			return err
		}

		w.logger.Info("重建媒体文件完成", zap.Int64("count", count))

		return nil
	})
}

func PublishVirtualFileRefresh(ctx context.Context, fileId int64, deep bool) error {
	return singletonBusWork.bus.Publish(ctx, TopicFileRefreshFile, TopicFileRefreshFileRequest{
		FileId: fileId,
		Deep:   deep,
	})
}

func PublishVirtualFileDelete(ctx context.Context, fileId int64) error {
	return singletonBusWork.bus.Publish(ctx, TopicFileDeleteFile, TopicFileDeleteRequest{
		FileId: fileId,
	})
}

func PublishVirtualFileScanTop(ctx context.Context) error {
	return singletonBusWork.bus.Publish(ctx, TopicFileScanTop, nil)
}

func PublishRebuildMediaFile(ctx context.Context, mediaTypes ...models.MediaType) error {
	return singletonBusWork.bus.Publish(ctx, TopicFileRebuildMediaFile, TopicFileRebuildMediaFileRequest{
		MediaTypes: mediaTypes,
	})
}
