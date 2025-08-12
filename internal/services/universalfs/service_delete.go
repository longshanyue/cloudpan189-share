package universalfs

import (
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/types"
	"go.uber.org/zap"
	"net/http"
)

func (s *service) Delete() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rawPath := ctx.Param("path")

		release, status, err := s.confirmLocks(rawPath, "")
		if err != nil {
			s.logger.Error("确认文件锁失败", zap.String("path", rawPath), zap.Error(err))

			ctx.JSON(status, types.ErrResponse{
				Code:    status,
				Message: err.Error(),
			})

			return
		}

		defer release()

		var fid = ctx.GetInt64(consts.CtxKeyFileId)
		if fid <= 0 {
			s.logger.Warn("尝试删除不存在的文件", zap.String("path", rawPath), zap.Int64("fileId", fid))

			ctx.JSON(http.StatusNotFound, types.ErrResponse{
				Code:    http.StatusNotFound,
				Message: "文件不存在",
			})

			return
		}

		file := new(models.VirtualFile)

		if err = s.db.WithContext(ctx).Where("id = ?", fid).First(file).Error; err != nil {
			s.logger.Error("查询文件信息失败", zap.String("path", rawPath), zap.Int64("fileId", fid), zap.Error(err))

			ctx.JSON(http.StatusNotFound, types.ErrResponse{
				Code:    http.StatusNotFound,
				Message: "文件不存在",
			})

			return
		}

		if file.IsTop == 1 {
			s.logger.Warn("尝试删除挂载点文件", zap.String("path", rawPath), zap.Int64("fileId", fid), zap.String("fileName", file.Name))

			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "挂载点文件请在后台存储管理删除",
			})

			return
		}

		if err = s.fs.Delete(ctx, fid); err != nil {
			s.logger.Error("删除文件失败", zap.String("path", rawPath), zap.Int64("fileId", fid), zap.String("fileName", file.Name), zap.Error(err))

			ctx.JSON(http.StatusInternalServerError, types.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "删除失败",
			})

			return
		}

		s.logger.Info("文件删除成功", zap.String("path", rawPath), zap.Int64("fileId", fid), zap.String("fileName", file.Name))

		ctx.JSON(http.StatusNoContent, types.ErrResponse{
			Code:    http.StatusNoContent,
			Message: "删除成功",
		})
	}
}
