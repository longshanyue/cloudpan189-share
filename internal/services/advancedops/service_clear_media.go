package advancedops

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/bus"
	"github.com/xxcheng123/cloudpan189-share/internal/types"
)

func (s *service) ClearMedia() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := bus.PublishMediaClearAllMedia(ctx); err != nil {
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
