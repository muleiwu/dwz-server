package interfaces

import (
	"time"
)

type ConfigInterface interface {
	Set(key string, value any)

	Get(key string, defaultValue any) any
	GetBool(key string, defaultValue bool) bool
	GetInt(key string, defaultValue int) int
	GetInt32(key string, defaultValue int32) int32
	GetInt64(key string, defaultValue int64) int64
	GetFloat64(key string, defaultValue float64) float64
	GetStringSlice(key string, defaultValue []string) []string
	GetString(key string, defaultValue string) string
	GetStringMapString(key string, defaultValue map[string]string) map[string]string
	GetStringMapStringSlice(key string, defaultValue map[string][]string) map[string][]string
	GetTime(key string, defaultValue time.Time) time.Time
}

type InitConfig interface {
	InitConfig(helper GetHelperInterface) map[string]any
}
