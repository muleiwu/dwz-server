package impl

import (
	"os"
	"sync"

	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
)

type Installed struct {
	helper         interfaces.HelperInterface
	isInstalled    bool
	installMutex   sync.RWMutex
	lockFilePath   string
	configFilePath string
}

func NewInstalled(lockFilePath, configFilePath string, helper interfaces.HelperInterface) *Installed {
	i := &Installed{
		helper:         helper,
		lockFilePath:   lockFilePath,
		configFilePath: configFilePath,
	}
	i.Init()
	return i
}

func (receiver *Installed) Init() {
	if receiver.fileExists(receiver.lockFilePath) && receiver.fileExists(receiver.configFilePath) {
		receiver.isInstalled = true
		receiver.helper.GetLogger().Info("dwz-server is installed")
	} else {
		receiver.isInstalled = false
		receiver.helper.GetLogger().Warn("dwz-server is not installed")
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
