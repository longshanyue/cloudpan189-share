package cloudtoken

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-interface/client"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
)

type checkQrcodeRequest struct {
	ID   int64  `json:"id" binding:"omitempty"`
	UUID string `json:"uuid" binding:"required"`
}

func (s *service) CheckQrcode() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req checkQrcodeRequest

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "参数错误",
			})

			return
		}

		resp, err := client.LoginQuery(req.UUID)
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  fmt.Sprintf("登录查询失败: %s", err.Error()),
			})

			return
		}

		if req.ID != 0 {
			if err = s.db.WithContext(ctx).Model(&models.CloudToken{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
				"status":       1,
				"access_token": resp.AccessToken,
				"expires_in":   resp.ExpiresIn,
			}).Error; err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code": http.StatusInternalServerError,
					"msg":  fmt.Sprintf("更新云盘令牌失败: %s", err.Error()),
				})

				return
			}
		} else {
			if err = s.db.WithContext(ctx).Model(&models.CloudToken{}).Create(&models.CloudToken{
				Name:        "云盘令牌",
				Status:      1,
				AccessToken: resp.AccessToken,
				ExpiresIn:   resp.ExpiresIn,
				LoginType:   models.LoginTypeScan,
				Addition:    map[string]interface{}{},
			}).Error; err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code": http.StatusInternalServerError,
					"msg":  fmt.Sprintf("创建云盘令牌失败: %s", err.Error()),
				})

				return
			}
		}

		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusOK,
			"msg":  "绑定成功",
		})
	}
}
