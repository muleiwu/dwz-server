package test

import (
	"cnb.cool/mliev/open/dwz-server/config"
	"sync"
	"testing"
)

// BenchmarkGetEnvWithDefault 测试当前实现的并发性能
func BenchmarkGetEnvWithDefault(b *testing.B) {
	// 确保初始化
	config.InitViper()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			config.GetEnvWithDefault("db.host", "localhost")
		}
	})
}

// BenchmarkGetEnvWithDefaultMultipleKeys 测试多个不同key的并发读取
func BenchmarkGetEnvWithDefaultMultipleKeys(b *testing.B) {
	// 确保初始化
	config.InitViper()

	keys := []string{
		"db.host",
		"db.port",
		"redis.host",
		"redis.port",
		"server.mode",
	}

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := keys[i%len(keys)]
			config.GetEnvWithDefault(key, "default")
			i++
		}
	})
}

// BenchmarkConcurrentFirstAccess 测试并发首次访问性能
func BenchmarkConcurrentFirstAccess(b *testing.B) {
	// 确保初始化
	config.InitViper()

	// 清空缓存，模拟首次访问
	config.ClearCache()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			config.GetEnvWithDefault("test.concurrent.key", "default")
		}
	})
}

// TestConcurrentSafety 测试并发安全性
func TestConcurrentSafety(t *testing.T) {
	config.InitViper()
	config.ClearCache()

	const numGoroutines = 200
	const numOperations = 2000

	var wg sync.WaitGroup

	// 启动多个goroutine并发读取配置
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				val := config.GetEnvWithDefault("db.host", "localhost")
				if val == nil {
					t.Errorf("GetEnvWithDefault returned nil")
					return
				}
			}
		}(i)
	}

	wg.Wait()
}

// TestPreloadingFeature 测试预加载功能
func TestPreloadingFeature(t *testing.T) {
	config.InitViper()

	// 获取所有缓存的键
	keys := config.GetAllCachedKeys()

	t.Logf("预加载了 %d 个配置项", len(keys))

	// 验证基本功能
	val := config.GetEnvWithDefault("server.port", 8080)
	t.Logf("server.port 的值: %v", val)
}
