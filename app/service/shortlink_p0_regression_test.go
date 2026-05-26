package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
	ipRegionImpl "cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/ip_region/impl"
	"github.com/glebarez/sqlite"
	"github.com/muleiwu/gsr"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func TestShortLinkP0RegressionCreateRedirectStatisticsAndABSplit(t *testing.T) {
	helper := newShortLinkRegressionHelper(t)
	db := helper.GetDatabase()

	if err := db.Create(&model.Domain{
		WorkspaceID:     1,
		Protocol:        "https",
		Domain:          "dwz.do",
		PassQueryParams: true,
		IsActive:        true,
	}).Error; err != nil {
		t.Fatalf("seed domain: %v", err)
	}

	campaign := model.Campaign{WorkspaceID: 1, Name: "P0 Launch", Status: model.CampaignStatusActive}
	if err := db.Create(&campaign).Error; err != nil {
		t.Fatalf("seed campaign: %v", err)
	}
	tag := model.Tag{WorkspaceID: 1, Name: "docs", Color: "#2563eb"}
	if err := db.Create(&tag).Error; err != nil {
		t.Fatalf("seed tag: %v", err)
	}

	shortLinkSvc := NewShortLinkService(helper, context.Background())
	createResp, err := shortLinkSvc.CreateShortLinkInWorkspace(&dto.CreateShortLinkRequest{
		OriginalURL: "https://example.com/landing?utm_source=old&keep=1",
		Domain:      "dwz.do",
		CustomCode:  "p0",
		CampaignID:  &campaign.ID,
		TagIDs:      []uint64{tag.ID},
		UTMSource:   "newsletter",
		UTMMedium:   "email",
		UTMCampaign: "launch",
		Notes:       "regression",
	}, "203.0.113.10", 1, 7)
	if err != nil {
		t.Fatalf("create short link: %v", err)
	}
	if createResp.WorkspaceID != 1 || createResp.CreatedBy == nil || *createResp.CreatedBy != 7 {
		t.Fatalf("workspace/creator not set on response: %+v", createResp)
	}
	if createResp.CampaignID == nil || *createResp.CampaignID != campaign.ID || len(createResp.Tags) != 1 {
		t.Fatalf("campaign/tag response mismatch: %+v", createResp)
	}
	if !strings.Contains(createResp.OriginalURL, "utm_source=newsletter") ||
		!strings.Contains(createResp.OriginalURL, "utm_medium=email") ||
		!strings.Contains(createResp.OriginalURL, "utm_campaign=launch") ||
		!strings.Contains(createResp.OriginalURL, "keep=1") {
		t.Fatalf("UTM fields were not merged into original_url: %s", createResp.OriginalURL)
	}

	if _, err := shortLinkSvc.CreateShortLinkInWorkspace(&dto.CreateShortLinkRequest{
		OriginalURL: "https://example.com",
		Domain:      "https://dwz.do",
		CustomCode:  "bad",
	}, "203.0.113.10", 1, 7); err == nil {
		t.Fatal("expected protocol-prefixed domain to be rejected")
	}

	targetURL, err := shortLinkSvc.RedirectShortLinkWithQuery(
		"dwz.do",
		"p0",
		"8.8.8.8",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
		"https://ref.example",
		"click_id=abc",
	)
	if err != nil {
		t.Fatalf("redirect short link: %v", err)
	}
	if !strings.Contains(targetURL, "click_id=abc") || !strings.Contains(targetURL, "utm_source=newsletter") {
		t.Fatalf("redirect target did not preserve expected query values: %s", targetURL)
	}

	waitForModelCount(t, db, &model.ClickStatistic{}, "short_link_id = ?", createResp.ID, 1)
	var stat model.ClickStatistic
	if err := db.Where("short_link_id = ?", createResp.ID).First(&stat).Error; err != nil {
		t.Fatalf("load click statistic: %v", err)
	}
	if stat.WorkspaceID != 1 || stat.CampaignID == nil || *stat.CampaignID != campaign.ID {
		t.Fatalf("statistic did not copy workspace/campaign: %+v", stat)
	}
	if stat.UTMSource != "newsletter" || stat.UTMMedium != "email" || stat.UTMCampaign != "launch" {
		t.Fatalf("statistic did not copy UTM fields: %+v", stat)
	}
	if stat.DeviceType == "" || stat.DeviceType == "unknown" || stat.Browser == "" || stat.OS == "" {
		t.Fatalf("user-agent metadata was not parsed: %+v", stat)
	}

	abSvc := NewABTestService(helper)
	abResp, err := abSvc.CreateABTestInWorkspace(&dto.CreateABTestRequest{
		ShortLinkID:  createResp.ID,
		Name:         "redirect split",
		TrafficSplit: "weighted",
		Variants: []dto.CreateABTestVariantRequest{
			{Name: "A", TargetURL: "https://example.com/a", Weight: 100, IsControl: true},
			{Name: "B", TargetURL: "https://example.com/b", Weight: 0, IsControl: false},
		},
	}, 1)
	if err == nil || !strings.Contains(err.Error(), "权重") {
		t.Fatalf("expected invalid variant weight error, got resp=%+v err=%v", abResp, err)
	}

	abResp, err = abSvc.CreateABTestInWorkspace(&dto.CreateABTestRequest{
		ShortLinkID:  createResp.ID,
		Name:         "redirect split",
		TrafficSplit: "weighted",
		Variants: []dto.CreateABTestVariantRequest{
			{Name: "A", TargetURL: "https://example.com/a", Weight: 60, IsControl: true},
			{Name: "B", TargetURL: "https://example.com/b", Weight: 40, IsControl: false},
		},
	}, 1)
	if err != nil {
		t.Fatalf("create ab test: %v", err)
	}
	if _, err := abSvc.StartABTestInWorkspace(abResp.ID, &dto.StartABTestRequest{}, 1); err != nil {
		t.Fatalf("start ab test: %v", err)
	}

	abTargetURL, err := shortLinkSvc.RedirectShortLinkWithQuery(
		"dwz.do",
		"p0",
		"8.8.4.4",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
		"https://ref.example",
		"",
	)
	if err != nil {
		t.Fatalf("redirect with ab test: %v", err)
	}
	if !strings.HasPrefix(abTargetURL, "https://example.com/") || strings.Contains(abTargetURL, "landing") {
		t.Fatalf("A/B redirect did not use a variant URL: %s", abTargetURL)
	}
	parsedABURL, err := url.Parse(abTargetURL)
	if err != nil {
		t.Fatalf("parse A/B target URL: %v", err)
	}
	feedbackToken := parsedABURL.Query().Get(ABTestFeedbackQueryParam)
	if feedbackToken == "" {
		t.Fatalf("A/B redirect did not append feedback token: %s", abTargetURL)
	}

	waitForModelCount(t, db, &model.ABTestClickStatistic{}, "ab_test_id = ?", abResp.ID, 1)
	var abStat model.ABTestClickStatistic
	if err := db.Where("ab_test_id = ?", abResp.ID).First(&abStat).Error; err != nil {
		t.Fatalf("load ab click statistic: %v", err)
	}
	if abStat.WorkspaceID != 1 || abStat.CampaignID == nil || *abStat.CampaignID != campaign.ID {
		t.Fatalf("ab statistic did not copy workspace/campaign: %+v", abStat)
	}
	if abStat.UTMSource != "newsletter" || abStat.DeviceType == "" {
		t.Fatalf("ab statistic did not copy traffic metadata: %+v", abStat)
	}

	value := 42.5
	feedback, err := abSvc.RecordABTestFeedback(&dto.ABTestFeedbackRequest{
		FeedbackToken: feedbackToken,
		EventID:       "order-abc",
		Value:         &value,
		Currency:      "usd",
		Metadata:      json.RawMessage(`{"plan":"pro"}`),
	}, "8.8.4.4", "Mozilla/5.0", "https://ref.example")
	if err != nil {
		t.Fatalf("record ab feedback: %v", err)
	}
	if feedback.Duplicate || feedback.ABTestID != abResp.ID || feedback.ShortLinkID != createResp.ID || feedback.SessionID == "" {
		t.Fatalf("unexpected feedback response: %+v", feedback)
	}
	duplicate, err := abSvc.RecordABTestFeedback(&dto.ABTestFeedbackRequest{
		FeedbackToken: feedbackToken,
		EventID:       "order-abc",
	}, "8.8.4.4", "Mozilla/5.0", "https://ref.example")
	if err != nil {
		t.Fatalf("record duplicate ab feedback: %v", err)
	}
	if !duplicate.Duplicate || duplicate.ID != feedback.ID {
		t.Fatalf("expected idempotent duplicate response, got %+v", duplicate)
	}
	if _, err := abSvc.RecordABTestFeedback(&dto.ABTestFeedbackRequest{
		FeedbackToken: feedbackToken + "x",
		EventID:       "tampered",
	}, "8.8.4.4", "Mozilla/5.0", ""); !errors.Is(err, ErrABTestFeedbackInvalidToken) {
		t.Fatalf("expected invalid token error, got %v", err)
	}
	expiredToken, err := abSvc.generateABTestFeedbackToken(1, abResp.ID, feedback.VariantID, createResp.ID, feedback.SessionID, time.Now().Add(-time.Minute))
	if err != nil {
		t.Fatalf("generate expired token: %v", err)
	}
	if _, err := abSvc.RecordABTestFeedback(&dto.ABTestFeedbackRequest{
		FeedbackToken: expiredToken,
		EventID:       "expired",
	}, "8.8.4.4", "Mozilla/5.0", ""); !errors.Is(err, ErrABTestFeedbackExpiredToken) {
		t.Fatalf("expected expired token error, got %v", err)
	}
	stats, err := abSvc.GetABTestStatisticsInWorkspace(abResp.ID, 30, 1)
	if err != nil {
		t.Fatalf("get ab statistics: %v", err)
	}
	if stats.TotalConversions != 1 || stats.ConversionRate != 100 || stats.ConversionValue != value {
		t.Fatalf("unexpected conversion statistics: %+v", stats)
	}
	var convertedVariantSeen bool
	for _, stat := range stats.VariantStats {
		if stat.Variant.ID == feedback.VariantID {
			convertedVariantSeen = true
			if stat.UniqueClicks != 1 || stat.ConversionCount != 1 || stat.ConversionRate != 100 || stat.ConversionValue != value {
				t.Fatalf("unexpected converted variant stats: %+v", stat)
			}
		} else if stat.ConversionRate != 0 || stat.ConversionCount != 0 {
			t.Fatalf("unexpected non-converted variant stats: %+v", stat)
		}
	}
	if !convertedVariantSeen || stats.WinningVariant == nil || stats.WinningVariant.ID != feedback.VariantID {
		t.Fatalf("winning variant should use conversion feedback, got stats=%+v feedback=%+v", stats, feedback)
	}
}

