package interfaces

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type GetHelperInterface interface {
	GetEnv() EnvInterface
	GetConfig() ConfigInterface
	GetLogger() LoggerInterface
	GetRedis() *redis.Client
	GetDatabase() *gorm.DB
	GetInstalled() Installed
}

type SetHelperInterface interface {
	SetEnv(env EnvInterface)
	SetConfig(config ConfigInterface)
	SetLogger(logger LoggerInterface)
	SetRedis(redis *redis.Client)
	SetDatabase(database *gorm.DB)
	SetInstalled(installed Installed)
}

type HelperInterface interface {
	GetHelperInterface
	SetHelperInterface
}
