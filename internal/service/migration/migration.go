package migration

import (
	"fmt"

	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
)

type Migration struct {
	Helper    interfaces.HelperInterface
	Migration []any
}

func (receiver *Migration) Run() error {

	if receiver.Helper.GetDatabase() == nil {
		return fmt.Errorf("[db migration err: database is nil]")
	}

	autoInstall := receiver.Helper.GetEnv().GetString("AUTO_INSTALL", "")

	if !receiver.Helper.GetInstalled().IsInstalled() && autoInstall != "install" {
		return nil
	}

	if len(receiver.Migration) > 0 {
		err := receiver.Helper.GetDatabase().AutoMigrate(receiver.Migration...)
		if err != nil {
			return fmt.Errorf("[db migration err:%s]", err.Error())
		}

		receiver.Helper.GetLogger().Info(fmt.Sprintf("[db migration success: %d models migrated]", len(receiver.Migration)))
	}
	return nil
}
