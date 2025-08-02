package user

import (
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"net/http"
)

type updateRequest struct {
	ID          int64   `json:"id" binding:"required,min=1"`
	Password    *string `json:"password" binding:"omitempty,min=6,max=20"`
	Permissions *uint8  `json:"permissions" binding:"omitempty,min=1"`
}

type updateResponse struct {
	RowsAffected int64 `json:"rowsAffected"`
}

func (s *service) Update() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := updateRequest{}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  err.Error(),
			})

			return
		}

		var mp = make(map[string]any)
		if req.Password != nil {
			mp["password"] = hash(*req.Password)
		}

		if req.Permissions != nil {
			mp["permissions"] = *req.Permissions
		}

		if len(mp) == 0 {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "请填写需要更新的字段",
			})

			return
		}

		result := s.db.WithContext(ctx).Model(new(models.User)).Where("id = ?", req.ID).Updates(mp)
		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "更新失败",
			})

			return
		}

		ctx.JSON(http.StatusOK, &updateResponse{
			RowsAffected: result.RowsAffected,
		})
	}
}