func TestParseTrafficMetadataUnknownAndBot(t *testing.T) {
	unknown := parseTrafficMetadata("   ")
	if unknown.DeviceType != "unknown" || unknown.Browser != "unknown" || unknown.OS != "unknown" || unknown.IsBot {
		t.Fatalf("unexpected empty user-agent metadata: %+v", unknown)
	}

	bot := parseTrafficMetadata("Googlebot/2.1 (+http://www.google.com/bot.html)")
	if !bot.IsBot || bot.DeviceType != "bot" || bot.BotName == "" {
		t.Fatalf("bot user-agent was not classified as bot: %+v", bot)
	}
}

func TestLinkSecurityURLRulesAndPasswordChallenge(t *testing.T) {
	helper := newShortLinkRegressionHelper(t)
	db := helper.GetDatabase()
	if err := db.Create(&model.Domain{
		WorkspaceID: 1,
		Protocol:    "https",
		Domain:      "dwz.do",
		IsActive:    true,
	}).Error; err != nil {
		t.Fatalf("seed domain: %v", err)
	}

	securitySvc := NewLinkSecurityService(helper)
	if _, err := securitySvc.CreateURLRule(1, 1, &dto.SecurityURLRuleRequest{
		RuleType: model.SecurityRuleTypeDomain,
		Action:   model.SecurityRuleActionBlock,
		Pattern:  "example.com",
	}); err != nil {
		t.Fatalf("create block rule: %v", err)
	}
	if _, err := securitySvc.CreateURLRule(1, 1, &dto.SecurityURLRuleRequest{
		RuleType: model.SecurityRuleTypeDomain,
		Action:   model.SecurityRuleActionAllow,
		Pattern:  "good.example.com",
	}); err != nil {
		t.Fatalf("create allow rule: %v", err)
	}

	shortLinkSvc := NewShortLinkService(helper, context.Background())
	if _, err := shortLinkSvc.CreateShortLinkInWorkspace(&dto.CreateShortLinkRequest{
		OriginalURL: "https://bad.example.com/path",
		Domain:      "dwz.do",
		CustomCode:  "blocked",
	}, "203.0.113.10", 1, 7); err == nil || !strings.Contains(err.Error(), "安全规则") {
		t.Fatalf("expected URL rule rejection, got %v", err)
	}

	password := "secret-pass"
	createResp, err := shortLinkSvc.CreateShortLinkInWorkspace(&dto.CreateShortLinkRequest{
		OriginalURL: "https://good.example.com/landing",
		Domain:      "dwz.do",
		CustomCode:  "safe",
		Security: &dto.LinkSecurityRequest{
			Password:      &password,
			ReportEnabled: boolPtr(true),
			BotPolicy:     model.LinkBotPolicyRecordOnly,
			IPPolicy:      model.LinkIPPolicyOff,
		},
	}, "203.0.113.10", 1, 7)
	if err != nil {
		t.Fatalf("create password link: %v", err)
	}
	if !createResp.SecurityEnabled || createResp.SecuritySummary == "未启用" || !createResp.ReportEnabled {
		t.Fatalf("security response not populated: %+v", createResp)
	}

	_, err = shortLinkSvc.ResolveRedirectWithSecurity(
		"dwz.do", "safe", "8.8.8.8", "Mozilla/5.0", "", "", "",
	)
	if !errors.Is(err, ErrSecurityPasswordRequired) {
		t.Fatalf("expected password challenge, got %v", err)
	}
	var clickCount int64
	if err := db.Model(&model.ClickStatistic{}).Where("short_link_id = ?", createResp.ID).Count(&clickCount).Error; err != nil {
		t.Fatalf("count clicks: %v", err)
	}
	if clickCount != 0 {
		t.Fatalf("password challenge should not count as click, got %d", clickCount)
	}
	waitForModelCount(t, db, &model.LinkSecurityEvent{}, "short_link_id = ?", createResp.ID, 1)

	_, token, _, err := securitySvc.VerifyPassword("dwz.do", "safe", password, "8.8.8.8", "Mozilla/5.0")
	if err != nil {
		t.Fatalf("verify password: %v", err)
	}
	decision, err := shortLinkSvc.ResolveRedirectWithSecurity(
		"dwz.do", "safe", "8.8.8.8", "Mozilla/5.0", "", "", token,
	)
	if err != nil {
		t.Fatalf("redirect after password: %v", err)
	}
	if decision.TargetURL != "https://good.example.com/landing" {
		t.Fatalf("unexpected target after password: %s", decision.TargetURL)
	}
	waitForModelCount(t, db, &model.ClickStatistic{}, "short_link_id = ?", createResp.ID, 1)
}

