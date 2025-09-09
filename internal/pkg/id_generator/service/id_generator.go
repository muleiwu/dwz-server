package service

import (
	"cnb.cool/mliev/open/dwz-server/internal/helper"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
)

type IDGenerator struct {
	Helper interfaces.HelperInterface
}

func (receiver *IDGenerator) Run() error {
	_, err := helper.InitIdGenerator(receiver.Helper)

	if err != nil {
		return err
	}

	return nil
}
