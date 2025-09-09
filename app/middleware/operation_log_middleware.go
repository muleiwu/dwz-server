package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"cnb.cool/mliev/open/dwz-server/app/service"
	envInterface "cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"github.com/gin-gonic/gin"
)

// OperationLogConfig 操作日志配置
type OperationLogConfig struct {
	Enable          bool     `yaml:"enable"`            // 是否启用日志记录
	MaxRequestSize  int64    `yaml:"max_request_size"`  // 最大请求体大小（字节），超过则不记录
	MaxResponseSize int64    `yaml:"max_response_size"` // 最大响应体大小（字节），超过则不记录
	SkipPaths       []string `yaml:"skip_paths"`        // 跳过记录的路径（支持正则）
	SensitiveFields []string `yaml:"sensitive_fields"`  // 敏感字段，记录时将被脱敏
	LogRequestBody  bool     `yaml:"log_request_body"`  // 是否记录请求体
	LogResponseBody bool     `yaml:"log_response_body"` // 是否记录响应体
	AsyncLogging    bool     `yaml:"async_logging"`     // 是否异步记录日志
	LogHealthCheck  bool     `yaml:"log_health_check"`  // 是否记录健康检查
}

// DefaultOperationLogConfig 默认配置
var DefaultOperationLogConfig = OperationLogConfig{
	Enable:          true,
	MaxRequestSize:  1024 * 1024, // 1MB
	MaxResponseSize: 1024 * 1024, // 1MB
	SkipPaths:       []string{},
	SensitiveFields: []string{"password", "token", "secret", "key", "passwd"},
	LogRequestBody:  true,
	LogResponseBody: true,
	AsyncLogging:    true,
	LogHealthCheck:  false,
}

type responseWriter struct {
	gin.ResponseWriter
	body    *bytes.Buffer
	config  *OperationLogConfig
	written bool
}

func (w *responseWriter) Write(b []byte) (int, error) {
	// 检查是否需要记录响应体以及大小限制
	if w.config.LogResponseBody && w.body.Len() < int(w.config.MaxResponseSize) {
		w.body.Write(b)
	}
	w.written = true
	return w.ResponseWriter.Write(b)
}

// OperationLogMiddleware 操作日志记录中间件
func OperationLogMiddleware(helper envInterface.HelperInterface, config ...*OperationLogConfig) gin.HandlerFunc {
	// 使用传入的配置或默认配置
	cfg := &DefaultOperationLogConfig
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	}

	// 如果未启用，返回空中间件
	if !cfg.Enable {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// 编译跳过路径的正则表达式
	skipPatterns := make([]*regexp.Regexp, 0, len(cfg.SkipPaths))
	for _, pattern := range cfg.SkipPaths {
		if re, err := regexp.Compile(pattern); err == nil {
			skipPatterns = append(skipPatterns, re)
		}
	}

	return func(c *gin.Context) {
		// 检查是否跳过该路径
		if shouldSkip(c.Request.URL.Path, skipPatterns, cfg) {
			c.Next()
			return
		}

		// 记录开始时间
		startTime := time.Now()

		// 读取并处理请求体
		requestBody := ""
		if cfg.LogRequestBody && c.Request.Body != nil {
			if bodyBytes, err := readRequestBody(c, cfg.MaxRequestSize); err == nil {
				requestBody = maskSensitiveData(string(bodyBytes), cfg.SensitiveFields)
				// 重新设置请求体
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// 包装响应写入器
		w := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
			config:         cfg,
		}
		c.Writer = w

		// 处理请求
		c.Next()

		// 计算执行时间
		executeTime := time.Since(startTime).Milliseconds()

		// 获取用户信息
		var userID uint64
		var username string
		if user := GetCurrentUser(c); user != nil {
			userID = user.ID
			username = user.Username
		}

		// 获取操作名称和资源
		operation, resource := getOperationAndResource(c.Request.Method, c.Request.URL.Path)

		// 获取资源ID
		resourceID := getResourceID(c)

		// 获取响应体
		responseBody := ""
		if cfg.LogResponseBody {
			responseBody = maskSensitiveData(w.body.String(), cfg.SensitiveFields)
		}

		// 创建日志记录函数
		logFunc := func() {
			if err := createOperationLog(
				helper,
				&userID,
				username,
				operation,
				resource,
				resourceID,
				c.Request.Method,
				c.Request.URL.Path,
				requestBody,
				responseBody,
				c.ClientIP(),
				c.GetHeader("User-Agent"),
				c.Writer.Status(),
				executeTime,
			); err != nil {
				helper.GetLogger().Error(fmt.Sprintf("记录操作日志失败: %s", err.Error()))
			}
		}

		// 根据配置决定同步还是异步记录
		if cfg.AsyncLogging {
			go logFunc()
		} else {
			logFunc()
		}
	}
}

// shouldSkip 检查是否应该跳过记录
func shouldSkip(path string, skipPatterns []*regexp.Regexp, config *OperationLogConfig) bool {
	// 健康检查路径
	if !config.LogHealthCheck && (path == "/health" || path == "/api/v1/health") {
		return true
	}

	// 检查跳过路径
	for _, pattern := range skipPatterns {
		if pattern.MatchString(path) {
			return true
		}
	}

	return false
}

// readRequestBody 读取请求体
func readRequestBody(c *gin.Context, maxSize int64) ([]byte, error) {
	if c.Request.Body == nil {
		return nil, nil
	}

	// 限制读取大小
	limitedReader := io.LimitReader(c.Request.Body, maxSize+1)
	bodyBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, err
	}

	// 如果超过限制，截断并添加提示
	if int64(len(bodyBytes)) > maxSize {
		bodyBytes = bodyBytes[:maxSize]
		bodyBytes = append(bodyBytes, []byte("...[TRUNCATED]")...)
	}

	return bodyBytes, nil
}

