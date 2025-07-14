package constants

// 错误码定义
const (
	// 成功
	ErrCodeSuccess = 0

	// 客户端错误 4xx
	ErrCodeBadRequest   = 400 // 请求参数错误
	ErrCodeUnauthorized = 401 // 未授权
	ErrCodeForbidden    = 403 // 禁止访问
	ErrCodeNotFound     = 404 // 资源不存在
	ErrCodeConflict     = 409 // 资源冲突

	// 服务器错误 5xx
	ErrCodeInternal    = 500 // 内部服务器错误
	ErrCodeBadGateway  = 502 // 网关错误
	ErrCodeUnavailable = 503 // 服务不可用

	// 短网址服务相关错误码
	ErrCodeShortLinkNotFound  = 1001 // 短网址不存在
	ErrCodeShortLinkExpired   = 1002 // 短网址已过期
	ErrCodeShortLinkDisabled  = 1003 // 短网址已禁用
	ErrCodeCustomCodeExists   = 1004 // 自定义短代码已存在
	ErrCodeInvalidCustomCode  = 1005 // 无效的自定义短代码
	ErrCodeDomainNotAllowed   = 1006 // 域名不允许
	ErrCodeGenerateCodeFailed = 1007 // 生成短代码失败
)

// 错误信息映射
var ErrMessages = map[int]string{
	ErrCodeSuccess:      "操作成功",
	ErrCodeBadRequest:   "请求参数错误",
	ErrCodeUnauthorized: "未授权访问",
	ErrCodeForbidden:    "禁止访问",
	ErrCodeNotFound:     "资源不存在",
	ErrCodeConflict:     "资源冲突",
	ErrCodeInternal:     "内部服务器错误",
	ErrCodeBadGateway:   "网关错误",
	ErrCodeUnavailable:  "服务暂不可用",

	// 短网址服务相关错误信息
	ErrCodeShortLinkNotFound:  "短网址不存在",
	ErrCodeShortLinkExpired:   "短网址已过期",
	ErrCodeShortLinkDisabled:  "短网址已禁用",
	ErrCodeCustomCodeExists:   "自定义短代码已存在",
	ErrCodeInvalidCustomCode:  "无效的自定义短代码",
	ErrCodeDomainNotAllowed:   "域名不允许",
	ErrCodeGenerateCodeFailed: "生成短代码失败",
}

// GetErrMessage 获取错误信息
func GetErrMessage(code int) string {
	if msg, exists := ErrMessages[code]; exists {
		return msg
	}
	return "未知错误"
}
