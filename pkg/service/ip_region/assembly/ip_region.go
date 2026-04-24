// Package assembly 组装 IP 归属地查询服务到依赖容器。
package assembly

import (
	"os"
	"reflect"

	"cnb.cool/mliev/dwz/dwz-server/v2/data"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/ip_region/impl"
	"cnb.cool/mliev/open/go-web/pkg/container"
	"github.com/muleiwu/gsr"
)

// IPRegion assembly：加载 IPv4 与 IPv6 两份 xdb，构造同时支持双栈查询的
// Searcher。每份 xdb 优先读取 ip_region.db_path_{v4,v6} 指向的外部文件，
// 失败或未配置时回落到二进制内嵌的 data/ip2region_v{4,6}.xdb。任一版本加载
// 成功即可正常工作；两者都加载失败才退到 impl.Noop，确保不阻塞 HTTP 启动 ——
// IP 归属地缺失仅影响分析面板，点击记录仍正常落库。
type IPRegion struct{}

func (IPRegion) Type() reflect.Type { return reflect.TypeFor[impl.IPRegion]() }

func (IPRegion) DependsOn() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[gsr.Provider](),
		reflect.TypeFor[gsr.Logger](),
	}
}

func (IPRegion) Assembly() (any, error) {
	logger := container.MustGet[gsr.Logger]()
	cfg := container.MustGet[gsr.Provider]()

	if !cfg.GetBool("ip_region.enabled", true) {
		logger.Info("[ip_region] 已禁用，使用 Noop 实现")
		return impl.Noop{}, nil
	}

	v4Content, v4Source := loadXDB(cfg, logger, "ip_region.db_path_v4", data.ReadIPv4XDB, "ipv4")
	v6Content, v6Source := loadXDB(cfg, logger, "ip_region.db_path_v6", data.ReadIPv6XDB, "ipv6")
	if v4Content == nil && v6Content == nil {
		logger.Warn("[ip_region] 未能加载任何 xdb 数据，归属地字段将为空")
		return impl.Noop{}, nil
	}

	searcher, err := impl.NewFromBuffers(v4Content, v6Content)
	if err != nil {
		logger.Error("[ip_region] 构建 searcher 失败，降级为 Noop：" + err.Error())
		return impl.Noop{}, nil
	}
	logger.Info("[ip_region] 归属地查询器就绪，ipv4 来源：" + v4Source + "，ipv6 来源：" + v6Source)
	return searcher, nil
}

// loadXDB 选择某一 IP 版本的 xdb 数据源：优先外部 pathKey（便于运维替换更新
// 过的 xdb），否则使用 //go:embed 打包在二进制中的默认数据。任一失败都会 log
// warn 但不终止 —— 上层决定是否降级。label 用于日志区分 ipv4/ipv6。
func loadXDB(cfg gsr.Provider, logger gsr.Logger, pathKey string, readEmbed func() ([]byte, error), label string) ([]byte, string) {
	if path := cfg.GetString(pathKey, ""); path != "" {
		bytes, err := os.ReadFile(path)
		if err == nil {
			return bytes, "file:" + path
		}
		logger.Warn("[ip_region] 读取 " + pathKey + " 失败，将尝试内嵌 " + label + " 数据：" + err.Error())
	}

	bytes, err := readEmbed()
	if err != nil {
		logger.Error("[ip_region] 读取内嵌 " + label + " xdb 失败：" + err.Error())
		return nil, ""
	}
	return bytes, "embed"
}
