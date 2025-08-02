package storage

import (
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"net/http"
)

type deepRefreshFileRequest struct {
	ID int64 `json:"id" binding:"required"`
}

func (s *service) DeepRefreshFile() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req = new(deepRefreshFileRequest)

		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "参数错误",
			})

			return
		}

		file := new(models.VirtualFile)
		if err := s.db.WithContext(ctx).Where("id = ?", req.ID).First(file).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "获取文件信息失败",
			})

			return
		}

		if file.IsFolder == 0 {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "只允许操作文件夹类型",
			})

			return
		}

		if err := shared.ScanJobPublish(shared.ScanJobTypeDeepRefresh, file); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "下方刷新指令失败，请稍后再试",
			})

			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusOK,
			"msg":  "刷新指令已发送",
		})
	}
}
