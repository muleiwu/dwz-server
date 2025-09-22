package assembly

import (
	"time"

	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"cnb.cool/mliev/open/dwz-server/internal/pkg/cache/impl"
	"github.com/patrickmn/go-cache"
)

type Cache struct {
	Helper interfaces.HelperInterface
}

func (receiver *Cache) Assembly() error {

	driver := receiver.Helper.GetConfig().GetString("cache.driver", "redis")

	if driver == "redis" {
		cacheRedis := impl.NewCacheRedis(receiver.Helper)
		receiver.Helper.SetCache(cacheRedis)
	} else {
		// 设置超时时间和清理时间
		c := cache.New(5*time.Minute, 10*time.Minute)
		cacheLocal := impl.NewCacheLocal(receiver.Helper, c)
		receiver.Helper.SetCache(cacheLocal)
	}

	return nil
}
