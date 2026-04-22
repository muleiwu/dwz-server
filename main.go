package main

import (
	"embed"

	"cnb.cool/mliev/dwz/dwz-server/config"
	"cnb.cool/mliev/open/go-web/cmd"
	"github.com/muleiwu/gomander"
)

//go:embed templates/**
var templateFS embed.FS

//go:embed static/**
var staticFS embed.FS

//go:embed all:migrations
var migrationsFS embed.FS

func main() {
	gomander.Run(func() {
		cmd.Start(
			cmd.WithTemplateFs(templateFS),
			cmd.WithWebStaticFs(staticFS),
			cmd.WithApp(config.App{MigrationsFS: migrationsFS}),
		)
	})
}
