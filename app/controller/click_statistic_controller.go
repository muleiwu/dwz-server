package controller

import (
	"strconv"
	"time"

	"cnb.cool/mliev/open/dwz-server/app/constants"
	"cnb.cool/mliev/open/dwz-server/app/dto"
	"cnb.cool/mliev/open/dwz-server/app/service"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"github.com/gin-gonic/gin"
)

type ClickStatisticController struct {
	BaseResponse
}

// GetClickStatisticList 获取点击统计列表
func (ctrl ClickStatisticController) GetClickStatisticList(c *gin.Context, helper interfaces.GetHelperInterface) {
	var req dto.ClickStatisticListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	clickStatisticService := service.NewClickStatisticService(helper)
	response, err := clickStatisticService.GetClickStatisticList(&req)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}

// GetClickStatisticAnalysis 获取点击统计分析
func (ctrl ClickStatisticController) GetClickStatisticAnalysis(c *gin.Context, helper interfaces.GetHelperInterface) {
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

		response, err := clickStatisticService.GetClickStatisticAnalysisByDateRange(shortLinkID, startDate, endDate)
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

	response, err := clickStatisticService.GetClickStatisticAnalysis(shortLinkID, days)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}
