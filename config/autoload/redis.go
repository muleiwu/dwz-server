package autoload

import (
	"cnb.cool/mliev/open/go-web/pkg/helper"
)

type Redis struct{}

func (Redis) InitConfig() map[string]any {
	env := helper.GetEnv()
	return map[string]any{
		"redis.host":     env.GetString("redis.host", "localhost"),
		"redis.port":     env.GetInt("redis.port", 6379),
		"redis.password": env.GetString("redis.password", ""),
		"redis.db":       env.GetInt("redis.db", 0),
	}
}
