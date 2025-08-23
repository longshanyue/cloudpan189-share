package universalfs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/enc"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/xxcheng123/cloudpan189-interface/client"
	"github.com/xxcheng123/cloudpan189-share/configs"
	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"github.com/xxcheng123/cloudpan189-share/internal/types"
	"github.com/xxcheng123/multistreamer"
)

type fileDownloadRequest struct {
	ID        int64  `form:"id" binding:"required"`
	TimeStamp int64  `form:"timestamp" binding:"required"`
	Random    string `form:"random" binding:"required"`
	Sign      string `form:"sign" binding:"required"`
}

type DoResult struct {
	Content  string
	HttpCode int
	Err      error
}

// 全局HTTP客户端，复用连接
var globalHTTPClient = &http.Client{
	Timeout: 0,
	Transport: &http.Transport{
		DisableKeepAlives:     false,
		MaxIdleConns:          200,
		MaxIdleConnsPerHost:   20,
		MaxConnsPerHost:       50,
		IdleConnTimeout:       120 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableCompression:    true,
		WriteBufferSize:       128 * 1024,
		ReadBufferSize:        128 * 1024,
		ForceAttemptHTTP2:     false,
	},
}

func (s *service) FileDownload() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req = new(fileDownloadRequest)

		if err := ctx.ShouldBindQuery(req); err != nil {
			s.logger.Warn("文件下载请求参数错误", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})

			return
		}

		if req.TimeStamp != -1 && req.TimeStamp < time.Now().Unix() {
			s.logger.Warn("文件下载请求时间戳已过期",
				zap.Int64("timestamp", req.TimeStamp),
				zap.Int64("fileId", req.ID))
			ctx.JSON(http.StatusUnauthorized, types.ErrResponse{
				Code:    http.StatusUnauthorized,
				Message: "请求时间戳已过期",
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

		if !enc.Verify(values, key) {
			s.logger.Warn("文件下载请求签名验证失败",
				zap.Int64("fileId", req.ID),
				zap.String("sign", req.Sign))
			ctx.JSON(http.StatusUnauthorized, types.ErrResponse{
				Code:    http.StatusUnauthorized,
				Message: "签名验证失败",
			})

			return
		}

		if v, ok := s.cache.Get(fmt.Sprintf("file::url::%d", req.ID)); ok {
			if downURL, ok := v.(string); ok {
				ctx.Header("X-Download-Url-Cache", "true")
				s.doResponse(ctx, downURL)
				return
			} else {
				s.logger.Warn("缓存中的URL格式错误", zap.Int64("fileId", req.ID))
				// 继续执行，重新获取下载链接
			}
		}

		file := &models.VirtualFile{}
		if err := s.db.WithContext(ctx).Where("id = ?", req.ID).First(file).Error; err != nil {
			s.logger.Error("查询文件信息失败", zap.Int64("fileId", req.ID), zap.Error(err))
			ctx.JSON(http.StatusNotFound, types.ErrResponse{
				Code:    http.StatusNotFound,
				Message: "文件未找到",
			})

			return
		}

		if file.OsType == models.OsTypeRealFile {
			s.handleRealFileDownload(ctx, file, req.ID)

			return
		}

		s.handleCloudFileDownload(ctx, req.ID)
	}
}

func (s *service) handleRealFileDownload(ctx *gin.Context, file *models.VirtualFile, fileID int64) {
	v, ok := file.Addition[consts.FileAdditionKeyFilePath]
	if !ok {
		s.logger.Error("真实文件路径信息缺失",
			zap.Int64("fileId", fileID),
			zap.String("fileName", file.Name))
		ctx.JSON(http.StatusInternalServerError, types.ErrResponse{
			Code:    http.StatusInternalServerError,
			Message: "文件路径不存在",
		})

		return
	}

	filePath, ok := v.(string)
	if !ok {
		s.logger.Error("真实文件路径格式错误",
			zap.Int64("fileId", fileID),
			zap.String("fileName", file.Name))
		ctx.JSON(http.StatusInternalServerError, types.ErrResponse{
			Code:    http.StatusInternalServerError,
			Message: "文件路径格式错误",
		})

		return
	}
	fullPath := path.Join(configs.GetConfig().FileDir, filePath)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		s.logger.Error("真实文件不存在",
			zap.Int64("fileId", fileID),
			zap.String("fileName", file.Name),
			zap.String("fullPath", fullPath))
		ctx.JSON(http.StatusNotFound, types.ErrResponse{
			Code:    http.StatusNotFound,
			Message: "文件不存在",
		})

		return
	}

	filename := file.Name
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s",
		filename, url.QueryEscape(filename)))

	s.logger.Info("开始下载真实文件",
		zap.Int64("fileId", fileID),
		zap.String("fileName", file.Name),
		zap.String("fullPath", fullPath))

	ctx.File(fullPath)
}

