package universalfs

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/samber/lo"
	"go.uber.org/zap"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"gorm.io/gorm"
)

type openRequest struct {
	IncludeAutoGenerateStrmFile bool `form:"includeAutoGenerateStrmFile"`
}

func (s *service) Open(prefix string, format string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rawPath := ctx.Param("path")

		paths, err := utils.SplitPath(rawPath)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": err.Error(),
			})

			return
		}

		req := new(openRequest)
		if err = ctx.ShouldBindQuery(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": err.Error(),
			})

			return
		}

		var pid int64
		var fullPaths = make([]string, 0)
		var session = new(ReadSession)
		var file = new(models.VirtualFile)

		var (
			groupId      = ctx.GetInt64("group_id")
			groupFileSet = mapset.NewSet[int64]()
		)

		if groupId != 0 {
			groupFiles := make([]*models.Group2File, 0)

			if err = s.db.WithContext(ctx).Model(new(models.Group2File)).Where("group_id", groupId).Find(&groupFiles).Error; err != nil {
				s.logger.Error("get group files failure", zap.Error(err))

				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code":    http.StatusInternalServerError,
					"message": "获取用户组文件关系失败",
				})

				return
			}

			for _, groupFile := range groupFiles {
				groupFileSet.Add(groupFile.FileId)
			}
		}

		if len(paths) == 0 {
			file = &models.VirtualFile{
				ID:         0,
				ParentId:   -1,
				Name:       "root",
				OsType:     models.OsTypeFolder,
				CreateDate: s.startTime.Format(time.DateTime),
				ModifyDate: s.startTime.Format(time.DateTime),
				CreatedAt:  s.startTime,
				UpdatedAt:  s.startTime,
				Rev:        s.startTime.Format("20060102150405"),
				IsFolder:   1,
				Addition:   map[string]interface{}{},
			}
		} else {
			for _, p := range paths {
				var tmpFile = new(models.VirtualFile)
				if err = s.db.WithContext(ctx).Where("parent_id", pid).Where("name", p).First(tmpFile).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						ctx.JSON(http.StatusNotFound, gin.H{
							"code":    http.StatusNotFound,
							"message": "file not found",
						})

						return
					}

					ctx.JSON(http.StatusBadRequest, gin.H{
						"code":    http.StatusBadRequest,
						"message": err.Error(),
					})

					return
				} else {
					if v, ok := tmpFile.Addition["cloud_token"]; ok {
						session.CloudTokenID, _ = v.(json.Number).Int64()
					}
				}

				if tmpFile.IsTop == 1 && groupId != 0 && !groupFileSet.Contains(tmpFile.ID) {
					// 没有权限
					ctx.JSON(http.StatusForbidden, gin.H{
						"code":    http.StatusForbidden,
						"message": "no permission",
					})

					return
				}

				fullPaths = append(fullPaths, p)
				pid = tmpFile.ID
				file = tmpFile
			}
		}

		f := &FileInfo{
			VirtualFile: file,
			Path:        path.Join(prefix, strings.Join(fullPaths, "/")),
			Href:        utils.PathEscape(prefix, strings.Join(fullPaths, "/")),
		}

		if file.IsFolder == 1 {
			var list = make([]*models.VirtualFile, 0)
			if err = s.db.WithContext(ctx).Where("parent_id", file.ID).Find(&list).Error; err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"code":    http.StatusBadRequest,
					"message": err.Error(),
				})

				return
			}

			for _, v := range list {
				f.Children = append(f.Children, &FileInfo{
					VirtualFile: v,
					Path:        path.Join(f.Path, v.Name),
					Href:        utils.PathEscape(f.Path, v.Name),
				})
			}

			if groupId != 0 {
				f.Children = lo.Filter(f.Children, func(item *FileInfo, _ int) bool {
					if item.IsTop == 1 && !groupFileSet.Contains(item.ID) {
						return false
					}

					return true
				})
			}

			// 处理 strm 的过滤问题
			switch format {
			case "dav":
				f.Children = lo.Filter(f.Children, func(item *FileInfo, _ int) bool {
					if item.OsType == models.OsTypeStrmFile {
						return false
					}

					return true
				})
			case "strm_dav":
				var linkIds = make([]int64, 0)
				for _, item := range f.Children {
					if item.OsType == models.OsTypeStrmFile {
						linkIds = append(linkIds, item.LinkId)
					}
				}

				f.Children = lo.Filter(f.Children, func(item *FileInfo, _ int) bool {
					if lo.IndexOf(linkIds, item.ID) > -1 {
						return false
					}

					return true
				})
			case "json":
				f.Children = lo.Filter(f.Children, func(item *FileInfo, _ int) bool {
					if !req.IncludeAutoGenerateStrmFile && item.OsType == models.OsTypeStrmFile {
						return false
					}

					return true
				})
			}

			sort.Slice(f.Children, func(i, j int) bool {
				if f.Children[i].IsFolder != f.Children[j].IsFolder {
					return f.Children[i].IsFolder > f.Children[j].IsFolder
				}

				return f.Children[i].Rev > f.Children[j].Rev
			})
		} else {
			f.DownloadURL = s.generateDownloadURL(file.ID)
		}

		switch format {
		case "dav":
			s.responseDav(ctx, f)
		default:
			ctx.JSON(http.StatusOK, f)
		}
	}
}

func (s *service) generateDownloadURL(fid int64) string {
	values := enc(url.Values{
		"id":     []string{fmt.Sprintf("%d", fid)},
		"random": []string{uuid.NewString()},
	}, shared.Setting.SaltKey)

	baseURL := shared.Setting.BaseURL

	return fmt.Sprintf("%s/api/file_download?%s", baseURL, values.Encode())
}

func (s *service) generateDownloadURLWithNeverExpire(fid int64) string {
	values := enc(url.Values{
		"id":        []string{fmt.Sprintf("%d", fid)},
		"random":    []string{uuid.NewString()},
		"timestamp": []string{"-1"},
	}, shared.Setting.SaltKey)

	baseURL := shared.Setting.BaseURL

	return fmt.Sprintf("%s/api/file_download?%s", baseURL, values.Encode())
}
