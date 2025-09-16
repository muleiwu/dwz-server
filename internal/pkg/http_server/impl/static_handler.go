package impl

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"github.com/gin-gonic/gin"
)

type StaticHandler struct {
	helper interfaces.HelperInterface
	engine *gin.Engine
}

func NewStaticHandler(helper interfaces.HelperInterface, engine *gin.Engine) *StaticHandler {
	return &StaticHandler{
		helper: helper,
		engine: engine,
	}
}

// setupStaticFileServers 为嵌入的静态文件设置HTTP服务
func (receiver *StaticHandler) setupStaticFileServers() {

	staticFs := receiver.helper.GetConfig().Get("static.fs", map[string]embed.FS{}).(map[string]embed.FS)

	if !receiver.helper.GetConfig().GetBool("http.load_static", false) {
		return
	}

	staticDirSlice := receiver.helper.GetConfig().GetStringSlice("http.static_dir", []string{})

	if len(staticDirSlice) == 0 {
		receiver.helper.GetLogger().Warn("没有配置需要加载的目录")
		return
	}

	webStaticFs, ok := staticFs["web.static"]

	if !ok {
		receiver.helper.GetLogger().Debug("不存在需要对Web暴露的静态资源")
		return
	}

	for i, dir := range staticDirSlice {
		receiver.loadStatic(webStaticFs, dir)
		receiver.helper.GetLogger().Debug(fmt.Sprintf("序号：%d 加载文件夹：%s", i, dir))
	}

}

func (receiver *StaticHandler) loadStatic(webStaticFs embed.FS, dir string) {

	relativePath := fmt.Sprintf("/%s", dir)
	subFS, err := fs.Sub(webStaticFs, fmt.Sprintf("static/%s", dir))
	if err == nil {
		// 将Vue前端文件映射到根路径
		receiver.engine.StaticFS(relativePath, http.FS(subFS))
		receiver.helper.GetLogger().Info(fmt.Sprintf("已启用 %s 静态文件服务", relativePath))
		// 处理Vue的单页面应用请求
		receiver.engine.NoRoute(func(c *gin.Context) {
			if c.Request.Method == "GET" && !receiver.excludeStaticPath(c.Request.URL.Path) {
				c.FileFromFS("/index.html", http.FS(subFS))
			}
		})
	} else {
		receiver.helper.GetLogger().Error(fmt.Sprintf("设置%s静态文件失败: %v", relativePath, err))
	}
}

// excludeStaticPath 判断是否排除某些路径不重定向到单页应用
func (receiver *StaticHandler) excludeStaticPath(path string) bool {
	// 排除API和admin路径
	prefixes := []string{"/admin/"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	// 排除已知的静态文件扩展名
	extensions := []string{".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", ".ttf", ".eot"}
	for _, ext := range extensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}

	return false
}
