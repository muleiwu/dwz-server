package impl

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"github.com/patrickmn/go-cache"
)

type CacheMemory struct {
	helper interfaces.HelperInterface
	cache  *cache.Cache
}

func NewCacheLocal(helper interfaces.HelperInterface, c *cache.Cache) *CacheMemory {
	return NewCacheMemory(helper, c)
}

func NewCacheMemory(helper interfaces.HelperInterface, c *cache.Cache) *CacheMemory {
	return &CacheMemory{helper: helper, cache: c}
}

func (c *CacheMemory) Exists(ctx context.Context, key string) bool {
	_, b := c.cache.Get(key)
	return b
}

func (c *CacheMemory) Get(ctx context.Context, key string, obj any) error {
	val, b := c.cache.Get(key)
	if !b {
		return errors.New("key not exists")
	}
	return c.assignValue(obj, val)
}

func (c *CacheMemory) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = -1
	}
	c.cache.Set(key, value, ttl)
	return nil
}

func (c *CacheMemory) GetSet(ctx context.Context, key string, ttl time.Duration, obj any, fun interfaces.CacheFun) error {

	err := fun(key, obj)
	if err != nil {
		return err
	}

	return c.Set(ctx, key, obj, ttl)
}

func (c *CacheMemory) Del(ctx context.Context, key string) error {
	c.cache.Delete(key)
	return nil
}

func (c *CacheMemory) ExpiresAt(ctx context.Context, key string, expiresAt time.Time) error {
	var obj any
	err := c.Get(ctx, key, &obj)
	if err != nil {
		return err
	}

	now := time.Now()

	c.cache.Set(key, obj, now.Sub(expiresAt))

	return nil
}

func (c *CacheMemory) ExpiresIn(ctx context.Context, key string, ttl time.Duration) error {
	var obj any
	err := c.Get(ctx, key, &obj)
	if err != nil {
		return err
	}

	c.cache.Set(key, obj, ttl)

	return nil
}

// assignValue 使用反射将值赋给目标对象
func (c *CacheMemory) assignValue(obj any, value interface{}) error {
	if obj == nil {
		return fmt.Errorf("obj cannot be nil")
	}

	objValue := reflect.ValueOf(obj)
	if objValue.Kind() != reflect.Ptr {
		return fmt.Errorf("obj must be a pointer")
	}

	objElem := objValue.Elem()
	if !objElem.CanSet() {
		return fmt.Errorf("obj cannot be set")
	}

	valueReflect := reflect.ValueOf(value)
	if !valueReflect.IsValid() {
		return fmt.Errorf("value is not valid")
	}

	// 确保类型匹配
	if objElem.Type() != valueReflect.Type() {
		return fmt.Errorf("type mismatch: expected %s, got %s", objElem.Type(), valueReflect.Type())
	}

	objElem.Set(valueReflect)
	return nil
}
