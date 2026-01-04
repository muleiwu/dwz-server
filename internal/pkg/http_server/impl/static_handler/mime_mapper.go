package static_handler

import (
	"path/filepath"
	"strings"
)

// MimeTypeMapper MIME 类型映射器
type MimeTypeMapper struct {
	mimeTypes map[string]string
}

// NewMimeTypeMapper 创建 MIME 类型映射器实例
func NewMimeTypeMapper() *MimeTypeMapper {
	return &MimeTypeMapper{
		mimeTypes: map[string]string{
			// 文本类型
			".html": "text/html; charset=utf-8",
			".htm":  "text/html; charset=utf-8",
			".css":  "text/css; charset=utf-8",
			".js":   "application/javascript; charset=utf-8",
			".mjs":  "application/javascript; charset=utf-8",
			".json": "application/json; charset=utf-8",
			".xml":  "application/xml; charset=utf-8",
			".txt":  "text/plain; charset=utf-8",
			".md":   "text/markdown; charset=utf-8",

			// 图片类型
			".png":  "image/png",
			".jpg":  "image/jpeg",
			".jpeg": "image/jpeg",
			".gif":  "image/gif",
			".svg":  "image/svg+xml",
			".ico":  "image/x-icon",
			".webp": "image/webp",
			".bmp":  "image/bmp",

			// 字体类型
			".woff":  "font/woff",
			".woff2": "font/woff2",
			".ttf":   "font/ttf",
			".otf":   "font/otf",
			".eot":   "application/vnd.ms-fontobject",

			// 音视频类型
			".mp3":  "audio/mpeg",
			".mp4":  "video/mp4",
			".webm": "video/webm",
			".ogg":  "audio/ogg",
			".wav":  "audio/wav",

			// 文档类型
			".pdf": "application/pdf",
			".zip": "application/zip",
			".tar": "application/x-tar",
			".gz":  "application/gzip",

			// 其他
			".map": "application/json; charset=utf-8",
		},
	}
}

// GetMimeType 根据文件路径获取 MIME 类型
func (m *MimeTypeMapper) GetMimeType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	if mimeType, ok := m.mimeTypes[ext]; ok {
		return mimeType
	}
	// 默认返回二进制流类型
	return "application/octet-stream"
}

