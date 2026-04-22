package assembly

import (
	"reflect"

	installedImpl "cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/installed/impl"
	"cnb.cool/mliev/open/go-web/pkg/container"
	"cnb.cool/mliev/open/go-web/pkg/server/database/config"
	dbDriver "cnb.cool/mliev/open/go-web/pkg/server/database/driver"
	"github.com/muleiwu/golog"
	"github.com/muleiwu/gsr"
	"gorm.io/gorm"
)

// Database is the dwz-private wrapper around go-web's database driver
// manager. Pre-install (no install.lock yet) it deliberately registers a
// typed-nil *gorm.DB so the rest of the assembly chain succeeds and the
// HTTP server boots into the install flow. After install completes, a
// SIGHUP reload re-runs assemblies and the real connection is built.
type Database struct{}

func (Database) Type() reflect.Type { return reflect.TypeFor[*gorm.DB]() }
func (Database) DependsOn() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[gsr.Provider](),
		reflect.TypeFor[gsr.Logger](),
		reflect.TypeFor[*installedImpl.Installed](),
	}
}

func (Database) Assembly() (any, error) {
	logger := container.MustGet[gsr.Logger]()
	installed := container.MustGet[*installedImpl.Installed]()
	if !installed.IsInstalled() {
		logger.Notice("[database] 应用未安装，跳过数据库连接，等待安装向导")
		return (*gorm.DB)(nil), nil
	}

	cfg := container.MustGet[gsr.Provider]()
	dbCfg := config.NewConfig(cfg)
	db, err := dbDriver.DatabaseDriverManager.Make(dbCfg.Driver, dbCfg)
	if err != nil {
		// Already installed but DB is unreachable: log loudly and still
		// register a nil — operators can fix the config and SIGHUP-reload.
		logger.Error("[database] 已安装但数据库连接失败，请检查配置或网络: " + err.Error())
		return (*gorm.DB)(nil), nil
	}
	logger.Info("[database] 数据库连接成功", golog.Field("driver", dbCfg.Driver))
	return db, nil
}
