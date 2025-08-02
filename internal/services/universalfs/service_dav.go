package universalfs

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

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
		case "OPTIONS":
			allow := "OPTIONS, HEAD, GET, POST, PROPFIND"
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

	// 添加当前文件/文件夹的响应
	s.addPropResponse(&xmlResponse, fileInfo, ctx.Request.URL.Path)

	// 如果是文件夹且depth不为0，添加子项
	if fileInfo.IsFolder == 1 && depth != "0" && len(fileInfo.Children) > 0 {
		for _, child := range fileInfo.Children {
			// 构建子项的URL路径
			childPath := path.Join(ctx.Request.URL.Path, url.PathEscape(child.Name))
			if child.IsFolder == 1 {
				childPath += "/"
			}
			s.addPropResponse(&xmlResponse, child, childPath)
		}
	}

	xmlResponse.WriteString(`</D:multistatus>`)

	ctx.Data(http.StatusMultiStatus, "application/xml; charset=utf-8", []byte(xmlResponse.String()))
}

// addPropResponse 添加单个文件/文件夹的属性响应
func (s *service) addPropResponse(xmlResponse *strings.Builder, fileInfo *FileInfo, href string) {
	xmlResponse.WriteString(`<D:response>`)
	xmlResponse.WriteString(fmt.Sprintf(`<D:href>%s</D:href>`, escapeXML(href)))
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
