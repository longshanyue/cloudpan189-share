package usergroup

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

type listResponse struct {
	Total       int64               `json:"total"`
	CurrentPage int                 `json:"currentPage"`
	PageSize    int                 `json:"pageSize"`
	Data        []*models.UserGroup `json:"data"`
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

		query := s.db.WithContext(ctx).Model(&models.UserGroup{})

		// 根据名称模糊查询
		if req.Name != "" {
			query = query.Where("name LIKE ?", "%"+req.Name+"%")
		}

		// 分页处理
		if !req.NoPaginate {
			if req.CurrentPage <= 0 {
				req.CurrentPage = 1
			}

			if req.PageSize <= 0 {
				req.PageSize = 10
			}

			query = query.Offset((req.CurrentPage - 1) * req.PageSize).Limit(req.PageSize)
		}

		// 获取总数
		var count int64
		if err := query.Count(&count).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "查询失败",
			})
			return
		}

		// 获取数据列表
		var list = make([]*models.UserGroup, 0)
		if err := query.Order("created_at DESC").Find(&list).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "查询失败",
			})
			return
		}

		ctx.JSON(http.StatusOK, &listResponse{
			Total:       count,
			CurrentPage: req.CurrentPage,
			PageSize:    req.PageSize,
			Data:        list,
		})
	}
}
