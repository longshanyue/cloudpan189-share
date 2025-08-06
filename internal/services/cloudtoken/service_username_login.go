package cloudtoken

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tickstep/cloudpan189-api/cloudpan"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type usernameLoginRequest struct {
	ID       int64  `json:"id" binding:"omitempty"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type addByUsernameResponse struct {
	ID int64 `json:"id"`
}
type modifyUsernameLoginResponse struct {
	RowsAffected int64 `json:"rowsAffected"`
}

func (s *service) UsernameLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := new(usernameLoginRequest)

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "参数错误",
			})

			return
		}

		loginResult, loginErr := cloudpan.AppLogin(req.Username, req.Password)
		if loginErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  fmt.Sprintf("登录失败: %s", loginErr.Error()),
			})

			return
		}

		if req.ID > 0 {
			// 检测信息
			var oldToken models.CloudToken
			if err := s.db.WithContext(ctx).Model(&models.CloudToken{}).Where("id = ?", req.ID).First(&oldToken).Error; err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code": http.StatusInternalServerError,
					"msg":  fmt.Sprintf("查询云盘令牌失败: %s", err.Error()),
				})

				return
			} else if oldToken.LoginType != models.LoginTypePassword {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"code": http.StatusBadRequest,
					"msg":  "云盘令牌类型错误",
				})

				return
			}

			addition := oldToken.Addition
			addition[models.CloudTokenAdditionAutoLoginResultKey] = fmt.Sprintf("%s, token 刷新成功", time.Now().Format(time.DateTime))
			addition[models.CloudTokenAdditionAutoLoginTimes] = 0

			updateMap := map[string]interface{}{
				"access_token": loginResult.SskAccessToken,
				"expires_in":   loginResult.SskAccessTokenExpiresIn,
				"username":     req.Username,
				"password":     req.Password,
				"addition":     addition,
			}

			result := s.db.Model(&models.CloudToken{}).Where("id = ?", oldToken.ID).Updates(updateMap)

			if result.Error != nil {
				s.logger.Error("failed to update cloud token", zap.Error(result.Error))

				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code": http.StatusInternalServerError,
					"msg":  "更新云盘令牌失败",
				})

				return
			}

			ctx.JSON(http.StatusOK, modifyUsernameLoginResponse{
				RowsAffected: result.RowsAffected,
			})
		} else {
			m := &models.CloudToken{
				Name:        "云盘令牌（账密）",
				Status:      1,
				AccessToken: loginResult.SskAccessToken,
				ExpiresIn:   loginResult.SskAccessTokenExpiresIn,
				Username:    req.Username,
				Password:    req.Password,
				LoginType:   models.LoginTypePassword,
				Addition:    map[string]interface{}{},
			}

			if err := s.db.WithContext(ctx).Model(&models.CloudToken{}).Create(m).Error; err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code": http.StatusInternalServerError,
					"msg":  fmt.Sprintf("创建云盘令牌失败: %s", err.Error()),
				})

				return
			}

			ctx.JSON(http.StatusOK, addByUsernameResponse{
				ID: m.ID,
			})
		}
	}
}
