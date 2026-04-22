package config

import (
	"embed"

	idGenerator "cnb.cool/mliev/dwz/dwz-server/pkg/service/id_generator/service"
	installedAssembly "cnb.cool/mliev/dwz/dwz-server/pkg/service/installed/assembly"
	"cnb.cool/mliev/dwz/dwz-server/pkg/service/migration"
	versionAssembly "cnb.cool/mliev/dwz/dwz-server/pkg/service/version/assembly"
	"cnb.cool/mliev/open/go-web/pkg/interfaces"
	cacheAssembly "cnb.cool/mliev/open/go-web/pkg/server/cache/assembly"
	configAssembly "cnb.cool/mliev/open/go-web/pkg/server/config/assembly"
	databaseAssembly "cnb.cool/mliev/open/go-web/pkg/server/database/assembly"
	envAssembly "cnb.cool/mliev/open/go-web/pkg/server/env/assembly"
	httpServer "cnb.cool/mliev/open/go-web/pkg/server/http_server/service"
	loggerAssembly "cnb.cool/mliev/open/go-web/pkg/server/logger/assembly"
	redisAssembly "cnb.cool/mliev/open/go-web/pkg/server/redis/assembly"
)

// App is the dwz-server AppProvider implementation. main.go populates
// MigrationsFS from the //go:embed directive so the migration server can
// resolve the dialect-specific SQL files.
type App struct {
	MigrationsFS embed.FS
}

func (a App) Assemblies() []interfaces.AssemblyInterface { return DefaultAssemblies() }
func (a App) Servers() []interfaces.ServerInterface      { return DefaultServers(a.MigrationsFS) }

// DefaultAssemblies returns the CE assembly chain. EE consumers wrap this and
// append their own assemblies before passing the combined slice to the
// AppProvider they hand to cmd.WithApp.
func DefaultAssemblies() []interfaces.AssemblyInterface {
	return []interfaces.AssemblyInterface{
		&envAssembly.Env{},
		&configAssembly.Config{DefaultConfigs: Config{}.Get()},
		&loggerAssembly.Logger{},
		&databaseAssembly.Database{},
		&redisAssembly.Redis{},
		&cacheAssembly.Cache{},
		&installedAssembly.Installed{},
		&versionAssembly.Version{},
	}
}

// DefaultServers returns the CE server chain (migration → id_generator →
// http_server). EE consumers can prepend / append their own servers around
// this slice. migrationsFS is the embedded SQL tree forwarded from main.go.
func DefaultServers(migrationsFS embed.FS) []interfaces.ServerInterface {
	return []interfaces.ServerInterface{
		&migration.Migration{BaseFS: migrationsFS},
		&idGenerator.IDGenerator{},
		&httpServer.HttpServer{},
	}
}

// DefaultConfigs returns the CE InitConfig list passed into the config
// assembly. EE consumers can append their own InitConfig providers and feed
// the combined slice into a custom configAssembly.Config{DefaultConfigs: ...}.
func DefaultConfigs() []interfaces.InitConfig { return Config{}.Get() }
