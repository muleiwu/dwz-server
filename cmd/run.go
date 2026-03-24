package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"cnb.cool/mliev/dwz/dwz-server/config"
	helper2 "cnb.cool/mliev/dwz/dwz-server/pkg/helper"
	configAssembly "cnb.cool/mliev/dwz/dwz-server/pkg/service/config/assembly"
)

// Start 启动应用程序
func Start(opts StartOptions) {
	initializeServices(opts)
	// 添加阻塞以保持主程序运行
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}

// initializeServices 初始化所有服务
// 实现8步初始化流程以支持CE和EE版本扩展
func initializeServices(opts StartOptions) {
	helper := helper2.GetHelper()

	// Step 1: 获取CE组装列表 (8个核心组装)
	assembly := config.Assembly{
		Helper: helper,
	}
	ceAssemblies := assembly.Get()

	// Step 2: 将EE额外配置注入到Config组装的DefaultConfigs字段中
	// 这确保EE配置在Config.Assembly()执行时一起加载
	if len(opts.ExtraConfigs) > 0 {
		for i, assemblyInterface := range ceAssemblies {
			if cfgAssembly, ok := assemblyInterface.(*configAssembly.Config); ok {
				cfgAssembly.DefaultConfigs = append(cfgAssembly.DefaultConfigs, opts.ExtraConfigs...)
				ceAssemblies[i] = cfgAssembly
				fmt.Printf("[load] 注入 EE 配置提供者: %d 个\n", len(opts.ExtraConfigs))
				break
			}
		}
	}

	// Step 3: 执行CE组装 (包括合并后的配置)
	for _, assemblyInterface := range ceAssemblies {
		startTime := time.Now()
		err := assemblyInterface.Assembly()
		if err != nil {
			if helper.GetLogger() != nil {
				helper.GetLogger().Error(err.Error())
			} else {
				fmt.Println(err.Error())
			}
		}
		// 记录启动耗时
		duration := time.Since(startTime)
		typeName := reflect.TypeOf(assemblyInterface).Elem().Name()
		fmt.Printf("[load] 加载: %s  完成，总耗时: %v \n", typeName, duration)
	}

	// Step 4: 执行EE组装 (如果提供)
	if opts.ExtraAssembly != nil {
		eeAssemblies := opts.ExtraAssembly(helper)
		for _, assemblyInterface := range eeAssemblies {
			startTime := time.Now()
			err := assemblyInterface.Assembly()
			if err != nil {
				helper.GetLogger().Error(err.Error())
			}
			// 记录启动耗时
			duration := time.Since(startTime)
			typeName := reflect.TypeOf(assemblyInterface).Elem().Name()
			fmt.Printf("[load] 加载 EE 组装: %s  完成，总耗时: %v \n", typeName, duration)
		}
	}

	// Step 5: 存储静态文件
	helper.GetConfig().Set("static.fs", opts.StaticFs)

	// Step 6: 设置版本信息
	helper.GetVersion().SetVersionInfo(opts.Version, opts.GitCommit, opts.BuildTime)

	helper.GetLogger().Info(fmt.Sprintf("【构建版本】：%s", helper.GetVersion().GetVersion()))
	helper.GetLogger().Info(fmt.Sprintf("【构建时间】：%s", helper.GetVersion().GetBuildTime()))
	helper.GetLogger().Info(fmt.Sprintf("【构建哈希】：%s", helper.GetVersion().GetGitCommit()))

	// Step 7: 存储EE钩子到配置中供后续使用
	if opts.ExtraRoutes != nil {
		helper.GetConfig().Set("ee.extra_routes", opts.ExtraRoutes)
	}
	if opts.ExtraModels != nil && len(opts.ExtraModels) > 0 {
		helper.GetConfig().Set("ee.extra_models", opts.ExtraModels)
	}
	if opts.ExtraMigrationsFS != (opts.ExtraMigrationsFS) {
		helper.GetConfig().Set("ee.extra_migrations_fs", opts.ExtraMigrationsFS)
	}

	// Step 8: 启动服务器 (Migration → IDGenerator → HttpServer)
	server := config.Server{
		Helper: helper,
	}
	for _, serverInterface := range server.Get() {
		startTime := time.Now()
		err := serverInterface.Run()
		if err != nil {
			helper.GetLogger().Error(err.Error())
		}

		// 记录启动耗时
		duration := time.Since(startTime)
		typeName := reflect.TypeOf(serverInterface).Elem().Name()
		helper.GetLogger().Debug(fmt.Sprintf("[启动] 服务: %s 总耗时: %v", typeName, duration))
	}
}
