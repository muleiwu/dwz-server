package main

import (
	"embed"

	"cnb.cool/mliev/dwz/dwz-server/v2/config"
	"cnb.cool/mliev/open/go-web/cmd"
	"github.com/muleiwu/gomander"
)

//go:embed templates/**
var templateFS embed.FS

//go:embed static/**
var staticFS embed.FS

//go:embed all:migrations
var migrationsFS embed.FS

var (
	Version   = "unknown"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	gomander.Run(func() {
		cmd.Start(
			cmd.WithTemplateFs(templateFS),
			cmd.WithWebStaticFs(staticFS),
			cmd.WithApp(config.App{
				MigrationsFS: migrationsFS,
				Version:      Version,
				BuildTime:    BuildTime,
				GitCommit:    GitCommit,
			}),
		)
	})
}
