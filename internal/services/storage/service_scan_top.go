package storage

import (
	"github.com/xxcheng123/cloudpan189-share/internal/bus"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
)

type scanTopResponse struct {
	Message string `json:"message"`
}

func (s *service) ScanTop() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := shared.FileBus.Publish(ctx, bus.TopicFileScanTop, nil); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "发布扫描顶层文件任务失败：" + err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, scanTopResponse{
			Message: "扫描顶层文件任务已发送",
		})
	}
}
