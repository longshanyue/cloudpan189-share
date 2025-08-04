package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"go.uber.org/zap"
)

type addRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6,max=20"`
	IsSuper  int8   `json:"is_super" binding:"oneof=0 1"`
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

		u := models.User{
			Username: req.Username,
			Password: hash(req.Password),
		}

		if err := s.db.WithContext(ctx).Create(&u).Error; err != nil {
			s.logger.Error("user create failure", zap.Error(err))

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "用户创建失败",
			})

			return
		}

		ctx.JSON(http.StatusOK, &addResponse{
			ID: u.ID,
		})
	}
}
