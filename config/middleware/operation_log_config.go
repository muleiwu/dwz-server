package middleware

import (
	"cnb.cool/mliev/open/dwz-server/app/middleware"
)

// OperationLogConfig 操作日志配置
type OperationLogConfig struct {
	// 基础配置
	Enable bool `yaml:"enable" json:"enable"` // 是否启用日志记录

	// 大小限制
	MaxRequestSize  int64 `yaml:"max_request_size" json:"max_request_size"`   // 最大请求体大小（字节）
	MaxResponseSize int64 `yaml:"max_response_size" json:"max_response_size"` // 最大响应体大小（字节）

	// 路径过滤
	SkipPaths []string `yaml:"skip_paths" json:"skip_paths"` // 跳过记录的路径（支持正则）

	// 敏感信息处理
	SensitiveFields []string `yaml:"sensitive_fields" json:"sensitive_fields"` // 敏感字段，记录时将被脱敏

	// 记录选项
	LogRequestBody  bool `yaml:"log_request_body" json:"log_request_body"`   // 是否记录请求体
	LogResponseBody bool `yaml:"log_response_body" json:"log_response_body"` // 是否记录响应体
	LogHealthCheck  bool `yaml:"log_health_check" json:"log_health_check"`   // 是否记录健康检查

	// 性能选项
	AsyncLogging bool `yaml:"async_logging" json:"async_logging"` // 是否异步记录日志
}

// GetOperationLogConfig 获取操作日志配置
func GetOperationLogConfig() *middleware.OperationLogConfig {
	// 这里可以从配置文件中读取配置
	// 目前使用默认配置
	return &middleware.OperationLogConfig{
		Enable:          true,
		MaxRequestSize:  1024 * 1024, // 1MB
		MaxResponseSize: 1024 * 1024, // 1MB
		SkipPaths: []string{
			"/health",
			"/api/v1/health",
			"/favicon.ico",
			"/assets/.*",
			"/static/.*",
		},
		SensitiveFields: []string{
			"password", "passwd", "pwd",
			"token", "access_token", "refresh_token",
			"secret", "key", "private_key",
			"authorization", "auth",
		},
		LogRequestBody:  true,
		LogResponseBody: true,
		AsyncLogging:    true,
		LogHealthCheck:  false,
	}
}

// GetDevelopmentConfig 获取开发环境配置
func GetDevelopmentConfig() *middleware.OperationLogConfig {
	config := GetOperationLogConfig()
	// 开发环境可以记录更多信息
	config.LogHealthCheck = true
	config.AsyncLogging = false // 开发环境同步记录便于调试
	return config
}

// GetProductionConfig 获取生产环境配置
func GetProductionConfig() *middleware.OperationLogConfig {
	config := GetOperationLogConfig()
	// 生产环境更注重性能和安全
	config.MaxRequestSize = 512 * 1024  // 512KB
	config.MaxResponseSize = 512 * 1024 // 512KB
	config.AsyncLogging = true
	config.LogHealthCheck = false

	// 生产环境可能需要跳过更多路径
	config.SkipPaths = append(config.SkipPaths,
		"/metrics",
		"/debug/.*",
		"/api/v1/short_links/.*/statistics", // 统计接口可能频繁调用
	)

	return config
}