func (s *service) handleCloudFileDownload(ctx *gin.Context, fileID int64) {
	_result, err, _ := s.g.Do(fmt.Sprintf("file::url::%d", fileID), func() (interface{}, error) {
		u, httpCode, err := s.getFileDownloadURL(ctx, fileID)

		return &DoResult{
			Content:  u,
			HttpCode: httpCode,
			Err:      err,
		}, nil
	})

	if err != nil {
		s.logger.Error("获取文件下载链接失败", zap.Int64("fileId", fileID), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, types.ErrResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})

		return
	}

	result, ok := _result.(*DoResult)
	if !ok {
		s.logger.Error("类型断言失败", zap.Int64("fileId", fileID))
		ctx.JSON(http.StatusInternalServerError, types.ErrResponse{
			Code:    http.StatusInternalServerError,
			Message: "内部错误",
		})
		return
	}

	if result.Err != nil {
		s.logger.Error("处理文件下载请求失败",
			zap.Int64("fileId", fileID),
			zap.Int("httpCode", result.HttpCode),
			zap.Error(result.Err))
		ctx.JSON(result.HttpCode, types.ErrResponse{
			Code:    result.HttpCode,
			Message: result.Err.Error(),
		})

		return
	}

	if result.HttpCode == http.StatusOK {
		ctx.String(http.StatusOK, result.Content)

		return
	}

	s.doResponse(ctx, result.Content)
}

func (s *service) doResponse(ctx *gin.Context, url string) {
	if shared.Setting.MultipleStream {
		s.handleMultiStreamResponse(ctx, url)
	} else if shared.Setting.LocalProxy {
		s.handleLocalProxy(ctx, url)
	} else {
		ctx.Header("X-Transfer-Type", "redirect")
		ctx.Redirect(http.StatusFound, url)
	}
}

func (s *service) handleMultiStreamResponse(ctx *gin.Context, url string) {
	ctx.Header("X-Transfer-Type", "multi_stream")
	ctx.Header("X-Transfer-Chunk-Size", strconv.FormatInt(shared.MultipleStreamChunkSize, 10))
	ctx.Header("X-Transfer-Chunk-Size-Format", utils.FormatBytes(shared.MultipleStreamChunkSize))
	ctx.Header("X-Transfer-Thread-Count", strconv.Itoa(shared.MultipleStreamThreadCount))

	httpReq := ctx.Request.Header.Clone()
	httpReq.Set("Accept-Encoding", "identity")
	httpReq.Del("Content-Type")

	streamer, err := multistreamer.NewStreamer(ctx,
		url,
		httpReq,
		multistreamer.WithLogger(s.logger),
		multistreamer.WithThreads(shared.MultipleStreamThreadCount),
		multistreamer.WithChunkSize(shared.MultipleStreamChunkSize),
	)
	if err != nil {
		s.logger.Error("多线程流初始化失败", zap.Error(err), zap.String("url", url))
		ctx.JSON(http.StatusInternalServerError, types.ErrResponse{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("多线程流初始化失败: %v", err),
		})

		return
	}

	for k, v := range streamer.GetResponseHeader() {
		ctx.Header(k, v[0])
	}

	ctx.Status(streamer.HTTPCode())

	if err = streamer.Transfer(ctx, ctx.Writer); err != nil {
		if s.isConnectionError(err) {
			s.logger.Info("客户端连接断开", zap.String("url", url))
		} else {
			s.logger.Error("多线程流文件传输失败", zap.Error(err), zap.String("url", url))
		}
	}
}

