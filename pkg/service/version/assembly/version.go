package assembly

import (
	"cnb.cool/mliev/open/dwz-server/pkg/interfaces"
	"cnb.cool/mliev/open/dwz-server/pkg/service/version/impl"
)

// Version 版本组装器
type Version struct {
	Helper interfaces.HelperInterface
}

// Assembly 组装版本管理器
func (receiver *Version) Assembly() error {
	version := impl.NewVersion()
	receiver.Helper.SetVersion(version)
	return nil
}
