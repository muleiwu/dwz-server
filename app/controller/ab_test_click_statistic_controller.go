package controller

import (
	"strconv"
	"time"

	"cnb.cool/mliev/open/dwz-server/app/dto"
	"cnb.cool/mliev/open/dwz-server/app/service"
	"cnb.cool/mliev/open/dwz-server/constants"
	"github.com/gin-gonic/gin"
)

type ABTestClickStatisticController struct {
	BaseResponse
}

// GetABTestClickStatisticList 获取AB测试点击统计列表
func (ctrl ABTestClickStatisticController) GetABTestClickStatisticList(c *gin.Context) {
	var req dto.ABTestClickStatisticListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	abTestClickStatisticService := service.NewABTestClickStatisticService()
	response, err := abTestClickStatisticService.GetABTestClickStatisticList(&req)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}

// GetABTestClickStatisticAnalysis 获取AB测试点击统计分析
func (ctrl ABTestClickStatisticController) GetABTestClickStatisticAnalysis(c *gin.Context) {
	abTestIDStr := c.Query("ab_test_id")
	daysStr := c.DefaultQuery("days", "7")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var abTestID uint64
	var err error

	if abTestIDStr != "" {
		abTestID, err = strconv.ParseUint(abTestIDStr, 10, 64)
		if err != nil {
			ctrl.Error(c, constants.ErrCodeBadRequest, "无效的AB测试ID格式")
			return
		}
	}

	abTestClickStatisticService := service.NewABTestClickStatisticService()

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

		response, err := abTestClickStatisticService.GetABTestClickStatisticAnalysisByDateRange(abTestID, startDate, endDate)
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

	response, err := abTestClickStatisticService.GetABTestClickStatisticAnalysis(abTestID, days)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}

// GetABTestVariantStatistics 获取AB测试版本统计
func (ctrl ABTestClickStatisticController) GetABTestVariantStatistics(c *gin.Context) {
	abTestIDStr := c.Param("id")
	daysStr := c.DefaultQuery("days", "7")

	abTestID, err := strconv.ParseUint(abTestIDStr, 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的AB测试ID格式")
		return
	}

	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 365 {
		days = 7
	}

	abTestClickStatisticService := service.NewABTestClickStatisticService()
	response, err := abTestClickStatisticService.GetVariantStatistics(abTestID, days)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}
