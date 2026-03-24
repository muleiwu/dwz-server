package autoload

import envInterface "cnb.cool/mliev/dwz/dwz-server/pkg/interfaces"

type Base struct {
}

func (receiver Base) InitConfig(helper envInterface.HelperInterface) map[string]any {
	return map[string]any{
		"app.base.app_name": helper.GetEnv().GetString("app.base.app_name", "go-web-app"),
	}
}
