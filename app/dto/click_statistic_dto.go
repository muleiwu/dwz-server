package dto

import "time"

// ClickStatisticListRequest 点击统计列表请求
type ClickStatisticListRequest struct {
	Page        int       `form:"page" binding:"min=1" example:"1"`
	PageSize    int       `form:"page_size" binding:"min=1,max=100" example:"10"`
	ShortLinkID uint64    `form:"short_link_id" example:"1"` // 短链接ID筛选
	CampaignID  uint64    `form:"campaign_id"`
	RouteID     uint64    `form:"route_id"`
	TagID       uint64    `form:"tag_id"`
	DeviceType  string    `form:"device_type"`
	IsBot       *bool     `form:"is_bot"`
	IP          string    `form:"ip" example:"192.168.1.1"`                                 // IP地址筛选
	Country     string    `form:"country" example:"中国"`                                     // 国家筛选
	Province    string    `form:"province" example:"广东省"`                                   // 省份筛选
	City        string    `form:"city" example:"北京"`                                        // 城市筛选
	ISP         string    `form:"isp" example:"电信"`                                         // 运营商筛选
	StartDate   time.Time `form:"start_date" time_format:"2006-01-02" example:"2023-01-01"` // 开始日期
	EndDate     time.Time `form:"end_date" time_format:"2006-01-02" example:"2023-12-31"`   // 结束日期
}

// ClickStatisticDetailResponse 点击统计详细响应
type ClickStatisticDetailResponse struct {
	ID          uint64    `json:"id"`
	WorkspaceID uint64    `json:"workspace_id"`
	CampaignID  *uint64   `json:"campaign_id"`
	RouteID     *uint64   `json:"route_id"`
	RouteName   string    `json:"route_name"`
	ShortLinkID uint64    `json:"short_link_id"`
	ShortCode   string    `json:"short_code,omitempty"`   // 短链代码
	Domain      string    `json:"domain,omitempty"`       // 域名
	OriginalURL string    `json:"original_url,omitempty"` // 原始URL
	IP          string    `json:"ip"`
	UserAgent   string    `json:"user_agent"`
	Referer     string    `json:"referer"`
	QueryParams string    `json:"query_params"`
	UTMSource   string    `json:"utm_source"`
	UTMMedium   string    `json:"utm_medium"`
	UTMCampaign string    `json:"utm_campaign"`
	UTMTerm     string    `json:"utm_term"`
	UTMContent  string    `json:"utm_content"`
	DeviceType  string    `json:"device_type"`
	Browser     string    `json:"browser"`
	OS          string    `json:"os"`
	IsBot       bool      `json:"is_bot"`
	BotName     string    `json:"bot_name"`
	Country     string    `json:"country"`
	Province    string    `json:"province"`
	City        string    `json:"city"`
	ISP         string    `json:"isp"`
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
	TotalClicks     int64               `json:"total_clicks"`  // 总点击数
	UniqueIPs       int64               `json:"unique_ips"`    // 独立IP数
	TopCountries    []CountryStatistic  `json:"top_countries"` // 热门国家
	TopProvinces    []ProvinceStatistic `json:"top_provinces"` // 热门省份
	TopCities       []CityStatistic     `json:"top_cities"`    // 热门城市
	TopISPs         []ISPStatistic      `json:"top_isps"`      // 热门运营商
	TopReferers     []RefererStatistic  `json:"top_referers"`  // 热门来源
	TopDevices      []DeviceStatistic   `json:"top_devices"`
	TopBrowsers     []BrowserStatistic  `json:"top_browsers"`
	TopOS           []OSStatistic       `json:"top_os"`
	BotStats        BotStatistic        `json:"bot_stats"`
	TopUTMSources   []UTMStatistic      `json:"top_utm_sources"`
	TopUTMCampaigns []UTMStatistic      `json:"top_utm_campaigns"`
	TopRoutes       []RouteStatistic    `json:"top_routes"`
	HourlyStats     []HourlyStatistic   `json:"hourly_stats"` // 小时统计
	DailyStats      []DailyStatistic    `json:"daily_stats"`  // 日统计
}

// ClickStatisticGeoAnalysisResponse 地理访问聚合响应
type ClickStatisticGeoAnalysisResponse struct {
	TotalClicks int64                `json:"total_clicks"`
	UniqueIPs   int64                `json:"unique_ips"`
	Level       string               `json:"level"`
	Country     string               `json:"country,omitempty"`
	Province    string               `json:"province,omitempty"`
	Regions     []GeoRegionStatistic `json:"regions"`
}

// GeoRegionStatistic 地理区域访问统计
type GeoRegionStatistic struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// CountryStatistic 国家统计
type CountryStatistic struct {
	Country string `json:"country"`
	Count   int64  `json:"count"`
}

// ProvinceStatistic 省份统计
type ProvinceStatistic struct {
	Province string `json:"province"`
	Count    int64  `json:"count"`
}

// CityStatistic 城市统计
type CityStatistic struct {
	City  string `json:"city"`
	Count int64  `json:"count"`
}

// ISPStatistic 运营商统计
type ISPStatistic struct {
	ISP   string `json:"isp"`
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

type DeviceStatistic struct {
	DeviceType string `json:"device_type"`
	Count      int64  `json:"count"`
}

type BrowserStatistic struct {
	Browser string `json:"browser"`
	Count   int64  `json:"count"`
}

type OSStatistic struct {
	OS    string `json:"os"`
	Count int64  `json:"count"`
}

type BotStatistic struct {
	BotClicks     int64 `json:"bot_clicks"`
	HumanClicks   int64 `json:"human_clicks"`
	UnknownClicks int64 `json:"unknown_clicks"`
}

type UTMStatistic struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}

type RouteStatistic struct {
	RouteID   uint64 `json:"route_id"`
	RouteName string `json:"route_name"`
	Count     int64  `json:"count"`
}
