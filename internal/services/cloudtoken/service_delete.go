package cloudtoken

import (
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"net/http"
)

type deleteRequest struct {
	ID int64 `json:"id" binding:"required,min=1"`
}

func (s *service) Delete() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req deleteRequest

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "参数错误",
			})

			return
		}

		// 检查有没有被绑定

		if err := s.db.Where("id = ?", req.ID).Delete(&models.CloudToken{}).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "删除失败",
			})

			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusOK,
			"msg":  "删除成功",
		})
	}
}
