package dto

import "time"

// ABTestClickStatisticListRequest AB测试点击统计列表请求
type ABTestClickStatisticListRequest struct {
	Page        int       `form:"page" binding:"min=1" example:"1"`
	PageSize    int       `form:"page_size" binding:"min=1,max=100" example:"10"`
	ABTestID    uint64    `form:"ab_test_id" example:"1"`                                   // AB测试ID筛选
	VariantID   uint64    `form:"variant_id" example:"1"`                                   // 版本ID筛选
	ShortLinkID uint64    `form:"short_link_id" example:"1"`                                // 短链接ID筛选
	IP          string    `form:"ip" example:"192.168.1.1"`                                 // IP地址筛选
	Country     string    `form:"country" example:"中国"`                                     // 国家筛选
	City        string    `form:"city" example:"北京"`                                        // 城市筛选
	StartDate   time.Time `form:"start_date" time_format:"2006-01-02" example:"2023-01-01"` // 开始日期
	EndDate     time.Time `form:"end_date" time_format:"2006-01-02" example:"2023-12-31"`   // 结束日期
}

// ABTestClickStatisticDetailResponse AB测试点击统计详细响应
type ABTestClickStatisticDetailResponse struct {
	ID          uint64    `json:"id"`
	ABTestID    uint64    `json:"ab_test_id"`
	ABTestName  string    `json:"ab_test_name,omitempty"` // AB测试名称
	VariantID   uint64    `json:"variant_id"`
	VariantName string    `json:"variant_name,omitempty"` // 版本名称
	TargetURL   string    `json:"target_url,omitempty"`   // 目标URL
	ShortLinkID uint64    `json:"short_link_id"`
	ShortCode   string    `json:"short_code,omitempty"` // 短链代码
	Domain      string    `json:"domain,omitempty"`     // 域名
	IP          string    `json:"ip"`
	UserAgent   string    `json:"user_agent"`
	Referer     string    `json:"referer"`
	QueryParams string    `json:"query_params"`
	Country     string    `json:"country"`
	City        string    `json:"city"`
	SessionID   string    `json:"session_id"`
	ClickDate   time.Time `json:"click_date"`
	CreatedAt   time.Time `json:"created_at"`
}

// ABTestClickStatisticListResponse AB测试点击统计列表响应
type ABTestClickStatisticListResponse struct {
	List  []ABTestClickStatisticDetailResponse `json:"list"`
	Total int64                                `json:"total"`
	Page  int                                  `json:"page"`
	Size  int                                  `json:"size"`
}

// ABTestClickStatisticAnalysisResponse AB测试点击统计分析响应
type ABTestClickStatisticAnalysisResponse struct {
	TotalClicks    int64                          `json:"total_clicks"`    // 总点击数
	UniqueIPs      int64                          `json:"unique_ips"`      // 独立IP数
	UniqueSessions int64                          `json:"unique_sessions"` // 独立会话数
	VariantStats   []ABTestVariantStatistic       `json:"variant_stats"`   // 版本统计
	TopCountries   []CountryStatistic             `json:"top_countries"`   // 热门国家
	TopCities      []CityStatistic                `json:"top_cities"`      // 热门城市
	TopReferers    []RefererStatistic             `json:"top_referers"`    // 热门来源
	HourlyStats    []HourlyStatistic              `json:"hourly_stats"`    // 小时统计
	DailyStats     []DailyStatistic               `json:"daily_stats"`     // 日统计
	ConversionRate map[string]ConversionRateStats `json:"conversion_rate"` // 转化率统计
}

// ABTestVariantStatistic AB测试版本统计
type ABTestVariantStatistic struct {
	VariantID      uint64  `json:"variant_id"`
	VariantName    string  `json:"variant_name"`
	TargetURL      string  `json:"target_url"`
	ClickCount     int64   `json:"click_count"`
	UniqueClicks   int64   `json:"unique_clicks"`   // 去重点击数（按session_id去重）
	TrafficPercent float64 `json:"traffic_percent"` // 流量占比
}

// ConversionRateStats 转化率统计
type ConversionRateStats struct {
	Impressions    int64   `json:"impressions"`     // 展示数
	Clicks         int64   `json:"clicks"`          // 点击数
	ConversionRate float64 `json:"conversion_rate"` // 转化率
}
