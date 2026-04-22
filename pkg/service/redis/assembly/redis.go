package assembly

import (
	"reflect"

	installedImpl "cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/installed/impl"
	"cnb.cool/mliev/open/go-web/pkg/container"
	redisConfig "cnb.cool/mliev/open/go-web/pkg/server/redis/config"
	redisDriver "cnb.cool/mliev/open/go-web/pkg/server/redis/driver"
	"github.com/muleiwu/gsr"
	"github.com/redis/go-redis/v9"
)

// Redis wraps go-web's redis driver, gated on install state so the boot
// path doesn't fail when redis is not yet configured.
type Redis struct{}

func (Redis) Type() reflect.Type { return reflect.TypeFor[*redis.Client]() }
func (Redis) DependsOn() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[gsr.Provider](),
		reflect.TypeFor[gsr.Logger](),
		reflect.TypeFor[*installedImpl.Installed](),
	}
}

func (Redis) Assembly() (any, error) {
	logger := container.MustGet[gsr.Logger]()
	installed := container.MustGet[*installedImpl.Installed]()
	if !installed.IsInstalled() {
		logger.Notice("[redis] 应用未安装，跳过 Redis 连接")
		return (*redis.Client)(nil), nil
	}

	cfg := container.MustGet[gsr.Provider]()
	rc := redisConfig.NewRedis(cfg)
	client, err := redisDriver.RedisDriverManager.Make("redis", rc)
	if err != nil {
		logger.Error("[redis] Redis 连接失败: " + err.Error())
		return (*redis.Client)(nil), nil
	}
	return client, nil
}
