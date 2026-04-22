package middleware

import (
	"net/http"
	"strconv"
	"strings"

	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
)

// CorsConfig 配置CORS中间件的选项
type CorsConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCorsConfig 返回默认CORS配置
func DefaultCorsConfig() CorsConfig {
	return CorsConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           86400,
	}
}

func CorsMiddleware() httpInterfaces.HandlerFunc {
	return CorsWithConfig(DefaultCorsConfig())
}

func CorsWithConfig(config CorsConfig) httpInterfaces.HandlerFunc {
	return func(c httpInterfaces.RouterContextInterface) {
		origin := c.GetHeader("Origin")

		if len(config.AllowOrigins) == 0 || config.AllowOrigins[0] == "*" {
			c.SetHeader("Access-Control-Allow-Origin", "*")
		} else if origin != "" {
			for _, allowOrigin := range config.AllowOrigins {
				if allowOrigin == origin {
					c.SetHeader("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		c.SetHeader("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
		c.SetHeader("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))

		if len(config.ExposeHeaders) > 0 {
			c.SetHeader("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
		}
		if config.AllowCredentials {
			c.SetHeader("Access-Control-Allow-Credentials", "true")
		}
		if config.MaxAge > 0 {
			c.SetHeader("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
		}

		if c.Method() == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
