package dto

// HealthStatus 健康状态结构
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp int64                  `json:"timestamp"`
	Services  map[string]interface{} `json:"services"`
}

// ServiceStatus 服务状态
type ServiceStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}
