package impl

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"cnb.cool/mliev/open/dwz-server/internal/helper"
	"github.com/spf13/viper"
)

type Env struct {
	cache       sync.Map  // 线程安全的并发map
	initialized int64     // 原子操作的初始化标志
	initOnce    sync.Once // 确保只初始化一次
	initError   error     // 初始化错误
	Helper      *helper.Helper
}

func NewEnv() *Env {
	e := Env{}
	e.initOnce.Do(func() {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")

		// 支持读取环境变量
		replace := strings.NewReplacer(".", "_") // 替换点为下划线
		viper.SetEnvKeyReplacer(replace)         // 设置环境变量的替换器
		viper.AutomaticEnv()

		// 尝试读取配置文件
		if err := viper.ReadInConfig(); err != nil {
			// 如果配置文件不存在，只记录日志但不返回错误
			fmt.Printf("警告: 配置文件未找到，将使用默认配置和环境变量: %v\n", err)
		}

		// 预加载所有配置到缓存
		e.preloadAllConfigs()

		// 原子设置初始化完成标志
		atomic.StoreInt64(&e.initialized, 1)
	})

	return &e
}

func (receiver *Env) preloadAllConfigs() {
	allSettings := viper.AllSettings()
	receiver.flattenAndCache("", allSettings)
}

// flattenAndCache 递归扁平化配置并缓存
func (receiver *Env) flattenAndCache(prefix string, settings map[string]interface{}) {
	for key, value := range settings {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]interface{}:
			// 递归处理嵌套配置
			receiver.flattenAndCache(fullKey, v)
		case map[interface{}]interface{}:
			// 处理viper可能返回的map[interface{}]interface{}类型
			convertedMap := make(map[string]interface{})
			for k, val := range v {
				if keyStr, ok := k.(string); ok {
					convertedMap[keyStr] = val
				}
			}
			receiver.flattenAndCache(fullKey, convertedMap)
		default:
			// 缓存配置值
			receiver.cache.Store(fullKey, value)
		}
	}
}

// GetEnvWithDefault 获取环境变量，如果不存在则返回默认值
func (receiver *Env) GetEnvWithDefault(name string, def any) any {

	// 从sync.Map缓存中获取，无锁操作
	if val, ok := receiver.cache.Load(name); ok {
		return val
	}

	// 缓存未命中时，从viper获取并缓存
	val := viper.Get(name)
	if val == nil {
		val = def
	}

	// 存储到缓存，sync.Map内部处理并发安全
	receiver.cache.Store(name, val)

	return val
}

func (receiver *Env) Get(key string, defaultValue any) any {
	return receiver.GetEnvWithDefault(key, defaultValue)
}

func (receiver *Env) GetBool(key string, defaultValue bool) bool {
	val := receiver.Get(key, defaultValue)
	if str, ok := val.(bool); ok {
		return str
	} else if str, ok := val.(string); ok {
		if strings.ToLower(str) == "true" {
			return true
		} else if strings.ToLower(str) == "1" {
			return true
		} else if strings.ToLower(str) == "false" {
			return false
		} else if strings.ToLower(str) == "0" {
			return false
		}
		return defaultValue
	} else if str, ok := val.(int); ok {
		if str == 0 {
			return false
		} else if str == 1 {
			return true
		}
		return defaultValue
	}
	return defaultValue
}

func (receiver *Env) GetInt(key string, defaultValue int) int {
	val := receiver.Get(key, defaultValue)
	if str, ok := val.(int); ok {
		return str
	}
	return defaultValue
}

func (receiver *Env) GetInt32(key string, defaultValue int32) int32 {
	val := receiver.Get(key, defaultValue)
	if str, ok := val.(int32); ok {
		return str
	}
	return defaultValue
}

func (receiver *Env) GetInt64(key string, defaultValue int64) int64 {
	val := receiver.Get(key, defaultValue)
	if str, ok := val.(int64); ok {
		return str
	}
	return defaultValue
}

func (receiver *Env) GetFloat64(key string, defaultValue float64) float64 {
	val := receiver.Get(key, defaultValue)
	if str, ok := val.(float64); ok {
		return str
	}
	return defaultValue
}

func (receiver *Env) GetStringSlice(key string, defaultValue []string) []string {
	val := receiver.Get(key, defaultValue)
	if str, ok := val.([]string); ok {
		return str
	}
	return defaultValue
}

func (receiver *Env) GetString(key string, defaultValue string) string {
	val := receiver.Get(key, defaultValue)
	if str, ok := val.(string); ok {
		return str
	}
	return defaultValue
}

func (receiver *Env) GetStringMapString(key string, defaultValue map[string]string) map[string]string {
	val := receiver.Get(key, defaultValue)
	if str, ok := val.(map[string]string); ok {
		return str
	}
	return defaultValue
}

func (receiver *Env) GetStringMapStringSlice(key string, defaultValue map[string][]string) map[string][]string {
	val := receiver.Get(key, defaultValue)
	if str, ok := val.(map[string][]string); ok {
		return str
	}
	return defaultValue
}

func (receiver *Env) GetTime(key string, defaultValue time.Time) time.Time {
	val := receiver.Get(key, defaultValue)
	if str, ok := val.(time.Time); ok {
		return str
	}
	return defaultValue
}
