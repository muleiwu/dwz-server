package autoload

import (
	"cnb.cool/mliev/open/go-web/pkg/helper"
)

type Http struct{}

func (Http) InitConfig() map[string]any {
	return map[string]any{
		"http.addr":           helper.GetEnv().GetString("http.addr", ":8080"),
		"http.mode":           helper.GetEnv().GetString("http.mode", "release"),
		"http.load_static":    helper.GetEnv().GetBool("http.load_static", true),
		"http.static_mode":    helper.GetEnv().GetString("http.static_mode", "embed"),
		"http.static_dir":     helper.GetEnv().GetString("http.static_dir", "static"),
		"http.templates_mode": helper.GetEnv().GetString("http.templates_mode", "embed"),
		"http.templates_dir":  helper.GetEnv().GetString("http.templates_dir", "templates"),
	}
}
