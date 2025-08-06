package usergroup

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"go.uber.org/zap"
)

type batchBindFilesRequest struct {
	GroupID int64   `json:"groupId" binding:"required,min=1"`
	FileIDs []int64 `json:"fileIds" binding:"required"`
}

type batchBindFilesResponse struct {
	GroupID      int64 `json:"groupId"`
	BindCount    int   `json:"bindCount"`
	DeletedCount int64 `json:"deletedCount"`
}

func (s *service) BatchBindFiles() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := new(batchBindFilesRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  err.Error(),
			})
			return
		}

		// 检查用户组是否存在
		var groupExists int64
		if err := s.db.WithContext(ctx).Model(&models.UserGroup{}).
			Where("id = ?", req.GroupID).
			Count(&groupExists).Error; err != nil {
			s.logger.Error("check user group existence failure", zap.Error(err))

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "检查用户组失败",
			})
			return
		}

		if groupExists == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code": http.StatusNotFound,
				"msg":  "用户组不存在",
			})
			return
		}

		// 先删除该组的所有旧文件绑定关系
		deleteResult := s.db.WithContext(ctx).Where("group_id = ?", req.GroupID).Delete(&models.Group2File{})
		if deleteResult.Error != nil {
			s.logger.Error("delete old group file bindings failure", zap.Error(deleteResult.Error))

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "删除旧文件绑定关系失败",
			})
			return
		}

		deletedCount := deleteResult.RowsAffected

		// 如果文件ID列表为空，只删除不创建新绑定
		if len(req.FileIDs) == 0 {
			ctx.JSON(http.StatusOK, &batchBindFilesResponse{
				GroupID:      req.GroupID,
				BindCount:    0,
				DeletedCount: deletedCount,
			})
			return
		}

		// 去重文件ID列表
		uniqueFileIDs := make([]int64, 0, len(req.FileIDs))
		fileIDMap := make(map[int64]bool)
		for _, fileID := range req.FileIDs {
			if fileID > 0 && !fileIDMap[fileID] {
				uniqueFileIDs = append(uniqueFileIDs, fileID)
				fileIDMap[fileID] = true
			}
		}

		// 如果去重后为空，只删除不创建新绑定
		if len(uniqueFileIDs) == 0 {
			ctx.JSON(http.StatusOK, &batchBindFilesResponse{
				GroupID:      req.GroupID,
				BindCount:    0,
				DeletedCount: deletedCount,
			})
			return
		}

		// 批量创建新的文件绑定关系
		bindings := make([]*models.Group2File, 0, len(uniqueFileIDs))
		for _, fileID := range uniqueFileIDs {
			bindings = append(bindings, &models.Group2File{
				FileId:  fileID,
				GroupId: req.GroupID,
			})
		}

		// 批量插入
		if err := s.db.WithContext(ctx).CreateInBatches(bindings, 100).Error; err != nil {
			s.logger.Error("batch create group file bindings failure",
				zap.Error(err),
				zap.Int64("group_id", req.GroupID),
				zap.Int("file_count", len(uniqueFileIDs)))

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "批量创建文件绑定关系失败",
			})
			return
		}

		s.logger.Info("batch bind files success",
			zap.Int64("group_id", req.GroupID),
			zap.Int("bind_count", len(uniqueFileIDs)),
			zap.Int64("deleted_count", deletedCount))

		ctx.JSON(http.StatusOK, &batchBindFilesResponse{
			GroupID:      req.GroupID,
			BindCount:    len(uniqueFileIDs),
			DeletedCount: deletedCount,
		})
	}
}
