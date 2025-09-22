package impl

import (
	"context"
	"time"

	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"cnb.cool/mliev/open/dwz-server/internal/pkg/cache/impl/cache_value"
)

type CacheRedis struct {
	helper interfaces.HelperInterface
}

func NewCacheRedis(helper interfaces.HelperInterface) *CacheRedis {
	return &CacheRedis{helper: helper}
}

func (c *CacheRedis) Exists(ctx context.Context, key string) bool {
	exists := c.helper.GetRedis().Exists(ctx, key)

	return exists.Val() != 0
}

func (c *CacheRedis) Get(ctx context.Context, key string, obj any) error {
	cmd := c.helper.GetRedis().Get(ctx, key)

	result, err := cmd.Result()

	if err != nil {
		return err
	}

	err = cache_value.Decode([]byte(result), obj)
	if err != nil {
		return err
	}

	return nil
}

func (c *CacheRedis) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	encode, err := cache_value.Encode(value)
	if err != nil {
		return err
	}
	if ttl <= 0 {
		ttl = 0
	}
	cmd := c.helper.GetRedis().Set(ctx, key, string(encode), ttl)
	return cmd.Err()
}

func (c *CacheRedis) GetSet(ctx context.Context, key string, ttl time.Duration, obj any, fun interfaces.CacheFun) error {

	err := fun(key, obj)
	if err != nil {
		return err
	}

	return c.Set(ctx, key, obj, ttl)
}

func (c *CacheRedis) Del(ctx context.Context, key string) error {
	return c.helper.GetRedis().Del(ctx, key).Err()
}

func (c *CacheRedis) ExpiresAt(ctx context.Context, key string, expiresAt time.Time) error {
	cmd := c.helper.GetRedis().ExpireAt(ctx, key, expiresAt)
	return cmd.Err()
}

func (c *CacheRedis) ExpiresIn(ctx context.Context, key string, ttl time.Duration) error {
	cmd := c.helper.GetRedis().Expire(ctx, key, ttl)
	return cmd.Err()
}
