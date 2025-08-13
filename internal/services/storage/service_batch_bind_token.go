package storage

import (
	"net/http"

	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/models"

	"github.com/gin-gonic/gin"
)

type batchBindTokenRequest struct {
	IDs        []int64 `json:"ids" binding:"required"`
	CloudToken int64   `json:"cloudToken" binding:"required"`
}

type batchBindTokenResponse struct {
	SuccessCount int     `json:"successCount"`
	FailedCount  int     `json:"failedCount"`
	FailedFiles  []int64 `json:"failedFiles"`
}

func (s *service) BatchBindToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req batchBindTokenRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  err.Error(),
			})
			return
		}

		if len(req.IDs) == 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "文件ID列表不能为空",
			})
			return
		}

		var files []models.VirtualFile
		if err := s.db.Where("id IN ? AND is_top = ?", req.IDs, 1).Find(&files).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "查询文件失败",
			})
			return
		}

		// 检查是否所有文件都存在且都是挂载点
		foundIDs := make(map[int64]bool)
		for _, file := range files {
			foundIDs[file.ID] = true
		}

		var failedFiles []int64
		for _, id := range req.IDs {
			if !foundIDs[id] {
				failedFiles = append(failedFiles, id)
			}
		}

		// 批量更新找到的文件
		successCount := 0
		for _, file := range files {
			if file.Addition == nil {
				file.Addition = make(map[string]interface{})
			}

			file.Addition[consts.FileAdditionKeyCloudToken] = req.CloudToken

			result := s.db.Model(&models.VirtualFile{}).Where("id = ?", file.ID).Update("addition", file.Addition)
			if result.Error == nil && result.RowsAffected > 0 {
				successCount++
			} else {
				failedFiles = append(failedFiles, file.ID)
			}
		}

		ctx.JSON(http.StatusOK, batchBindTokenResponse{
			SuccessCount: successCount,
			FailedCount:  len(failedFiles),
			FailedFiles:  failedFiles,
		})
	}
}
