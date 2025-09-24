package assembly

import (
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"cnb.cool/mliev/open/dwz-server/internal/pkg/id_generator/impl"
)

func GetDriver(helper interfaces.HelperInterface, driver string) (interfaces.IDGenerator, error) {
	if driver == "redis" {
		return impl.NewIDGeneratorRedis(helper), nil
	} else {
		// local/base implementation
		return impl.NewIDGeneratorLocal(), nil
	}
}
