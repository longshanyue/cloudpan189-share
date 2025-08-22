package bus

import (
	"context"

	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/eventbus"
	"go.uber.org/zap"
)

func (w *busWorker) doSubscribeTopicAddStrmFile() eventbus.Subscription {
	return w.bus.Subscribe(TopicMediaAddStrmFile, func(ctx context.Context, data interface{}) error {
		req, ok := data.(TopicMediaAddStrmFileRequest)
		if !ok {
			return ErrRequestDataFormat
		}

		return w.addStrmMediaFile(ctx, req.FileID, req.Path)
	})
}

func (w *busWorker) doSubscribeTopicDeleteLinkVirtualFile() eventbus.Subscription {
	return w.bus.Subscribe(TopicMediaDeleteLinkFile, func(ctx context.Context, data interface{}) error {
		req, ok := data.(TopicMediaDeleteLinkFileRequest)
		if !ok {
			return ErrRequestDataFormat
		}

		return w.delMediaFile(ctx, req.FileId)
	})
}

func (w *busWorker) doSubscribeTopicMediaClearEmptyDir() eventbus.Subscription {
	return w.bus.Subscribe(TopicMediaClearEmptyDir, func(ctx context.Context, data interface{}) error {
		_, ok := data.(TopicMediaClearEmptyDirRequest)
		if !ok {
			return ErrRequestDataFormat
		}

		count, err := w.clearEmptyDirs(ctx)
		if err != nil {
			return err
		}

		w.logger.Debug("清理空文件夹完成", zap.Int64("count", count))

		return nil
	})
}

func (w *busWorker) doSubscribeTopicMediaClearAllMedia() eventbus.Subscription {
	return w.bus.Subscribe(TopicMediaClearAllMedia, func(ctx context.Context, data interface{}) error {
		req, ok := data.(TopicMediaClearAllMediaRequest)
		if !ok {
			return ErrRequestDataFormat
		}

		count, err := w.clearAllMediaFiles(ctx, req.MediaTypes...)
		if err != nil {
			return err
		}

		w.logger.Debug("清理所有媒体文件完成", zap.Int64("count", count))

		return nil
	})
}

func PublishMediaAddStrmFile(ctx context.Context, fileID int64, path string) error {
	return singletonBusWork.bus.Publish(ctx, TopicMediaAddStrmFile, TopicMediaAddStrmFileRequest{
		FileID: fileID,
		Path:   path,
	})
}

func PublishMediaClearEmptyDir(ctx context.Context) error {
	return singletonBusWork.bus.Publish(ctx, TopicMediaClearEmptyDir, TopicMediaClearEmptyDirRequest{})
}

func PublishMediaClearAllMedia(ctx context.Context, mediaTypes ...models.MediaType) error {
	return singletonBusWork.bus.Publish(ctx, TopicMediaClearAllMedia, TopicMediaClearAllMediaRequest{
		MediaTypes: mediaTypes,
	})
}
