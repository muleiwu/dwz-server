package impl

import (
	"strconv"
	"time"
)

type Config struct {
	data map[string]any
}

func NewConfig() *Config {
	return &Config{data: map[string]any{}}
}

func (c *Config) Set(key string, value any) {
	c.data[key] = value
}

func (c *Config) Get(key string, defaultValue any) any {

	data, ok := c.data[key]

	if !ok {
		return defaultValue
	}

	return data
}

func (c *Config) GetBool(key string, defaultValue bool) bool {
	val := c.Get(key, defaultValue)
	if b, ok := val.(bool); ok {
		return b
	}
	return defaultValue
}

func (c *Config) GetInt(key string, defaultValue int) int {
	val := c.Get(key, defaultValue)
	switch v := val.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		tv, err := strconv.Atoi(val.(string))
		if err != nil {
			return defaultValue
		}
		return tv
	default:
		return defaultValue
	}
}

func (c *Config) GetInt32(key string, defaultValue int32) int32 {
	val := c.Get(key, defaultValue)
	switch v := val.(type) {
	case int32:
		return v
	case int:
		return int32(v)
	case int64:
		return int32(v)
	case float64:
		return int32(v)
	default:
		return defaultValue
	}
}

func (c *Config) GetInt64(key string, defaultValue int64) int64 {
	val := c.Get(key, defaultValue)
	switch v := val.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case float64:
		return int64(v)
	default:
		return defaultValue
	}
}

func (c *Config) GetFloat64(key string, defaultValue float64) float64 {
	val := c.Get(key, defaultValue)
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return defaultValue
	}
}

func (c *Config) GetStringSlice(key string, defaultValue []string) []string {
	val := c.Get(key, defaultValue)
	if ss, ok := val.([]string); ok {
		return ss
	}
	return defaultValue
}

func (c *Config) GetString(key string, defaultValue string) string {
	val := c.Get(key, defaultValue)
	if s, ok := val.(string); ok {
		return s
	}
	return defaultValue
}

func (c *Config) GetStringMapString(key string, defaultValue map[string]string) map[string]string {
	val := c.Get(key, defaultValue)
	if sms, ok := val.(map[string]string); ok {
		return sms
	}
	return defaultValue
}

func (c *Config) GetStringMapStringSlice(key string, defaultValue map[string][]string) map[string][]string {
	val := c.Get(key, defaultValue)
	if smss, ok := val.(map[string][]string); ok {
		return smss
	}
	return defaultValue
}

func (c *Config) GetTime(key string, defaultValue time.Time) time.Time {
	val := c.Get(key, defaultValue)
	if t, ok := val.(time.Time); ok {
		return t
	}
	return defaultValue
}
