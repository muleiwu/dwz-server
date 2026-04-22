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
		// 每个条目对应 embed 里 static/<name>/ 子目录，URL 前缀就是 /<name>。
		// 默认只暴露 admin 后台界面，其它子目录可用环境变量追加。
		"http.static_dir":     helper.GetEnv().GetStringSlice("http.static_dir", []string{"admin"}),
		"http.templates_mode": helper.GetEnv().GetString("http.templates_mode", "embed"),
		"http.templates_dir":  helper.GetEnv().GetString("http.templates_dir", "templates"),
	}
}
