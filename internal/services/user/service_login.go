package user

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
)

// 登录请求结构
type loginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6,max=20"`
}

// 登录响应结构
type loginResponse struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	TokenType    string       `json:"tokenType"`
	ExpiresIn    int64        `json:"expiresIn"`
	User         *models.User `json:"user"`
}

// Login 用户登录
func (s *service) Login() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := new(loginRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  err.Error(),
			})

			return
		}

		var user = new(models.User)
		if err := s.db.WithContext(ctx).Where("username", req.Username).First(user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				s.logger.Warn("login failed - user not found", zap.String("username", req.Username))

				ctx.JSON(http.StatusUnauthorized, gin.H{
					"code": http.StatusUnauthorized,
					"msg":  "用户名或密码错误",
				})

				return
			}

			s.logger.Error("database error during login", zap.Error(err))

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "登录失败",
			})

			return
		}

		// 验证密码
		if user.Password != hash(req.Password) {
			s.logger.Warn("login failed - invalid password",
				zap.String("username", req.Username),
				zap.Int64("user_id", user.ID))

			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "用户名或密码错误",
			})
			return
		}

		if user.Status != 1 {
			s.logger.Warn("login failed - user is inactive",
				zap.String("username", req.Username),
				zap.Int64("user_id", user.ID))

			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "用户被禁用",
			})

			return
		}

		// 生成访问Token
		accessToken, err := s.generateAccessToken(user.ID, user.Username, user.Version)
		if err != nil {
			s.logger.Error("failed to generate access token", zap.Error(err))

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "Token生成失败",
			})

			return
		}

		// 生成刷新Token
		refreshToken, err := s.generateRefreshToken(user.ID, user.Username, user.Version)
		if err != nil {
			s.logger.Error("failed to generate refresh token", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "Token生成失败",
			})
			return
		}

		ctx.JSON(http.StatusOK, &loginResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(AccessTokenExpire.Seconds()),
			User:         user,
		})
	}
}
