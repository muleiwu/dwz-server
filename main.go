package main

import (
	"cnb.cool/mliev/open/dwz-server/cmd"

	"embed"
)

var (
	Version   string
	GitCommit string
	BuildTime string
)

//go:embed templates/**
var templateFS embed.FS

//go:embed static/**
var staticFs embed.FS

func main() {
	staticFs := map[string]embed.FS{
		"templates":  templateFS,
		"web.static": staticFs,
	}
	cmd.Start(staticFs, Version, GitCommit, BuildTime)
}
