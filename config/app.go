package config

import (
	"embed"

	cacheAssembly "cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/cache/assembly"
	databaseAssembly "cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/database/assembly"
	idGenerator "cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/id_generator/service"
	installedAssembly "cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/installed/assembly"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/migration"
	redisAssembly "cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/redis/assembly"
	versionAssembly "cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/version/assembly"
	"cnb.cool/mliev/open/go-web/pkg/interfaces"
	configAssembly "cnb.cool/mliev/open/go-web/pkg/server/config/assembly"
	envAssembly "cnb.cool/mliev/open/go-web/pkg/server/env/assembly"
	httpServer "cnb.cool/mliev/open/go-web/pkg/server/http_server/service"
	loggerAssembly "cnb.cool/mliev/open/go-web/pkg/server/logger/assembly"
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
//
// The chain is ordered so that `installed` resolves before any service that
// depends on external resources (DB / Redis / Cache). The dwz-private
// database / redis / cache assemblies short-circuit pre-install so a fresh
// download boots straight into the install wizard regardless of whether
// MySQL or Redis is running.
func DefaultAssemblies() []interfaces.AssemblyInterface {
	return []interfaces.AssemblyInterface{
		&envAssembly.Env{},
		&configAssembly.Config{DefaultConfigs: Config{}.Get()},
		&loggerAssembly.Logger{},
		&installedAssembly.Installed{},
		&versionAssembly.Version{},
		&databaseAssembly.Database{},
		&redisAssembly.Redis{},
		&cacheAssembly.Cache{},
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
