package controller

import (
	"fmt"
	"net/http"
	"strconv"
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
	shortLinkIDStr := c.Query("short_link_id")
	daysStr := c.DefaultQuery("days", "7")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var shortLinkID uint64
	var err error

	if shortLinkIDStr != "" {
		shortLinkID, err = strconv.ParseUint(shortLinkIDStr, 10, 64)
		if err != nil {
			ctrl.Error(c, constants.ErrCodeBadRequest, "无效的短链接ID格式")
			return
		}
	}

	clickStatisticService := service.NewClickStatisticService(helper)
	filterReq := &dto.ClickStatisticListRequest{
		ShortLinkID: shortLinkID,
		CampaignID:  parseUintQuery(c.Query("campaign_id")),
		RouteID:     parseUintQuery(c.Query("route_id")),
		TagID:       parseUintQuery(c.Query("tag_id")),
		DeviceType:  c.Query("device_type"),
	}
	if isBotStr := c.Query("is_bot"); isBotStr != "" {
		if isBot, err := strconv.ParseBool(isBotStr); err == nil {
			filterReq.IsBot = &isBot
		}
	}

	// 如果指定了日期范围，优先使用日期范围
	if startDateStr != "" && endDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			ctrl.Error(c, constants.ErrCodeBadRequest, "开始日期格式错误")
			return
		}

		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			ctrl.Error(c, constants.ErrCodeBadRequest, "结束日期格式错误")
			return
		}

		// 结束日期加一天，以包含结束日期的全天数据
		endDate = endDate.Add(24 * time.Hour)
		filterReq.StartDate = startDate
		filterReq.EndDate = endDate

		response, err := clickStatisticService.GetClickStatisticAnalysisInWorkspace(middleware.GetCurrentWorkspaceID(c), filterReq, 7)
		if err != nil {
			ctrl.Error(c, constants.ErrCodeInternal, err.Error())
			return
		}

		ctrl.Success(c, response)
		return
	}

	// 使用天数范围
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 365 {
		days = 7
	}

	response, err := clickStatisticService.GetClickStatisticAnalysisInWorkspace(middleware.GetCurrentWorkspaceID(c), filterReq, days)
	if err != nil {
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
