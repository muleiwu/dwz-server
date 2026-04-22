package assembly

import (
	"reflect"

	installedImpl "cnb.cool/mliev/dwz/dwz-server/pkg/service/installed/impl"
	"cnb.cool/mliev/open/go-web/pkg/container"
	cacheDriver "cnb.cool/mliev/open/go-web/pkg/server/cache/driver"
	"github.com/muleiwu/gsr"
	"github.com/redis/go-redis/v9"
)

// Cache wraps go-web's cache driver. Pre-install we always fall back to the
// in-memory driver so the HTTP server can boot and serve the install page.
// After install completes, a SIGHUP reload picks up the configured driver.
type Cache struct{}

func (Cache) Type() reflect.Type { return reflect.TypeFor[gsr.Cacher]() }
func (Cache) DependsOn() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[gsr.Provider](),
		reflect.TypeFor[gsr.Logger](),
		reflect.TypeFor[*installedImpl.Installed](),
		reflect.TypeFor[*redis.Client](),
	}
}

func (Cache) Assembly() (any, error) {
	logger := container.MustGet[gsr.Logger]()
	installed := container.MustGet[*installedImpl.Installed]()

	driverName := "memory"
	if installed.IsInstalled() {
		driverName = container.MustGet[gsr.Provider]().GetString("cache.driver", "memory")
	}

	var driverConfig any
	if driverName == "redis" {
		client, err := container.Get[*redis.Client]()
		if err != nil || client == nil {
			logger.Warn("[cache] 配置为 redis 但 Redis 不可用，回退到 memory 驱动")
			driverName = "memory"
		} else {
			driverConfig = client
		}
	}

	c, err := cacheDriver.CacheDriverManager.Make(driverName, driverConfig)
	if err != nil {
		logger.Error("[cache] 加载驱动失败，降级到 memory: " + err.Error())
		c, err = cacheDriver.CacheDriverManager.Make("memory", nil)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}
