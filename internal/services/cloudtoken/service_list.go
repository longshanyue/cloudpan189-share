package cloudtoken

import (
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"net/http"
)

type listRequest struct {
	Name string `form:"name" binding:"omitempty"`
}

func (s *service) List() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req listRequest

		if err := ctx.ShouldBindQuery(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "参数错误",
			})

			return
		}

		var cloudTokens = make([]*models.CloudToken, 0)

		query := s.db.WithContext(ctx)
		if req.Name != "" {
			query = query.Where("name like ?", "%"+req.Name+"%")
		}

		if err := query.Find(&cloudTokens).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "查询失败",
			})

			return
		}

		ctx.JSON(http.StatusOK, cloudTokens)
	}
}
