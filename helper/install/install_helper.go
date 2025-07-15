package install

import (
	"os"
	"sync"
)

var (
	isInstalled    bool
	installMutex   sync.RWMutex
	installChecked bool
)

// IsInstalled 检查系统是否已经安装
func IsInstalled() bool {
	installMutex.RLock()
	defer installMutex.RUnlock()
	return isInstalled
}

// SetInstalled 设置安装状态
func SetInstalled(installed bool) {
	installMutex.Lock()
	defer installMutex.Unlock()
	isInstalled = installed
}

// CheckInstallStatus 检查并初始化安装状态
func CheckInstallStatus() bool {
	installMutex.Lock()
	defer installMutex.Unlock()

	// 如果已经检查过，直接返回状态
	if installChecked {
		return isInstalled
	}

	// 检查安装文件
	lockFile := "./config/install.lock"
	configFile := "./config/config.yaml"

	// 检查锁文件和配置文件是否存在
	lockExists := fileExists(lockFile)
	configExists := fileExists(configFile)

	isInstalled = lockExists && configExists
	installChecked = true

	return isInstalled
}

// MarkAsInstalled 标记系统为已安装
func MarkAsInstalled() {
	installMutex.Lock()
	defer installMutex.Unlock()
	isInstalled = true
}

// fileExists 检查文件是否存在
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// GetInstallStatus 获取详细的安装状态信息
func GetInstallStatus() map[string]interface{} {
	installMutex.RLock()
	defer installMutex.RUnlock()

	return map[string]interface{}{
		"installed":     isInstalled,
		"checked":       installChecked,
		"lock_exists":   fileExists("./config/install.lock"),
		"config_exists": fileExists("./config/config.yaml"),
	}
}
