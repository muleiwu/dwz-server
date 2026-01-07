package impl

import (
	"embed"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"cnb.cool/mliev/open/dwz-server/internal/pkg/http_server/impl/static_handler"
	"github.com/gin-gonic/gin"
)

type StaticHandler struct {
	helper      interfaces.HelperInterface
	engine      *gin.Engine
	driver      static_handler.StaticFileDriver
	mimeMapper  *static_handler.MimeTypeMapper
	directories map[string]string // 记录所有目录路径：URL路径 -> 静态目录路径
}

func NewStaticHandler(helper interfaces.HelperInterface, engine *gin.Engine) *StaticHandler {
	return &StaticHandler{
		helper:      helper,
		engine:      engine,
		mimeMapper:  static_handler.NewMimeTypeMapper(),
		directories: make(map[string]string),
	}
}

// setupStaticFileServers 为嵌入的静态文件设置HTTP服务
func (receiver *StaticHandler) setupStaticFileServers() {

	if !receiver.helper.GetConfig().GetBool("http.load_static", false) {
		return
	}

	staticDir := receiver.helper.GetConfig().GetString("http.static_dir", "")

	if staticDir == "" {
		receiver.helper.GetLogger().Warn("没有配置需要加载的静态目录")
		return
	}

	// 初始化驱动（只判断一次）
	staticMode := receiver.helper.GetConfig().GetString("http.static_mode", "embed")
	httpMode := receiver.helper.GetConfig().GetString("http.mode", "release")
	receiver.helper.GetLogger().Debug(fmt.Sprintf("当前静态文件模式：%s, HTTP模式：%s", staticMode, httpMode))
	staticFs := receiver.helper.GetConfig().Get("static.fs", map[string]embed.FS{}).(map[string]embed.FS)

	if staticMode == "disk" {
		// disk 模式下使用磁盘驱动
		// 根据 http.mode 决定是否启用缓存
		enableCache := (httpMode == "release")
		receiver.driver = static_handler.NewDiskStaticDriver(".", enableCache)

		if enableCache {
			receiver.helper.GetLogger().Info("Disk 模式（缓存启用）：首次读取后缓存到内存，提升性能")
		} else {
			receiver.helper.GetLogger().Info("Disk 模式（实时读取）：每次请求从磁盘读取，支持热更新")
		}
	} else {
		// embed 模式下使用 embed 驱动
		embeddedFs, ok := staticFs["web.static"]
		if !ok {
			receiver.helper.GetLogger().Debug("不存在需要对Web暴露的静态资源")
			return
		}
		receiver.driver = static_handler.NewEmbedStaticDriver(embeddedFs)
		receiver.helper.GetLogger().Info("Embed 模式：使用 embed 驱动加载静态文件")
	}

	receiver.loadStatic(staticDir)

	// 注册目录路由（处理 /admin/ 这样的路径，尝试返回 index.html 或 index.htm）
	receiver.registerDirectoryRoutes()

	// 统一处理未匹配的路由
	// 由于所有静态文件和 API 路由都已单独注册，未匹配的路径直接返回 404
	receiver.engine.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "Not Found")
	})

}

func (receiver *StaticHandler) loadStatic(rootDir string) {
	// 创建错误回调函数用于记录遍历过程中的错误
	errorCallback := func(path string, err error) {
		receiver.helper.GetLogger().Error(fmt.Sprintf("遍历文件时出错 [路径: %s]: %v", path, err))
	}

	// 使用 WalkFiles 获取所有文件
	files, err := receiver.driver.WalkFiles(rootDir, errorCallback)
	if err != nil {
		receiver.helper.GetLogger().Error(fmt.Sprintf("遍历目录 %s 失败: %v", rootDir, err))
		return
	}

	// 为每个文件注册路由，并收集目录信息
	for _, relativePath := range files {
		receiver.registerFileRoute(rootDir, relativePath)

		// 收集目录信息
		dir := filepath.Dir(relativePath)
		if dir != "." && dir != "" {
			// 提取URL路径（去掉第一级目录，因为 relativePath 是相对于 static 的）
			urlPath := "/" + dir
			receiver.directories[urlPath] = rootDir
		}

		// 如果文件在一级子目录下，也记录该一级目录
		parts := strings.Split(relativePath, "/")
		if len(parts) > 1 {
			firstDir := "/" + parts[0]
			receiver.directories[firstDir] = rootDir
		}
	}

	// 记录注册的路由数量
	receiver.helper.GetLogger().Info(fmt.Sprintf("已为目录 %s 注册 %d 个静态文件路由 (%s驱动)", rootDir, len(files), receiver.driver.GetDriverName()))
}

