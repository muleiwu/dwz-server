package autoload

import (
	envInterface "cnb.cool/mliev/open/dwz-server/internal/interfaces"
)

type Redis struct {
	env envInterface.EnvInterface
}

func (receiver Redis) InitConfig(helper envInterface.HelperInterface) map[string]any {
	return map[string]any{
		"redis.host":     helper.GetEnv().GetString("redis.host", "localhost"),
		"redis.port":     helper.GetEnv().GetInt("redis.port", 6379),
		"redis.password": helper.GetEnv().GetString("redis.password", ""),
		"redis.db":       helper.GetEnv().GetInt("redis.db", 0),
	}
}
