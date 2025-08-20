package universalfs

import (
	"fmt"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/enc"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"go.uber.org/zap"

	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"github.com/xxcheng123/cloudpan189-share/internal/types"
)

type openRequest struct {
}

func (s *service) Open(prefix string, format string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req = new(openRequest)
		if err := ctx.ShouldBindQuery(req); err != nil {
			s.logger.Warn("文件浏览请求参数错误", zap.Error(err))

			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})

			return
		}

		var (
			fid = ctx.GetInt64(consts.CtxKeyFileId)
			pid = ctx.GetInt64(consts.CtxKeyParentId)
			gid = ctx.GetInt64(consts.CtxKeyGroupId)
		)

		file, err := s.getFileInfo(ctx, fid, pid)
		if err != nil {
			s.logger.Error("获取文件信息失败",
				zap.Int64("fileId", fid),
				zap.Int64("parentId", pid),
				zap.Error(err))
			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})

			return
		}

		var (
			vPath, _         = ctx.Get(consts.CtxKeyFullPaths)
			fullPaths        = utils.StringSlice(vPath)
			vGroupFileSet, _ = ctx.Get(consts.CtxKeyGroupFileSet)
		)

		groupFileSet, ok := vGroupFileSet.(mapset.Set[int64])
		if !ok {
			s.logger.Error("获取用户组文件集合失败",
				zap.Int64("fileId", fid),
				zap.Int64("parentId", pid))
			ctx.JSON(http.StatusInternalServerError, types.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "获取用户组文件集合失败",
			})
			return
		}

		f := &FileInfo{
			VirtualFile: file,
			Path:        path.Join(prefix, strings.Join(fullPaths, "/")),
			Href:        utils.PathEscape(prefix, strings.Join(fullPaths, "/")),
		}

		if file.IsFolder == 1 {
			if err := s.loadFolderChildren(ctx, f, gid, groupFileSet, format); err != nil {
				s.logger.Error("加载文件夹子项失败",
					zap.Int64("fileId", file.ID),
					zap.String("fileName", file.Name),
					zap.Error(err))
				ctx.JSON(http.StatusBadRequest, types.ErrResponse{
					Code:    http.StatusBadRequest,
					Message: err.Error(),
				})

				return
			}
		} else {
			f.DownloadURL = s.generateDownloadURL(file.ID)
		}

		s.responseByFormat(ctx, f, format)
	}
}

func (s *service) getFileInfo(ctx *gin.Context, fid, pid int64) (*models.VirtualFile, error) {
	if pid == -1 && fid == 0 {
		// 根目录特殊处理
		return &models.VirtualFile{
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
		}, nil
	}

	file := new(models.VirtualFile)
	if err := s.db.WithContext(ctx).Where("id", fid).First(&file).Error; err != nil {
		return nil, err
	}

	return file, nil
}

func (s *service) loadFolderChildren(ctx *gin.Context, f *FileInfo, gid int64, groupFileSet mapset.Set[int64], format string) error {
	var list = make([]*models.VirtualFile, 0)
	if err := s.db.WithContext(ctx).Where("parent_id", f.ID).Find(&list).Error; err != nil {
		return err
	}

	// 构建子项列表
	for _, v := range list {
		f.Children = append(f.Children, &FileInfo{
			VirtualFile: v,
			Path:        path.Join(f.Path, v.Name),
			Href:        utils.PathEscape(f.Path, v.Name),
		})
	}

	// 应用权限过滤
	if gid != 0 {
		f.Children = lo.Filter(f.Children, func(item *FileInfo, _ int) bool {
			if item.IsTop == 1 && !groupFileSet.Contains(item.ID) {
				return false
			}

			return true
		})
	}

	// 根据格式过滤STRM文件
	s.filterByFormat(f, format)

	// 排序：文件夹优先，然后按修改时间倒序
	sort.Slice(f.Children, func(i, j int) bool {
		if f.Children[i].IsFolder != f.Children[j].IsFolder {
			return f.Children[i].IsFolder > f.Children[j].IsFolder
		}

		return f.Children[i].Rev > f.Children[j].Rev
	})

	return nil
}

func (s *service) filterByFormat(f *FileInfo, format string) {
	switch format {
	case "dav":
		// DAV格式：过滤掉STRM文件
		f.Children = lo.Filter(f.Children, func(item *FileInfo, _ int) bool {
			return item.OsType != models.OsTypeStrmFile
		})
	case "strm_dav":
		// STRM DAV格式：过滤掉被STRM文件链接的原文件
		var linkIds = make([]int64, 0)
		for _, item := range f.Children {
			if item.OsType == models.OsTypeStrmFile {
				linkIds = append(linkIds, item.LinkId)
			}
		}

		f.Children = lo.Filter(f.Children, func(item *FileInfo, _ int) bool {
			return lo.IndexOf(linkIds, item.ID) == -1
		})
	case "json":
		// JSON格式：不包含自动生成的STRM文件
		f.Children = lo.Filter(f.Children, func(item *FileInfo, _ int) bool {
			return item.OsType != models.OsTypeStrmFile
		})
	}
}

func (s *service) responseByFormat(ctx *gin.Context, f *FileInfo, format string) {
	switch format {
	case "dav", "strm_dav":
		s.responseDav(ctx, f)
	default:
		ctx.JSON(http.StatusOK, f)
	}
}

func (s *service) generateDownloadURL(fid int64) string {
	values := enc.Enc(url.Values{
		"id":     []string{fmt.Sprintf("%d", fid)},
		"random": []string{uuid.NewString()},
	}, shared.Setting.SaltKey)

	baseURL := shared.Setting.BaseURL

	return fmt.Sprintf("%s/api/file_download?%s", baseURL, values.Encode())
}

func (s *service) generateDownloadURLWithNeverExpire(fid int64) string {
	values := enc.Enc(url.Values{
		"id":        []string{fmt.Sprintf("%d", fid)},
		"random":    []string{uuid.NewString()},
		"timestamp": []string{"-1"},
	}, shared.Setting.SaltKey)

	baseURL := shared.Setting.BaseURL

	return fmt.Sprintf("%s/api/file_download?%s", baseURL, values.Encode())
}
