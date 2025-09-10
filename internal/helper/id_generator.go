package helper

import (
	"errors"
	"fmt"
	"sync"

	"cnb.cool/mliev/open/dwz-server/app/dao"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"cnb.cool/mliev/open/dwz-server/pkg/id_generator/base"
	"cnb.cool/mliev/open/dwz-server/pkg/id_generator/redis"
)

var idGeneratorHelper interfaces.IDGenerator
var idGeneratorOnce sync.Once

// GetIdGenerator returns a singleton IDGenerator instance
// The initialization will only happen once, even if called concurrently
func GetIdGenerator() interfaces.IDGenerator {
	// Fast path: if already initialized, return it
	if idGeneratorHelper != nil {
		return idGeneratorHelper
	}

	GetHelper().GetLogger().Error("Failed to initialize ID generator")

	return base.NewIdGeneratorBase()
}

func InitIdGenerator(helper interfaces.HelperInterface) (interfaces.IDGenerator, error) {
	// Initialize only once using sync.Once
	var errMsg error
	idGeneratorOnce.Do(func() {
		idGenerator, err := initializeDomainCounters(helper)
		if err != nil {
			helper.GetLogger().Error(fmt.Sprintf("Failed to initialize ID generator: %v", err))
			errMsg = err
			return
		}
		idGeneratorHelper = idGenerator
	})

	if errMsg != nil {
		helper.GetLogger().Error(errMsg.Error())
		helper.GetLogger().Error("ID generator initialization failed, retrying...")
		// Reset once to allow retrying initialization
		idGeneratorOnce = sync.Once{}

		return nil, errMsg
	}

	return idGeneratorHelper, nil
}

// initializeDomainCounters 初始化域名计数器
func initializeDomainCounters(helper interfaces.HelperInterface) (interfaces.IDGenerator, error) {

	driver := helper.GetConfig().GetString("id_generator.driver", "redis")
	if driver == "redis" {
		idGeneratorHelper = redis.NewIdGeneratorRedis(helper)
	} else {
		idGeneratorHelper = base.NewIdGeneratorBase()
	}

	if helper.GetDatabase() == nil {
		return nil, errors.New("数据库连接获取失败，初始化失败")
	}

	domainDao := dao.NewDomainDao(helper)
	shortLinkDao := dao.NewShortLinkDao(helper)

	// 获取所有活跃域名
	domains, err := domainDao.GetActiveDomains()
	if err != nil {
		return nil, fmt.Errorf("获取活跃域名失败: %v", err)
	}

	helper.GetLogger().Info(fmt.Sprintf("开始初始化%d个域名的计数器", len(domains)))

	// 为每个域名初始化计数器
	for _, domain := range domains {
		// 查询该域名下的最大short_link ID
		maxID, err := shortLinkDao.GetMaxIDByDomain(domain.Domain)
		if err != nil {
			return nil, fmt.Errorf("查询域名%s最大ID失败: %v", domain.Domain, err)
		}

		// 初始化计数器
		if err := idGeneratorHelper.InitializeDomainCounter(domain.ID, maxID); err != nil {
			return nil, fmt.Errorf("初始化域名%s Redis计数器失败: %v", domain.Domain, err)
		}

		helper.GetLogger().Info(fmt.Sprintf("域名%s(ID:%d)计数器初始化完成，起始值:%d", domain.Domain, domain.ID, maxID))
	}

	helper.GetLogger().Info("所有域名计数器初始化完成")
	return idGeneratorHelper, nil
}
