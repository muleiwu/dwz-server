package config

import (
	"cnb.cool/mliev/open/dwz-server/pkg/interfaces"
	cacheAssembly "cnb.cool/mliev/open/dwz-server/pkg/service/cache/assembly"
	configAssembly "cnb.cool/mliev/open/dwz-server/pkg/service/config/assembly"
	databaseAssembly "cnb.cool/mliev/open/dwz-server/pkg/service/database/assembly"
	envAssembly "cnb.cool/mliev/open/dwz-server/pkg/service/env/assembly"
	installedAssembly "cnb.cool/mliev/open/dwz-server/pkg/service/installed/assembly"
	loggerAssembly "cnb.cool/mliev/open/dwz-server/pkg/service/logger/assembly"
	redisAssembly "cnb.cool/mliev/open/dwz-server/pkg/service/redis/assembly"
	versionAssembly "cnb.cool/mliev/open/dwz-server/pkg/service/version/assembly"
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
		&versionAssembly.Version{Helper: receiver.Helper},                               // 版本信息
		&installedAssembly.Installed{Helper: receiver.Helper},                           // 安装检测
		&databaseAssembly.Database{Helper: receiver.Helper},                             // 数据库配置
		&redisAssembly.Redis{Helper: receiver.Helper},                                   // redis 配置
		&cacheAssembly.Cache{Helper: receiver.Helper},                                   // 缓存驱动
	}
}
