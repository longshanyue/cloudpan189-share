package setting

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"net/http"
)

type modifyMultipleStreamThreadCountRequest struct {
	MultipleStreamThreadCount int `json:"multipleStreamThreadCount" binding:"required,min=1,max=64"`
}

type modifyMultipleStreamThreadCountResponse struct {
	RowsAffected int64 `json:"rowsAffected"`
}

func (s *service) ModifyMultipleStreamThreadCount() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req = new(modifyMultipleStreamThreadCountRequest)

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "参数错误，线程数必须在1-64之间",
			})
			return
		}

		result := new(models.SettingDict).SetMultipleStreamThreadCount(s.db.WithContext(ctx), req.MultipleStreamThreadCount)
		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  fmt.Sprintf("修改失败：%s", result.Error.Error()),
			})
			return
		}

		shared.MultipleStreamThreadCount = req.MultipleStreamThreadCount

		ctx.JSON(http.StatusOK, modifyMultipleStreamThreadCountResponse{
			RowsAffected: result.RowsAffected,
		})
	}
}