func (s *service) handleLocalProxy(ctx *gin.Context, url string) {
	start := time.Now()
	ctx.Header("X-Transfer-Type", "local_proxy")

	req, err := http.NewRequestWithContext(ctx.Request.Context(), http.MethodGet, url, nil)
	if err != nil {
		s.logger.Error("创建本地代理请求失败", zap.Error(err), zap.String("url", url))
		ctx.JSON(http.StatusInternalServerError, types.ErrResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})

		return
	}

	s.copyOptimizedHeaders(ctx.Request.Header, req.Header)

	rangeHeader := ctx.Request.Header.Get("Range")
	isRangeRequest := rangeHeader != ""

	if isRangeRequest {
		req.Header.Set("Connection", "keep-alive")
	}

	resp, err := globalHTTPClient.Do(req)
	if err != nil {
		if ctx.Request.Context().Err() != nil {
			s.logger.Info("客户端断开连接", zap.String("url", url))

			return
		}

		s.logger.Error("本地代理请求失败", zap.Error(err), zap.String("url", url))
		ctx.JSON(http.StatusInternalServerError, types.ErrResponse{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("本地代理请求失败: %v", err),
		})

		return
	}
	defer resp.Body.Close()

	s.copyOptimizedResponseHeaders(resp.Header, ctx)
	ctx.Status(resp.StatusCode)

	if flusher, ok := ctx.Writer.(http.Flusher); ok {
		flusher.Flush()
	}

	var copyErr error
	if isRangeRequest || resp.ContentLength < 1024*1024 {
		copyErr = s.fastCopy(ctx, ctx.Writer, resp.Body)
	} else {
		copyErr = s.streamCopy(ctx, ctx.Writer, resp.Body)
	}

	duration := time.Since(start)
	if copyErr != nil {
		if s.isConnectionError(copyErr) {
			s.logger.Info("客户端连接断开",
				zap.String("url", url),
				zap.Duration("duration", duration))
		} else {
			s.logger.Error("本地代理文件传输失败",
				zap.Error(copyErr),
				zap.String("url", url),
				zap.Duration("duration", duration))
		}
	} else {
		s.logger.Info("本地代理请求完成",
			zap.String("url", url),
			zap.Duration("duration", duration),
			zap.String("user_agent", ctx.Request.UserAgent()))

		if duration > 5*time.Second {
			s.logger.Warn("本地代理请求响应较慢",
				zap.String("url", url),
				zap.Duration("duration", duration))
		}
	}
}

