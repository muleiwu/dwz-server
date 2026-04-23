package controller

import (
	"context"
	"time"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/constants"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
	helperPkg "cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
)

type HealthController struct {
	BaseResponse
}

// GetHealth 健康检查接口
func (receiver HealthController) GetHealth(c httpInterfaces.RouterContextInterface) {
	helper := helperPkg.GetHelper()
	_ = helper
	healthStatus := dto.HealthStatus{
		Status:    "UP",
		Timestamp: time.Now().Unix(),
		Services:  make(map[string]interface{}),
	}

	// 添加版本信息
	if helper.GetVersion() != nil {
		healthStatus.Version = &dto.VersionInfo{
			Version:   helper.GetVersion().GetVersion(),
			GitCommit: helper.GetVersion().GetGitCommit(),
			BuildTime: helper.GetVersion().GetBuildTime(),
		}
	}

	// 检查数据库连接
	dbStatus := receiver.checkDatabase()
	healthStatus.Services["database"] = dbStatus
	healthStatus.Services["database_driver"] = helper.GetConfig().GetString("database.driver", "")

	// 检查Redis连接
	redisStatus := receiver.checkRedis(c.Request().Context())
	healthStatus.Services["redis"] = redisStatus

	// 如果任何必要服务不健康，整体状态设为DOWN（忽略DISABLED状态的服务）
	if dbStatus.Status == "DOWN" || (redisStatus.Status == "DOWN") {
		healthStatus.Status = "DOWN"
		var baseResponse BaseResponse
		baseResponse.Error(c, constants.ErrCodeUnavailable, "服务不健康")
		return
	}

	var baseResponse BaseResponse
	baseResponse.Success(c, healthStatus)
}

// GetHealthSimple 简单健康检查接口
func (receiver HealthController) GetHealthSimple(c httpInterfaces.RouterContextInterface) {
	helper := helperPkg.GetHelper()
	_ = helper
	var baseResponse BaseResponse
	baseResponse.Success(c, map[string]any{
		"status":    "UP",
		"timestamp": time.Now().Unix(),
	})
}

// checkDatabase 检查数据库连接
func (receiver HealthController) checkDatabase() dto.ServiceStatus {
	gormDB := helperPkg.GetHelper().GetDatabase()
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
func (receiver HealthController) checkRedis(ctx context.Context) dto.ServiceStatus {
	helper := helperPkg.GetHelper()
	redisHelper := helper.GetRedis()
	if redisHelper == nil {
		return dto.ServiceStatus{
			Status:  "DISABLED",
			Message: "禁用",
		}
	}
	if err := redisHelper.Ping(ctx); err.Err() != nil {
		helper.GetLogger().Error(err.Err().Error())
		return dto.ServiceStatus{
			Status:  "DOWN",
			Message: "Redis ping失败: " + err.Err().Error(),
		}
	}

	return dto.ServiceStatus{
		Status: "UP",
	}
}
