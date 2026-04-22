package assembly

import (
	"reflect"

	"cnb.cool/mliev/dwz/dwz-server/pkg/service/version/impl"
)

// Version assembles the build-info holder. No dependencies; populated by
// AppProvider after assembly via SetVersionInfo (see config/app.go).
type Version struct{}

func (Version) Type() reflect.Type        { return reflect.TypeFor[*impl.Version]() }
func (Version) DependsOn() []reflect.Type { return nil }
func (Version) Assembly() (any, error) {
	return impl.NewVersion(), nil
}
