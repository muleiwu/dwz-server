package assembly

import (
	"errors"
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

	if receiver.Helper.GetInstalled() == nil || !receiver.Helper.GetInstalled().IsInstalled() {
		receiver.Helper.GetLogger().Warn("应用未安装，停止初始化缓存")
		return nil
	}

	driver := receiver.Helper.GetConfig().GetString("cache.driver", "redis")
	receiver.Helper.GetLogger().Debug("加载缓存驱动" + driver)

	if driver == "redis" && receiver.Helper.GetRedis() == nil {
		panic(errors.New("缓存服务驱动配置为：redis，但Redis服务不可用，拒绝启动"))
	}

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
