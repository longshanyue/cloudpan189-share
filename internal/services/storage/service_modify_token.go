package storage

import (
	"net/http"

	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/models"

	"github.com/gin-gonic/gin"
)

type modifyTokenRequest struct {
	ID         int64 `json:"id" binding:"required"`
	CloudToken int64 `json:"cloudToken" binding:"required"`
}

type modifyTokenResponse struct {
	RowsAffected int64 `json:"rowsAffected"`
}

func (s *service) ModifyToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req modifyTokenRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  err.Error(),
			})

			return
		}

		file := new(models.VirtualFile)
		if err := s.db.First(file, req.ID).Error; err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code": http.StatusNotFound,
				"msg":  "文件不存在",
			})

			return
		}

		if file.IsTop != 1 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "不是挂载点",
			})

			return
		}

		file.Addition[consts.FileAdditionKeyCloudToken] = req.CloudToken

		result := s.db.Model(&models.VirtualFile{}).Where("id = ?", req.ID).Update("addition", file.Addition)

		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "修改失败",
			})

			return
		}

		ctx.JSON(http.StatusOK, modifyTokenResponse{
			RowsAffected: result.RowsAffected,
		})
	}
}
