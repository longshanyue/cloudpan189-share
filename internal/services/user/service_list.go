package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
)

type listRequest struct {
	CurrentPage int    `form:"currentPage" binding:"omitempty"`
	PageSize    int    `form:"pageSize" binding:"omitempty"`
	NoPaginate  bool   `form:"noPaginate" binding:"omitempty"`
	Username    string `form:"username" binding:"omitempty"`
}

// 自定义用户响应结构体
type userResponse struct {
	*models.User
	GroupName string `json:"groupName"`
}

type listResponse struct {
	Total       int64           `json:"total"`
	CurrentPage int             `json:"currentPage"`
	PageSize    int             `json:"pageSize"`
	Data        []*userResponse `json:"data"`
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

		query := s.db.WithContext(ctx).Model(&models.User{})
		if req.Username != "" {
			query = query.Where("username LIKE ?", "%"+req.Username+"%")
		}

		// 先获取总数
		var count int64
		if err := query.Count(&count).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "查询失败",
			})

			return
		}

		// 应用分页
		if !req.NoPaginate {
			if req.CurrentPage <= 0 {
				req.CurrentPage = 1
			}
			if req.PageSize <= 0 {
				req.PageSize = 10
			}
			query = query.Offset((req.CurrentPage - 1) * req.PageSize).Limit(req.PageSize)
		}

		// 查询用户列表
		var list = make([]*models.User, 0)
		if err := query.Find(&list).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "查询失败",
			})
			return
		}

		// 获取所有用户组ID
		groupIDs := make([]int64, 0)
		for _, user := range list {
			if user.GroupID > 0 {
				groupIDs = append(groupIDs, user.GroupID)
			}
		}

		// 查询用户组信息
		groupMap := make(map[int64]string)
		if len(groupIDs) > 0 {
			var groups []*models.UserGroup
			if err := s.db.WithContext(ctx).Where("id IN ?", groupIDs).Find(&groups).Error; err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code": http.StatusInternalServerError,
					"msg":  "查询用户组失败",
				})
				return
			}

			for _, group := range groups {
				groupMap[group.ID] = group.Name
			}
		}

		// 构建响应数据
		responseData := make([]*userResponse, 0, len(list))
		for _, user := range list {
			userResp := &userResponse{
				User: user,
			}

			if user.GroupID == 0 {
				userResp.GroupName = "默认用户组"
			} else {
				if groupName, exists := groupMap[user.GroupID]; exists {
					userResp.GroupName = groupName
				} else {
					userResp.GroupName = "用户组查询失败"
				}
			}

			responseData = append(responseData, userResp)
		}

		ctx.JSON(http.StatusOK, &listResponse{
			Total:       count,
			CurrentPage: req.CurrentPage,
			PageSize:    req.PageSize,
			Data:        responseData,
		})
	}
}
