package advancedops

import (
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/bus"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"github.com/xxcheng123/cloudpan189-share/internal/types"
	"net/http"
)

func (s *service) ClearMedia() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		err := shared.MediaBus.Publish(ctx, bus.TopicMediaClearAllMedia, bus.TopicMediaClearAllMediaRequest{})

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, types.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "清理媒体文件失败",
			})

			return
		}

		ctx.JSON(http.StatusOK, types.SuccessResponse{
			Code:    http.StatusOK,
			Message: "清理媒体文件成功",
		})
	}
}
