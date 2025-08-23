package setting

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
)

type modifyMultipleStreamChunkSizeRequest struct {
	MultipleStreamChunkSize int64 `json:"multipleStreamChunkSize" binding:"required,min=524288,max=33554432"`
}

type modifyMultipleStreamChunkSizeResponse struct {
	RowsAffected int64 `json:"rowsAffected"`
}

func (s *service) ModifyMultipleStreamChunkSize() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req = new(modifyMultipleStreamChunkSizeRequest)

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "参数错误，必须在512KB-32MB之间",
			})

			return
		}

		result := new(models.SettingDict).SetMultipleStreamChunkSize(s.db.WithContext(ctx), req.MultipleStreamChunkSize)
		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  fmt.Sprintf("修改失败：%s", result.Error.Error()),
			})

			return
		}

		shared.MultipleStreamChunkSize = req.MultipleStreamChunkSize

		ctx.JSON(http.StatusOK, modifyMultipleStreamChunkSizeResponse{
			RowsAffected: result.RowsAffected,
		})
	}
}
