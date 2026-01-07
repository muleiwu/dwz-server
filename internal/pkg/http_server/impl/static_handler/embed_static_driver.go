package static_handler

import (
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// EmbedStaticDriver embed.FS 驱动实现
type EmbedStaticDriver struct {
	embedFS embed.FS
}

// NewEmbedStaticDriver 创建 embed 驱动实例
func NewEmbedStaticDriver(embedFS embed.FS) *EmbedStaticDriver {
	return &EmbedStaticDriver{
		embedFS: embedFS,
	}
}

// FileExists 检查文件是否存在于 embed.FS 中
func (d *EmbedStaticDriver) FileExists(path string) bool {
	// 移除开头的斜杠
	path = strings.TrimPrefix(path, "/")
	_, err := fs.Stat(d.embedFS, path)
	return err == nil
}

// GetFS 获取指定目录的文件系统
func (d *EmbedStaticDriver) GetFS(dir string) (http.FileSystem, error) {
	// embed.FS 的根目录就是 static，直接使用 dir
	subFS, err := fs.Sub(d.embedFS, dir)
	if err != nil {
		return nil, err
	}
	return http.FS(subFS), nil
}

// ServeFile 提供文件到 gin.Context
func (d *EmbedStaticDriver) ServeFile(c *gin.Context, dir string, relativePath string) error {
	// embed.FS 的根目录就是 static，直接使用 dir
	subFS, err := fs.Sub(d.embedFS, dir)
	if err != nil {
		return err
	}

	// 移除开头的斜杠
	relativePath = strings.TrimPrefix(relativePath, "/")

	// 先检查文件本身是否存在
	if d.fileExistsInFS(subFS, relativePath) {
		c.FileFromFS(relativePath, http.FS(subFS))
		return nil
	}

	// 检查 路径/index.html 是否存在
	indexPath := strings.TrimSuffix(relativePath, "/") + "/index.html"
	if d.fileExistsInFS(subFS, indexPath) {
		c.FileFromFS(indexPath, http.FS(subFS))
		return nil
	}

	return errors.New("file not found")
}

// fileExistsInFS 检查文件是否存在于文件系统中
func (d *EmbedStaticDriver) fileExistsInFS(filesystem fs.FS, path string) bool {
	path = strings.TrimPrefix(path, "/")
	_, err := fs.Stat(filesystem, path)
	return err == nil
}

// GetDriverName 获取驱动名称
func (d *EmbedStaticDriver) GetDriverName() string {
	return "embed"
}

// WalkFiles 递归遍历目录，返回所有文件的相对路径
// errorCallback: 可选的错误回调函数，用于记录遍历过程中的错误
func (d *EmbedStaticDriver) WalkFiles(dir string, errorCallback WalkErrorCallback) ([]string, error) {
	var files []string
	// embed.FS 的根目录就是 static，所以直接使用 dir 作为 rootPath
	rootPath := dir

	err := fs.WalkDir(d.embedFS, rootPath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			// 目录读取失败时记录日志并继续处理其他目录
			if errorCallback != nil {
				errorCallback(path, err)
			}
			return nil
		}

		// 只处理文件，跳过目录
		if entry.IsDir() {
			return nil
		}

		// 计算相对路径（移除 rootPath 前缀）
		relativePath := strings.TrimPrefix(path, rootPath)
		relativePath = strings.TrimPrefix(relativePath, "/")

		if relativePath != "" {
			files = append(files, relativePath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

