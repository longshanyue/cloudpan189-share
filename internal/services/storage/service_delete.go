package storage

import (
	"errors"
	"net/http"

	"github.com/xxcheng123/cloudpan189-share/internal/bus"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"gorm.io/gorm"
)

type deleteRequest struct {
	ID int64 `json:"id" binding:"required"`
}

func (s *service) Delete() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req = new(deleteRequest)

		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": err.Error(),
			})
		}

		var file = models.VirtualFile{}

		if err := s.db.WithContext(ctx).Where("id", req.ID).First(&file).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{
					"code": http.StatusNotFound,
					"msg":  "挂载点不存在",
				})

				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "挂载点信息查询失败",
			})

			return
		}

		if file.IsTop != 1 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "非挂载点禁止删除",
			})

			return
		}

		var scanFile = file
		for scanFile.ParentId != 0 {
			var tmpParent = models.VirtualFile{}
			if err := s.db.WithContext(ctx).Where("id", scanFile.ParentId).First(&tmpParent).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					ctx.JSON(http.StatusNotFound, gin.H{
						"code": http.StatusNotFound,
						"msg":  "挂载点不存在",
					})

					return
				}

				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code": http.StatusInternalServerError,
					"msg":  "挂载点信息查询失败",
				})

				return
			}

			if tmpParent.IsTop == 1 {
				break
			}

			// 检查下属文件
			var count int64
			s.db.WithContext(ctx).Model(&models.VirtualFile{}).Where("parent_id", tmpParent.ID).Count(&count)
			if count > 1 {
				break
			}

			scanFile = tmpParent
		}

		file = scanFile

		if err := bus.PublishVirtualFileDelete(ctx, file.ID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  err.Error(),
			})

			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"message": "已添加到删除任务队列，请稍后查看删除结果",
		})
	}
}
