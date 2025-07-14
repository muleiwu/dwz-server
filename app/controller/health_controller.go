package controller

import (
	"cnb.cool/mliev/open/dwz-server/helper/database"
	"cnb.cool/mliev/open/dwz-server/helper/logger"
	"cnb.cool/mliev/open/dwz-server/helper/redis"
	"context"
	"go.uber.org/zap"
	"time"

	"cnb.cool/mliev/open/dwz-server/app/dto"
	"cnb.cool/mliev/open/dwz-server/constants"
	"github.com/gin-gonic/gin"
)

type HealthController struct {
	BaseResponse
}

// GetHealth 健康检查接口
func (receiver HealthController) GetHealth(c *gin.Context) {
	healthStatus := dto.HealthStatus{
		Status:    "UP",
		Timestamp: time.Now().Unix(),
		Services:  make(map[string]interface{}),
	}

	// 检查数据库连接
	dbStatus := receiver.checkDatabase()
	healthStatus.Services["database"] = dbStatus

	// 检查Redis连接
	redisStatus := receiver.checkRedis()
	healthStatus.Services["redis"] = redisStatus

	// 如果任何服务不健康，整体状态设为DOWN
	if dbStatus.Status == "DOWN" || redisStatus.Status == "DOWN" {
		logger.Logger().Error("服务不健康日志", zap.Any("healthStatus", healthStatus))
		healthStatus.Status = "DOWN"
		var baseResponse BaseResponse
		baseResponse.Error(c, constants.ErrCodeUnavailable, "服务不健康")
		return
	}

	var baseResponse BaseResponse
	baseResponse.Success(c, healthStatus)
}

// GetHealthSimple 简单健康检查接口
func (receiver HealthController) GetHealthSimple(c *gin.Context) {
	var baseResponse BaseResponse
	baseResponse.Success(c, gin.H{
		"status":    "UP",
		"timestamp": time.Now().Unix(),
	})
}

// checkDatabase 检查数据库连接
func (receiver HealthController) checkDatabase() dto.ServiceStatus {
	database := database.GetDB()
	if database == nil {
		return dto.ServiceStatus{
			Status:  "DOWN",
			Message: "数据库连接失败",
		}
	}

	sqlDB, err := database.DB()
	if err != nil {
		return dto.ServiceStatus{
			Status:  "DOWN",
			Message: "获取数据库连接失败: " + err.Error(),
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
func (receiver HealthController) checkRedis() dto.ServiceStatus {
	redis := redis.GetRedis()
	if redis == nil {
		return dto.ServiceStatus{
			Status:  "DOWN",
			Message: "Redis连接失败",
		}
	}

	ctx := context.Background()
	if err := redis.Ping(ctx).Err(); err != nil {
		return dto.ServiceStatus{
			Status:  "DOWN",
			Message: "Redis ping失败: " + err.Error(),
		}
	}

	return dto.ServiceStatus{
		Status: "UP",
	}
}
