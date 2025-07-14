package dto

import "time"

// CreateABTestRequest 创建AB测试请求
type CreateABTestRequest struct {
	ShortLinkID  uint64                       `json:"short_link_id" binding:"required"`
	Name         string                       `json:"name" binding:"required" example:"首页Banner测试"`
	Description  string                       `json:"description" example:"测试不同banner的点击率"`
	TrafficSplit string                       `json:"traffic_split" example:"equal"` // equal, weighted, custom
	StartTime    *time.Time                   `json:"start_time"`
	EndTime      *time.Time                   `json:"end_time"`
	Variants     []CreateABTestVariantRequest `json:"variants" binding:"required,min=2"`
}

// CreateABTestVariantRequest 创建AB测试变体请求
type CreateABTestVariantRequest struct {
	Name        string `json:"name" binding:"required" example:"版本A"`
	TargetURL   string `json:"target_url" binding:"required,url" example:"https://example.com/page-a"`
	Weight      int    `json:"weight" example:"50"` // 权重百分比
	IsControl   bool   `json:"is_control" example:"false"`
	Description string `json:"description" example:"原始版本"`
}

// UpdateABTestRequest 更新AB测试请求
type UpdateABTestRequest struct {
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	Status       string     `json:"status" example:"running"` // draft, running, paused, completed
	TrafficSplit string     `json:"traffic_split"`
	StartTime    *time.Time `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	IsActive     *bool      `json:"is_active"`
}

// UpdateABTestVariantRequest 更新AB测试变体请求
type UpdateABTestVariantRequest struct {
	Name        string `json:"name"`
	TargetURL   string `json:"target_url" binding:"omitempty,url"`
	Weight      int    `json:"weight"`
	IsControl   bool   `json:"is_control"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

// ABTestResponse AB测试响应
type ABTestResponse struct {
	ID           uint64                  `json:"id"`
	ShortLinkID  uint64                  `json:"short_link_id"`
	Name         string                  `json:"name"`
	Description  string                  `json:"description"`
	Status       string                  `json:"status"`
	TrafficSplit string                  `json:"traffic_split"`
	StartTime    *time.Time              `json:"start_time"`
	EndTime      *time.Time              `json:"end_time"`
	IsActive     bool                    `json:"is_active"`
	CreatedAt    time.Time               `json:"created_at"`
	UpdatedAt    time.Time               `json:"updated_at"`
	Variants     []ABTestVariantResponse `json:"variants"`
}

// ABTestVariantResponse AB测试变体响应
type ABTestVariantResponse struct {
	ID          uint64    `json:"id"`
	ABTestID    uint64    `json:"ab_test_id"`
	Name        string    `json:"name"`
	TargetURL   string    `json:"target_url"`
	Weight      int       `json:"weight"`
	IsControl   bool      `json:"is_control"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ABTestListRequest AB测试列表请求
type ABTestListRequest struct {
	Page        int    `form:"page" binding:"min=1" example:"1"`
	PageSize    int    `form:"page_size" binding:"min=1,max=100" example:"10"`
	ShortLinkID uint64 `form:"short_link_id"`
	Status      string `form:"status" example:"running"`
}

// ABTestListResponse AB测试列表响应
type ABTestListResponse struct {
	List  []ABTestResponse `json:"list"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
}

// ABTestStatisticResponse AB测试统计响应
type ABTestStatisticResponse struct {
	ABTestID       uint64                      `json:"ab_test_id"`
	TotalClicks    int64                       `json:"total_clicks"`
	VariantStats   []ABTestVariantStatResponse `json:"variant_stats"`
	DailyStats     []ABTestDailyStatResponse   `json:"daily_stats"`
	ConversionRate float64                     `json:"conversion_rate"`
	WinningVariant *ABTestVariantResponse      `json:"winning_variant,omitempty"`
}

// ABTestVariantStatResponse AB测试变体统计响应
type ABTestVariantStatResponse struct {
	Variant        ABTestVariantResponse `json:"variant"`
	ClickCount     int64                 `json:"click_count"`
	ConversionRate float64               `json:"conversion_rate"`
	Percentage     float64               `json:"percentage"`
}

// ABTestDailyStatResponse AB测试每日统计响应
type ABTestDailyStatResponse struct {
	Date     string           `json:"date"`
	Variants map[uint64]int64 `json:"variants"` // variant_id -> click_count
}

// StartABTestRequest 启动AB测试请求
type StartABTestRequest struct {
	StartTime *time.Time `json:"start_time"`
}

// StopABTestRequest 停止AB测试请求
type StopABTestRequest struct {
	EndTime *time.Time `json:"end_time"`
}

// ABTestRedirectInfo AB测试重定向信息（内部使用）
type ABTestRedirectInfo struct {
	ABTestID    uint64 `json:"ab_test_id"`
	VariantID   uint64 `json:"variant_id"`
	TargetURL   string `json:"target_url"`
	VariantName string `json:"variant_name"`
	SessionID   string `json:"session_id"`
}
