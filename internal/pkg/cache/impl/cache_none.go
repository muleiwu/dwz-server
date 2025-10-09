package impl

import (
	"context"
	"errors"
	"time"

	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
)

type CacheNone struct {
	helper interfaces.HelperInterface
}

func NewCacheNone(helper interfaces.HelperInterface) *CacheNone {
	return &CacheNone{helper: helper}
}

func (c *CacheNone) Exists(ctx context.Context, key string) bool {
	return false
}

func (c *CacheNone) Get(ctx context.Context, key string, obj any) error {
	return errors.New("not implemented")
}

func (c *CacheNone) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return nil
}

func (c *CacheNone) GetSet(ctx context.Context, key string, ttl time.Duration, obj any, fun interfaces.CacheFun) error {
	return errors.New("not implemented")
}

func (c *CacheNone) Del(ctx context.Context, key string) error {
	return nil
}

func (c *CacheNone) ExpiresAt(ctx context.Context, key string, expiresAt time.Time) error {
	return nil
}

func (c *CacheNone) ExpiresIn(ctx context.Context, key string, ttl time.Duration) error {
	return nil
}
