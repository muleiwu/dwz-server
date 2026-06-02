package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	ceInterfaces "cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
	ipRegionImpl "cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/ip_region/impl"
	"github.com/muleiwu/gsr"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUpdateUserSystemAdminPermissionAndLastAdminProtection(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	helper := &userServiceTestHelper{db: db, settings: userServiceTestSettings{}}
	svc := NewUserService(helper)

	admin := &model.User{Username: "admin", Email: "admin@example.com", Status: 1, IsSystemAdmin: true}
	if err := db.Create(admin).Error; err != nil {
		t.Fatalf("create admin: %v", err)
	}
	regular := &model.User{Username: "regular", Email: "regular@example.com", Status: 1}
	if err := db.Create(regular).Error; err != nil {
		t.Fatalf("create regular: %v", err)
	}

	enable := true
	if _, err := svc.UpdateUser(regular.ID, &dto.UpdateUserRequest{IsSystemAdmin: &enable}, false); err == nil || !strings.Contains(err.Error(), "无权限") {
		t.Fatalf("expected non-system admin grant to fail, got %v", err)
	}

	disable := false
	if _, err := svc.UpdateUser(admin.ID, &dto.UpdateUserRequest{IsSystemAdmin: &disable}, true); err == nil || !strings.Contains(err.Error(), "至少保留") {
		t.Fatalf("expected last system admin revoke to fail, got %v", err)
	}

	otherAdmin := &model.User{Username: "other-admin", Email: "other-admin@example.com", Status: 1, IsSystemAdmin: true}
	if err := db.Create(otherAdmin).Error; err != nil {
		t.Fatalf("create other admin: %v", err)
	}
	if _, err := svc.UpdateUser(admin.ID, &dto.UpdateUserRequest{IsSystemAdmin: &disable}, true); err != nil {
		t.Fatalf("expected revoke with another system admin to succeed: %v", err)
	}
}

type userServiceTestHelper struct {
	db       *gorm.DB
	settings userServiceTestSettings
}

func (h *userServiceTestHelper) GetEnv() ceInterfaces.EnvInterface       { return h.settings }
func (h *userServiceTestHelper) GetConfig() ceInterfaces.ConfigInterface { return h.settings }
func (h *userServiceTestHelper) GetLogger() ceInterfaces.LoggerInterface {
	return userServiceTestLogger{}
}
func (h *userServiceTestHelper) GetCache() gsr.Cacher                 { return userServiceTestCache{} }
func (h *userServiceTestHelper) GetRedis() *redis.Client              { return nil }
func (h *userServiceTestHelper) GetDatabase() *gorm.DB                { return h.db }
func (h *userServiceTestHelper) GetInstalled() ceInterfaces.Installed { return nil }
func (h *userServiceTestHelper) GetVersion() ceInterfaces.VersionInterface {
	return nil
}
func (h *userServiceTestHelper) GetIPRegion() ipRegionImpl.IPRegion { return ipRegionImpl.Noop{} }

type userServiceTestSettings map[string]any

func (s userServiceTestSettings) Get(key string, defaultValue any) any {
	if value, ok := s[key]; ok {
		return value
	}
	return defaultValue
}
func (s userServiceTestSettings) GetBool(key string, defaultValue bool) bool {
	if value, ok := s[key].(bool); ok {
		return value
	}
	return defaultValue
}
func (s userServiceTestSettings) GetInt(key string, defaultValue int) int {
	if value, ok := s[key].(int); ok {
		return value
	}
	return defaultValue
}
func (s userServiceTestSettings) GetInt32(key string, defaultValue int32) int32 {
	if value, ok := s[key].(int32); ok {
		return value
	}
	return defaultValue
}
func (s userServiceTestSettings) GetInt64(key string, defaultValue int64) int64 {
	if value, ok := s[key].(int64); ok {
		return value
	}
	return defaultValue
}
func (s userServiceTestSettings) GetFloat64(key string, defaultValue float64) float64 {
	if value, ok := s[key].(float64); ok {
		return value
	}
	return defaultValue
}
func (s userServiceTestSettings) GetStringSlice(key string, defaultValue []string) []string {
	if value, ok := s[key].([]string); ok {
		return value
	}
	return defaultValue
}
func (s userServiceTestSettings) GetString(key string, defaultValue string) string {
	if value, ok := s[key].(string); ok {
		return value
	}
	return defaultValue
}
func (s userServiceTestSettings) GetStringMapString(key string, defaultValue map[string]string) map[string]string {
	if value, ok := s[key].(map[string]string); ok {
		return value
	}
	return defaultValue
}
func (s userServiceTestSettings) GetStringMapStringSlice(key string, defaultValue map[string][]string) map[string][]string {
	if value, ok := s[key].(map[string][]string); ok {
		return value
	}
	return defaultValue
}
func (s userServiceTestSettings) GetTime(key string, defaultValue time.Time) time.Time {
	if value, ok := s[key].(time.Time); ok {
		return value
	}
	return defaultValue
}
func (s userServiceTestSettings) Set(key string, value any) { s[key] = value }

type userServiceTestLogger struct{}

func (userServiceTestLogger) Debug(string, ...gsr.LoggerField)  {}
func (userServiceTestLogger) Info(string, ...gsr.LoggerField)   {}
func (userServiceTestLogger) Notice(string, ...gsr.LoggerField) {}
func (userServiceTestLogger) Error(string, ...gsr.LoggerField)  {}
func (userServiceTestLogger) Warn(string, ...gsr.LoggerField)   {}
func (userServiceTestLogger) Fatal(string, ...gsr.LoggerField)  {}

type userServiceTestCache struct{}

func (userServiceTestCache) Exists(context.Context, string) bool { return false }
func (userServiceTestCache) Get(context.Context, string, any) error {
	return errors.New("cache miss")
}
func (userServiceTestCache) Set(context.Context, string, any, time.Duration) error { return nil }
func (userServiceTestCache) GetSet(_ context.Context, key string, _ time.Duration, obj any, callback gsr.CacheCallback) error {
	return callback(key, obj)
}
func (userServiceTestCache) Del(context.Context, string) error { return nil }
func (userServiceTestCache) ExpiresAt(context.Context, string, time.Time) error {
	return nil
}
func (userServiceTestCache) ExpiresIn(context.Context, string, time.Duration) error {
	return nil
}
