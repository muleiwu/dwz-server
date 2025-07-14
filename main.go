package main

import (
	"cnb.cool/mliev/open/dwz-server/cmd"
	"embed"
)

//go:embed templates/*
var templateFS embed.FS

func main() {
	cmd.Start(templateFS)
}
