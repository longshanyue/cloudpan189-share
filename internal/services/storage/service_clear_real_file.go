package storage

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
)

type clearRealFileResponse struct {
	Message string `json:"message"`
}

func (s *service) ClearRealFile() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 使用根虚拟文件来触发清空任务
		rootFile := new(models.VirtualFile)

		if err := shared.ScanJobPublish(shared.ScanJobClearRealFile, rootFile); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  fmt.Sprintf("下发清空本地真实存储任务失败：%s", err.Error()),
			})

			return
		}

		ctx.JSON(http.StatusOK, clearRealFileResponse{
			Message: "清空本地真实存储任务已发送",
		})
	}
}
