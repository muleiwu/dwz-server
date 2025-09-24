package helper

import (
	"sync"

	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var helperOnce sync.Once
var helperData interfaces.HelperInterface

type Helper struct {
	env       interfaces.EnvInterface
	config    interfaces.ConfigInterface
	logger    interfaces.LoggerInterface
	cache     interfaces.ICache
	redis     *redis.Client
	database  *gorm.DB
	installed interfaces.Installed
}

func GetHelper() interfaces.HelperInterface {
	helperOnce.Do(func() {
		helperData = &Helper{}
	})
	return helperData
}

func (receiver *Helper) GetEnv() interfaces.EnvInterface {
	return receiver.env
}

func (receiver *Helper) GetConfig() interfaces.ConfigInterface {
	return receiver.config
}

func (receiver *Helper) GetLogger() interfaces.LoggerInterface {
	return receiver.logger
}

func (receiver *Helper) GetCache() interfaces.ICache {
	return receiver.cache
}

func (receiver *Helper) GetRedis() *redis.Client {
	return receiver.redis
}

func (receiver *Helper) GetDatabase() *gorm.DB {
	return receiver.database
}

func (receiver *Helper) GetInstalled() interfaces.Installed {
	return receiver.installed
}

func (receiver *Helper) SetEnv(env interfaces.EnvInterface) {
	receiver.env = env
}

func (receiver *Helper) SetConfig(config interfaces.ConfigInterface) {
	receiver.config = config
}

func (receiver *Helper) SetLogger(logger interfaces.LoggerInterface) {
	receiver.logger = logger
}

func (receiver *Helper) SetCache(cache interfaces.ICache) {
	receiver.cache = cache
}

func (receiver *Helper) SetRedis(redis *redis.Client) {
	receiver.redis = redis
}

func (receiver *Helper) SetDatabase(database *gorm.DB) {
	receiver.database = database
}

func (receiver *Helper) SetInstalled(installed interfaces.Installed) {
	receiver.installed = installed
}
