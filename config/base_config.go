package config

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/spf13/viper"
)

// Config 高性能配置管理器
type Config struct {
	cache       sync.Map  // 线程安全的并发map
	initialized int64     // 原子操作的初始化标志
	initOnce    sync.Once // 确保只初始化一次
	initError   error     // 初始化错误
}

// 全局实例
var globalConfig = &Config{}

// InitViper 初始化配置，返回错误以便调用方处理
func InitViper() error {
	globalConfig.initOnce.Do(func() {
		viper.SetConfigName("config")
		viper.SetConfigType("yml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")

		// 支持读取环境变量
		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

		// 尝试读取配置文件
		if err := viper.ReadInConfig(); err != nil {
			// 如果配置文件不存在，只记录日志但不返回错误
			fmt.Printf("警告: 配置文件未找到，将使用默认配置和环境变量: %v\n", err)
		}

		// 预加载所有配置到缓存
		globalConfig.preloadAllConfigs()

		// 原子设置初始化完成标志
		atomic.StoreInt64(&globalConfig.initialized, 1)
	})

	return globalConfig.initError
}

// InitViper 初始化配置，返回错误以便调用方处理
func ReloadViper() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// 支持读取环境变量
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 如果配置文件不存在，只记录日志但不返回错误
		fmt.Printf("警告: 配置文件未找到，将使用默认配置和环境变量: %v\n", err)
	}

	// 预加载所有配置到缓存
	globalConfig.preloadAllConfigs()

	// 原子设置初始化完成标志
	atomic.StoreInt64(&globalConfig.initialized, 1)

	return globalConfig.initError
}

// preloadAllConfigs 预加载所有配置项到缓存
func (c *Config) preloadAllConfigs() {
	allSettings := viper.AllSettings()
	c.flattenAndCache("", allSettings)
}

// flattenAndCache 递归扁平化配置并缓存
func (c *Config) flattenAndCache(prefix string, settings map[string]interface{}) {
	for key, value := range settings {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]interface{}:
			// 递归处理嵌套配置
			c.flattenAndCache(fullKey, v)
		case map[interface{}]interface{}:
			// 处理viper可能返回的map[interface{}]interface{}类型
			convertedMap := make(map[string]interface{})
			for k, val := range v {
				if keyStr, ok := k.(string); ok {
					convertedMap[keyStr] = val
				}
			}
			c.flattenAndCache(fullKey, convertedMap)
		default:
			// 缓存配置值
			c.cache.Store(fullKey, value)
		}
	}
}

// GetEnvWithDefault 获取环境变量，如果不存在则返回默认值
func GetEnvWithDefault(name string, def any) any {
	// 使用原子操作检查初始化状态
	if atomic.LoadInt64(&globalConfig.initialized) == 0 {
		if err := InitViper(); err != nil {
			return def
		}
	}

	// 从sync.Map缓存中获取，无锁操作
	if val, ok := globalConfig.cache.Load(name); ok {
		return val
	}

	// 缓存未命中时，从viper获取并缓存
	val := viper.Get(name)
	if val == nil {
		val = def
	}

	// 存储到缓存，sync.Map内部处理并发安全
	globalConfig.cache.Store(name, val)

	return val
}

// GetString 获取字符串配置值
func GetString(key string, defaultValue string) string {
	val := GetEnvWithDefault(key, defaultValue)
	if str, ok := val.(string); ok {
		return str
	}
	return defaultValue
}

// GetInt 获取整数配置值
func GetInt(key string, defaultValue int) int {
	val := GetEnvWithDefault(key, defaultValue)
	if i, ok := val.(string); ok {
		num, err := strconv.Atoi(i)
		if err != nil {
			return defaultValue
		}
		return num
	}
	if i, ok := val.(int); ok {
		return i
	}
	return defaultValue
}

// GetBool 获取布尔配置值
func GetBool(key string, defaultValue bool) bool {
	val := GetEnvWithDefault(key, defaultValue)
	if b, ok := val.(bool); ok {
		return b
	}
	return defaultValue
}

// IsInitialized 检查配置是否已初始化
func IsInitialized() bool {
	return atomic.LoadInt64(&globalConfig.initialized) == 1
}

// ClearCache 清空配置缓存（用于测试或重新加载配置）
func ClearCache() {
	globalConfig.cache.Range(func(key, value interface{}) bool {
		globalConfig.cache.Delete(key)
		return true
	})
	atomic.StoreInt64(&globalConfig.initialized, 0)
}

// GetAllCachedKeys 获取所有已缓存的配置键（用于调试）
func GetAllCachedKeys() []string {
	keys := make([]string, 0)
	globalConfig.cache.Range(func(key, value interface{}) bool {
		if keyStr, ok := key.(string); ok {
			keys = append(keys, keyStr)
		}
		return true
	})
	return keys
}
