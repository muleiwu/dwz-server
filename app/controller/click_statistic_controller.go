package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/constants"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/middleware"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/service"
	helperPkg "cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
)

type ClickStatisticController struct {
	BaseResponse
}

// GetClickStatisticList 获取点击统计列表
func (ctrl ClickStatisticController) GetClickStatisticList(c httpInterfaces.RouterContextInterface) {
	helper := helperPkg.GetHelper()
	_ = helper
	var req dto.ClickStatisticListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	if err := normalizeClickStatisticQueryDateRange(c, &req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, err.Error())
		return
	}

	clickStatisticService := service.NewClickStatisticService(helper)
	response, err := clickStatisticService.GetClickStatisticListInWorkspace(&req, middleware.GetCurrentWorkspaceID(c))
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}

// GetClickStatisticAnalysis 获取点击统计分析
func (ctrl ClickStatisticController) GetClickStatisticAnalysis(c httpInterfaces.RouterContextInterface) {
	helper := helperPkg.GetHelper()
	_ = helper
	filterReq, days, err := buildClickStatisticAnalysisRequest(c)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, err.Error())
		return
	}

	response, err := service.NewClickStatisticService(helper).GetClickStatisticAnalysisInWorkspace(middleware.GetCurrentWorkspaceID(c), filterReq, days)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}

func (ctrl ClickStatisticController) GetClickStatisticGeoAnalysis(c httpInterfaces.RouterContextInterface) {
	helper := helperPkg.GetHelper()
	filterReq, days, err := buildClickStatisticAnalysisRequest(c)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, err.Error())
		return
	}

	response, err := service.NewClickStatisticService(helper).GetClickStatisticGeoAnalysisInWorkspace(middleware.GetCurrentWorkspaceID(c), filterReq, c.DefaultQuery("level", "country"), days)
	if err != nil {
		if strings.Contains(err.Error(), "地理统计级别") {
			ctrl.Error(c, constants.ErrCodeBadRequest, err.Error())
			return
		}
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}

func (ctrl ClickStatisticController) ExportCSV(c httpInterfaces.RouterContextInterface) {
	helper := helperPkg.GetHelper()
	var req dto.ClickStatisticListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	if err := normalizeClickStatisticQueryDateRange(c, &req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, err.Error())
		return
	}
	data, err := service.NewClickStatisticService(helper).ExportCSV(middleware.GetCurrentWorkspaceID(c), &req)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, err.Error())
		return
	}
	filename := fmt.Sprintf("click-statistics-%s.csv", time.Now().Format("20060102150405"))
	c.SetHeader("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

func parseUintQuery(value string) uint64 {
	if value == "" {
		return 0
	}
	id, _ := strconv.ParseUint(value, 10, 64)
	return id
}

func buildClickStatisticAnalysisRequest(c httpInterfaces.RouterContextInterface) (*dto.ClickStatisticListRequest, int, error) {
	shortLinkIDStr := c.Query("short_link_id")
	var shortLinkID uint64
	var err error
	if shortLinkIDStr != "" {
		shortLinkID, err = strconv.ParseUint(shortLinkIDStr, 10, 64)
		if err != nil {
			return nil, 0, fmt.Errorf("无效的短链接ID格式")
		}
	}

	days, err := strconv.Atoi(c.DefaultQuery("days", "7"))
	if err != nil || days < 1 || days > 365 {
		days = 7
	}

	filterReq := &dto.ClickStatisticListRequest{
		ShortLinkID: shortLinkID,
		CampaignID:  parseUintQuery(c.Query("campaign_id")),
		RouteID:     parseUintQuery(c.Query("route_id")),
		TagID:       parseUintQuery(c.Query("tag_id")),
		DeviceType:  c.Query("device_type"),
		IP:          c.Query("ip"),
		Country:     c.Query("country"),
		Province:    c.Query("province"),
		City:        c.Query("city"),
		ISP:         c.Query("isp"),
	}
	if isBotStr := c.Query("is_bot"); isBotStr != "" {
		if isBot, err := strconv.ParseBool(isBotStr); err == nil {
			filterReq.IsBot = &isBot
		}
	}

	if c.Query("start_date") != "" || c.Query("end_date") != "" {
		if c.Query("start_date") == "" || c.Query("end_date") == "" {
			return nil, 0, fmt.Errorf("开始日期和结束日期必须同时提供")
		}
		if err := normalizeClickStatisticQueryDateRange(c, filterReq); err != nil {
			return nil, 0, err
		}
	}

	return filterReq, days, nil
}

func normalizeClickStatisticQueryDateRange(c httpInterfaces.RouterContextInterface, req *dto.ClickStatisticListRequest) error {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	location := time.Now().Location()

	if startDateStr != "" {
		startDate, err := time.ParseInLocation("2006-01-02", startDateStr, location)
		if err != nil {
			return fmt.Errorf("开始日期格式错误")
		}
		req.StartDate = startDate
	}

	if endDateStr != "" {
		endDate, err := time.ParseInLocation("2006-01-02", endDateStr, location)
		if err != nil {
			return fmt.Errorf("结束日期格式错误")
		}
		req.EndDate = endDate.AddDate(0, 0, 1)
	}

	if !req.StartDate.IsZero() && !req.EndDate.IsZero() && !req.EndDate.After(req.StartDate) {
		return fmt.Errorf("结束日期不能早于开始日期")
	}
	return nil
}
