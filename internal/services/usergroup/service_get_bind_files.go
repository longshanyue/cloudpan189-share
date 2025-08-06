package usergroup

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"go.uber.org/zap"
)

type getBindFilesRequest struct {
	GroupID int64 `form:"groupId" binding:"required,min=1"`
}

type bindFileInfo struct {
	*models.Group2File
	Name string `json:"name,omitempty"`
}

type getBindFilesResponse struct {
	GroupID   int64           `json:"groupId"`
	FileCount int             `json:"fileCount"`
	Files     []*bindFileInfo `json:"files"`
}

func (s *service) GetBindFiles() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := new(getBindFilesRequest)
		if err := ctx.ShouldBindQuery(req); err != nil {
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

		// 查询该组绑定的所有文件
		var bindings []*models.Group2File
		if err := s.db.WithContext(ctx).
			Where("group_id = ?", req.GroupID).
			Find(&bindings).Error; err != nil {
			s.logger.Error("query group file bindings failure",
				zap.Error(err),
				zap.Int64("group_id", req.GroupID))

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "查询文件绑定关系失败",
			})
			return
		}

		// 如果没有绑定任何文件
		if len(bindings) == 0 {
			ctx.JSON(http.StatusOK, &getBindFilesResponse{
				GroupID:   req.GroupID,
				FileCount: 0,
				Files:     []*bindFileInfo{},
			})
			return
		}

		// 提取文件ID列表
		fileIDs := make([]int64, 0, len(bindings))
		for _, binding := range bindings {
			fileIDs = append(fileIDs, binding.FileId)
		}

		// 查询文件信息
		var fileList []*models.VirtualFile
		if err := s.db.WithContext(ctx).
			Where("id IN ?", fileIDs).
			Find(&fileList).Error; err != nil {
			s.logger.Error("query file info failure",
				zap.Error(err),
				zap.Int64s("file_ids", fileIDs))

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "查询文件信息失败",
			})
			return
		}

		// 创建文件ID到文件名的映射
		fileMap := make(map[int64]string)
		for _, file := range fileList {
			fileMap[file.ID] = file.Name
		}

		// 构建响应数据
		files := make([]*bindFileInfo, 0, len(bindings))
		for _, binding := range bindings {
			fileInfo := &bindFileInfo{
				Group2File: binding,
			}
			if fileName, exists := fileMap[binding.FileId]; exists {
				fileInfo.Name = fileName
			}
			files = append(files, fileInfo)
		}

		s.logger.Info("get bind files success",
			zap.Int64("group_id", req.GroupID),
			zap.Int("file_count", len(files)))

		ctx.JSON(http.StatusOK, &getBindFilesResponse{
			GroupID:   req.GroupID,
			FileCount: len(files),
			Files:     files,
		})
	}
}
