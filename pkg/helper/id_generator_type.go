package helper

import (
	"reflect"

	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
)

func idGeneratorType() reflect.Type { return reflect.TypeFor[interfaces.IDGenerator]() }
