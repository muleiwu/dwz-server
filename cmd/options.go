package cmd

import (
	"cnb.cool/mliev/open/dwz-server/pkg/interfaces"
	"cnb.cool/mliev/open/dwz-server/pkg/service/http_server/impl"
	"embed"
	"github.com/gin-gonic/gin"
)

// StartOptions defines configuration options for starting the dwz-server
// Supports both CE (Community Edition) and EE (Enterprise Edition) deployments
type StartOptions struct {
	// Build information (required for CE/EE)
	Version   string              // Application version
	GitCommit string              // Git commit hash
	BuildTime string              // Build timestamp
	StaticFs  map[string]embed.FS // Embedded static files (templates, web assets)

	// EE extension hooks (optional, nil-safe)
	ExtraRoutes       func(*gin.Engine, *impl.HttpDeps)                               // Additional HTTP routes for EE
	ExtraModels       []any                                                           // Additional GORM models for EE database migrations
	ExtraAssembly     func(interfaces.HelperInterface) []interfaces.AssemblyInterface // Additional service assemblies for EE
	ExtraConfigs      []interfaces.InitConfig                                         // Additional configuration providers for EE
	ExtraMigrationsFS embed.FS                                                        // EE-specific database migration files
}
