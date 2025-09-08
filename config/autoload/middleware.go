package autoload

import (
	"cnb.cool/mliev/open/dwz-server/app/middleware"
	envInterface "cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"github.com/gin-gonic/gin"
)

type Middleware struct {
}

func (receiver Middleware) InitConfig(helper envInterface.GetHelperInterface) map[string]any {
	return map[string]any{
		"http.middleware": []gin.HandlerFunc{
			middleware.CorsMiddleware(helper),
		},
	}
}
