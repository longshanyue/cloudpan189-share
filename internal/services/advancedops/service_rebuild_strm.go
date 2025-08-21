package advancedops

import (
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/bus"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"github.com/xxcheng123/cloudpan189-share/internal/types"
	"net/http"
)

// RebuildStrm 重建 strm 文件
func (s *service) RebuildStrm() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		err := shared.FileBus.Publish(ctx, bus.TopicFileRebuildMediaFile, bus.TopicFileRebuildMediaFileRequest{
			MediaTypes: []models.MediaType{models.MediaTypeStrm},
		})

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, types.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "重建 strm 文件失败",
			})

			return
		}

		ctx.JSON(http.StatusOK, types.SuccessResponse{
			Code:    http.StatusOK,
			Message: "重建 strm 文件成功",
		})
	}
}
