package model

import (
	"time"
)

// OperationLog 操作日志模型
type OperationLog struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`               // 自增主键
	UserID       *uint64   `gorm:"index" json:"user_id"`               // 操作用户ID，可为空（匿名操作）
	Username     string    `gorm:"size:50" json:"username"`            // 用户名（冗余存储便于查询）
	Operation    string    `gorm:"size:100;not null" json:"operation"` // 操作名称
	Resource     string    `gorm:"size:100" json:"resource"`           // 操作资源
	ResourceID   string    `gorm:"size:100" json:"resource_id"`        // 资源ID
	Method       string    `gorm:"size:10" json:"method"`              // HTTP方法
	Path         string    `gorm:"size:255" json:"path"`               // 请求路径
	RequestBody  string    `gorm:"type:text" json:"request_body"`      // 请求体
	ResponseCode int       `gorm:"default:0" json:"response_code"`     // 响应状态码
	ResponseBody string    `gorm:"type:text" json:"response_body"`     // 响应体
	IP           string    `gorm:"size:45" json:"ip"`                  // 操作IP
	UserAgent    string    `gorm:"size:500" json:"user_agent"`         // 用户代理
	ExecuteTime  int64     `gorm:"default:0" json:"execute_time"`      // 执行耗时（毫秒）
	Status       int8      `gorm:"default:1" json:"status"`            // 状态：1-成功，0-失败
	ErrorMessage string    `gorm:"size:1000" json:"error_message"`     // 错误信息
	CreatedAt    time.Time `json:"created_at"`                         // 创建时间

	// 关联查询
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (ol *OperationLog) TableName() string {
	return "operation_logs"
}

// SetSuccess 设置操作成功
func (ol *OperationLog) SetSuccess(responseCode int, responseBody string, executeTime int64) {
	ol.Status = 1
	ol.ResponseCode = responseCode
	ol.ResponseBody = responseBody
	ol.ExecuteTime = executeTime
}

// SetFailed 设置操作失败
func (ol *OperationLog) SetFailed(responseCode int, errorMessage string, executeTime int64) {
	ol.Status = 0
	ol.ResponseCode = responseCode
	ol.ErrorMessage = errorMessage
	ol.ExecuteTime = executeTime
}
