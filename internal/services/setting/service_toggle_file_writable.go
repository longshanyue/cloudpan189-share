package setting

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"net/http"
)

type toggleFileWritableRequest struct {
	FileWritable bool `json:"fileWritable"`
}

type toggleFileWritableResponse struct {
	RowsAffected int64 `json:"rowsAffected"`
}

func (s *service) ToggleFileWritable() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req = new(toggleFileWritableRequest)

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "参数错误",
			})

			return
		}

		result := new(models.SettingDict).SetFileWritable(s.db.WithContext(ctx), req.FileWritable)
		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  fmt.Sprintf("修改失败：%s", result.Error.Error()),
			})

			return
		}

		shared.FileWritable = req.FileWritable

		ctx.JSON(http.StatusOK, toggleFileWritableResponse{
			RowsAffected: result.RowsAffected,
		})
	}
}
