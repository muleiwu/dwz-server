package autoload

import "cnb.cool/mliev/open/go-web/pkg/helper"

// IPRegion 暴露归属地查询相关的可配置项。
//
//   - ip_region.enabled     = true  启用查询；false 则始终返回空归属地
//   - ip_region.db_path_v4  = ""    非空时从外部 IPv4 xdb 文件加载（优先级高于内嵌）
//   - ip_region.db_path_v6  = ""    非空时从外部 IPv6 xdb 文件加载（优先级高于内嵌）
type IPRegion struct{}

func (IPRegion) InitConfig() map[string]any {
	env := helper.GetEnv()
	return map[string]any{
		"ip_region.enabled":    env.GetBool("ip_region.enabled", true),
		"ip_region.db_path_v4": env.GetString("ip_region.db_path_v4", ""),
		"ip_region.db_path_v6": env.GetString("ip_region.db_path_v6", ""),
	}
}
