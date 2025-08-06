package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"go.uber.org/zap"
)

type bindGroupRequest struct {
	UserID  int64 `json:"userId" binding:"required,min=1"`
	GroupID int64 `json:"groupId" binding:"min=0"`
}

type bindGroupResponse struct {
	UserID       int64  `json:"userId"`
	GroupID      int64  `json:"groupId"`
	GroupName    string `json:"groupName"`
	RowsAffected int64  `json:"rowsAffected"`
}

func (s *service) BindGroup() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := new(bindGroupRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  err.Error(),
			})
			return
		}

		// 检查用户是否存在
		var user models.User
		if err := s.db.WithContext(ctx).Where("id = ?", req.UserID).First(&user).Error; err != nil {
			s.logger.Error("check user existence failure",
				zap.Error(err),
				zap.Int64("user_id", req.UserID))

			ctx.JSON(http.StatusNotFound, gin.H{
				"code": http.StatusNotFound,
				"msg":  "用户不存在",
			})
			return
		}

		var groupName string

		// 如果 GroupID 不为 0，检查用户组是否存在
		if req.GroupID > 0 {
			var group models.UserGroup
			if err := s.db.WithContext(ctx).Where("id = ?", req.GroupID).First(&group).Error; err != nil {
				s.logger.Error("check user group existence failure",
					zap.Error(err),
					zap.Int64("group_id", req.GroupID))

				ctx.JSON(http.StatusNotFound, gin.H{
					"code": http.StatusNotFound,
					"msg":  "用户组不存在",
				})
				return
			}
			groupName = group.Name
		} else {
			groupName = "默认用户组"
		}

		// 更新用户的 group_id
		result := s.db.WithContext(ctx).Model(&user).Update("group_id", req.GroupID)
		if result.Error != nil {
			s.logger.Error("bind user to group failure",
				zap.Error(result.Error),
				zap.Int64("user_id", req.UserID),
				zap.Int64("group_id", req.GroupID))

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "绑定用户组失败",
			})
			return
		}

		s.logger.Info("bind user to group success",
			zap.Int64("user_id", req.UserID),
			zap.Int64("group_id", req.GroupID),
			zap.String("group_name", groupName),
			zap.Int64("rows_affected", result.RowsAffected))

		ctx.JSON(http.StatusOK, &bindGroupResponse{
			UserID:       req.UserID,
			GroupID:      req.GroupID,
			GroupName:    groupName,
			RowsAffected: result.RowsAffected,
		})
	}
}
