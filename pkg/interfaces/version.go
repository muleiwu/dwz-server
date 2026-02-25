package interfaces

// VersionInterface 版本信息接口
type VersionInterface interface {
	// GetVersion 获取版本号
	GetVersion() string
	// GetGitCommit 获取 Git 提交哈希
	GetGitCommit() string
	// GetBuildTime 获取构建时间
	GetBuildTime() string
	// SetVersionInfo 设置版本信息
	SetVersionInfo(version, commit, buildTime string)
}
