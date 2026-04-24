package interfaces

import (
	ipRegionImpl "cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/ip_region/impl"
	"github.com/muleiwu/gsr"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// HelperInterface is the legacy aggregate handle that controllers, services
// and DAOs receive via constructor injection. Methods resolve to go-web's
// container-managed singletons.
type HelperInterface interface {
	GetEnv() EnvInterface
	GetConfig() ConfigInterface
	GetLogger() LoggerInterface
	GetCache() gsr.Cacher
	GetRedis() *redis.Client
	GetDatabase() *gorm.DB
	GetInstalled() Installed
	GetVersion() VersionInterface
	// GetIPRegion 返回 IP 归属地查询器；未配置或加载失败时返回 Noop 实现，
	// 调用方可直接 Lookup 而无需判空。
	GetIPRegion() ipRegionImpl.IPRegion
}
