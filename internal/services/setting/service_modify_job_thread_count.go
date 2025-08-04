package setting

import (
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"net/http"
)

type modifyJobThreadCountRequest struct {
	ThreadCount int `json:"threadCount" binding:"required,min=1,max=8"`
}

func (s *service) ModifyJobThreadCount() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req modifyJobThreadCountRequest

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "参数错误，线程数必须在1-8之间",
			})

			return
		}

		record, err := s.get(ctx)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "查询配置失败",
			})

			return
		}

		if err = s.db.WithContext(ctx).Model(&models.Setting{}).Where("id = ?", record.ID).Update("job_thread_count", req.ThreadCount).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "修改失败",
			})

			return
		}

		shared.Setting.JobThreadCount = req.ThreadCount

		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusOK,
			"msg":  "修改成功",
		})
	}
}
