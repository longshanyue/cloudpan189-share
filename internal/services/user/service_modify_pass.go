package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"go.uber.org/zap"
)

type modifyPassRequest struct {
	ID       int64  `json:"id" binding:"required,min=1"`
	Password string `json:"password" binding:"required,min=6,max=20"`
}

type modifyPassResponse struct {
	RowsAffected int64 `json:"rowsAffected"`
}

func (s *service) ModifyPass() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := new(modifyPassRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  err.Error(),
			})
			return
		}

		// 构建更新数据，包含加密后的密码和版本号+1
		updateData := map[string]any{
			"password": hash(req.Password),
			"version":  s.db.Raw("version + 1"), // 版本号自增1
		}

		// 执行更新操作
		result := s.db.WithContext(ctx).Model(new(models.User)).
			Where("id = ?", req.ID).
			Updates(updateData)

		if result.Error != nil {
			s.logger.Error("user modify password failure", zap.Error(result.Error))

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

		ctx.JSON(http.StatusOK, &modifyPassResponse{
			RowsAffected: result.RowsAffected,
		})
	}
}
