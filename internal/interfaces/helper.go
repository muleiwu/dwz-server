package interfaces

import (
	"github.com/muleiwu/gsr"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type GetHelperInterface interface {
	GetEnv() EnvInterface
	GetConfig() ConfigInterface
	GetLogger() LoggerInterface
	GetCache() gsr.Cacher
	GetRedis() *redis.Client
	GetDatabase() *gorm.DB
	GetInstalled() Installed
	GetVersion() VersionInterface
}

type SetHelperInterface interface {
	SetEnv(env EnvInterface)
	SetConfig(config ConfigInterface)
	SetLogger(logger LoggerInterface)
	SetCache(cache gsr.Cacher)
	SetRedis(redis *redis.Client)
	SetDatabase(database *gorm.DB)
	SetInstalled(installed Installed)
	SetVersion(version VersionInterface)
}

type HelperInterface interface {
	GetHelperInterface
	SetHelperInterface
}
