package setting

import "github.com/gin-gonic/gin"

import (
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"net/http"
)

type toggleEnableTopFileAutoRefreshRequest struct {
	Disable bool `json:"disable"`
}

func (s *service) ToggleEnableTopFileAutoRefresh() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req toggleEnableTopFileAutoRefreshRequest

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

		if err := s.db.WithContext(ctx).Model(&models.Setting{}).Where("id = ?", record.ID).Update("enable_top_file_auto_refresh", !req.Disable).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "修改失败",
			})

			return
		}

		shared.Setting.EnableTopFileAutoRefresh = !req.Disable

		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusOK,
			"msg":  "修改成功",
		})
	}
}
