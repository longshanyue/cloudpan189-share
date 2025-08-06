package usergroup

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"go.uber.org/zap"
)

type deleteRequest struct {
	ID int64 `json:"id" binding:"required,min=1"`
}

type deleteResponse struct {
	RowsAffected int64 `json:"rowsAffected"`
}

func (s *service) Delete() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := new(deleteRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  err.Error(),
			})
			return
		}

		// 检查是否有用户绑定到这个组
		var userCount int64
		if err := s.db.WithContext(ctx).Model(&models.User{}).
			Where("group_id = ?", req.ID).
			Count(&userCount).Error; err != nil {
			s.logger.Error("check user group binding failure", zap.Error(err))

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "检查用户组绑定失败",
			})
			return
		}

		// 如果有用户绑定到这个组，不允许删除
		if userCount > 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "该用户组下还有用户，请先解绑用户后再删除",
			})
			return
		}

		// 先删除该组的所有文件绑定关系
		if err := s.db.WithContext(ctx).Where("group_id = ?", req.ID).Delete(&models.Group2File{}).Error; err != nil {
			s.logger.Error("delete group file bindings failure", zap.Error(err))

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "删除用户组文件绑定失败",
			})
			return
		}

		// 再删除用户组
		result := s.db.WithContext(ctx).Where("id = ?", req.ID).Delete(&models.UserGroup{})
		if result.Error != nil {
			s.logger.Error("user group delete failure", zap.Error(result.Error))

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "用户组删除失败",
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

		ctx.JSON(http.StatusOK, &deleteResponse{
			RowsAffected: result.RowsAffected,
		})
	}
}
