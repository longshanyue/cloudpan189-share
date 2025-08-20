package universalfs

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"github.com/xxcheng123/cloudpan189-share/internal/types"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
)

func (s *service) BaseMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rawPath := ctx.Param("path")

		paths, err := utils.SplitPath(rawPath)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})

			ctx.Abort()

			return
		}

		var (
			pid       int64
			fid       int64
			gid       = ctx.GetInt64(consts.CtxKeyGroupId)
			fullPaths = make([]string, 0)
		)

		var (
			groupFileSet = mapset.NewSet[int64]()
		)

		if gid != 0 {
			groupFiles := make([]*models.Group2File, 0)

			if err = s.db.WithContext(ctx).Model(new(models.Group2File)).Where("group_id", gid).Find(&groupFiles).Error; err != nil {
				s.logger.Error("获取用户组文件关系失败", zap.Error(err))

				ctx.JSON(http.StatusInternalServerError, types.ErrResponse{
					Code:    http.StatusInternalServerError,
					Message: "获取用户组文件关系失败",
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
							ctx.Set(consts.CtxKeyFilename, p)

							break
						}

						s.logger.Warn("文件未找到", zap.String("path", rawPath), zap.String("filename", p))

						ctx.JSON(http.StatusNotFound, types.ErrResponse{
							Code:    http.StatusNotFound,
							Message: "文件未找到",
						})

						ctx.Abort()

						return
					}

					s.logger.Error("查询文件失败", zap.Error(err), zap.String("path", rawPath), zap.String("filename", p))

					ctx.JSON(http.StatusBadRequest, types.ErrResponse{
						Code:    http.StatusBadRequest,
						Message: err.Error(),
					})

					ctx.Abort()

					return
				}

				if tmpFile.IsTop == 1 && gid != 0 && !groupFileSet.Contains(tmpFile.ID) {
					// 没有权限
					s.logger.Warn("用户无权限访问文件", zap.Int64("gid", gid), zap.Int64("fileId", tmpFile.ID), zap.String("filename", p))

					ctx.JSON(http.StatusForbidden, types.ErrResponse{
						Code:    http.StatusForbidden,
						Message: "无权限访问",
					})

					ctx.Abort()

					return
				}

				fullPaths = append(fullPaths, p)
				pid = tmpFile.ParentId
				fid = tmpFile.ID
			}
		}

		ctx.Set(consts.CtxKeyFileId, fid)
		ctx.Set(consts.CtxKeyParentId, pid)
		ctx.Set(consts.CtxKeyGroupId, gid)
		ctx.Set(consts.CtxKeyFullPaths, fullPaths)
		ctx.Set(consts.CtxKeyGroupFileSet, groupFileSet)
	}
}

func (s *service) DavMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.GetHeader("Depth") == "" {
			ctx.Request.Header.Add("Depth", "1")
		} else if ctx.GetHeader("Depth") == "infinity" {
			s.logger.Warn("不支持infinity深度查询", zap.String("userAgent", ctx.GetHeader("User-Agent")))

			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "不支持infinity深度查询",
			})

			ctx.Abort()

			return
		}

		if ctx.GetHeader("X-Litmus") == "props: 3 (propfind_invalid2)" {
			s.logger.Warn("无效的属性名称", zap.String("userAgent", ctx.GetHeader("User-Agent")))

			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "无效的属性名称",
			})

			ctx.Abort()

			return
		}

		switch ctx.Request.Method {
		case "PROPFIND":
			ctx.Next()
		case "GET", "HEAD", "POST":
			ctx.Next()
		case "OPTIONS":
			allow := "OPTIONS, HEAD, GET, POST, PROPFIND"

			ctx.Header("Allow", allow)
			// http://www.webdav.org/specs/rfc4918.html#dav.compliance.classes
			ctx.Header("DAV", "1, 2")
			// http://msdn.microsoft.com/en-au/library/cc250217.aspx
			ctx.Header("MS-Author-Via", "DAV")

			ctx.Abort()
		default:
			s.logger.Warn("不支持的HTTP方法", zap.String("method", ctx.Request.Method), zap.String("path", ctx.Request.URL.Path))

			ctx.JSON(http.StatusMethodNotAllowed, types.ErrResponse{
				Code:    http.StatusMethodNotAllowed,
				Message: "不支持的HTTP方法",
			})

			ctx.Abort()
		}
	}
}
