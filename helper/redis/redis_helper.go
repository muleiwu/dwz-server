package redis

import (
	redis2 "cnb.cool/mliev/open/dwz-server/config/redis"
	"github.com/redis/go-redis/v9"
	"sync"
)

var (
	rdb     *redis.Client
	rdbOnce sync.Once
)

// GetRedis initializes and returns a Redis client.
func GetRedis() *redis.Client {
	rdbOnce.Do(func() {
		redisConfig := redis2.GetRedisConfig()
		rdb = redis.NewClient(redisConfig.GetOptions())
	})

	return rdb
}
