package dto

import "time"

// CreateShortLinkRequest 创建短网址请求
type CreateShortLinkRequest struct {
	OriginalURL string     `json:"original_url" binding:"required,url" example:"https://www.example.com"`
	Domain      string     `json:"domain" example:"https://short.ly"`
	CustomCode  string     `json:"custom_code" example:"abc123"`
	Title       string     `json:"title" example:"示例网站"`
	Description string     `json:"description" example:"这是一个示例网站"`
	ExpireAt    *time.Time `json:"expire_at" example:"2024-12-31T23:59:59Z"`
}

// UpdateShortLinkRequest 更新短网址请求
type UpdateShortLinkRequest struct {
	OriginalURL string     `json:"original_url" binding:"omitempty,url"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	ExpireAt    *time.Time `json:"expire_at"`
	IsActive    *bool      `json:"is_active"`
}

// ShortLinkResponse 短网址响应
type ShortLinkResponse struct {
	ID          uint64     `json:"id"`
	ShortCode   string     `json:"short_code"`
	Domain      string     `json:"domain"`
	ShortURL    string     `json:"short_url"`
	OriginalURL string     `json:"original_url"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	ExpireAt    *time.Time `json:"expire_at"`
	IsActive    bool       `json:"is_active"`
	ClickCount  int64      `json:"click_count"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ShortLinkListRequest 短网址列表请求
type ShortLinkListRequest struct {
	Page     int    `form:"page" binding:"min=1" example:"1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100" example:"10"`
	Domain   string `form:"domain" example:"https://short.ly"`
	Keyword  string `form:"keyword" example:"example"`
}

// ShortLinkListResponse 短网址列表响应
type ShortLinkListResponse struct {
	List  []ShortLinkResponse `json:"list"`
	Total int64               `json:"total"`
	Page  int                 `json:"page"`
	Size  int                 `json:"size"`
}

// ClickStatisticResponse 点击统计响应
type ClickStatisticResponse struct {
	Date       string `json:"date"`
	ClickCount int64  `json:"click_count"`
}

// ShortLinkStatisticResponse 短网址统计响应
type ShortLinkStatisticResponse struct {
	TotalClicks     int64                    `json:"total_clicks"`
	TodayClicks     int64                    `json:"today_clicks"`
	WeekClicks      int64                    `json:"week_clicks"`
	MonthClicks     int64                    `json:"month_clicks"`
	DailyStatistics []ClickStatisticResponse `json:"daily_statistics"`
}

// BatchCreateShortLinkRequest 批量创建短网址请求
type BatchCreateShortLinkRequest struct {
	URLs   []string `json:"urls" binding:"required,min=1,max=100"`
	Domain string   `json:"domain"`
}

// BatchCreateShortLinkResponse 批量创建短网址响应
type BatchCreateShortLinkResponse struct {
	Success []ShortLinkResponse `json:"success"`
	Failed  []BatchFailedItem   `json:"failed"`
}

// BatchFailedItem 批量创建失败项
type BatchFailedItem struct {
	URL   string `json:"url"`
	Error string `json:"error"`
}

// DomainResponse 域名响应
type DomainResponse struct {
	ID                   uint64    `json:"id"`
	Domain               string    `json:"domain"`
	Protocol             string    `json:"protocol" example:"https"`
	SiteName             string    `json:"site_name"`     // 网站名称
	ICPNumber            string    `json:"icp_number"`    // ICP备案号码
	PoliceNumber         string    `json:"police_number"` // 公安备案号码
	IsActive             bool      `json:"is_active"`
	PassQueryParams      bool      `json:"pass_query_params"`
	RandomSuffixLength   int       `json:"random_suffix_length"`   // 随机后缀位数 (0-10)
	EnableChecksum       bool      `json:"enable_checksum"`        // 是否启用校验位
	EnableXorObfuscation bool      `json:"enable_xor_obfuscation"` // 是否启用XOR混淆
	XorSecret            string    `json:"xor_secret"`             // XOR密钥（字符串格式）
	XorRot               int       `json:"xor_rot"`                // 旋转位数
	Description          string    `json:"description"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// DomainRequest 域名请求（通用）
type DomainRequest struct {
	Domain               string  `json:"domain" binding:"required" example:"dwz.do"`                        // 域名
	Protocol             string  `json:"protocol" binding:"required,oneof=http https" example:"https"`      // 协议
	SiteName             string  `json:"site_name"`                                                         // 网站名称
	ICPNumber            string  `json:"icp_number"`                                                        // ICP备案号码
	PoliceNumber         string  `json:"police_number"`                                                     // 公安备案号码
	IsActive             bool    `json:"is_active" example:"true"`                                          // 是否激活
	PassQueryParams      bool    `json:"pass_query_params" example:"false"`                                 // 透传参数
	RandomSuffixLength   *int    `json:"random_suffix_length" binding:"omitempty,min=0,max=10" example:"2"` // 随机后缀位数 (0-10)，使用指针以支持0值
	EnableChecksum       *bool   `json:"enable_checksum" example:"true"`                                    // 是否启用校验位，使用指针以支持false值
	EnableXorObfuscation *bool   `json:"enable_xor_obfuscation" example:"false"`                            // 是否启用XOR混淆，使用指针以支持false值
	XorSecret            *string `json:"xor_secret" example:"11817553067636239985"`                         // XOR密钥（字符串格式），不填写时随机生成
	XorRot               *int    `json:"xor_rot" binding:"omitempty,min=1,max=63" example:"17"`             // 旋转位数 (1-63)，不填写时随机生成
	Description          string  `json:"description" example:"主要短链域名"`                                      // 描述
}

// CreateDomainRequest 创建域名请求
type CreateDomainRequest struct {
	Domain          string `json:"domain" binding:"required,url" example:"https://short.ly"`
	PassQueryParams bool   `json:"pass_query_params" example:"false"`
	Description     string `json:"description" example:"主要短链域名"`
}

// UpdateDomainRequest 更新域名请求
type UpdateDomainRequest struct {
	Domain          string `json:"domain" binding:"required" example:"https://short.ly"`
	IsActive        bool   `json:"is_active" example:"true"`
	PassQueryParams bool   `json:"pass_query_params" example:"false"`
	Description     string `json:"description"`
}

// UpdateStatusDomainRequest 更新域名状态请求
type UpdateStatusDomainRequest struct {
	IsActive bool `json:"is_active" example:"true"`
}

// DomainListResponse 域名列表响应
type DomainListResponse struct {
	List []DomainResponse `json:"list"`
}
