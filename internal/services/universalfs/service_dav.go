package universalfs

import (
	"errors"
	"fmt"
	"golang.org/x/net/webdav"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *service) responseDav(ctx *gin.Context, fileInfo *FileInfo) {
	switch ctx.Request.Method {
	case "GET", "HEAD", "POST":
		if fileInfo.IsFolder == 1 {
			ctx.JSON(http.StatusMethodNotAllowed, gin.H{
				"code":    http.StatusMethodNotAllowed,
				"message": "Method not allowed",
			})

			return
		}

		ctx.Redirect(http.StatusFound, fileInfo.DownloadURL)
	case "PROPFIND":
		// 转webdav格式
		s.generatePropfindResponse(ctx, fileInfo)
	}
}

// generatePropfindResponse 生成WebDAV PROPFIND响应
func (s *service) generatePropfindResponse(ctx *gin.Context, fileInfo *FileInfo) {
	depth := ctx.GetHeader("Depth")
	if depth == "" {
		depth = "1"
	}

	// 设置WebDAV响应头
	ctx.Header("Content-Type", "application/xml; charset=utf-8")
	ctx.Header("DAV", "1, 2")

	// 构建XML响应
	var xmlResponse strings.Builder
	xmlResponse.WriteString(`<?xml version="1.0" encoding="utf-8"?>`)
	xmlResponse.WriteString(`<D:multistatus xmlns:D="DAV:">`)

	// 当前路径处理
	currentPath := s.normalizeWebDAVPath(ctx.Request.URL.Path)
	if fileInfo.IsFolder == 1 && !strings.HasSuffix(currentPath, "/") {
		currentPath += "/"
	}

	s.addPropResponse(&xmlResponse, fileInfo, currentPath)

	// 如果是文件夹且depth不为0，添加子项
	if fileInfo.IsFolder == 1 && depth != "0" && len(fileInfo.Children) > 0 {
		for _, child := range fileInfo.Children {
			childPath := s.buildChildPath(currentPath, child.Name, child.IsFolder == 1)
			s.addPropResponse(&xmlResponse, child, childPath)
		}
	}

	xmlResponse.WriteString(`</D:multistatus>`)

	ctx.Data(http.StatusMultiStatus, "application/xml; charset=utf-8", []byte(xmlResponse.String()))
}

// addPropResponse 添加单个文件/文件夹的属性响应
func (s *service) addPropResponse(xmlResponse *strings.Builder, fileInfo *FileInfo, href string) {
	xmlResponse.WriteString(`<D:response>`)

	// 正确编码 href
	encodedHref := s.encodeWebDAVPath(href)
	xmlResponse.WriteString(fmt.Sprintf(`<D:href>%s</D:href>`, escapeXML(encodedHref)))

	xmlResponse.WriteString(`<D:propstat>`)
	xmlResponse.WriteString(`<D:prop>`)

	// 资源类型
	if fileInfo.IsFolder == 1 {
		xmlResponse.WriteString(`<D:resourcetype><D:collection/></D:resourcetype>`)
	} else {
		xmlResponse.WriteString(`<D:resourcetype/>`)
	}

	// 显示名称
	xmlResponse.WriteString(fmt.Sprintf(`<D:displayname>%s</D:displayname>`, escapeXML(fileInfo.Name)))

	// 内容长度（文件大小）
	if fileInfo.IsFolder == 0 {
		xmlResponse.WriteString(fmt.Sprintf(`<D:getcontentlength>%d</D:getcontentlength>`, fileInfo.Size))
	}

	// 内容类型
	if fileInfo.IsFolder == 0 {
		contentType := s.getContentType(fileInfo.Name)
		xmlResponse.WriteString(fmt.Sprintf(`<D:getcontenttype>%s</D:getcontenttype>`, escapeXML(contentType)))
	}

	// 最后修改时间
	if fileInfo.ModifyDate != "" {
		if modTime, err := time.Parse("2006-01-02 15:04:05", fileInfo.ModifyDate); err == nil {
			rfc1123Time := modTime.UTC().Format(time.RFC1123)
			xmlResponse.WriteString(fmt.Sprintf(`<D:getlastmodified>%s</D:getlastmodified>`, escapeXML(rfc1123Time)))
		}
	}

	// 创建时间
	if fileInfo.CreateDate != "" {
		if createTime, err := time.Parse("2006-01-02 15:04:05", fileInfo.CreateDate); err == nil {
			rfc3339Time := createTime.UTC().Format(time.RFC3339)
			xmlResponse.WriteString(fmt.Sprintf(`<D:creationdate>%s</D:creationdate>`, escapeXML(rfc3339Time)))
		}
	}

	// ETag（使用文件哈希或修改时间）
	if fileInfo.Hash != "" {
		xmlResponse.WriteString(fmt.Sprintf(`<D:getetag>"%s"</D:getetag>`, escapeXML(fileInfo.Hash)))
	} else if fileInfo.ModifyDate != "" {
		// 如果没有哈希，使用修改时间作为ETag
		etag := fmt.Sprintf("%d-%s", fileInfo.ID, fileInfo.ModifyDate)
		xmlResponse.WriteString(fmt.Sprintf(`<D:getetag>"%s"</D:getetag>`, escapeXML(etag)))
	}

	xmlResponse.WriteString(`</D:prop>`)
	xmlResponse.WriteString(`<D:status>HTTP/1.1 200 OK</D:status>`)
	xmlResponse.WriteString(`</D:propstat>`)
	xmlResponse.WriteString(`</D:response>`)
}

