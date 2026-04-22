package autoload

import "embed"

type StaticFs struct{}

// InitConfig provides a placeholder for the static FS map.
// The real value is injected by cmd.WithTemplateFs / cmd.WithWebStaticFs in main.go.
func (StaticFs) InitConfig() map[string]any {
	return map[string]any{
		"static.fs": map[string]embed.FS{},
	}
}
