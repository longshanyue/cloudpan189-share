package setting

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"net/http"
)

type toggleStrmFileEnableRequest struct {
	StrmFileEnable bool `json:"strmFileEnable"`
}

type toggleStrmFileEnableResponse struct {
	RowsAffected int64 `json:"rowsAffected"`
}

func (s *service) ToggleStrmFileEnable() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req = new(toggleStrmFileEnableRequest)

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "参数错误",
			})

			return
		}

		result := new(models.SettingDict).SetStrmFileEnable(s.db.WithContext(ctx), req.StrmFileEnable)
		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  fmt.Sprintf("修改失败：%s", result.Error.Error()),
			})

			return
		}

		//if req.StrmFileEnable {
		//	if err := shared.FileBus.Publish(ctx, bus.TopicFileRebuildMediaFile, bus.TopicFileRebuildMediaFileRequest{}); err != nil {
		//		ctx.JSON(http.StatusInternalServerError, gin.H{
		//			"code": http.StatusInternalServerError,
		//			"msg":  fmt.Sprintf("下发扫描任务失败：%s", err.Error()),
		//		})
		//
		//		return
		//	}
		//} else {
		//	if err := shared.MediaBus.Publish(ctx, bus.TopicMediaClearAllMedia, bus.TopicMediaClearAllMediaRequest{}); err != nil {
		//		ctx.JSON(http.StatusInternalServerError, gin.H{
		//			"code": http.StatusInternalServerError,
		//			"msg":  fmt.Sprintf("下发清除任务失败：%s", err.Error()),
		//		})
		//
		//		return
		//	}
		//}

		shared.StrmFileEnable = req.StrmFileEnable

		ctx.JSON(http.StatusOK, toggleStrmFileEnableResponse{
			RowsAffected: result.RowsAffected,
		})
	}
}
