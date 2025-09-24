package test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"cnb.cool/mliev/open/dwz-server/app/dto"
	"cnb.cool/mliev/open/dwz-server/config"
	helper2 "cnb.cool/mliev/open/dwz-server/internal/helper"
)

func TestCache(t *testing.T) {
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

	key := "15555da:sdada"
	ctx := context.Background()

	pagination := dto.Pagination{
		Total:    1,
		Page:     1231,
		PageSize: 1234234234234,
		Pages:    5343453,
	}

	err := helper.GetCache().Set(ctx, key, pagination, 3600)

	if err != nil {
		panic(err)
	}

	d := dto.Pagination{}

	err = helper.GetCache().Get(ctx, key, &d)

	if err != nil {
		panic(err)
	}

	fmt.Println(pagination)
	fmt.Println(d)

}
