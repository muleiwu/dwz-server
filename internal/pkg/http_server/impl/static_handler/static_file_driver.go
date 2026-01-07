package static_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// WalkErrorCallback 遍历错误回调函数类型
// path: 发生错误的路径
// err: 错误信息
type WalkErrorCallback func(path string, err error)

// StaticFileDriver 静态文件驱动接口
type StaticFileDriver interface {
	// FileExists 检查文件是否存在
	FileExists(path string) bool

	// GetFS 获取文件系统（用于 gin.StaticFS）
	GetFS(dir string) (http.FileSystem, error)

	// ServeFile 直接提供文件到 gin.Context
	ServeFile(c *gin.Context, dir string, relativePath string) error

	// GetDriverName 获取驱动名称（用于日志）
	GetDriverName() string

	// WalkFiles 递归遍历目录，返回所有文件的相对路径
	// errorCallback: 可选的错误回调函数，用于记录遍历过程中的错误
	WalkFiles(dir string, errorCallback WalkErrorCallback) ([]string, error)
}

