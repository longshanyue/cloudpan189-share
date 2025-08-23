package storage

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"gorm.io/datatypes"
)

type toggleAutoScanRequest struct {
	ID              int64 `json:"id" binding:"required"`
	DisableAutoScan bool  `json:"disableAutoScan"`
}

func (s *service) ToggleAutoScan() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req = new(toggleAutoScanRequest)

		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "参数错误",
			})
			return
		}

		file := new(models.VirtualFile)
		if err := s.db.WithContext(ctx).Where("id = ?", req.ID).First(file).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "获取文件信息失败",
			})
			return
		}

		// 只允许对顶层文件夹（is_top=1）进行操作
		if file.IsTop != 1 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "只允许对顶层文件夹设置自动扫描选项",
			})
			return
		}

		// 初始化 addition 如果为空
		if file.Addition == nil {
			file.Addition = make(datatypes.JSONMap)
		}

		// 设置 disable_auto_scan 标志
		file.Addition[consts.FileAdditionKeyDisableAutoScan] = req.DisableAutoScan

		// 更新数据库
		if err := s.db.WithContext(ctx).Model(file).Update("addition", file.Addition).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "更新文件信息失败",
			})
			return
		}

		var msg string
		if req.DisableAutoScan {
			msg = "已设置禁用自动扫描队列"
		} else {
			msg = "已取消禁用自动扫描队列设置"
		}

		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusOK,
			"msg":  msg,
		})
	}
}
