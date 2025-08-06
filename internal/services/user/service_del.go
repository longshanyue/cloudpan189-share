package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
)

type delRequest struct {
	ID int64 `json:"id" binding:"required,min=1"`
}

type delResponse struct {
	RowsAffected int64 `json:"rowsAffected"`
}

func (s *service) Del() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req = new(delRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  err.Error(),
			})

			return
		}

		if req.ID == 1 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "创始人不能删除",
			})

			return
		}

		result := s.db.WithContext(ctx).Where("id = ?", req.ID).Delete(&models.User{})
		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "用户删除失败",
			})

			return
		}

		ctx.JSON(http.StatusOK, &delResponse{
			RowsAffected: result.RowsAffected,
		})
	}
}
