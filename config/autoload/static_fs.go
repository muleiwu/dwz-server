package autoload

import (
	"embed"

	envInterface "cnb.cool/mliev/open/dwz-server/internal/interfaces"
)

type StaticFs struct {
}

func (receiver StaticFs) InitConfig(helper envInterface.GetHelperInterface) map[string]any {
	return map[string]any{
		"static.fs": func() map[string]embed.FS {
			return map[string]embed.FS{
				// 这里会在启动的时候在 main.go 注入静态资源进来，请在main.go 添加静态资源
				//"tempe": embed.FS{},
			}
		},
	}
}
