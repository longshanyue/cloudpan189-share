package setting

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
)

type modifyBaseURLRequest struct {
	BaseURL string `json:"baseURL" binding:"required,url,min=1,max=255"`
}

func (s *service) ModifyBaseURL() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req modifyBaseURLRequest

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "参数错误",
			})

			return
		}

		record, err := s.get(ctx)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "查询配置失败",
			})

			return
		}

		if err := s.db.WithContext(ctx).Model(&models.Setting{}).Where("id = ?", record.ID).Update("base_url", req.BaseURL).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "修改失败",
			})

			return
		}

		shared.Setting.BaseURL = req.BaseURL

		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusOK,
			"msg":  "修改成功",
		})
	}
}
