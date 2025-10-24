package middleware

import (
	"net/http"
	"strings"

	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"github.com/gin-gonic/gin"
)

// InstallMiddleware 安装检查中间件
func InstallMiddleware(helper interfaces.HelperInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 如果系统已经安装，继续处理
		if helper.GetInstalled().IsInstalled() {
			c.Next()
			return
		}

		// 如果系统未安装，只允许访问安装相关的页面
		allowedPaths := []string{
			"/install/index",
			"/api/v1/install",
			"/favicon.ico",
			"/health",
			"/health/simple",
		}

		// 检查是否为允许的路径
		for _, allowedPath := range allowedPaths {
			if path == allowedPath || strings.HasPrefix(path, allowedPath) {
				c.Next()
				return
			}
		}

		// 静态资源放行
		if strings.HasPrefix(path, "/static/") ||
			strings.HasPrefix(path, "/assets/") ||
			strings.HasPrefix(path, "/css/") ||
			strings.HasPrefix(path, "/js/") ||
			strings.HasPrefix(path, "/images/") {
			c.Next()
			return
		}

		// 其他路径重定向到安装页面
		if c.Request.Method == "GET" {
			c.Redirect(http.StatusFound, "/install/index")
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"message": "系统尚未安装，请先完成安装",
			})
		}
		c.Abort()
	}
}
