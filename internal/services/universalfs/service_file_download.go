package universalfs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/xxcheng123/cloudpan189-interface/client"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"github.com/xxcheng123/multistreamer"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type fileDownloadRequest struct {
	ID        int64  `form:"id" binding:"required"`
	TimeStamp int64  `form:"timestamp" binding:"required"`
	Random    string `form:"random" binding:"required"`
	Sign      string `form:"sign" binding:"required"`
}

func (s *service) FileDownload() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req = new(fileDownloadRequest)

		if err := ctx.ShouldBindQuery(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": err.Error(),
			})

			return
		}

		if req.TimeStamp < time.Now().Unix() {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "timeStamp is expired",
			})

			return
		}

		key := shared.Setting.SaltKey

		values := url.Values{
			"id":        []string{strconv.FormatInt(req.ID, 10)},
			"timestamp": []string{strconv.FormatInt(req.TimeStamp, 10)},
			"random":    []string{req.Random},
			"sign":      []string{req.Sign},
		}

		if !verify(values, key) {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "sign is invalid",
			})

			return
		}

		if v, ok := s.cache.Get(fmt.Sprintf("file::url::%d", req.ID)); ok {
			ctx.Header("X-Download-Url-Cache", "true")

			s.doResponse(ctx, v.(string))

			return
		}

		_result, err, _ := s.g.Do(fmt.Sprintf("file::url::%d", req.ID), func() (interface{}, error) {
			u, httpCode, err := s.getFileDownloadURL(ctx, req.ID)
			return &DoResult{
				URL:      u,
				HttpCode: httpCode,
				Err:      err,
			}, nil
		})

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": err.Error(),
			})

			return
		}

		result := _result.(*DoResult)

		if result.Err != nil {
			ctx.JSON(result.HttpCode, gin.H{
				"code":    result.HttpCode,
				"message": result.Err.Error(),
			})

			return
		}

		s.doResponse(ctx, result.URL)
	}
}

func (s *service) doResponse(ctx *gin.Context, url string) {
	if shared.Setting.MultipleStream {
		ctx.Header("X-Transfer-Type", "multi_stream")

		opts := []multistreamer.OptionFunc{
			multistreamer.WithHeader(ctx.Request.Header),
			multistreamer.WithContext(ctx),
		}

		httpCode := http.StatusOK

		rangeSpec, yes, err := multistreamer.ParseRangeHeader(ctx.Request.Header.Get("Range"))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": fmt.Sprintf("解析Range失败: %v", err),
			})

			return
		} else if yes {
			opts = append(opts, multistreamer.WithRange(rangeSpec.Start, rangeSpec.End))
			httpCode = http.StatusPartialContent
		}

		mt := multistreamer.New(url, ctx.Writer, opts...)

		if err = mt.Init(); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": fmt.Sprintf("启动失败: %v", err),
			})

			return
		}

		if header, err := mt.GetResponseHeader(); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": fmt.Sprintf("获取响应头失败: %v", err),
			})

			return
		} else {
			for k, v := range header {
				ctx.Header(k, v[0])
			}
		}

		// 设置 http code
		ctx.Status(httpCode)

		_ = mt.Execute()
	} else if shared.Setting.LocalProxy {
		ctx.Header("X-Transfer-Type", "local_proxy")
		// 创建请求
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": err.Error(),
			})
			return
		}

		for k, v := range ctx.Request.Header {
			req.Header.Set(k, v[0])
		}

		// 设置超时
		httpClient := &http.Client{
			Timeout: 30 * time.Second,
		}

		// 发送请求
		resp, err := httpClient.Do(req)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": fmt.Sprintf("代理请求失败: %v", err),
			})
			return
		}
		defer resp.Body.Close()

		for k, v := range resp.Header {
			ctx.Header(k, v[0])
		}

		ctx.Status(resp.StatusCode)

		_, err = io.Copy(ctx.Writer, resp.Body)
		if err != nil {
			s.logger.Error("文件下载转发失败: %v", zap.Error(err))
		}
	} else {
		ctx.Header("X-Transfer-Type", "redirect")
		ctx.Redirect(http.StatusFound, url)
	}

}

type DoResult struct {
	URL      string
	HttpCode int
	Err      error
}

func (s *service) getFileDownloadURL(ctx context.Context, id int64) (string, int, error) {
	var file = new(models.VirtualFile)
	if err := s.db.WithContext(ctx).Where("id", id).First(file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", http.StatusNotFound, errors.New("file not found")
		}

		return "", http.StatusInternalServerError, err
	}

	var cloudTokenId int64

	var scanFile = file

	for {
		if v, ok := scanFile.Addition["cloud_token"]; ok {
			cloudTokenId, _ = utils.Int64(v)

			break
		}

		if scanFile.ParentId == 0 {
			return "", http.StatusBadRequest, errors.New("当前资源没有绑定用于获取播放链接的令牌")
		}

		var parent = new(models.VirtualFile)
		if err := s.db.WithContext(ctx).Where("id", scanFile.ParentId).First(parent).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", http.StatusNotFound, errors.New("file not found")
			}

			return "", http.StatusBadRequest, errors.New("当前资源没有绑定用于获取播放链接的令牌")
		}

		scanFile = parent
	}

	var ct = new(models.CloudToken)
	if err := s.db.WithContext(ctx).Where("id", cloudTokenId).First(ct).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", http.StatusBadRequest, errors.New("绑定的令牌查找失败，可能被删除或隐藏")
		}

		return "", http.StatusInternalServerError, err
	}

	fileId := utils.String(file.Addition["file_id"])

	result, err := client.New().WithToken(client.NewAuthToken(ct.AccessToken, ct.ExpiresIn)).GetFileDownload(ctx, client.String(fileId), func(req *client.GetFileDownloadRequest) {
		if v, ok := file.Addition["share_id"]; ok {
			req.ShareId, _ = utils.Int64(v)
		}
	})

	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	resp, err := http.Get(result.FileDownloadUrl)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	s.cache.Set(fmt.Sprintf("file::url::%d", file.ID), resp.Request.URL.String(), time.Minute)

	return resp.Request.URL.String(), http.StatusFound, nil
}