// maskSensitiveData 脱敏敏感数据
func maskSensitiveData(data string, sensitiveFields []string) string {
	if data == "" || len(sensitiveFields) == 0 {
		return data
	}

	// 尝试解析为JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &jsonData); err == nil {
		// JSON数据脱敏
		maskJSONSensitiveFields(jsonData, sensitiveFields)
		if maskedBytes, err := json.Marshal(jsonData); err == nil {
			return string(maskedBytes)
		}
	}

	// 非JSON数据，使用正则表达式脱敏
	result := data
	for _, field := range sensitiveFields {
		// 脱敏pattern: "field":"value" 或 field=value
		patterns := []string{
			fmt.Sprintf(`"%s"\s*:\s*"[^"]*"`, field),
			fmt.Sprintf(`%s\s*=\s*[^\s&]*`, field),
		}

		for _, pattern := range patterns {
			if re, err := regexp.Compile(pattern); err == nil {
				result = re.ReplaceAllStringFunc(result, func(match string) string {
					if strings.Contains(match, ":") {
						return fmt.Sprintf(`"%s":"***"`, field)
					}
					return fmt.Sprintf(`%s=***`, field)
				})
			}
		}
	}

	return result
}

// maskJSONSensitiveFields 脱敏JSON中的敏感字段
func maskJSONSensitiveFields(data map[string]interface{}, sensitiveFields []string) {
	for key, value := range data {
		// 检查是否是敏感字段
		for _, sensitive := range sensitiveFields {
			if strings.EqualFold(key, sensitive) {
				data[key] = "***"
				break
			}
		}

		// 递归处理嵌套对象
		if nestedMap, ok := value.(map[string]interface{}); ok {
			maskJSONSensitiveFields(nestedMap, sensitiveFields)
		}
	}
}

// createOperationLog 创建操作日志
func createOperationLog(helper envInterface.HelperInterface, userID *uint64, username, operation, resource, resourceID, method, path, requestBody, responseBody, ip, userAgent string, responseCode int, executeTime int64) error {
	logService := service.NewOperationLogService(helper)

	status := int8(1) // 成功
	errorMessage := ""

	// 判断是否成功
	if responseCode >= 400 {
		status = 0 // 失败
		// 从响应体中提取错误信息
		if responseBody != "" {
			var respMap map[string]interface{}
			if err := json.Unmarshal([]byte(responseBody), &respMap); err == nil {
				if msg, ok := respMap["message"].(string); ok {
					errorMessage = msg
				}
			}
		}
	}

	return logService.CreateLog(
		userID,
		username,
		operation,
		resource,
		resourceID,
		method,
		path,
		requestBody,
		responseBody,
		ip,
		userAgent,
		responseCode,
		executeTime,
		status,
		errorMessage,
	)
}

// OperationMapping 操作映射配置
type OperationMapping struct {
	Method    string `json:"method"`
	Pattern   string `json:"pattern"`
	Operation string `json:"operation"`
	Resource  string `json:"resource"`
}

