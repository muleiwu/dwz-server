package middleware

import (
	"net/http"
	"strings"

	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
)

func InstallMiddleware() httpInterfaces.HandlerFunc {
	return func(c httpInterfaces.RouterContextInterface) {
		if installed := helper.GetHelper().GetInstalled(); installed != nil && installed.IsInstalled() {
			c.Next()
			return
		}

		path := c.Path()

		allowedPaths := []string{
			"/install/index",
			"/api/v1/install",
			"/favicon.ico",
			"/health",
			"/health/simple",
		}
		for _, allowedPath := range allowedPaths {
			if path == allowedPath || strings.HasPrefix(path, allowedPath) {
				c.Next()
				return
			}
		}

		if strings.HasPrefix(path, "/static/") ||
			strings.HasPrefix(path, "/assets/") ||
			strings.HasPrefix(path, "/css/") ||
			strings.HasPrefix(path, "/js/") ||
			strings.HasPrefix(path, "/images/") {
			c.Next()
			return
		}

		if c.Method() == http.MethodGet {
			c.Redirect(http.StatusFound, "/install/index")
		} else {
			c.JSON(http.StatusServiceUnavailable, map[string]any{
				"code":    503,
				"message": "系统尚未安装，请先完成安装",
			})
		}
		c.Abort()
	}
}
