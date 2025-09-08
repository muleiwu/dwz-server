package interfaces

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisInterface interface {
	// 基础连接方法
	GetClient() *redis.Client
	Ping(ctx context.Context) error
	Close() error

	// 字符串操作
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	GetSet(ctx context.Context, key string, value interface{}) (string, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) (int64, error)
	Exists(ctx context.Context, keys ...string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) (bool, error)
	TTL(ctx context.Context, key string) (time.Duration, error)

	// 列表操作
	LPush(ctx context.Context, key string, values ...interface{}) (int64, error)
	RPush(ctx context.Context, key string, values ...interface{}) (int64, error)
	LPop(ctx context.Context, key string) (string, error)
	RPop(ctx context.Context, key string) (string, error)
	LLen(ctx context.Context, key string) (int64, error)
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)

	// 哈希操作
	HSet(ctx context.Context, key string, values ...interface{}) (int64, error)
	HGet(ctx context.Context, key, field string) (string, error)
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HDel(ctx context.Context, key string, fields ...string) (int64, error)
	HExists(ctx context.Context, key, field string) (bool, error)
	HLen(ctx context.Context, key string) (int64, error)

	// 集合操作
	SAdd(ctx context.Context, key string, members ...interface{}) (int64, error)
	SRem(ctx context.Context, key string, members ...interface{}) (int64, error)
	SMembers(ctx context.Context, key string) ([]string, error)
	SIsMember(ctx context.Context, key string, member interface{}) (bool, error)
	SCard(ctx context.Context, key string) (int64, error)

	// 有序集合操作
	ZAdd(ctx context.Context, key string, members ...redis.Z) (int64, error)
	ZRem(ctx context.Context, key string, members ...interface{}) (int64, error)
	ZRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error)
	ZScore(ctx context.Context, key, member string) (float64, error)
	ZCard(ctx context.Context, key string) (int64, error)

	// 发布订阅
	Publish(ctx context.Context, channel string, message interface{}) (int64, error)
	Subscribe(ctx context.Context, channels ...string) *redis.PubSub

	// 事务
	TxPipeline() redis.Pipeliner
	Pipeline() redis.Pipeliner

	// 分布式锁
	SetNXWithExpire(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)

	// 计数器
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
	IncrBy(ctx context.Context, key string, value int64) (int64, error)
	DecrBy(ctx context.Context, key string, value int64) (int64, error)
}
