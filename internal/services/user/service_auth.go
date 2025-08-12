package user

import (
	"errors"
	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (s *service) AuthMiddleware(permission uint8) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "缺少Authorization头",
			})

			ctx.Abort()

			return
		}

		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "Authorization头格式错误",
			})

			ctx.Abort()

			return
		}

		uid, username, version, err := s.ParseAccessToken(tokenParts[1])
		if err != nil {
			s.logger.Warn("invalid token", zap.Error(err))
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "无效的Token",
			})

			ctx.Abort()

			return
		}

		// 验证用户是否存在且激活
		var user = new(models.User)
		if err = s.db.WithContext(ctx).Where("id", uid).Where("status", 1).First(user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				s.logger.Warn("user not found or inactive", zap.Int64("user_id", uid))
				ctx.JSON(http.StatusUnauthorized, gin.H{
					"code": http.StatusUnauthorized,
					"msg":  "用户不存在或已禁用",
				})

				ctx.Abort()

				return
			}
			s.logger.Error("database error during token check", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "Token验证失败",
			})

			ctx.Abort()

			return
		}

		if user.Version > version {
			s.logger.Warn("user version mismatch",
				zap.Int64("user_id", uid),
				zap.Int("user_version", user.Version),
				zap.Int("token_version", version))

			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "用户信息已更新，请重新登录",
			})

			ctx.Abort()

			return
		}

		//检查权限够不够 (位计算)
		if user.Permissions&permission == 0 {
			s.logger.Warn("insufficient permissions",
				zap.Int64("user_id", uid),
				zap.String("username", username),
				zap.Uint8("required_permissions", permission),
				zap.Uint8("user_permissions", user.Permissions))

			ctx.JSON(http.StatusForbidden, gin.H{
				"code": http.StatusForbidden,
				"msg":  "权限不足",
			})

			ctx.Abort()

			return
		}

		ctx.Set("user_id", uid)
		ctx.Set("username", username)
		ctx.Set("permissions", user.Permissions)
		ctx.Set("group_id", user.GroupID)
		ctx.Set(consts.CtxKeyGroupId, user.GroupID)

		ctx.Next()
	}
}

func (s *service) BasicAuthMiddleware(permission uint8) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !shared.Setting.EnableAuth {
			ctx.Next()

			return
		}

		username, password, ok := ctx.Request.BasicAuth()
		if !ok {
			ctx.Header("WWW-Authenticate", `Basic realm="Restricted"`)

			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "请输入用户名和密码",
			})

			ctx.Abort()

			return
		}

		// 验证用户是否存在且激活
		var user = new(models.User)
		if err := s.db.WithContext(ctx).Where("username", username).Where("status", 1).First(user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusUnauthorized, gin.H{
					"code": http.StatusUnauthorized,
					"msg":  "用户不存在或已禁用",
				})

				ctx.Abort()

				return
			}

			s.logger.Error("database error during token check", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "验证失败",
			})

			ctx.Abort()

			return
		}

		if hash(password) != user.Password {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "密码错误",
			})

			ctx.Abort()

			return
		}

		//检查权限够不够 (位计算)
		if user.Permissions&permission == 0 {
			s.logger.Warn("insufficient permissions",
				zap.Int64("user_id", user.ID),
				zap.String("username", username),
				zap.Uint8("required_permissions", permission),
				zap.Uint8("user_permissions", user.Permissions))

			ctx.JSON(http.StatusForbidden, gin.H{
				"code": http.StatusForbidden,
				"msg":  "权限不足",
			})

			ctx.Abort()

			return
		}

		ctx.Set("user_id", user.ID)
		ctx.Set("username", username)
		ctx.Set("permissions", user.Permissions)
		ctx.Set("group_id", user.GroupID)

		ctx.Next()
	}
}
