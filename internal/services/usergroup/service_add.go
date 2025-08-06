package usergroup

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"go.uber.org/zap"
)

type addRequest struct {
	Name string `json:"name" binding:"required,min=1,max=255"`
}

type addResponse struct {
	ID int64 `json:"id"`
}

func (s *service) Add() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := new(addRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  err.Error(),
			})
			return
		}

		group := models.UserGroup{
			Name: req.Name,
		}

		if err := s.db.WithContext(ctx).Create(&group).Error; err != nil {
			s.logger.Error("user group create failure", zap.Error(err))

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "用户组创建失败",
			})
			return
		}

		ctx.JSON(http.StatusOK, &addResponse{
			ID: group.ID,
		})
	}
}
