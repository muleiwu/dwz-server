package migrations

import "embed"

//go:embed all:mysql all:postgresql all:sqlite
var FS embed.FS
