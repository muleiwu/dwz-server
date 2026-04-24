// Package data 封装随二进制分发的静态数据文件。
//
// 当前包含 IP 归属地查询数据库（ip2region xdb 格式）的 IPv4 与 IPv6 两份，
// 通过 //go:embed 一并打包进二进制。部署时零额外步骤即可同时支持 IPv4 与
// IPv6 查询。若运行时设置了 ip_region.db_path_v4 / ip_region.db_path_v6，则
// 对应版本会优先从外部文件加载（便于运维侧单独热更新某一版）。
package data

import (
	"embed"
	"io/fs"
)

//go:embed ip2region_v4.xdb ip2region_v6.xdb
var ipRegionFS embed.FS

const (
	ipv4File = "ip2region_v4.xdb"
	ipv6File = "ip2region_v6.xdb"
)

// IPRegionFS 返回包含两份 xdb 的嵌入文件系统，方便外部按需读取。
func IPRegionFS() embed.FS {
	return ipRegionFS
}

// ReadIPv4XDB 读取内嵌的 IPv4 xdb 内容。
func ReadIPv4XDB() ([]byte, error) {
	return fs.ReadFile(ipRegionFS, ipv4File)
}

// ReadIPv6XDB 读取内嵌的 IPv6 xdb 内容。
func ReadIPv6XDB() ([]byte, error) {
	return fs.ReadFile(ipRegionFS, ipv6File)
}