// 默认操作映射
var defaultOperationMappings = []OperationMapping{
	// 短网址相关
	{"POST", "/api/v1/short_links", "创建", "短网址"},
	{"POST", "/api/v1/short_links/batch", "批量创建", "短网址"},
	{"GET", "/api/v1/short_links", "查看列表", "短网址"},
	{"GET", "/api/v1/short_links/[^/]+", "查看详情", "短网址"},
	{"PUT", "/api/v1/short_links/[^/]+", "更新", "短网址"},
	{"DELETE", "/api/v1/short_links/[^/]+", "删除", "短网址"},
	{"GET", "/api/v1/short_links/[^/]+/statistics", "查看统计", "短网址"},

	// 域名相关
	{"POST", "/api/v1/domains", "创建", "域名"},
	{"GET", "/api/v1/domains", "查看列表", "域名"},
	{"GET", "/api/v1/domains/[^/]+", "查看详情", "域名"},
	{"PUT", "/api/v1/domains/[^/]+", "更新", "域名"},
	{"DELETE", "/api/v1/domains/[^/]+", "删除", "域名"},

	// AB测试相关
	{"POST", "/api/v1/ab_tests", "创建", "AB测试"},
	{"GET", "/api/v1/ab_tests", "查看列表", "AB测试"},
	{"GET", "/api/v1/ab_tests/[^/]+", "查看详情", "AB测试"},
	{"PUT", "/api/v1/ab_tests/[^/]+", "更新", "AB测试"},
	{"DELETE", "/api/v1/ab_tests/[^/]+", "删除", "AB测试"},

	// 用户相关
	{"POST", "/api/v1/users", "创建", "用户"},
	{"GET", "/api/v1/users", "查看列表", "用户"},
	{"GET", "/api/v1/users/[^/]+", "查看详情", "用户"},
	{"PUT", "/api/v1/users/[^/]+", "更新", "用户"},
	{"DELETE", "/api/v1/users/[^/]+", "删除", "用户"},
	{"POST", "/api/v1/users/[^/]+/reset-password", "重置密码", "用户"},

	// 当前用户相关
	{"GET", "/api/v1/profile", "查看资料", "个人信息"},
	{"POST", "/api/v1/profile/change-password", "修改密码", "个人信息"},

	// Token相关
	{"POST", "/api/v1/tokens", "创建", "Token"},
	{"GET", "/api/v1/tokens", "查看列表", "Token"},
	{"DELETE", "/api/v1/tokens/[^/]+", "删除", "Token"},

	// 操作日志
	{"GET", "/api/v1/logs", "查看列表", "操作日志"},

	// 认证相关
	{"POST", "/api/v1/login", "登录", "认证"},
	{"POST", "/api/v1/auth/login", "登录", "认证"},
	{"POST", "/api/v1/auth/logout", "登出", "认证"},
}

// getOperationAndResource 根据HTTP方法和路径获取操作名称和资源类型
func getOperationAndResource(method, path string) (string, string) {
	// 使用配置的映射规则
	for _, mapping := range defaultOperationMappings {
		if mapping.Method == method {
			if matched, _ := regexp.MatchString("^"+mapping.Pattern+"$", path); matched {
				return mapping.Operation, mapping.Resource
			}
		}
	}

	// 兜底逻辑：基于路径解析
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	operation := method
	resource := "未知"

	if len(pathParts) >= 3 && pathParts[0] == "api" && pathParts[1] == "v1" {
		resource = pathParts[2]

		// 根据方法确定操作
		switch method {
		case "POST":
			if strings.Contains(path, "/batch") {
				operation = "批量创建"
			} else {
				operation = "创建"
			}
		case "GET":
			if strings.Contains(path, "/statistics") {
				operation = "查看统计"
			} else if len(pathParts) == 3 {
				operation = "查看列表"
			} else {
				operation = "查看详情"
			}
		case "PUT", "PATCH":
			operation = "更新"
		case "DELETE":
			operation = "删除"
		}

		// 资源名称本地化
		resourceMap := map[string]string{
			"short_links": "短网址",
			"domains":     "域名",
			"ab_tests":    "AB测试",
			"users":       "用户",
			"tokens":      "Token",
			"logs":        "操作日志",
			"profile":     "个人信息",
		}

		if localized, exists := resourceMap[resource]; exists {
			resource = localized
		}
	}

	return operation, resource
}

// getResourceID 获取资源ID
func getResourceID(c *gin.Context) string {
	// 尝试从路径参数中获取各种可能的ID
	idParams := []string{"id", "code", "token_id", "domain", "test_id"}

	for _, param := range idParams {
		if value := c.Param(param); value != "" {
			return value
		}
	}

	// 尝试从查询参数中获取ID
	for _, param := range idParams {
		if value := c.Query(param); value != "" {
			return value
		}
	}

	return ""
}
