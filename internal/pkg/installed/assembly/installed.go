package assembly

import (
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"cnb.cool/mliev/open/dwz-server/internal/pkg/installed/impl"
)

type Installed struct {
	Helper interfaces.HelperInterface
}

var lockFilePath = "./config/install.lock"
var configFilePath = "./config/config.yaml"

func (receiver *Installed) Assembly() error {
	installed := impl.NewInstalled(lockFilePath, configFilePath)
	receiver.Helper.SetInstalled(installed)
	return nil
}
