package impl

import (
	"testing"

	"cnb.cool/mliev/dwz/dwz-server/v2/data"
)

// newTestSearcher 用打包在二进制里的默认 xdb（v4 + v6）构造 Searcher。
// 这两份数据来自 ip2region 官方，跨测试环境稳定可用。
func newTestSearcher(t *testing.T) *Searcher {
	t.Helper()
	v4, err := data.ReadIPv4XDB()
	if err != nil {
		t.Fatalf("read embed ipv4 xdb: %v", err)
	}
	v6, err := data.ReadIPv6XDB()
	if err != nil {
		t.Fatalf("read embed ipv6 xdb: %v", err)
	}
	s, err := NewFromBuffers(v4, v6)
	if err != nil {
		t.Fatalf("new searcher: %v", err)
	}
	return s
}

func TestSearcher_Lookup_IPv4PublicIP(t *testing.T) {
	s := newTestSearcher(t)
	// 223.5.5.5 是阿里公共 DNS，归属稳定为中国浙江杭州阿里云。
	// 这里同时断言省/市/ISP 非空 —— 如果 xdb 字段顺序变化（例如 ISP 被错位
	// 解析成国家代码），只断言 Country 会漏掉这种退化。
	region := s.Lookup("223.5.5.5")
	if region.Country != "中国" {
		t.Fatalf("Country 错误：%+v", region)
	}
	if region.Province == "" || region.City == "" || region.ISP == "" {
		t.Fatalf("省/市/ISP 应全部有值，可能字段错位：%+v", region)
	}
	// ISP 不应是 "CN" —— 那是国家代码字段（索引 4），说明解析偏移了。
	if region.ISP == "CN" {
		t.Fatalf("ISP 被错位成国家代码 CN，parseRegion 下标错误：%+v", region)
	}
}

func TestSearcher_Lookup_IPv6PublicIP(t *testing.T) {
	s := newTestSearcher(t)
	// 2400:3200::1 是阿里公共 DNS 的 IPv6 地址，归属稳定为中国。
	region := s.Lookup("2400:3200::1")
	if region.Country == "" {
		t.Fatalf("期望拿到国家，得到空：%+v", region)
	}
	if region.Country != "中国" {
		t.Logf("公网 IPv6 2400:3200::1 归属国：%s（库变更可接受）", region.Country)
	}
}

func TestSearcher_Lookup_PrivateOrInvalid_ReturnsEmpty(t *testing.T) {
	s := newTestSearcher(t)
	cases := []string{"", "127.0.0.1", "10.0.0.1", "192.168.1.1", "::1", "fe80::1", "not-an-ip"}
	for _, ip := range cases {
		if got := s.Lookup(ip); (got != Region{}) {
			t.Errorf("%q 应该返回空 Region，得到 %+v", ip, got)
		}
	}
}

func TestSearcher_OnlyIPv4_IPv6ReturnsEmpty(t *testing.T) {
	v4, err := data.ReadIPv4XDB()
	if err != nil {
		t.Fatalf("read embed ipv4 xdb: %v", err)
	}
	s, err := NewFromBuffers(v4, nil)
	if err != nil {
		t.Fatalf("new searcher: %v", err)
	}
	if got := s.Lookup("2400:3200::1"); (got != Region{}) {
		t.Errorf("缺失 ipv6 数据时 IPv6 查询应返回空 Region，得到 %+v", got)
	}
	// IPv4 仍应正常工作。
	if got := s.Lookup("223.5.5.5"); got.Country == "" {
		t.Errorf("单栈 v4 模式下 IPv4 查询仍应有结果，得到 %+v", got)
	}
}

func TestNewFromBuffers_AllNil_ReturnsError(t *testing.T) {
	if _, err := NewFromBuffers(nil, nil); err == nil {
		t.Fatalf("全部 nil 应该返回错误")
	}
}

func TestNoop_ReturnsEmpty(t *testing.T) {
	if got := (Noop{}).Lookup("223.5.5.5"); (got != Region{}) {
		t.Errorf("Noop 必须始终返回空 Region，得到 %+v", got)
	}
}

func TestParseRegion_ZeroPlaceholderStripped(t *testing.T) {
	// ip2region v3 xdb 标准格式：国家|省份|城市|ISP|国家代码
	r := parseRegion("中国|广东省|深圳市|中国电信|CN")
	if r.Country != "中国" || r.Province != "广东省" || r.City != "深圳市" || r.ISP != "中国电信" {
		t.Fatalf("解析失败：%+v", r)
	}

	// 省 / 市 / ISP 缺失时上游用 "0" 占位（参见 Australia|Queensland|0|0|AU）。
	r2 := parseRegion("Australia|Queensland|0|0|AU")
	if r2.Country != "Australia" || r2.Province != "Queensland" {
		t.Fatalf("country/province 应保留：%+v", r2)
	}
	if r2.City != "" || r2.ISP != "" {
		t.Fatalf("0 占位应被清空：%+v", r2)
	}
}
