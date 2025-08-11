package universalfs

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"
)

type openRequest struct {
	IncludeAutoGenerateStrmFile bool `form:"includeAutoGenerateStrmFile"`
}

func (s *service) Open(prefix string, format string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			req = new(openRequest)
			err error
		)

		if err = ctx.ShouldBindQuery(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": err.Error(),
			})

			return
		}

		var (
			fid  = ctx.GetInt64("x_fid")
			pid  = ctx.GetInt64("x_pid")
			gid  = ctx.GetInt64("group_id")
			file = new(models.VirtualFile)
		)

		if pid == -1 && fid == 0 {
			// 特殊处理
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
			if err = s.db.WithContext(ctx).Where("id", fid).First(&file).Error; err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"code":    http.StatusBadRequest,
					"message": err.Error(),
				})

				return
			}
		}

		var (
			vPath, _         = ctx.Get("x_full_paths")
			fullPaths        = utils.StringSlice(vPath)
			vGroupFileSet, _ = ctx.Get("x_group_file_set")
			groupFileSet     = vGroupFileSet.(mapset.Set[int64])
		)

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

			if gid != 0 {
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
		case "strm_dav":
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
