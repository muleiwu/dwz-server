package helper

import (
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
	installedImpl "cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/installed/impl"
	versionImpl "cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/version/impl"
	"cnb.cool/mliev/open/go-web/pkg/container"
	gowebHelper "cnb.cool/mliev/open/go-web/pkg/helper"
	"github.com/muleiwu/gsr"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Helper proxies the legacy dwz HelperInterface to go-web's container. All
// Get* methods resolve via container.MustGet[T](). The container's lifecycle
// is owned by go-web's cmd.Start; runtime swaps go through SIGHUP reload.
type Helper struct{}

var helperInstance interfaces.HelperInterface = &Helper{}

func GetHelper() interfaces.HelperInterface { return helperInstance }

func (Helper) GetEnv() interfaces.EnvInterface       { return gowebHelper.GetEnv() }
func (Helper) GetConfig() interfaces.ConfigInterface { return gowebHelper.GetConfig() }
func (Helper) GetLogger() interfaces.LoggerInterface { return gowebHelper.GetLogger() }
func (Helper) GetCache() gsr.Cacher                  { return gowebHelper.GetCache() }

func (Helper) GetRedis() *redis.Client {
	if c, err := container.Get[*redis.Client](); err == nil {
		return c
	}
	return nil
}

func (Helper) GetDatabase() *gorm.DB {
	if db, err := container.Get[*gorm.DB](); err == nil {
		return db
	}
	return nil
}

func (Helper) GetInstalled() interfaces.Installed {
	if v, err := container.Get[*installedImpl.Installed](); err == nil {
		return v
	}
	return nil
}

func (Helper) GetVersion() interfaces.VersionInterface {
	if v, err := container.Get[*versionImpl.Version](); err == nil {
		return v
	}
	return nil
}
