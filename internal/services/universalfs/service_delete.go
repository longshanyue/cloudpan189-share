package universalfs

import (
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/configs"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"net/http"
	"os"
	"path"
)

func (s *service) Delete() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rawPath := ctx.Param("path")

		release, status, err := s.confirmLocks(rawPath, "")
		if err != nil {
			ctx.JSON(status, gin.H{
				"code":    status,
				"message": err.Error(),
			})

			return
		}

		defer release()

		var fid = ctx.GetInt64("x_fid")
		if fid <= 0 {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code":    http.StatusNotFound,
				"message": "文件不存在",
			})

			return
		}

		file := new(models.VirtualFile)

		if err = s.db.WithContext(ctx).Where("id = ?", fid).First(file).Error; err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code":    http.StatusNotFound,
				"message": "文件不存在",
			})

			return
		}

		if !models.OsTypeAllowDelete(file.OsType) {
			ctx.JSON(http.StatusMethodNotAllowed, gin.H{
				"code":    http.StatusMethodNotAllowed,
				"message": "不支持删除此文件",
			})

			return
		}

		if file.OsType == models.OsTypeRealFile {
			v, ok := file.Addition["file_path"]
			if !ok {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code":    http.StatusInternalServerError,
					"message": "文件路径不存在",
				})

				return
			}

			filePath := v.(string)

			if err = os.Remove(path.Join(configs.GetConfig().FileDir, filePath)); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code":    http.StatusInternalServerError,
					"message": "删除文件失败",
				})

				return
			}
		}

		if err = s.db.WithContext(ctx).Where("id = ?", fid).Delete(&models.VirtualFile{}).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "删除文件失败",
			})

			return
		}

		ctx.JSON(http.StatusNoContent, gin.H{
			"code":    http.StatusNoContent,
			"message": "删除成功",
		})
	}
}
