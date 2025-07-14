package dto

import "time"

// ClickStatisticListRequest 点击统计列表请求
type ClickStatisticListRequest struct {
	Page        int       `form:"page" binding:"min=1" example:"1"`
	PageSize    int       `form:"page_size" binding:"min=1,max=100" example:"10"`
	ShortLinkID uint64    `form:"short_link_id" example:"1"`                                // 短链接ID筛选
	IP          string    `form:"ip" example:"192.168.1.1"`                                 // IP地址筛选
	Country     string    `form:"country" example:"中国"`                                     // 国家筛选
	City        string    `form:"city" example:"北京"`                                        // 城市筛选
	StartDate   time.Time `form:"start_date" time_format:"2006-01-02" example:"2023-01-01"` // 开始日期
	EndDate     time.Time `form:"end_date" time_format:"2006-01-02" example:"2023-12-31"`   // 结束日期
}

// ClickStatisticDetailResponse 点击统计详细响应
type ClickStatisticDetailResponse struct {
	ID          uint64    `json:"id"`
	ShortLinkID uint64    `json:"short_link_id"`
	ShortCode   string    `json:"short_code,omitempty"`   // 短链代码
	Domain      string    `json:"domain,omitempty"`       // 域名
	OriginalURL string    `json:"original_url,omitempty"` // 原始URL
	IP          string    `json:"ip"`
	UserAgent   string    `json:"user_agent"`
	Referer     string    `json:"referer"`
	QueryParams string    `json:"query_params"`
	Country     string    `json:"country"`
	City        string    `json:"city"`
	ClickDate   time.Time `json:"click_date"`
	CreatedAt   time.Time `json:"created_at"`
}

// ClickStatisticListResponse 点击统计列表响应
type ClickStatisticListResponse struct {
	List  []ClickStatisticDetailResponse `json:"list"`
	Total int64                          `json:"total"`
	Page  int                            `json:"page"`
	Size  int                            `json:"size"`
}

// ClickStatisticAnalysisResponse 点击统计分析响应
type ClickStatisticAnalysisResponse struct {
	TotalClicks  int64              `json:"total_clicks"`  // 总点击数
	UniqueIPs    int64              `json:"unique_ips"`    // 独立IP数
	TopCountries []CountryStatistic `json:"top_countries"` // 热门国家
	TopCities    []CityStatistic    `json:"top_cities"`    // 热门城市
	TopReferers  []RefererStatistic `json:"top_referers"`  // 热门来源
	HourlyStats  []HourlyStatistic  `json:"hourly_stats"`  // 小时统计
	DailyStats   []DailyStatistic   `json:"daily_stats"`   // 日统计
}

// CountryStatistic 国家统计
type CountryStatistic struct {
	Country string `json:"country"`
	Count   int64  `json:"count"`
}

// CityStatistic 城市统计
type CityStatistic struct {
	City  string `json:"city"`
	Count int64  `json:"count"`
}

// RefererStatistic 来源统计
type RefererStatistic struct {
	Referer string `json:"referer"`
	Count   int64  `json:"count"`
}

// HourlyStatistic 小时统计
type HourlyStatistic struct {
	Hour  int   `json:"hour"`
	Count int64 `json:"count"`
}

// DailyStatistic 日统计
type DailyStatistic struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}
