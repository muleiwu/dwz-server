package cmd

import (
	"embed"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"cnb.cool/mliev/open/dwz-server/config"
	helper2 "cnb.cool/mliev/open/dwz-server/internal/helper"
)

// Start 启动应用程序
func Start(staticFs map[string]embed.FS) {
	initializeServices(staticFs)
	// 添加阻塞以保持主程序运行
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}

// initializeServices 初始化所有服务
func initializeServices(staticFs map[string]embed.FS) {

	helper := helper2.GetHelper()

	assembly := config.Assembly{
		Helper: helper,
	}
	for _, assemblyInterface := range assembly.Get() {
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

	helper.GetConfig().Set("static.fs", staticFs)

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
