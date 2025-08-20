package setting

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"net/http"
)

type modifyStrmBaseURLRequest struct {
	StrmBaseURL string `json:"strmBaseURL"`
}

type modifyStrmBaseURLResponse struct {
	RowsAffected int64 `json:"rowsAffected"`
}

func (s *service) ModifyStrmBaseURL() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req = new(modifyStrmBaseURLRequest)

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "参数错误",
			})

			return
		}

		result := new(models.SettingDict).SetStrmBaseURL(s.db.WithContext(ctx), req.StrmBaseURL)
		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  fmt.Sprintf("修改失败：%s", result.Error.Error()),
			})

			return
		}

		shared.StrmBaseURL = req.StrmBaseURL

		ctx.JSON(http.StatusOK, modifyStrmBaseURLResponse{
			RowsAffected: result.RowsAffected,
		})
	}
}
