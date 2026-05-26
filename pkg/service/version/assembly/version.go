package assembly

import (
	"reflect"

	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/version/impl"
)

// Version assembles the build-info holder.
type Version struct {
	Version   string
	GitCommit string
	BuildTime string
}

func (Version) Type() reflect.Type        { return reflect.TypeFor[*impl.Version]() }
func (Version) DependsOn() []reflect.Type { return nil }
func (v Version) Assembly() (any, error) {
	version := impl.NewVersion()
	version.SetVersionInfo(v.Version, v.GitCommit, v.BuildTime)
	return version, nil
}
