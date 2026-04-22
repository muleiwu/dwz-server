package autoload

import (
	"cnb.cool/mliev/dwz/dwz-server/app/middleware"
	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
)

type Middleware struct{}

func (Middleware) InitConfig() map[string]any {
	return map[string]any{
		"http.middleware": []httpInterfaces.HandlerFunc{
			middleware.CorsMiddleware(),
			shortCodeDispatch(),
			middleware.InstallMiddleware(),
		},
	}
}
