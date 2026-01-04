package static_handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// DiskStaticDriver 磁盘文件系统驱动实现
type DiskStaticDriver struct {
	baseDir string
}

// NewDiskStaticDriver 创建磁盘驱动实例
func NewDiskStaticDriver(baseDir string) *DiskStaticDriver {
	return &DiskStaticDriver{
		baseDir: baseDir,
	}
}

// FileExists 检查文件是否存在于磁盘中
func (d *DiskStaticDriver) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetFS 获取指定目录的文件系统
func (d *DiskStaticDriver) GetFS(dir string) (http.FileSystem, error) {
	diskPath := fmt.Sprintf("%s/%s", d.baseDir, dir)
	return http.Dir(diskPath), nil
}

// ServeFile 提供文件到 gin.Context
func (d *DiskStaticDriver) ServeFile(c *gin.Context, dir string, relativePath string) error {
	filePath := fmt.Sprintf("%s/%s%s", d.baseDir, dir, relativePath)

	// 先检查文件本身是否存在
	if d.FileExists(filePath) {
		c.File(filePath)
		return nil
	}

	// 检查 路径/index.html 是否存在
	indexPath := fmt.Sprintf("%s/index.html", strings.TrimSuffix(filePath, "/"))
	if d.FileExists(indexPath) {
		c.File(indexPath)
		return nil
	}

	return fmt.Errorf("file not found")
}

// GetDriverName 获取驱动名称
func (d *DiskStaticDriver) GetDriverName() string {
	return "disk"
}

// WalkFiles 递归遍历目录，返回所有文件的相对路径
// errorCallback: 可选的错误回调函数，用于记录遍历过程中的错误
func (d *DiskStaticDriver) WalkFiles(dir string, errorCallback WalkErrorCallback) ([]string, error) {
	var files []string
	rootPath := filepath.Join(d.baseDir, dir)

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// 目录读取失败时记录日志并继续处理其他目录
			if errorCallback != nil {
				errorCallback(path, err)
			}
			return nil
		}

		// 只处理文件，跳过目录
		if info.IsDir() {
			return nil
		}

		// 计算相对路径
		relativePath, err := filepath.Rel(rootPath, path)
		if err != nil {
			if errorCallback != nil {
				errorCallback(path, err)
			}
			return nil
		}

		// 转换为 URL 路径格式（使用正斜杠）
		relativePath = filepath.ToSlash(relativePath)
		files = append(files, relativePath)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

