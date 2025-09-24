package helper

import (
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
)

var idGeneratorHelper interfaces.IDGenerator

func GetIdGenerator() interfaces.IDGenerator {
	if idGeneratorHelper == nil {
		GetHelper().GetLogger().Error("发号器未初始化，请检查")
	}
	return idGeneratorHelper
}

func SetIdGenerator(idGenerator interfaces.IDGenerator) {
	idGeneratorHelper = idGenerator
}
