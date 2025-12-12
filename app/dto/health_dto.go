package dto

// HealthStatus 健康状态结构
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp int64                  `json:"timestamp"`
	Version   *VersionInfo           `json:"version,omitempty"`
	Services  map[string]interface{} `json:"services"`
}

// VersionInfo 版本信息
type VersionInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildTime string `json:"build_time"`
}

// ServiceStatus 服务状态
type ServiceStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}
