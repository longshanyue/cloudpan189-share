package setting

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"net/http"
)

type toggleLinkFileAutoDeleteRequest struct {
	LinkFileAutoDelete bool `json:"linkFileAutoDelete"`
}

type toggleLinkFileAutoDeleteResponse struct {
	RowsAffected int64 `json:"rowsAffected"`
}

func (s *service) ToggleLinkFileAutoDelete() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req = new(toggleLinkFileAutoDeleteRequest)

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "参数错误",
			})

			return
		}

		result := new(models.SettingDict).SetLinkFileAutoDelete(s.db.WithContext(ctx), req.LinkFileAutoDelete)
		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  fmt.Sprintf("修改失败：%s", result.Error.Error()),
			})

			return
		}

		shared.LinkFileAutoDelete = req.LinkFileAutoDelete

		ctx.JSON(http.StatusOK, toggleLinkFileAutoDeleteResponse{
			RowsAffected: result.RowsAffected,
		})
	}
}