func (s *service) copyOptimizedHeaders(src, dst http.Header) {
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

func (s *service) copyOptimizedResponseHeaders(src http.Header, ctx *gin.Context) {
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

	ctx.Header("Connection", "keep-alive")
	ctx.Header("Keep-Alive", "timeout=120, max=100")
}

func (s *service) fastCopy(ctx *gin.Context, dst io.Writer, src io.Reader) error {
	buf := make([]byte, 256*1024)
	_, err := io.CopyBuffer(dst, src, buf)

	if flusher, ok := dst.(http.Flusher); ok {
		flusher.Flush()
	}

	return err
}

func (s *service) streamCopy(ctx *gin.Context, dst io.Writer, src io.Reader) error {
	buf := make([]byte, 128*1024)
	flusher, canFlush := dst.(http.Flusher)

	var written int64
	flushInterval := 0

	for {
		select {
		case <-ctx.Request.Context().Done():
			return ctx.Request.Context().Err()
		default:
		}

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

	if canFlush {
		flusher.Flush()
	}

	return nil
}

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

func (s *service) ConnectionMonitorMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxWithCancel, cancel := context.WithCancel(ctx.Request.Context())
		defer cancel()

		ctx.Request = ctx.Request.WithContext(ctxWithCancel)

		defer func() {
			if r := recover(); r != nil {
				if s.isConnectionError(fmt.Errorf("%v", r)) {
					s.logger.Info("连接断开导致的异常", zap.Any("error", r))
				} else {
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

func (s *service) getFileDownloadURL(ctx context.Context, id int64) (string, int, error) {
	var file = new(models.VirtualFile)
	if err := s.db.WithContext(ctx).Where("id", id).First(file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("获取下载链接时文件未找到", zap.Int64("fileId", id))

			return "", http.StatusNotFound, errors.New("文件未找到")
		}

		s.logger.Error("查询文件信息失败", zap.Int64("fileId", id), zap.Error(err))

		return "", http.StatusInternalServerError, err
	}

	var (
		familyId string
	)

	if file.OsType == models.OsTypeStrmFile {
		s.logger.Info("生成STRM文件下载链接",
			zap.Int64("fileId", id),
			zap.Int64("linkId", file.LinkId))

		return s.generateDownloadURLWithNeverExpire(file.LinkId), http.StatusOK, nil
	} else if file.OsType == models.OsTypeCloudFamilyFile {
		familyId = utils.GetString(file.Addition, consts.FileAdditionKeyFamilyId)
		if familyId == "" {
			return "", http.StatusBadRequest, errors.New("familyId is empty")
		}
	}

	cloudTokenId, err := s.findCloudTokenId(ctx, file)
	if err != nil {
		return "", http.StatusBadRequest, err
	}

	ct := new(models.CloudToken)
	if err := s.db.WithContext(ctx).Where("id", cloudTokenId).First(ct).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("云盘令牌未找到",
				zap.Int64("fileId", id),
				zap.Int64("cloudTokenId", cloudTokenId))

			return "", http.StatusBadRequest, errors.New("绑定的令牌查找失败，可能被删除或隐藏")
		}

		s.logger.Error("查询云盘令牌失败",
			zap.Int64("fileId", id),
			zap.Int64("cloudTokenId", cloudTokenId),
			zap.Error(err))

		return "", http.StatusInternalServerError, err
	}

	fileId := utils.String(file.Addition[consts.FileAdditionKeyFileId])
	s.logger.Info("开始获取云盘文件下载链接",
		zap.Int64("fileId", id),
		zap.String("cloudFileId", fileId))

	var (
		downloadURL string
	)

	if file.OsType == models.OsTypeCloudFamilyFile {
		result, err := client.New().WithToken(client.NewAuthToken(ct.AccessToken, ct.ExpiresIn)).
			FamilyGetFileDownload(ctx, client.String(familyId), client.String(fileId))

		if err != nil {
			s.logger.Error("获取云盘文件下载链接失败",
				zap.Int64("fileId", id),
				zap.String("cloudFileId", fileId),
				zap.Error(err))
			return "", http.StatusInternalServerError, err
		}

		downloadURL = html.UnescapeString(result.FileDownloadUrl)
	} else {
		result, err := client.New().WithToken(client.NewAuthToken(ct.AccessToken, ct.ExpiresIn)).
			GetFileDownload(ctx, client.String(fileId), func(req *client.GetFileDownloadRequest) {
				if v, ok := file.Addition[consts.FileAdditionKeyShareId]; ok {
					req.ShareId, _ = utils.Int64(v)
				}
			})

		if err != nil {
			s.logger.Error("获取云盘文件下载链接失败",
				zap.Int64("fileId", id),
				zap.String("cloudFileId", fileId),
				zap.Error(err))

			return "", http.StatusInternalServerError, err
		}

		downloadURL = result.FileDownloadUrl
	}

	resp, err := http.Get(downloadURL)
	if err != nil {
		s.logger.Error("请求云盘下载链接失败",
			zap.Int64("fileId", id),
			zap.String("downloadUrl", downloadURL),
			zap.Error(err))

		return "", http.StatusInternalServerError, err
	}
	defer resp.Body.Close()

	finalUrl := resp.Request.URL.String()
	s.cache.Set(fmt.Sprintf("file::url::%d", file.ID), finalUrl, time.Minute)

	s.logger.Info("成功获取文件下载链接",
		zap.Int64("fileId", id),
		zap.String("finalUrl", finalUrl))

	return finalUrl, http.StatusFound, nil
}

func (s *service) findCloudTokenId(ctx context.Context, file *models.VirtualFile) (int64, error) {
	var cloudTokenId int64
	scanFile := file

	for {
		if v, ok := scanFile.Addition[consts.FileAdditionKeyCloudToken]; ok {
			cloudTokenId, _ = utils.Int64(v)
			break
		}

		if scanFile.ParentId == 0 {
			s.logger.Error("文件未绑定云盘令牌",
				zap.Int64("fileId", file.ID),
				zap.String("fileName", file.Name))

			return 0, errors.New("当前资源没有绑定用于获取播放链接的令牌")
		}

		var parent = new(models.VirtualFile)
		if err := s.db.WithContext(ctx).Where("id", scanFile.ParentId).First(parent).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				s.logger.Error("查找父级文件失败",
					zap.Int64("fileId", file.ID),
					zap.Int64("parentId", scanFile.ParentId))

				return 0, errors.New("文件未找到")
			}

			s.logger.Error("查询父级文件信息失败",
				zap.Int64("fileId", file.ID),
				zap.Int64("parentId", scanFile.ParentId),
				zap.Error(err))

			return 0, errors.New("当前资源没有绑定用于获取播放链接的令牌")
		}

		scanFile = parent
	}

	return cloudTokenId, nil
}
