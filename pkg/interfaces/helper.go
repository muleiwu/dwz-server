package interfaces

import (
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
}
