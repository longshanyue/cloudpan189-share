package setting

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
)

func (s *service) RefreshKey() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		record, err := s.get(ctx)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "查询配置失败",
			})

			return
		}

		key := utils.GenerateRandomPassword(16)

		if err := s.db.Model(&models.Setting{}).Where("id = ?", record.ID).Update("salt_key", key).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "刷新密钥失败",
			})

			return
		}

		shared.Setting.SaltKey = key

		ctx.JSON(http.StatusOK, gin.H{
			"msg": "刷新密钥成功",
		})
	}
}
