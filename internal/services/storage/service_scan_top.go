package storage

import (
	"net/http"

	"github.com/xxcheng123/cloudpan189-share/internal/bus"

	"github.com/gin-gonic/gin"
)

type scanTopResponse struct {
	Message string `json:"message"`
}

func (s *service) ScanTop() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := bus.PublishVirtualFileScanTop(ctx); err != nil {
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
