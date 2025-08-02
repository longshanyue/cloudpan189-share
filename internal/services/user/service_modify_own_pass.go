package user

import (
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"go.uber.org/zap"
	"net/http"
)

type modifyOwnPassRequest struct {
	OldPassword string `json:"oldPassword" binding:"required,min=6,max=20"`
	Password    string `json:"password" binding:"required,min=6,max=20"`
}

type modifyOwnPassResponse struct {
	RowsAffected int64 `json:"rowsAffected"`
}

func (s *service) ModifyOwnPass() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := new(modifyOwnPassRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  err.Error(),
			})
			return
		}

		// 从上下文获取当前用户ID
		userID, exists := ctx.Get("user_id")
		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "用户未登录",
			})
			return
		}

		// 类型断言，确保userID是正确的类型
		uid, ok := userID.(int64)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "用户ID格式错误",
			})
			return
		}

		// 先查询用户当前密码进行验证
		var user models.User
		if err := s.db.WithContext(ctx).Select("password").Where("id = ?", uid).First(&user).Error; err != nil {
			s.logger.Error("user query failure", zap.Error(err))
			ctx.JSON(http.StatusNotFound, gin.H{
				"code": http.StatusNotFound,
				"msg":  "用户不存在",
			})
			return
		}

		// 验证旧密码是否正确
		if user.Password != hash(req.OldPassword) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "旧密码错误",
			})
			return
		}

		// 构建更新数据，包含加密后的新密码和版本号+1
		updateData := map[string]any{
			"password": hash(req.Password),
			"version":  s.db.Raw("version + 1"), // 版本号自增1
		}

		// 执行更新操作
		result := s.db.WithContext(ctx).Model(new(models.User)).
			Where("id = ?", uid).
			Updates(updateData)

		if result.Error != nil {
			s.logger.Error("user modify own password failure", zap.Error(result.Error))

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "修改密码失败",
			})
			return
		}

		// 检查是否找到了对应的用户
		if result.RowsAffected == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code": http.StatusNotFound,
				"msg":  "用户不存在",
			})
			return
		}

		ctx.JSON(http.StatusOK, &modifyOwnPassResponse{
			RowsAffected: result.RowsAffected,
		})
	}
}