// encodeWebDAVPath 对 WebDAV 路径进行正确的 URL 编码
func (s *service) encodeWebDAVPath(rawPath string) string {
	// 分割路径为各个部分
	parts := strings.Split(strings.Trim(rawPath, "/"), "/")

	// 对每个部分进行 URL 编码
	encodedParts := make([]string, len(parts))
	for i, part := range parts {
		if part != "" {
			encodedParts[i] = url.PathEscape(part)
		}
	}

	// 重新组合路径
	encodedPath := "/" + strings.Join(encodedParts, "/")

	// 如果原路径以 / 结尾（文件夹），保持这个特征
	if strings.HasSuffix(rawPath, "/") && !strings.HasSuffix(encodedPath, "/") {
		encodedPath += "/"
	}

	return encodedPath
}

// normalizeWebDAVPath 标准化 WebDAV 路径
func (s *service) normalizeWebDAVPath(rawPath string) string {
	// 清理路径
	cleanPath := path.Clean(rawPath)

	// 确保以 / 开头
	if !strings.HasPrefix(cleanPath, "/") {
		cleanPath = "/" + cleanPath
	}

	return cleanPath
}

// buildChildPath 构建子项路径
func (s *service) buildChildPath(parentPath, childName string, isFolder bool) string {
	// 标准化父路径
	parentPath = s.normalizeWebDAVPath(parentPath)

	// 移除末尾的 /
	parentPath = strings.TrimSuffix(parentPath, "/")

	// 构建子路径
	childPath := parentPath + "/" + childName

	// 如果是文件夹，添加末尾的 /
	if isFolder {
		childPath += "/"
	}

	return childPath
}

// getContentType 根据文件扩展名获取MIME类型
func (s *service) getContentType(filename string) string {
	ext := strings.ToLower(path.Ext(filename))
	switch ext {
	case ".txt":
		return "text/plain"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".pdf":
		return "application/pdf"
	case ".zip":
		return "application/zip"
	case ".rar":
		return "application/x-rar-compressed"
	case ".7z":
		return "application/x-7z-compressed"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".bmp":
		return "image/bmp"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".flac":
		return "audio/flac"
	case ".aac":
		return "audio/aac"
	case ".ogg":
		return "audio/ogg"
	case ".m4a":
		return "audio/mp4"
	case ".mp4":
		return "video/mp4"
	case ".avi":
		return "video/x-msvideo"
	case ".mkv":
		return "video/x-matroska"
	case ".mov":
		return "video/quicktime"
	case ".wmv":
		return "video/x-ms-wmv"
	case ".flv":
		return "video/x-flv"
	case ".webm":
		return "video/webm"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".ppt":
		return "application/vnd.ms-powerpoint"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	default:
		return "application/octet-stream"
	}
}

// escapeXML 转义XML特殊字符
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

const (
	lockTimeout = 30 * time.Minute
)

func (s *service) lock(now time.Time, root string) (token string, status int, err error) {
	token, err = s.LockSystem.Create(now, webdav.LockDetails{
		Root:      root,
		Duration:  lockTimeout,
		ZeroDepth: true,
	})
	if err != nil {
		if errors.Is(err, webdav.ErrLocked) {
			return "", webdav.StatusLocked, err
		}
		return "", http.StatusInternalServerError, err
	}
	return token, 0, nil
}

func (s *service) confirmLocks(src, dst string) (release func(), status int, err error) {
	now, srcToken, dstToken := time.Now(), "", ""
	if src != "" {
		srcToken, status, err = s.lock(now, src)
		if err != nil {
			return nil, status, err
		}
	}
	if dst != "" {
		dstToken, status, err = s.lock(now, dst)
		if err != nil {
			if srcToken != "" {
				_ = s.LockSystem.Unlock(now, srcToken)
			}
			return nil, status, err
		}
	}

	return func() {
		if dstToken != "" {
			_ = s.LockSystem.Unlock(now, dstToken)
		}
		if srcToken != "" {
			_ = s.LockSystem.Unlock(now, srcToken)
		}
	}, 0, nil
}
