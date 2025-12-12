package assembly

import (
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"cnb.cool/mliev/open/dwz-server/internal/pkg/version/impl"
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
