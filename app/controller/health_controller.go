package controller

import (
	"context"
	"time"

	"cnb.cool/mliev/open/dwz-server/app/constants"
	"cnb.cool/mliev/open/dwz-server/app/dto"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"github.com/gin-gonic/gin"
)

type HealthController struct {
	BaseResponse
}

// GetHealth 健康检查接口
func (receiver HealthController) GetHealth(c *gin.Context, helper interfaces.GetHelperInterface) {
	healthStatus := dto.HealthStatus{
		Status:    "UP",
		Timestamp: time.Now().Unix(),
		Services:  make(map[string]interface{}),
	}

	// 检查数据库连接
	dbStatus := receiver.checkDatabase(helper)
	healthStatus.Services["database"] = dbStatus

	// 检查Redis连接
	redisStatus := receiver.checkRedis(helper)
	healthStatus.Services["redis"] = redisStatus

	// 如果任何服务不健康，整体状态设为DOWN
	if dbStatus.Status == "DOWN" || redisStatus.Status == "DOWN" {
		healthStatus.Status = "DOWN"
		var baseResponse BaseResponse
		baseResponse.Error(c, constants.ErrCodeUnavailable, "服务不健康")
		return
	}

	var baseResponse BaseResponse
	baseResponse.Success(c, healthStatus)
}

// GetHealthSimple 简单健康检查接口
func (receiver HealthController) GetHealthSimple(c *gin.Context, helper interfaces.GetHelperInterface) {
	var baseResponse BaseResponse
	baseResponse.Success(c, gin.H{
		"status":    "UP",
		"timestamp": time.Now().Unix(),
	})
}

// checkDatabase 检查数据库连接
func (receiver HealthController) checkDatabase(helper interfaces.GetHelperInterface) dto.ServiceStatus {

	gormDB := helper.GetDatabase()
	if gormDB == nil {
		return dto.ServiceStatus{
			Status:  "DOWN",
			Message: "获取数据库连接失败",
		}
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return dto.ServiceStatus{
			Status:  "DOWN",
			Message: "获取底层数据库连接失败: " + err.Error(),
		}
	}

	if err := sqlDB.Ping(); err != nil {
		return dto.ServiceStatus{
			Status:  "DOWN",
			Message: "数据库ping失败: " + err.Error(),
		}
	}

	return dto.ServiceStatus{
		Status: "UP",
	}
}

// checkRedis 检查Redis连接
func (receiver HealthController) checkRedis(helper interfaces.GetHelperInterface) dto.ServiceStatus {
	redisHelper := helper.GetRedis()
	if redisHelper == nil {
		return dto.ServiceStatus{
			Status:  "DOWN",
			Message: "Redis连接失败",
		}
	}
	ctx := context.Background()
	if err := redisHelper.Ping(ctx); err != nil {
		return dto.ServiceStatus{
			Status:  "DOWN",
			Message: "Redis ping失败: " + err.Err().Error(),
		}
	}

	return dto.ServiceStatus{
		Status: "UP",
	}
}
