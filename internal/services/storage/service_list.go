package storage

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
)

type listRequest struct {
	CurrentPage int    `form:"currentPage" binding:"omitempty"`
	PageSize    int    `form:"pageSize" binding:"omitempty"`
	NoPaginate  bool   `form:"noPaginate" binding:"omitempty"`
	Name        string `form:"name" binding:"omitempty"`
}

type FileItem struct {
	*models.VirtualFile
	LocalPath string `json:"localPath"`
}

type listResponse struct {
	Total       int64       `json:"total"`
	CurrentPage int         `json:"currentPage"`
	PageSize    int         `json:"pageSize"`
	Data        []*FileItem `json:"data"`
}

func (s *service) List() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := new(listRequest)

		if err := ctx.ShouldBindQuery(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  err.Error(),
			})

			return
		}

		query := s.db.WithContext(ctx).Model(&models.VirtualFile{}).Where("is_top", 1).Order("id asc")

		if req.Name != "" {
			query = query.Where("name like ?", "%"+req.Name+"%")
		}

		if !req.NoPaginate {
			if req.CurrentPage <= 0 {
				req.CurrentPage = 1
			}

			if req.PageSize <= 0 {
				req.PageSize = 10
			}

			query = query.Offset((req.CurrentPage - 1) * req.PageSize).Limit(req.PageSize)
		}

		var count int64

		if err := query.Count(&count).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "查询失败",
			})

			return
		}

		var list = make([]*models.VirtualFile, 0)
		if err := query.Find(&list).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "查询失败",
			})

			return
		}

		var fileList = make([]*FileItem, 0)
		for _, v := range list {
			p, _ := s.getFullPath(ctx, v)

			fileList = append(fileList, &FileItem{
				VirtualFile: v,
				LocalPath:   p,
			})
		}

		ctx.JSON(http.StatusOK, &listResponse{
			Total:       count,
			CurrentPage: req.CurrentPage,
			PageSize:    req.PageSize,
			Data:        fileList,
		})
	}
}
