package middleware

import (
	"cnb.cool/mliev/open/dwz-server/app/middleware"
	"github.com/gin-gonic/gin"
)

type MiddlewareConfig struct {
}

func (m MiddlewareConfig) Get() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.CorsMiddleware(),
	}
}
