package dto

import "time"

// OperationLogInfo 操作日志信息
type OperationLogInfo struct {
	ID           uint64    `json:"id"`
	UserID       *uint64   `json:"user_id"`
	Username     string    `json:"username"`
	Operation    string    `json:"operation"`
	Resource     string    `json:"resource"`
	ResourceID   string    `json:"resource_id"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	RequestBody  string    `json:"request_body"`
	ResponseCode int       `json:"response_code"`
	ResponseBody string    `json:"response_body"`
	IP           string    `json:"ip"`
	UserAgent    string    `json:"user_agent"`
	ExecuteTime  int64     `json:"execute_time"`
	Status       int8      `json:"status"`
	ErrorMessage string    `json:"error_message"`
	CreatedAt    time.Time `json:"created_at"`
}

// OperationLogListRequest 操作日志列表请求
type OperationLogListRequest struct {
	Page      int     `form:"page,default=1" binding:"min=1" example:"1"`
	PageSize  int     `form:"page_size,default=10" binding:"min=1,max=100" example:"10"`
	UserID    *uint64 `form:"user_id" example:"1"`
	Username  string  `form:"username" example:"admin"`
	Operation string  `form:"operation" example:"创建短网址"`
	Resource  string  `form:"resource" example:"short_link"`
	Method    string  `form:"method" example:"POST"`
	Status    *int8   `form:"status" binding:"omitempty,oneof=0 1" example:"1"`
	StartTime string  `form:"start_time" example:"2024-01-01 00:00:00"`
	EndTime   string  `form:"end_time" example:"2024-12-31 23:59:59"`
}

// OperationLogListResponse 操作日志列表响应
type OperationLogListResponse struct {
	List       []OperationLogInfo `json:"list"`
	Pagination Pagination         `json:"pagination"`
}
