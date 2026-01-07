package autoload

import envInterface "cnb.cool/mliev/open/dwz-server/internal/interfaces"

type Http struct {
}

func (receiver Http) InitConfig(helper envInterface.HelperInterface) map[string]any {
	return map[string]any{
		"http.addr":           helper.GetEnv().GetString("http.addr", ":8080"),
		"http.mode":           helper.GetEnv().GetString("http.mode", "release"),
		"http.load_static":    helper.GetEnv().GetBool("http.load_static", true),
		"http.static_mode":    helper.GetEnv().GetString("http.static_mode", "embed"),       // embed 或 disk
		"http.static_dir":     helper.GetEnv().GetString("http.static_dir", "static"),       // 静态资源根目录
		"http.templates_mode": helper.GetEnv().GetString("http.templates_mode", "embed"),    // embed 或 disk
		"http.templates_dir":  helper.GetEnv().GetString("http.templates_dir", "templates"), // 模板目录
	}
}