func TestAdvancedRoutingPriorityFallbackAndABPrecedence(t *testing.T) {
	helper := newShortLinkRegressionHelper(t)
	db := helper.GetDatabase()
	if err := db.Create(&model.Domain{
		WorkspaceID:     1,
		Protocol:        "https",
		Domain:          "dwz.do",
		PassQueryParams: true,
		IsActive:        true,
	}).Error; err != nil {
		t.Fatalf("seed domain: %v", err)
	}

	shortLinkSvc := NewShortLinkService(helper, context.Background())
	createResp, err := shortLinkSvc.CreateShortLinkInWorkspace(&dto.CreateShortLinkRequest{
		OriginalURL:  "https://example.com/default",
		FallbackURL:  "https://example.com/fallback",
		RedirectCode: 307,
		Domain:       "dwz.do",
		CustomCode:   "route",
	}, "203.0.113.10", 1, 7)
	if err != nil {
		t.Fatalf("create routed short link: %v", err)
	}
	if createResp.RedirectCode != 307 || createResp.FallbackURL == "" {
		t.Fatalf("routing fields missing in response: %+v", createResp)
	}

	routeSvc := NewLinkRouteService(helper)
	routeResp, err := routeSvc.CreateRoute(createResp.ID, 1, 7, &dto.LinkRouteRequest{
		Name:      "移动端广告",
		Priority:  10,
		TargetURL: "https://example.com/mobile",
		ConditionGroups: []dto.LinkRouteConditionGroupRequest{
			{
				Conditions: []dto.LinkRouteConditionRequest{
					{ConditionType: model.RouteConditionDeviceType, Operator: model.RouteOperatorEq, ConditionValue: "mobile"},
					{ConditionType: model.RouteConditionQueryParam, Operator: model.RouteOperatorEq, ConditionKey: "utm", ConditionValue: "ad"},
				},
			},
			{
				Conditions: []dto.LinkRouteConditionRequest{
					{ConditionType: model.RouteConditionLanguage, Operator: model.RouteOperatorPrefix, ConditionValue: "zh"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("create route: %v", err)
	}

	abSvc := NewABTestService(helper)
	abResp, err := abSvc.CreateABTestInWorkspace(&dto.CreateABTestRequest{
		ShortLinkID:  createResp.ID,
		Name:         "fallback split",
		TrafficSplit: "weighted",
		Variants: []dto.CreateABTestVariantRequest{
			{Name: "A", TargetURL: "https://example.com/ab-a", Weight: 99, IsControl: true},
			{Name: "B", TargetURL: "https://example.com/ab-b", Weight: 1, IsControl: false},
		},
	}, 1)
	if err != nil {
		t.Fatalf("create ab test: %v", err)
	}
	if _, err := abSvc.StartABTestInWorkspace(abResp.ID, &dto.StartABTestRequest{}, 1); err != nil {
		t.Fatalf("start ab test: %v", err)
	}

	decision, err := shortLinkSvc.ResolveRedirectWithSecurityAndLanguage(
		"dwz.do",
		"route",
		"8.8.8.8",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
		"https://ref.example",
		"utm=ad&click_id=1",
		"",
		"en-US,en;q=0.9",
	)
	if err != nil {
		t.Fatalf("resolve route: %v", err)
	}
	if decision.TargetURL != "https://example.com/mobile?click_id=1&utm=ad" || decision.StatusCode != 307 || decision.Route == nil || decision.Route.ID != routeResp.ID {
		t.Fatalf("route decision mismatch: %+v", decision)
	}
	waitForModelCount(t, db, &model.ClickStatistic{}, "short_link_id = ?", createResp.ID, 1)
	var stat model.ClickStatistic
	if err := db.Where("short_link_id = ?", createResp.ID).First(&stat).Error; err != nil {
		t.Fatalf("load route statistic: %v", err)
	}
	if stat.RouteID == nil || *stat.RouteID != routeResp.ID || stat.RouteName != "移动端广告" {
		t.Fatalf("route statistic not recorded: %+v", stat)
	}
	var abClicks int64
	if err := db.Model(&model.ABTestClickStatistic{}).Where("ab_test_id = ?", abResp.ID).Count(&abClicks).Error; err != nil {
		t.Fatalf("count ab clicks: %v", err)
	}
	if abClicks != 0 {
		t.Fatalf("route hit should skip A/B statistics, got %d", abClicks)
	}

	fallbackURL, err := shortLinkSvc.RedirectShortLinkWithQuery(
		"dwz.do",
		"route",
		"8.8.4.4",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/125.0.0.0 Safari/537.36",
		"https://ref.example",
		"click_id=2",
	)
	if err != nil {
		t.Fatalf("resolve fallback route: %v", err)
	}
	if fallbackURL != "https://example.com/fallback?click_id=2" {
		t.Fatalf("fallback should be used when active routes miss, got %s", fallbackURL)
	}
}

func waitForModelCount(t *testing.T, db *gorm.DB, modelValue any, query string, id uint64, want int64) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	var count int64
	for {
		if err := db.Model(modelValue).Where(query, id).Count(&count).Error; err != nil {
			t.Fatalf("count model: %v", err)
		}
		if count >= want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %d rows, got %d", want, count)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func boolPtr(value bool) *bool {
	return &value
}

func newShortLinkRegressionHelper(t *testing.T) *shortLinkRegressionHelper {
	t.Helper()
	dbName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open("file:"+dbName+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(
		&model.Workspace{},
		&model.WorkspaceMember{},
		&model.Domain{},
		&model.Campaign{},
		&model.Tag{},
		&model.ShortLinkTag{},
		&model.ShortLink{},
		&model.LinkRoute{},
		&model.LinkRouteConditionGroup{},
		&model.LinkRouteCondition{},
		&model.LinkSecuritySetting{},
		&model.LinkSecurityIPRule{},
		&model.SecurityURLRule{},
		&model.AbuseReport{},
		&model.LinkSecurityEvent{},
		&model.ClickStatistic{},
		&model.ABTest{},
		&model.ABTestVariant{},
		&model.ABTestClickStatistic{},
		&model.ABTestFeedback{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	if !db.Migrator().HasColumn(&model.ShortLinkTag{}, "CreatedAt") {
		if err := db.Migrator().AddColumn(&model.ShortLinkTag{}, "CreatedAt"); err != nil {
			t.Fatalf("add short_link_tags.created_at: %v", err)
		}
	}
	return &shortLinkRegressionHelper{
		db:       db,
		settings: shortLinkRegressionSettings{"database.driver": "sqlite", "jwt.secret": "test-secret-for-ab-feedback"},
		cache:    newShortLinkRegressionCache(),
	}
}

type shortLinkRegressionHelper struct {
	db       *gorm.DB
	settings shortLinkRegressionSettings
	cache    *shortLinkRegressionCache
}

func (h *shortLinkRegressionHelper) GetEnv() interfaces.EnvInterface       { return h.settings }
func (h *shortLinkRegressionHelper) GetConfig() interfaces.ConfigInterface { return h.settings }
func (h *shortLinkRegressionHelper) GetLogger() interfaces.LoggerInterface {
	return shortLinkRegressionLogger{}
}
func (h *shortLinkRegressionHelper) GetCache() gsr.Cacher                    { return h.cache }
func (h *shortLinkRegressionHelper) GetRedis() *redis.Client                 { return nil }
func (h *shortLinkRegressionHelper) GetDatabase() *gorm.DB                   { return h.db }
func (h *shortLinkRegressionHelper) GetInstalled() interfaces.Installed      { return nil }
func (h *shortLinkRegressionHelper) GetVersion() interfaces.VersionInterface { return nil }
func (h *shortLinkRegressionHelper) GetIPRegion() ipRegionImpl.IPRegion      { return ipRegionImpl.Noop{} }

type shortLinkRegressionSettings map[string]any

func (s shortLinkRegressionSettings) Get(key string, defaultValue any) any {
	if value, ok := s[key]; ok {
		return value
	}
	return defaultValue
}

func (s shortLinkRegressionSettings) GetBool(key string, defaultValue bool) bool {
	if value, ok := s[key].(bool); ok {
		return value
	}
	return defaultValue
}

func (s shortLinkRegressionSettings) GetInt(key string, defaultValue int) int {
	if value, ok := s[key].(int); ok {
		return value
	}
	return defaultValue
}

func (s shortLinkRegressionSettings) GetInt32(key string, defaultValue int32) int32 {
	if value, ok := s[key].(int32); ok {
		return value
	}
	return defaultValue
}

func (s shortLinkRegressionSettings) GetInt64(key string, defaultValue int64) int64 {
	if value, ok := s[key].(int64); ok {
		return value
	}
	return defaultValue
}

func (s shortLinkRegressionSettings) GetFloat64(key string, defaultValue float64) float64 {
	if value, ok := s[key].(float64); ok {
		return value
	}
	return defaultValue
}

func (s shortLinkRegressionSettings) GetStringSlice(key string, defaultValue []string) []string {
	if value, ok := s[key].([]string); ok {
		return value
	}
	return defaultValue
}

func (s shortLinkRegressionSettings) GetString(key string, defaultValue string) string {
	if value, ok := s[key].(string); ok {
		return value
	}
	return defaultValue
}

func (s shortLinkRegressionSettings) GetStringMapString(key string, defaultValue map[string]string) map[string]string {
	if value, ok := s[key].(map[string]string); ok {
		return value
	}
	return defaultValue
}

func (s shortLinkRegressionSettings) GetStringMapStringSlice(key string, defaultValue map[string][]string) map[string][]string {
	if value, ok := s[key].(map[string][]string); ok {
		return value
	}
	return defaultValue
}

func (s shortLinkRegressionSettings) GetTime(key string, defaultValue time.Time) time.Time {
	if value, ok := s[key].(time.Time); ok {
		return value
	}
	return defaultValue
}

func (s shortLinkRegressionSettings) Set(key string, value any) {
	s[key] = value
}

type shortLinkRegressionLogger struct{}

func (shortLinkRegressionLogger) Debug(string, ...gsr.LoggerField)  {}
func (shortLinkRegressionLogger) Info(string, ...gsr.LoggerField)   {}
func (shortLinkRegressionLogger) Notice(string, ...gsr.LoggerField) {}
func (shortLinkRegressionLogger) Error(string, ...gsr.LoggerField)  {}
func (shortLinkRegressionLogger) Warn(string, ...gsr.LoggerField)   {}
func (shortLinkRegressionLogger) Fatal(string, ...gsr.LoggerField)  {}

type shortLinkRegressionCache struct {
	mu     sync.Mutex
	values map[string][]byte
}

func newShortLinkRegressionCache() *shortLinkRegressionCache {
	return &shortLinkRegressionCache{values: map[string][]byte{}}
}

func (c *shortLinkRegressionCache) Exists(_ context.Context, key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.values[key]
	return ok
}

func (c *shortLinkRegressionCache) Get(_ context.Context, key string, obj any) error {
	c.mu.Lock()
	data, ok := c.values[key]
	c.mu.Unlock()
	if !ok {
		return errors.New("cache miss")
	}
	return json.Unmarshal(data, obj)
}

func (c *shortLinkRegressionCache) Set(_ context.Context, key string, value any, _ time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.values[key] = data
	c.mu.Unlock()
	return nil
}

func (c *shortLinkRegressionCache) GetSet(ctx context.Context, key string, ttl time.Duration, obj any, callback gsr.CacheCallback) error {
	if err := c.Get(ctx, key, obj); err == nil {
		return nil
	}
	if err := callback(key, obj); err != nil {
		return err
	}
	return c.Set(ctx, key, obj, ttl)
}

func (c *shortLinkRegressionCache) Del(_ context.Context, key string) error {
	c.mu.Lock()
	delete(c.values, key)
	c.mu.Unlock()
	return nil
}

func (c *shortLinkRegressionCache) ExpiresAt(context.Context, string, time.Time) error {
	return nil
}

func (c *shortLinkRegressionCache) ExpiresIn(context.Context, string, time.Duration) error {
	return nil
}
