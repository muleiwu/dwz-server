// Package impl 提供 IP 归属地查询的实现，基于 lionsoul2014/ip2region v2 (xdb) 格式。
//
// 设计要点：
//  1. IPRegion 是业务侧面向的窄接口，仅暴露 Lookup；这样数据库未加载成功时
//     assembly 层可以返回 Noop 实现，调用方无需关心失败路径。
//  2. ip2region Searcher 文档自述「非线程安全」（会在每次 Search 中写入
//     ioCount 等状态），因此每个 version 的 Searcher 各加互斥锁串行化 ——
//     xdb 查询本身是内存纯 CPU 操作，亚毫秒级，加锁的竞争开销可以忽略。
//  3. IPv4 与 IPv6 走两份独立 xdb；Searcher 结构体同时持有两个底层 searcher，
//     Lookup 根据解析后的 IP 类型自动分发。任一 version 的数据缺失都会让对应
//     类型的 IP 落到空 Region，不影响另一类型。
//  4. 非公网 / 保留地址 / 解析失败 / 两份 xdb 都缺失的输入一律返回空 Region，
//     不阻塞点击记录主流程；调用方直接写入数据库即可。
package impl

import (
	"errors"
	"net"
	"strings"
	"sync"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

// Region 描述一次归属地查询结果。各字段在无数据时为空字符串。
type Region struct {
	Country  string
	Province string
	City     string
	ISP      string
}

// IPRegion 是对外暴露的查询接口。Lookup 不返回错误：任何异常（解析失败、
// 私网地址、未加载数据库等）都以空 Region 返回，调用方据此决定是否写库。
type IPRegion interface {
	Lookup(ip string) Region
}

// versionedSearcher 包装某一 IP 版本的 xdb searcher，带独立锁。
type versionedSearcher struct {
	mu sync.Mutex
	s  *xdb.Searcher
}

func (v *versionedSearcher) search(ip string) (string, error) {
	if v == nil || v.s == nil {
		return "", nil
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.s.Search(ip)
}

// Searcher 基于 ip2region xdb 内容缓冲实现 IPRegion，同时支持 IPv4 与 IPv6。
type Searcher struct {
	v4 *versionedSearcher
	v6 *versionedSearcher
}

// NewFromBuffers 从两份 xdb 内容字节构造 Searcher。ipv4Content / ipv6Content 均可
// 为 nil —— 对应类型的 IP 此时查询一律返回空 Region，但另一类型仍正常工作。
// 两者都为 nil 时返回错误，由上层决定是否降级为 Noop。
func NewFromBuffers(ipv4Content, ipv6Content []byte) (*Searcher, error) {
	s := &Searcher{}
	if ipv4Content != nil {
		searcher, err := xdb.NewWithBuffer(xdb.IPv4, ipv4Content)
		if err != nil {
			return nil, err
		}
		s.v4 = &versionedSearcher{s: searcher}
	}
	if ipv6Content != nil {
		searcher, err := xdb.NewWithBuffer(xdb.IPv6, ipv6Content)
		if err != nil {
			return nil, err
		}
		s.v6 = &versionedSearcher{s: searcher}
	}
	if s.v4 == nil && s.v6 == nil {
		return nil, errors.New("ip_region: 未提供任何 xdb 数据")
	}
	return s, nil
}

// Lookup 查询 IP 归属地。私网 / 环回 / 无法解析 / 对应版本无数据等情况均返回空 Region。
func (s *Searcher) Lookup(ip string) Region {
	if s == nil {
		return Region{}
	}
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return Region{}
	}
	if parsed.IsLoopback() || parsed.IsPrivate() || parsed.IsUnspecified() ||
		parsed.IsLinkLocalUnicast() || parsed.IsLinkLocalMulticast() {
		return Region{}
	}

	var (
		target   *versionedSearcher
		queryIP  string
		isIPv4   = parsed.To4() != nil
	)
	if isIPv4 {
		target = s.v4
		queryIP = parsed.To4().String()
	} else {
		target = s.v6
		queryIP = parsed.String()
	}

	raw, err := target.search(queryIP)
	if err != nil || raw == "" {
		return Region{}
	}
	return parseRegion(raw)
}

// parseRegion 把 ip2region 的 "国家|区域|省份|城市|ISP" 字符串切成 Region。
// 库中用 "0" 作为「无数据」的占位，这里统一替换为空串，便于上层判空。
func parseRegion(raw string) Region {
	parts := strings.SplitN(raw, "|", 5)
	get := func(idx int) string {
		if idx >= len(parts) {
			return ""
		}
		v := strings.TrimSpace(parts[idx])
		if v == "0" {
			return ""
		}
		return v
	}
	// 格式顺序：country | region | province | city | isp
	return Region{
		Country:  get(0),
		Province: get(2),
		City:     get(3),
		ISP:      get(4),
	}
}

// Noop 是一个始终返回空 Region 的实现，用于数据库加载失败 / 功能禁用时兜底。
type Noop struct{}

// Lookup 永远返回空 Region。
func (Noop) Lookup(string) Region { return Region{} }
