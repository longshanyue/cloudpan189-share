package universalfs

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
)

func (s *service) BaseMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rawPath := ctx.Param("path")

		paths, err := utils.SplitPath(rawPath)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": err.Error(),
			})

			ctx.Abort()

			return
		}

		var (
			pid       int64
			fid       int64
			gid       = ctx.GetInt64("group_id")
			fullPaths = make([]string, 0)
		)

		var (
			groupFileSet = mapset.NewSet[int64]()
		)

		if gid != 0 {
			groupFiles := make([]*models.Group2File, 0)

			if err = s.db.WithContext(ctx).Model(new(models.Group2File)).Where("group_id", gid).Find(&groupFiles).Error; err != nil {
				s.logger.Error("get group files failure", zap.Error(err))

				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code":    http.StatusInternalServerError,
					"message": "获取用户组文件关系失败",
				})

				ctx.Abort()

				return
			}

			for _, groupFile := range groupFiles {
				groupFileSet.Add(groupFile.FileId)
			}
		}

		if len(paths) == 0 {
			fid = 0
			pid = -1
		} else {
			for idx, p := range paths {
				var tmpFile = new(models.VirtualFile)
				if err = s.db.WithContext(ctx).Where("parent_id", fid).Where("name", p).First(tmpFile).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						if ctx.Request.Method == http.MethodPut && len(paths)-1 == idx {
							// 这个情况表示要写入文件，但是文件不存在
							// 取到父级目录的ID
							pid = fid
							fid = -1
							ctx.Set("x_file_name", p)

							break
						}

						ctx.JSON(http.StatusNotFound, gin.H{
							"code":    http.StatusNotFound,
							"message": "file not found",
						})

						ctx.Abort()

						return
					}

					ctx.JSON(http.StatusBadRequest, gin.H{
						"code":    http.StatusBadRequest,
						"message": err.Error(),
					})

					ctx.Abort()

					return
				}

				if tmpFile.IsTop == 1 && gid != 0 && !groupFileSet.Contains(tmpFile.ID) {
					// 没有权限
					ctx.JSON(http.StatusForbidden, gin.H{
						"code":    http.StatusForbidden,
						"message": "no permission",
					})

					ctx.Abort()

					return
				}

				fullPaths = append(fullPaths, p)
				pid = tmpFile.ParentId
				fid = tmpFile.ID
			}
		}

		ctx.Set("x_fid", fid)
		ctx.Set("x_pid", pid)
		ctx.Set("x_gid", gid)
		ctx.Set("x_full_paths", fullPaths)
		ctx.Set("x_group_file_set", groupFileSet)
	}
}

func (s *service) DavMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.GetHeader("Depth") == "" {
			ctx.Request.Header.Add("Depth", "1")
		} else if ctx.GetHeader("Depth") == "infinity" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "`infinity` is not allowed",
			})

			ctx.Abort()

			return
		}

		if ctx.GetHeader("X-Litmus") == "props: 3 (propfind_invalid2)" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "Invalid property name",
			})

			ctx.Abort()

			return
		}

		switch ctx.Request.Method {
		case "PROPFIND":
			ctx.Next()
		case "GET", "HEAD", "POST":
			ctx.Next()
		case "PUT", "DELETE":
			if !shared.FileWritable {
				ctx.JSON(http.StatusMethodNotAllowed, gin.H{
					"code":    http.StatusMethodNotAllowed,
					"message": "Method not allowed",
				})

				ctx.Abort()
			}

			ctx.Next()
		case "OPTIONS":
			allow := "OPTIONS, HEAD, GET, POST, PROPFIND"
			if shared.FileWritable {
				allow += ", PUT, DELETE"
			}
			
			ctx.Header("Allow", allow)
			// http://www.webdav.org/specs/rfc4918.html#dav.compliance.classes
			ctx.Header("DAV", "1, 2")
			// http://msdn.microsoft.com/en-au/library/cc250217.aspx
			ctx.Header("MS-Author-Via", "DAV")

			ctx.Abort()
		default:
			ctx.JSON(http.StatusMethodNotAllowed, gin.H{
				"code":    http.StatusMethodNotAllowed,
				"message": "Method not allowed",
			})

			ctx.Abort()
		}
	}
}
