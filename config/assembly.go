package config

import (
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	configAssembly "cnb.cool/mliev/open/dwz-server/internal/pkg/config/assembly"
	databaseAssembly "cnb.cool/mliev/open/dwz-server/internal/pkg/database/assembly"
	envAssembly "cnb.cool/mliev/open/dwz-server/internal/pkg/env/assembly"
	installedAssembly "cnb.cool/mliev/open/dwz-server/internal/pkg/installed/assembly"
	loggerAssembly "cnb.cool/mliev/open/dwz-server/internal/pkg/logger/assembly"
	redisAssembly "cnb.cool/mliev/open/dwz-server/internal/pkg/redis/assembly"
)

type Assembly struct {
	Helper interfaces.HelperInterface
}

// Get 注入反转(确保注入顺序，防止依赖为空或者循环依赖)
func (receiver *Assembly) Get() []interfaces.AssemblyInterface {

	return []interfaces.AssemblyInterface{
		&envAssembly.Env{Helper: receiver.Helper},                                       // 环境变量
		&configAssembly.Config{Helper: receiver.Helper, DefaultConfigs: Config{}.Get()}, // 代码中的配置(可使用环境变量)
		&loggerAssembly.Logger{Helper: receiver.Helper},                                 // 日志驱动
		&installedAssembly.Installed{Helper: receiver.Helper},                           // 安装检测
		&databaseAssembly.Database{Helper: receiver.Helper},                             // 数据库配置
		&redisAssembly.Redis{Helper: receiver.Helper},                                   // redis 配置
	}
}
