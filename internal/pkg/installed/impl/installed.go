package impl

import (
	"os"
	"sync"
)

type Installed struct {
	isInstalled    bool
	installMutex   sync.RWMutex
	lockFilePath   string
	configFilePath string
}

func NewInstalled(lockFilePath, configFilePath string) *Installed {
	i := &Installed{
		lockFilePath:   lockFilePath,
		configFilePath: configFilePath,
	}
	i.Init()
	return i
}

func (receiver *Installed) Init() {
	if receiver.fileExists(lockFilePath) && receiver.fileExists(configFilePath) {
		receiver.isInstalled = true
	} else {
		receiver.isInstalled = false
	}
}

func (receiver *Installed) SetInstalled(installed bool) {
	receiver.isInstalled = installed
}

func (receiver *Installed) IsInstalled() bool {
	receiver.installMutex.RLock()
	defer receiver.installMutex.RUnlock()
	return receiver.isInstalled
}

func (receiver *Installed) Install(fun func() error) error {
	receiver.installMutex.RLock()
	defer receiver.installMutex.RUnlock()
	if !receiver.isInstalled {
		err := fun()
		if err != nil {
			return err
		}

		receiver.Init()
	}

	return nil
}

// fileExists 检查文件是否存在
func (receiver *Installed) fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
