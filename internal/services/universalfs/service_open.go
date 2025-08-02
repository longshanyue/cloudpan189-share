package universalfs

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"gorm.io/gorm"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"
)

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

		var pid int64
		var fullPaths = make([]string, 0)
		var session = new(ReadSession)
		var file = new(models.VirtualFile)

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

			sort.Slice(f.Children, func(i, j int) bool {
				if f.Children[i].IsFolder != f.Children[j].IsFolder {
					return f.Children[i].IsFolder > f.Children[j].IsFolder
				}

				return f.Children[i].Rev > f.Children[j].Rev
			})

		} else {
			values := enc(url.Values{
				"id":     []string{fmt.Sprintf("%d", file.ID)},
				"random": []string{uuid.NewString()},
			}, shared.Setting.SaltKey)

			baseURL := shared.Setting.BaseURL
			if baseURL == "" {
				scheme := "http"
				if ctx.Request.TLS != nil {
					scheme = "https"
				}
				baseURL = fmt.Sprintf("%s://%s", scheme, ctx.Request.Host)
			}

			f.DownloadURL = fmt.Sprintf("%s/api/file_download?%s", baseURL, values.Encode())
		}

		switch format {
		case "dav":
			s.responseDav(ctx, f)
		default:
			ctx.JSON(http.StatusOK, f)
		}
	}
}
