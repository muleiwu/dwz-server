package autoload

import envInterface "cnb.cool/mliev/open/dwz-server/internal/interfaces"

type IdGenerator struct {
}

func (receiver IdGenerator) InitConfig(helper envInterface.HelperInterface) map[string]any {
	return map[string]any{
		"id_generator.driver": helper.GetEnv().GetString("id_generator.driver", "redis"),
	}
}
