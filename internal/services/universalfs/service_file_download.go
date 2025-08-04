package universalfs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

// 全局HTTP客户端，复用连接
var globalHTTPClient = &http.Client{
	Timeout: 0,
	Transport: &http.Transport{
		DisableKeepAlives:     false,
		MaxIdleConns:          200, // 增加连接池
		MaxIdleConnsPerHost:   20,  // 每个host更多连接
		MaxConnsPerHost:       50,  // 总连接数限制
		IdleConnTimeout:       120 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableCompression:    true,       // 禁用压缩
		WriteBufferSize:       128 * 1024, // 128KB写缓冲
		ReadBufferSize:        128 * 1024, // 128KB读缓冲
		ForceAttemptHTTP2:     false,      // 禁用HTTP2
	},
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

		ctx.Status(httpCode)
		_ = mt.Execute()

	} else if shared.Setting.LocalProxy {
		s.handleLocalProxy(ctx, url)
	} else {
		ctx.Header("X-Transfer-Type", "redirect")
		ctx.Redirect(http.StatusFound, url)
	}
}

// 单独处理 LocalProxy 逻辑
func (s *service) handleLocalProxy(ctx *gin.Context, url string) {
	start := time.Now()
	ctx.Header("X-Transfer-Type", "local_proxy")

	// 使用请求的上下文，支持取消
	req, err := http.NewRequestWithContext(ctx.Request.Context(), http.MethodGet, url, nil)
	if err != nil {
		s.logger.Error("创建代理请求失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	// 智能复制请求头
	s.copyOptimizedHeaders(ctx.Request.Header, req.Header)

	// 检查是否是Range请求
	rangeHeader := ctx.Request.Header.Get("Range")
	isRangeRequest := rangeHeader != ""

	if isRangeRequest {
		// Range请求使用更激进的优化
		req.Header.Set("Connection", "keep-alive")
	}

	// 发送请求
	resp, err := globalHTTPClient.Do(req)
	if err != nil {
		// 检查是否是上下文取消（客户端断开）
		if ctx.Request.Context().Err() != nil {
			s.logger.Info("客户端断开连接", zap.String("url", url))
			return
		}

		s.logger.Error("代理请求失败", zap.Error(err), zap.String("url", url))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": fmt.Sprintf("代理请求失败: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	// 优化响应头复制
	s.copyOptimizedResponseHeaders(resp.Header, ctx)

	ctx.Status(resp.StatusCode)

	// 立即刷新响应头
	if flusher, ok := ctx.Writer.(http.Flusher); ok {
		flusher.Flush()
	}

	// 根据请求类型选择不同的传输策略
	var copyErr error
	if isRangeRequest || resp.ContentLength < 1024*1024 { // 小于1MB
		// 小文件或Range请求：快速传输
		copyErr = s.fastCopy(ctx, ctx.Writer, resp.Body)
	} else {
		// 大文件：流式传输
		copyErr = s.streamCopy(ctx, ctx.Writer, resp.Body)
	}

	// 性能监控
	duration := time.Since(start)
	if copyErr != nil {
		if s.isConnectionError(copyErr) {
			s.logger.Info("客户端连接断开",
				zap.String("url", url),
				zap.Duration("duration", duration),
			)
		} else {
			s.logger.Error("文件下载转发失败",
				zap.Error(copyErr),
				zap.String("url", url),
				zap.Duration("duration", duration),
			)
		}
	} else {
		s.logger.Info("代理请求完成",
			zap.String("url", url),
			zap.Duration("duration", duration),
			zap.String("user_agent", ctx.Request.UserAgent()),
		)

		// 如果超过5秒，记录警告
		if duration > 5*time.Second {
			s.logger.Warn("代理请求较慢",
				zap.String("url", url),
				zap.Duration("duration", duration),
			)
		}
	}
}

// 优化的请求头复制
func (s *service) copyOptimizedHeaders(src, dst http.Header) {
	// 只复制必要的头
	importantHeaders := []string{
		"Range", "If-Range", "If-Modified-Since", "If-None-Match",
		"User-Agent", "Accept", "Accept-Encoding", "Authorization",
		"Referer", "Origin",
	}

	for _, header := range importantHeaders {
		if value := src.Get(header); value != "" {
			dst.Set(header, value)
		}
	}
}

// 优化的响应头复制
func (s *service) copyOptimizedResponseHeaders(src http.Header, ctx *gin.Context) {
	// 重要的响应头
	importantHeaders := []string{
		"Content-Type", "Content-Length", "Content-Range",
		"Accept-Ranges", "Last-Modified", "ETag", "Cache-Control",
		"Content-Disposition", "Content-Encoding",
	}

	for _, header := range importantHeaders {
		if value := src.Get(header); value != "" {
			ctx.Header(header, value)
		}
	}

	// 添加性能优化头
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Keep-Alive", "timeout=120, max=100")
}

// 快速复制（小文件）
func (s *service) fastCopy(ctx *gin.Context, dst io.Writer, src io.Reader) error {
	// 使用更大的缓冲区一次性读取
	buf := make([]byte, 256*1024) // 256KB
	_, err := io.CopyBuffer(dst, src, buf)

	if flusher, ok := dst.(http.Flusher); ok {
		flusher.Flush()
	}

	return err
}

// 优化的流式复制
func (s *service) streamCopy(ctx *gin.Context, dst io.Writer, src io.Reader) error {
	// 使用更大的缓冲区 (128KB)
	buf := make([]byte, 128*1024)
	flusher, canFlush := dst.(http.Flusher)

	var written int64
	flushInterval := 0

	for {
		// 检查连接状态
		select {
		case <-ctx.Request.Context().Done():
			return ctx.Request.Context().Err()
		default:
		}

		// 设置读取超时
		if conn, ok := src.(interface{ SetReadDeadline(time.Time) error }); ok {
			conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		}

		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if ew != nil {
				return ew
			}
			if nr != nw {
				return io.ErrShortWrite
			}

			written += int64(nw)
			flushInterval++

			// 每16KB或每10次写入就刷新一次
			if canFlush && (flushInterval%10 == 0 || written%16384 == 0) {
				flusher.Flush()
			}
		}

		if er != nil {
			if er != io.EOF {
				return er
			}
			break
		}
	}

	// 最终刷新
	if canFlush {
		flusher.Flush()
	}

	return nil
}

// 检查是否是连接相关的错误
func (s *service) isConnectionError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	connectionErrors := []string{
		"connection was forcibly closed",
		"wsasend",
		"broken pipe",
		"connection reset by peer",
		"client disconnected",
		"context canceled",
		"context deadline exceeded",
		"use of closed network connection",
		"connection refused",
		"no route to host",
	}

	for _, connErr := range connectionErrors {
		if strings.Contains(errStr, connErr) {
			return true
		}
	}

	return false
}

// ConnectionMonitorMiddleware 连接监控中间件
func (s *service) ConnectionMonitorMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 创建一个可以取消的上下文
		ctxWithCancel, cancel := context.WithCancel(ctx.Request.Context())
		defer cancel()

		// 替换请求上下文
		ctx.Request = ctx.Request.WithContext(ctxWithCancel)

		// 使用 recover 捕获 panic
		defer func() {
			if r := recover(); r != nil {
				if s.isConnectionError(fmt.Errorf("%v", r)) {
					// 连接断开导致的 panic，记录日志但不报错
					s.logger.Info("连接断开导致的异常", zap.Any("error", r))
				} else {
					// 其他 panic 正常处理
					s.logger.Error("处理请求时发生异常", zap.Any("error", r))
					if !ctx.Writer.Written() {
						ctx.AbortWithStatus(http.StatusInternalServerError)
					}
				}
			}
		}()

		ctx.Next()
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
