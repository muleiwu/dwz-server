package interfaces

import (
	"context"
	"time"
)

type CacheFun func(key string, obj any) error

type ICache interface {
	Exists(ctx context.Context, key string) bool
	Get(ctx context.Context, key string, obj any) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	GetSet(ctx context.Context, key string, ttl time.Duration, obj any, fun CacheFun) error
	Del(ctx context.Context, key string) error
	ExpiresAt(ctx context.Context, key string, expiresAt time.Time) error
	ExpiresIn(ctx context.Context, key string, ttl time.Duration) error
}
