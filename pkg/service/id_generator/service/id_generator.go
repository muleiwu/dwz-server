package service

import (
	"errors"
	"fmt"
	"reflect"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/dao"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/id_generator/assembly"
	"cnb.cool/mliev/open/go-web/pkg/container"
)

// IDGenerator implements go-web's ServerInterface. Run() builds the configured
// driver, walks active domains to seed counters, and registers the resulting
// generator into the container so helper.GetIdGenerator() can resolve it.
type IDGenerator struct{}

func (s *IDGenerator) Run() error {
	h := helper.GetHelper()
	logger := h.GetLogger()
	logger.Info("加载发号器")

	if h.GetInstalled() == nil || !h.GetInstalled().IsInstalled() {
		logger.Warn("应用未安装，初始化发号器停止")
		return nil
	}

	driver := h.GetConfig().GetString("id_generator.driver", "redis")
	logger.Info("加载ID发号器驱动: " + driver)

	if driver == "redis" && h.GetRedis() == nil {
		panic(errors.New("ID发号器驱动配置为：redis，但Redis服务不可用，拒绝启动"))
	}

	idGenerator, err := s.initializeDomainCounters(driver)
	if err != nil {
		return err
	}

	container.Register(container.NewSimpleProvider(reflect.TypeFor[interfaces.IDGenerator](), idGenerator))
	return nil
}

func (s *IDGenerator) Stop() error { return nil }

func (s *IDGenerator) initializeDomainCounters(driver string) (interfaces.IDGenerator, error) {
	h := helper.GetHelper()

	idGenerator, err := assembly.GetDriver(h, driver)
	if err != nil {
		return nil, fmt.Errorf("创建ID发号器失败: %v", err)
	}

	if h.GetDatabase() == nil {
		return nil, errors.New("数据库连接获取失败，初始化失败")
	}

	domainDao := dao.NewDomainDao(h)
	shortLinkDao := dao.NewShortLinkDao(h)

	domains, err := domainDao.GetActiveDomains()
	if err != nil {
		return nil, fmt.Errorf("获取活跃域名失败: %v", err)
	}

	logger := h.GetLogger()
	logger.Info(fmt.Sprintf("开始初始化%d个域名的计数器", len(domains)))

	for _, domain := range domains {
		maxID, err := shortLinkDao.GetMaxIDByDomain(domain.Domain)
		if err != nil {
			return nil, fmt.Errorf("查询域名%s最大ID失败: %v", domain.Domain, err)
		}
		if err := idGenerator.InitializeDomainCounter(domain.ID, maxID); err != nil {
			return nil, fmt.Errorf("初始化域名%s计数器失败: %v", domain.Domain, err)
		}
		logger.Info(fmt.Sprintf("域名%s(ID:%d)计数器初始化完成，起始值:%d", domain.Domain, domain.ID, maxID))
	}

	logger.Info("所有域名计数器初始化完成")
	return idGenerator, nil
}
