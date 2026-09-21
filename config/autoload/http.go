package autoload

import (
	"slices"

	"cnb.cool/mliev/open/go-web/pkg/helper"
)

type Http struct{}

func (Http) InitConfig() map[string]any {
	staticDirs := helper.GetEnv().GetStringSlice("http.static_dir", []string{"admin", "assets"})
	// The public homepage also needs its assets on existing installations that
	// explicitly configured only admin. A narrower home mount already serves them.
	if !slices.Contains(staticDirs, "assets") && !slices.Contains(staticDirs, "assets/home") {
		staticDirs = append(slices.Clone(staticDirs), "assets")
	}

	return map[string]any{
		"http.addr":        helper.GetEnv().GetString("http.addr", ":8080"),
		"http.mode":        helper.GetEnv().GetString("http.mode", "release"),
		"http.load_static": helper.GetEnv().GetBool("http.load_static", true),
		"http.static_mode": helper.GetEnv().GetString("http.static_mode", "embed"),
		// 每个条目对应 embed 里 static/<name>/ 子目录，URL 前缀就是 /<name>。
		// admin 提供后台，assets 提供公开首页的图片、样式和脚本。
		"http.static_dir":     staticDirs,
		"http.templates_mode": helper.GetEnv().GetString("http.templates_mode", "embed"),
		"http.templates_dir":  helper.GetEnv().GetString("http.templates_dir", "templates"),
	}
}