// registerFileRoute 为单个文件注册路由
// rootDir: 静态资源根目录（如 "static"）
// relativePath: 文件相对于 rootDir 的路径（如 "admin/css/main.css"）
// URL 路径如 /admin/css/main.css
func (receiver *StaticHandler) registerFileRoute(rootDir string, relativePath string) {
	// 构建 URL 路径
	urlPath := "/" + relativePath

	// 获取 MIME 类型
	mimeType := receiver.mimeMapper.GetMimeType(relativePath)

	// 注册 GET 路由
	receiver.engine.GET(urlPath, func(c *gin.Context) {
		// 设置 Content-Type 头
		c.Header("Content-Type", mimeType)

		// 使用 driver 的 ServeFile 方法返回文件内容
		err := receiver.driver.ServeFile(c, rootDir, "/"+relativePath)
		if err != nil {
			c.String(http.StatusNotFound, "Not Found")
			return
		}
	})

	receiver.helper.GetLogger().Debug(fmt.Sprintf("注册静态文件路由: %s (Content-Type: %s)", urlPath, mimeType))
}

// registerDirectoryRoutes 为目录路径注册路由，尝试返回 index.html 或 index.htm
func (receiver *StaticHandler) registerDirectoryRoutes() {
	// 去重后的目录集合
	uniqueDirs := make(map[string]string)
	for urlPath, rootDir := range receiver.directories {
		// 统一移除尾部斜杠
		cleanPath := strings.TrimSuffix(urlPath, "/")
		uniqueDirs[cleanPath] = rootDir
	}

	for urlPath, rootDir := range uniqueDirs {
		// 为不带斜杠的路径注册路由
		receiver.registerDirectoryIndexRoute(urlPath, rootDir)
		// 为带斜杠的路径注册路由
		receiver.registerDirectoryIndexRoute(urlPath+"/", rootDir)

		receiver.helper.GetLogger().Debug(fmt.Sprintf("注册目录路由: %s 和 %s/ -> %s", urlPath, urlPath, rootDir))
	}
}

// registerDirectoryIndexRoute 注册单个目录的 index 路由
func (receiver *StaticHandler) registerDirectoryIndexRoute(urlPath string, rootDir string) {
	// 使用闭包捕获参数，避免变量共享问题
	capturedURLPath := urlPath
	capturedRootDir := rootDir

	receiver.engine.GET(capturedURLPath, func(c *gin.Context) {
		// 提取目录路径（去掉开头的 / 和结尾的 /）
		dirPath := strings.Trim(capturedURLPath, "/")

		// 尝试返回 index.html
		indexHTMLPath := filepath.Join(dirPath, "index.html")
		err := receiver.driver.ServeFile(c, capturedRootDir, "/"+indexHTMLPath)
		if err == nil {
			c.Header("Content-Type", "text/html; charset=utf-8")
			return
		}

		// 尝试返回 index.htm
		indexHTMPath := filepath.Join(dirPath, "index.htm")
		err = receiver.driver.ServeFile(c, capturedRootDir, "/"+indexHTMPath)
		if err == nil {
			c.Header("Content-Type", "text/html; charset=utf-8")
			return
		}

		// 都不存在，返回 404
		c.String(http.StatusNotFound, "Not Found")
	})
}
