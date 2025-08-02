package storage

import (
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"net/http"
)

type searchRequest struct {
	Keyword     string `form:"keyword" binding:"required"`
	PID         int64  `form:"pid"`    // 父级ID
	Global      bool   `form:"global"` // 全局搜索（如果为true，忽略pid）
	PageSize    int    `form:"pageSize" binding:"required,min=1,max=100"`
	CurrentPage int    `form:"currentPage" binding:"required,min=1"`
}

type searchResponse struct {
	Total       int64       `json:"total"`
	CurrentPage int         `json:"currentPage"`
	PageSize    int         `json:"pageSize"`
	Data        []*FileItem `json:"data"`
}

func (s *service) Search() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req searchRequest

		if err := ctx.ShouldBindQuery(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "参数错误",
			})

			return
		}

		query := s.db.WithContext(ctx).Model(&models.VirtualFile{})
		if !req.Global {
			query = query.Where("parent_id = ?", req.PID)
		}

		if req.Keyword != "" {
			query = query.Where("name LIKE ?", "%"+req.Keyword+"%")
		}

		var count int64

		if err := query.Count(&count).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "查询失败",
			})

			return
		}

		if req.CurrentPage <= 0 {
			req.CurrentPage = 1
		}

		if req.PageSize <= 0 {
			req.PageSize = 10
		}

		query = query.Offset((req.CurrentPage - 1) * req.PageSize).Limit(req.PageSize)

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

		ctx.JSON(http.StatusOK, &searchResponse{
			Total:       count,
			CurrentPage: req.CurrentPage,
			PageSize:    req.PageSize,
			Data:        fileList,
		})
	}
}
