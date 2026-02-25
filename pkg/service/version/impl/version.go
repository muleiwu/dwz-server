package impl

import (
	"sync"
)

// Version 版本信息管理器
type Version struct {
	version   string
	gitCommit string
	buildTime string
	mutex     sync.RWMutex
}

// NewVersion 创建版本管理器实例
func NewVersion() *Version {
	return &Version{
		version:   "unknown",
		gitCommit: "unknown",
		buildTime: "unknown",
	}
}

// GetVersion 获取版本号
func (v *Version) GetVersion() string {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return v.version
}

// GetGitCommit 获取 Git 提交哈希
func (v *Version) GetGitCommit() string {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return v.gitCommit
}

// GetBuildTime 获取构建时间
func (v *Version) GetBuildTime() string {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return v.buildTime
}

// SetVersionInfo 设置版本信息
func (v *Version) SetVersionInfo(version, commit, buildTime string) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if version != "" {
		v.version = version
	}
	if commit != "" {
		v.gitCommit = commit
	}
	if buildTime != "" {
		v.buildTime = buildTime
	}
}
