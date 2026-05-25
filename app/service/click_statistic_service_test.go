package service

import (
	"fmt"
	"testing"
	"time"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"gorm.io/gorm"
)

func TestClickStatisticGeoAnalysisReturnsCompleteRegions(t *testing.T) {
	helper := newShortLinkRegressionHelper(t)
	db := helper.GetDatabase()
	clickedAt := time.Date(2026, 5, 10, 12, 0, 0, 0, time.Local)

	for i := 0; i < 12; i++ {
		stat := model.ClickStatistic{
			WorkspaceID: 1,
			ShortLinkID: uint64(i + 1),
			IP:          fmt.Sprintf("10.0.0.%d", i+1),
			Country:     "中国",
			Province:    fmt.Sprintf("测试省%02d", i+1),
			City:        fmt.Sprintf("测试市%02d", i+1),
			ClickDate:   clickedAt,
		}
		if err := db.Create(&stat).Error; err != nil {
			t.Fatalf("seed click statistic: %v", err)
		}
	}

	req := &dto.ClickStatisticListRequest{
		Country:   "中国",
		StartDate: clickedAt.Add(-time.Hour),
		EndDate:   clickedAt.Add(24 * time.Hour),
	}
	service := NewClickStatisticService(helper)

	geo, err := service.GetClickStatisticGeoAnalysisInWorkspace(1, req, "province", 7)
	if err != nil {
		t.Fatalf("geo analysis: %v", err)
	}
	if geo.TotalClicks != 12 {
		t.Fatalf("expected 12 total clicks, got %d", geo.TotalClicks)
	}
	if len(geo.Regions) != 12 {
		t.Fatalf("expected all 12 province regions, got %d", len(geo.Regions))
	}

	analysis, err := service.GetClickStatisticAnalysisInWorkspace(1, req, 7)
	if err != nil {
		t.Fatalf("summary analysis: %v", err)
	}
	if len(analysis.TopProvinces) != 10 {
		t.Fatalf("expected summary analysis to keep Top 10 provinces, got %d", len(analysis.TopProvinces))
	}
}

func TestClickStatisticGeoAnalysisCacheAndFilterIsolation(t *testing.T) {
	helper := newShortLinkRegressionHelper(t)
	db := helper.GetDatabase()
	clickedAt := time.Date(2026, 5, 11, 10, 0, 0, 0, time.Local)

	seedClickStatistic(t, db, model.ClickStatistic{
		WorkspaceID: 1,
		ShortLinkID: 1,
		IP:          "10.1.0.1",
		Country:     "中国",
		Province:    "广东省",
		City:        "广州市",
		ClickDate:   clickedAt,
	})

	service := NewClickStatisticService(helper)
	cityReq := &dto.ClickStatisticListRequest{
		Country:   "中国",
		Province:  "广东省",
		StartDate: clickedAt.Add(-time.Hour),
		EndDate:   clickedAt.Add(24 * time.Hour),
	}
	first, err := service.GetClickStatisticGeoAnalysisInWorkspace(1, cityReq, "city", 7)
	if err != nil {
		t.Fatalf("first geo analysis: %v", err)
	}
	if first.TotalClicks != 1 || len(first.Regions) != 1 {
		t.Fatalf("expected first city analysis to see one click/city, got total=%d regions=%d", first.TotalClicks, len(first.Regions))
	}

	seedClickStatistic(t, db, model.ClickStatistic{
		WorkspaceID: 1,
		ShortLinkID: 2,
		IP:          "10.1.0.2",
		Country:     "中国",
		Province:    "广东省",
		City:        "深圳市",
		ClickDate:   clickedAt,
	})
	seedClickStatistic(t, db, model.ClickStatistic{
		WorkspaceID: 1,
		ShortLinkID: 3,
		IP:          "10.1.0.3",
		Country:     "中国",
		Province:    "浙江省",
		City:        "杭州市",
		ClickDate:   clickedAt,
	})

	second, err := service.GetClickStatisticGeoAnalysisInWorkspace(1, cityReq, "city", 7)
	if err != nil {
		t.Fatalf("cached city geo analysis: %v", err)
	}
	if second.TotalClicks != 1 || len(second.Regions) != 1 {
		t.Fatalf("expected identical city query to use cached result, got total=%d regions=%d", second.TotalClicks, len(second.Regions))
	}

	provinceReq := &dto.ClickStatisticListRequest{
		Country:   "中国",
		StartDate: clickedAt.Add(-time.Hour),
		EndDate:   clickedAt.Add(24 * time.Hour),
	}
	province, err := service.GetClickStatisticGeoAnalysisInWorkspace(1, provinceReq, "province", 7)
	if err != nil {
		t.Fatalf("province geo analysis: %v", err)
	}
	if province.TotalClicks != 3 || len(province.Regions) != 2 {
		t.Fatalf("expected separate province query to see fresh data, got total=%d regions=%d", province.TotalClicks, len(province.Regions))
	}
}

func seedClickStatistic(t *testing.T, db *gorm.DB, stat model.ClickStatistic) {
	t.Helper()
	if err := db.Create(&stat).Error; err != nil {
		t.Fatalf("seed click statistic: %v", err)
	}
}
