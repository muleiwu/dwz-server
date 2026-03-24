package assembly

import (
	"cnb.cool/mliev/dwz/dwz-server/pkg/interfaces"
	"cnb.cool/mliev/dwz/dwz-server/pkg/service/id_generator/impl"
)

func GetDriver(helper interfaces.HelperInterface, driver string) (interfaces.IDGenerator, error) {
	if driver == "redis" {
		return impl.NewIDGeneratorRedis(helper), nil
	} else {
		// local/base implementation
		return impl.NewIDGeneratorLocal(), nil
	}
}
