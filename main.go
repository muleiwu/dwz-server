package main

import (
	"cnb.cool/mliev/open/dwz-server/cmd"

	"embed"
)

//go:embed templates/**
var templateFS embed.FS

//go:embed public/admin/**
var publicAdminFs embed.FS

func main() {
	staticFs := map[string]embed.FS{
		"templates":    templateFS,
		"public/admin": publicAdminFs,
	}
	cmd.Start(staticFs)
}
