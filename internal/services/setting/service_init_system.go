package setting

import (
	"crypto/md5"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"go.uber.org/zap"
	"net/http"
)

type initSystemRequest struct {
	Title         string `json:"title" binding:"required"`
	EnableAuth    bool   `json:"enableAuth" binding:"required"`
	BaseURL       string `json:"baseURL" binding:"required,url"`
	SuperUsername string `json:"superUsername" binding:"required,min=3,max=20"`
	SuperPassword string `json:"superPassword" binding:"required,min=6,max=20"`
}

func (s *service) InitSystem() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req = new(initSystemRequest)

		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "参数错误",
			})

			return
		}

		if shared.Setting.Initialized {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "系统已初始化",
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

		u := &models.User{
			Username:    req.SuperUsername,
			Password:    hash(req.SuperPassword),
			Permissions: models.PermissionAdmin | models.PermissionDavRead | models.PermissionBase,
		}

		if err = s.db.WithContext(ctx).Create(u).Error; err != nil {
			s.logger.Error("user create failure", zap.Error(err))

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "用户创建失败",
			})

			return
		}

		if err = s.db.WithContext(ctx).Model(&models.Setting{}).Where("id = ?", record.ID).Updates(map[string]interface{}{
			"initialized": true,
			"title":       req.Title,
			"base_url":    req.BaseURL,
			"enable_auth": req.EnableAuth,
		}).Error; err != nil {
			s.logger.Error("setting update failure", zap.Error(err))

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "初始化失败",
			})

			return
		}

		var setting = new(models.Setting)

		if err = s.db.Model(setting).Where("id = ?", 1).First(setting).Error; err != nil {
			s.logger.Error("failed to get setting", zap.Error(err))

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "系统错误，请尝试重启",
			})

			return
		}

		shared.Setting = setting

		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusOK,
			"msg":  "初始化成功",
		})
	}
}

func hash(input string) string {
	data := []byte(input)
	has := md5.Sum(data)

	return fmt.Sprintf("%x", has)
}
