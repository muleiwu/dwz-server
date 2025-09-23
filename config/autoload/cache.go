package autoload

import envInterface "cnb.cool/mliev/open/dwz-server/internal/interfaces"

type Config struct {
}

// cache.driver

func (receiver Config) InitConfig(helper envInterface.HelperInterface) map[string]any {
	return map[string]any{
		"cache.driver": helper.GetEnv().GetString("cache.driver", "redis"),
	}
}
