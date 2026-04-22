package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"cnb.cool/mliev/dwz/dwz-server/app/service"
	"cnb.cool/mliev/dwz/dwz-server/pkg/helper"
	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
)

// OperationLogConfig 操作日志配置
type OperationLogConfig struct {
	Enable          bool     `yaml:"enable"`
	MaxRequestSize  int64    `yaml:"max_request_size"`
	SkipPaths       []string `yaml:"skip_paths"`
	SensitiveFields []string `yaml:"sensitive_fields"`
	LogRequestBody  bool     `yaml:"log_request_body"`
	AsyncLogging    bool     `yaml:"async_logging"`
	LogHealthCheck  bool     `yaml:"log_health_check"`
}

// DefaultOperationLogConfig holds the defaults. Response-body capture is no
// longer supported because RouterContextInterface does not expose a writer
// hook; status code, request body and metadata are still recorded.
var DefaultOperationLogConfig = OperationLogConfig{
	Enable:          true,
	MaxRequestSize:  1024 * 1024,
	SkipPaths:       []string{},
	SensitiveFields: []string{"password", "token", "secret", "key", "passwd"},
	LogRequestBody:  true,
	AsyncLogging:    true,
	LogHealthCheck:  false,
}

func OperationLogMiddleware(config ...*OperationLogConfig) httpInterfaces.HandlerFunc {
	cfg := &DefaultOperationLogConfig
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	}

	if !cfg.Enable {
		return func(c httpInterfaces.RouterContextInterface) { c.Next() }
	}

	skipPatterns := make([]*regexp.Regexp, 0, len(cfg.SkipPaths))
	for _, pattern := range cfg.SkipPaths {
		if re, err := regexp.Compile(pattern); err == nil {
			skipPatterns = append(skipPatterns, re)
		}
	}

	return func(c httpInterfaces.RouterContextInterface) {
		if shouldSkip(c.Path(), skipPatterns, cfg) {
			c.Next()
			return
		}

		startTime := time.Now()

		requestBody := ""
		if cfg.LogRequestBody && c.Request().Body != nil {
			if bodyBytes, err := readRequestBody(c, cfg.MaxRequestSize); err == nil {
				requestBody = maskSensitiveData(string(bodyBytes), cfg.SensitiveFields)
				c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		c.Next()

		executeTime := time.Since(startTime).Milliseconds()

		var userID uint64
		var username string
		if user := GetCurrentUser(c); user != nil {
			userID = user.ID
			username = user.Username
		}

		operation, resource := getOperationAndResource(c.Method(), c.Path())
		resourceID := getResourceID(c)

		method := c.Method()
		path := c.Path()
		clientIP := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		status := c.GetStatus()

		logFunc := func() {
			if err := createOperationLog(
				&userID,
				username,
				operation,
				resource,
				resourceID,
				method,
				path,
				requestBody,
				"",
				clientIP,
				userAgent,
				status,
				executeTime,
			); err != nil {
				helper.GetHelper().GetLogger().Error(fmt.Sprintf("记录操作日志失败: %s", err.Error()))
			}
		}

		if cfg.AsyncLogging {
			go logFunc()
		} else {
			logFunc()
		}
	}
}

func shouldSkip(path string, skipPatterns []*regexp.Regexp, config *OperationLogConfig) bool {
	if !config.LogHealthCheck && (path == "/health" || path == "/api/v1/health") {
		return true
	}
	for _, pattern := range skipPatterns {
		if pattern.MatchString(path) {
			return true
		}
	}
	return false
}

func readRequestBody(c httpInterfaces.RouterContextInterface, maxSize int64) ([]byte, error) {
	body := c.Request().Body
	if body == nil {
		return nil, nil
	}
	limited := io.LimitReader(body, maxSize+1)
	bodyBytes, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(bodyBytes)) > maxSize {
		bodyBytes = bodyBytes[:maxSize]
		bodyBytes = append(bodyBytes, []byte("...[TRUNCATED]")...)
	}
	return bodyBytes, nil
}

