package user

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
)

func (s *service) Info() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, exists := ctx.Get("user_id")

		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "用户未认证",
			})

			return
		}

		var user models.User
		if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				s.logger.Warn("user not found", zap.Uint("user_id", userID.(uint)))
				ctx.JSON(http.StatusNotFound, gin.H{
					"code": http.StatusNotFound,
					"msg":  "用户不存在",
				})

				return
			}
			s.logger.Error("database error getting user info", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "获取用户信息失败",
			})

			return
		}

		s.logger.Info("user info retrieved successfully",
			zap.Int64("user_id", user.ID),
			zap.String("username", user.Username))

		ctx.JSON(http.StatusOK, user)
	}
}
