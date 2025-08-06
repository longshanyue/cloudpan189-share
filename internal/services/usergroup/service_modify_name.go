package usergroup

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"go.uber.org/zap"
)

type modifyNameRequest struct {
	ID   int64  `json:"id" binding:"required,min=1"`
	Name string `json:"name" binding:"required,min=1,max=255"`
}

type modifyNameResponse struct {
	RowsAffected int64 `json:"rowsAffected"`
}

func (s *service) ModifyName() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := new(modifyNameRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  err.Error(),
			})
			return
		}

		// 检查新名称是否已存在（排除当前记录）
		var existCount int64
		if err := s.db.WithContext(ctx).Model(&models.UserGroup{}).
			Where("name = ? AND id != ?", req.Name, req.ID).
			Count(&existCount).Error; err != nil {
			s.logger.Error("check user group name existence failure", zap.Error(err))

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "检查用户组名称失败",
			})
			return
		}

		if existCount > 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "用户组名称已存在",
			})
			return
		}

		// 执行更新操作
		result := s.db.WithContext(ctx).Model(&models.UserGroup{}).
			Where("id = ?", req.ID).
			Update("name", req.Name)

		if result.Error != nil {
			s.logger.Error("user group modify name failure", zap.Error(result.Error))

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "修改用户组名称失败",
			})
			return
		}

		// 检查是否找到了对应的用户组
		if result.RowsAffected == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code": http.StatusNotFound,
				"msg":  "用户组不存在",
			})
			return
		}

		ctx.JSON(http.StatusOK, &modifyNameResponse{
			RowsAffected: result.RowsAffected,
		})
	}
}
