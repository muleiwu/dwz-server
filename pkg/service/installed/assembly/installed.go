package assembly

import (
	"cnb.cool/mliev/dwz/dwz-server/pkg/interfaces"
	"cnb.cool/mliev/dwz/dwz-server/pkg/service/installed/impl"
)

type Installed struct {
	Helper interfaces.HelperInterface
}

var lockFilePath = "./config/install.lock"
var configFilePath = "./config/config.yaml"

func (receiver *Installed) Assembly() error {
	installed := impl.NewInstalled(lockFilePath, configFilePath, receiver.Helper)
	receiver.Helper.SetInstalled(installed)
	return nil
}
