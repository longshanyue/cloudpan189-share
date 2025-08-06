package setting

import (
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"net/http"
)

type modifyAutoRefreshMinutesRequest struct {
	AutoRefreshMinutes int `json:"autoRefreshMinutes" binding:"required,min=5,max=120"`
}

func (s *service) ModifyAutoRefreshMinutes() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req modifyAutoRefreshMinutesRequest

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "参数错误，自动刷新间隔必须在5-120分钟之间",
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

		if err = s.db.WithContext(ctx).Model(&models.Setting{}).Where("id = ?", record.ID).Update("auto_refresh_minutes", req.AutoRefreshMinutes).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "修改失败",
			})

			return
		}

		shared.Setting.AutoRefreshMinutes = req.AutoRefreshMinutes

		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusOK,
			"msg":  "修改成功",
		})
	}
}
