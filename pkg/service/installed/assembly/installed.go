package assembly

import (
	"reflect"

	"cnb.cool/mliev/dwz/dwz-server/pkg/helper"
	"cnb.cool/mliev/dwz/dwz-server/pkg/service/installed/impl"
	"github.com/muleiwu/gsr"
)

const (
	lockFilePath   = "./config/install.lock"
	configFilePath = "./config/config.yaml"
)

// Installed assembles the install-state checker. Depends on the logger so
// the underlying impl can log "installed/not installed" during Init().
type Installed struct{}

func (Installed) Type() reflect.Type { return reflect.TypeFor[*impl.Installed]() }
func (Installed) DependsOn() []reflect.Type {
	return []reflect.Type{reflect.TypeFor[gsr.Logger]()}
}
func (Installed) Assembly() (any, error) {
	return impl.NewInstalled(lockFilePath, configFilePath, helper.GetHelper()), nil
}
