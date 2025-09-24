package assembly

import (
	"fmt"
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
	receiver.Helper.GetLogger().Debug("加载缓存驱动" + driver)

	cacheDriver, err := receiver.GetDriver(driver)
	if err != nil {
		fmt.Printf("[cache] 加载缓存驱动失败: %s", err.Error())
	}
	receiver.Helper.SetCache(cacheDriver)

	return nil
}

func (receiver *Cache) GetDriver(driver string) (interfaces.ICache, error) {

	if driver == "redis" {
		return impl.NewCacheRedis(receiver.Helper), nil
	} else {
		// 设置超时时间和清理时间
		c := cache.New(5*time.Minute, 10*time.Minute)
		return impl.NewCacheLocal(receiver.Helper, c), nil
	}
}