func maskSensitiveData(data string, sensitiveFields []string) string {
	if data == "" || len(sensitiveFields) == 0 {
		return data
	}

	var jsonData map[string]any
	if err := json.Unmarshal([]byte(data), &jsonData); err == nil {
		maskJSONSensitiveFields(jsonData, sensitiveFields)
		if maskedBytes, err := json.Marshal(jsonData); err == nil {
			return string(maskedBytes)
		}
	}

	result := data
	for _, field := range sensitiveFields {
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

func maskJSONSensitiveFields(data map[string]any, sensitiveFields []string) {
	for key, value := range data {
		for _, sensitive := range sensitiveFields {
			if strings.EqualFold(key, sensitive) {
				data[key] = "***"
				break
			}
		}
		if nestedMap, ok := value.(map[string]any); ok {
			maskJSONSensitiveFields(nestedMap, sensitiveFields)
		}
	}
}

func createOperationLog(userID *uint64, username, operation, resource, resourceID, method, path, requestBody, responseBody, ip, userAgent string, responseCode int, executeTime int64) error {
	logService := service.NewOperationLogService(helper.GetHelper())
	status := int8(1)
	errorMessage := ""
	if responseCode >= 400 {
		status = 0
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

type OperationMapping struct {
	Method    string
	Pattern   string
	Operation string
	Resource  string
}

var defaultOperationMappings = []OperationMapping{
	{"POST", "/api/v1/short_links", "创建", "短网址"},
	{"POST", "/api/v1/short_links/batch", "批量创建", "短网址"},
	{"GET", "/api/v1/short_links", "查看列表", "短网址"},
	{"GET", "/api/v1/short_links/[^/]+", "查看详情", "短网址"},
	{"PUT", "/api/v1/short_links/[^/]+", "更新", "短网址"},
	{"DELETE", "/api/v1/short_links/[^/]+", "删除", "短网址"},
	{"GET", "/api/v1/short_links/[^/]+/statistics", "查看统计", "短网址"},
	{"POST", "/api/v1/domains", "创建", "域名"},
	{"GET", "/api/v1/domains", "查看列表", "域名"},
	{"GET", "/api/v1/domains/[^/]+", "查看详情", "域名"},
	{"PUT", "/api/v1/domains/[^/]+", "更新", "域名"},
	{"DELETE", "/api/v1/domains/[^/]+", "删除", "域名"},
	{"POST", "/api/v1/ab_tests", "创建", "AB测试"},
	{"GET", "/api/v1/ab_tests", "查看列表", "AB测试"},
	{"GET", "/api/v1/ab_tests/[^/]+", "查看详情", "AB测试"},
	{"PUT", "/api/v1/ab_tests/[^/]+", "更新", "AB测试"},
	{"DELETE", "/api/v1/ab_tests/[^/]+", "删除", "AB测试"},
	{"POST", "/api/v1/users", "创建", "用户"},
	{"GET", "/api/v1/users", "查看列表", "用户"},
	{"GET", "/api/v1/users/[^/]+", "查看详情", "用户"},
	{"PUT", "/api/v1/users/[^/]+", "更新", "用户"},
	{"DELETE", "/api/v1/users/[^/]+", "删除", "用户"},
	{"POST", "/api/v1/users/[^/]+/reset-password", "重置密码", "用户"},
	{"GET", "/api/v1/profile", "查看资料", "个人信息"},
	{"POST", "/api/v1/profile/change-password", "修改密码", "个人信息"},
	{"POST", "/api/v1/tokens", "创建", "Token"},
	{"GET", "/api/v1/tokens", "查看列表", "Token"},
	{"DELETE", "/api/v1/tokens/[^/]+", "删除", "Token"},
	{"GET", "/api/v1/logs", "查看列表", "操作日志"},
	{"POST", "/api/v1/login", "登录", "认证"},
	{"POST", "/api/v1/auth/login", "登录", "认证"},
	{"POST", "/api/v1/auth/logout", "登出", "认证"},
}

func getOperationAndResource(method, path string) (string, string) {
	for _, mapping := range defaultOperationMappings {
		if mapping.Method == method {
			if matched, _ := regexp.MatchString("^"+mapping.Pattern+"$", path); matched {
				return mapping.Operation, mapping.Resource
			}
		}
	}

	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	operation := method
	resource := "未知"

	if len(pathParts) >= 3 && pathParts[0] == "api" && pathParts[1] == "v1" {
		resource = pathParts[2]
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

func getResourceID(c httpInterfaces.RouterContextInterface) string {
	idParams := []string{"id", "code", "token_id", "domain", "test_id"}
	for _, param := range idParams {
		if value := c.Param(param); value != "" {
			return value
		}
	}
	for _, param := range idParams {
		if value := c.Query(param); value != "" {
			return value
		}
	}
	return ""
}
