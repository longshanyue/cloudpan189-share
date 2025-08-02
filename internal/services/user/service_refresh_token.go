package user

import (
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// RefreshToken 刷新Token模型
type RefreshToken struct {
	UserID    int64     `json:"userId"`
	Token     string    `json:"token"`
	Version   int       `json:"version"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// 刷新Token请求结构
type refreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// RefreshToken 刷新访问Token
func (s *service) RefreshToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := new(refreshRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  err.Error(),
			})

			return
		}

		// 解析刷新Token
		uid, _, version, err := s.parseRefreshToken(req.RefreshToken)
		if err != nil {
			s.logger.Warn("invalid refresh token", zap.Error(err))
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "无效的刷新Token",
			})

			return
		}

		// 获取用户信息
		var user = new(models.User)
		if err := s.db.WithContext(ctx).First(user, uid).Error; err != nil {
			s.logger.Error("user not found during token refresh", zap.Error(err))
			ctx.JSON(http.StatusNotFound, gin.H{
				"code": http.StatusNotFound,
				"msg":  "用户不存在",
			})

			return
		}

		if user.Status != 1 {
			s.logger.Warn("inactive user attempted token refresh", zap.Int64("user_id", uid))
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "用户账户已禁用",
			})

			return
		}

		if user.Version > version {
			s.logger.Warn("user version mismatch during token refresh", zap.Int64("user_id", uid))
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "用户信息已更新，请重新登录",
			})

			return
		}

		// 生成新的Token
		accessToken, err := s.generateAccessToken(user.ID, user.Username, user.Version)
		if err != nil {
			s.logger.Error("failed to generate access token", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "Token生成失败",
			})

			return
		}

		newRefreshToken, err := s.generateRefreshToken(user.ID, user.Username, user.Version)
		if err != nil {
			s.logger.Error("failed to generate refresh token", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "Token生成失败",
			})

			return
		}

		s.logger.Info("token refreshed successfully", zap.Int64("user_id", user.ID))

		ctx.JSON(http.StatusOK, &loginResponse{
			AccessToken:  accessToken,
			RefreshToken: newRefreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(AccessTokenExpire.Seconds()),
			User:         user,
		})
	}
}
