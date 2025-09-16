package controller

import (
	"time"

	"cnb.cool/mliev/open/dwz-server/app/dao"
	"cnb.cool/mliev/open/dwz-server/app/model"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"github.com/gin-gonic/gin"
)

type StatisticsController struct {
	BaseResponse
}

// SystemStatistics 系统统计数据
type SystemStatistics struct {
	TotalUsers      int64 `json:"total_users"`       // 总用户数
	ActiveUsers     int64 `json:"active_users"`      // 活跃用户数
	TotalShortLinks int64 `json:"total_short_links"` // 总短链数量
	TotalClicks     int64 `json:"total_clicks"`      // 总点击数
	TodayClicks     int64 `json:"today_clicks"`      // 今日点击数
	WeekClicks      int64 `json:"week_clicks"`       // 本周点击数
	MonthClicks     int64 `json:"month_clicks"`      // 本月点击数
	TotalDomains    int64 `json:"total_domains"`     // 总域名数量
	ActiveDomains   int64 `json:"active_domains"`    // 活跃域名数量
	TotalAbTests    int64 `json:"total_ab_tests"`    // 总AB测试数量
	RunningAbTests  int64 `json:"running_ab_tests"`  // 运行中的AB测试数量
	TotalTokens     int64 `json:"total_tokens"`      // 总Token数量
	ActiveTokens    int64 `json:"active_tokens"`     // 活跃Token数量
}

// RecentActivity 最近活动
type RecentActivity struct {
	ID          uint64    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	User        string    `json:"user"`
}

// TopLink 热门短链
type TopLink struct {
	ID          uint64 `json:"id"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	ClickCount  int64  `json:"click_count"`
	Title       string `json:"title,omitempty"`
}

// DashboardData 仪表盘数据
type DashboardData struct {
	Statistics       SystemStatistics `json:"statistics"`
	RecentActivities []RecentActivity `json:"recent_activities"`
	TopLinks         []TopLink        `json:"top_links"`
}

func (s StatisticsController) GetSystem(c *gin.Context, helper interfaces.HelperInterface) {
	// 创建各种DAO实例
	shortLinkDao := dao.NewShortLinkDao(helper)
	clickStatisticDao := dao.NewClickStatisticDao(helper)
	userDao := dao.NewUserDAO(helper)

	// 获取当前时间以及相关时间范围
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.AddDate(0, 0, 1)

	// 计算本周开始时间（从周一开始）
	weekday := int(now.Weekday())
	if weekday == 0 { // 周日
		weekday = 7
	}
	weekStart := today.AddDate(0, 0, -(weekday - 1))

	// 计算本月开始时间
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// 获取用户统计数据
	totalUsers, _ := userDao.CountAll()
	activeUsers, _ := userDao.CountActive()

	// 获取短链接统计数据
	totalShortLinks, _ := shortLinkDao.CountAll()

	// 获取点击统计数据
	totalClicks, _ := clickStatisticDao.CountAll()
	todayClicks, _ := clickStatisticDao.CountByDateRange(today, tomorrow)
	weekClicks, _ := clickStatisticDao.CountByDateRange(weekStart, tomorrow)
	monthClicks, _ := clickStatisticDao.CountByDateRange(monthStart, tomorrow)

	// 获取域名统计数据
	var totalDomains int64
	helper.GetDatabase().Model(&model.Domain{}).Count(&totalDomains)

	var activeDomains int64
	helper.GetDatabase().Model(&model.Domain{}).Where("status = ? AND deleted_at IS NULL", 1).Count(&activeDomains)

	// 获取AB测试统计数据
	var totalAbTests int64
	helper.GetDatabase().Model(&model.ABTest{}).Count(&totalAbTests)

	var runningAbTests int64
	helper.GetDatabase().Model(&model.ABTest{}).Where("status = ? AND deleted_at IS NULL", 1).Count(&runningAbTests)

	// 获取Token统计数据
	var totalTokens int64
	helper.GetDatabase().Model(&model.UserToken{}).Count(&totalTokens)

	var activeTokens int64
	helper.GetDatabase().Model(&model.UserToken{}).Where("status = ? AND deleted_at IS NULL", 1).Count(&activeTokens)

	// 构建系统统计数据
	statistics := SystemStatistics{
		TotalUsers:      totalUsers,
		ActiveUsers:     activeUsers,
		TotalShortLinks: totalShortLinks,
		TotalClicks:     totalClicks,
		TodayClicks:     todayClicks,
		WeekClicks:      weekClicks,
		MonthClicks:     monthClicks,
		TotalDomains:    totalDomains,
		ActiveDomains:   activeDomains,
		TotalAbTests:    totalAbTests,
		RunningAbTests:  runningAbTests,
		TotalTokens:     totalTokens,
		ActiveTokens:    activeTokens,
	}

	// 返回数据
	s.Success(c, statistics)
}

func (s StatisticsController) GetDashboard(c *gin.Context, helper interfaces.HelperInterface) {
	// 创建短链接服务实例
	shortLinkDao := dao.NewShortLinkDao(helper)
	clickStatisticDao := dao.NewClickStatisticDao(helper)
	operationLogDao := dao.NewOperationLogDAO(helper)
	userDao := dao.NewUserDAO(helper)

	// 获取当天日期范围
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.AddDate(0, 0, 1)

	// 获取系统统计数据
	totalLinks, _ := shortLinkDao.CountAll()
	totalClicks, _ := clickStatisticDao.CountAll()
	todayClicks, _ := clickStatisticDao.CountByDateRange(today, tomorrow)
	totalUsers, _ := userDao.CountAll()
	activeUsers, _ := userDao.CountActive()

	// 获取最近活动
	var status *int8
	recentLogs, _, _ := operationLogDao.GetList(0, 10, nil, "", "", "", "", status, nil, nil)
	recentActivities := make([]RecentActivity, 0)
	for _, log := range recentLogs {
		activity := RecentActivity{
			ID:          log.ID,
			CreatedAt:   log.CreatedAt,
			Description: log.Operation + " " + log.Resource,
			Type:        log.Method,
			User:        log.Username,
		}
		recentActivities = append(recentActivities, activity)
	}

	// 获取热门短链
	topShortLinks, _ := shortLinkDao.GetTopClicked(10)
	topLinks := make([]TopLink, 0)
	for _, link := range topShortLinks {
		topLink := TopLink{
			ID:          link.ID,
			ShortURL:    link.GetFullURL(),
			OriginalURL: link.OriginalURL,
			ClickCount:  int64(link.ClickCount),
			Title:       link.Title,
		}
		topLinks = append(topLinks, topLink)
	}

	// 构建仪表盘数据
	dashboardData := DashboardData{
		Statistics: SystemStatistics{
			TotalUsers:      totalUsers,
			ActiveUsers:     activeUsers,
			TotalShortLinks: totalLinks,
			TotalClicks:     totalClicks,
			TodayClicks:     todayClicks,
			WeekClicks:      0,
			MonthClicks:     0,
			TotalDomains:    0,
			ActiveDomains:   0,
			TotalAbTests:    0,
			RunningAbTests:  0,
			TotalTokens:     0,
			ActiveTokens:    0,
		},
		RecentActivities: recentActivities,
		TopLinks:         topLinks,
	}

	// 返回数据
	s.Success(c, dashboardData)
}

func (s StatisticsController) GetShortLinks(c *gin.Context, helper interfaces.HelperInterface) {
	// 创建短链接服务实例
	// shortLinkService := service.NewShortLinkService(helper, c)

	// days := c.DefaultQuery("days", "30")

}
