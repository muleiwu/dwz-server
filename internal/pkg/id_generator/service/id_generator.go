package service

import (
	"errors"
	"fmt"

	"cnb.cool/mliev/open/dwz-server/app/dao"
	"cnb.cool/mliev/open/dwz-server/internal/helper"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"cnb.cool/mliev/open/dwz-server/internal/pkg/id_generator/assembly"
)

type IDGenerator struct {
	Helper interfaces.HelperInterface
}

func (receiver *IDGenerator) Run() error {

	receiver.Helper.GetLogger().Info("加载发号器")

	if receiver.Helper.GetInstalled() == nil || !receiver.Helper.GetInstalled().IsInstalled() {
		receiver.Helper.GetLogger().Warn("应用未安装，初始化发号器停止")
		return nil
	}

	driver := receiver.Helper.GetConfig().GetString("id_generator.driver", "redis")
	receiver.Helper.GetLogger().Info("加载ID发号器驱动: " + driver)

	idGenerator, err := receiver.InitializeDomainCounters(driver)

	if err != nil {
		return err
	}

	helper.SetIdGenerator(idGenerator)

	return nil
}

// InitializeDomainCounters 初始化域名计数器
func (receiver *IDGenerator) InitializeDomainCounters(driver string) (interfaces.IDGenerator, error) {

	idGenerator, err := assembly.GetDriver(receiver.Helper, driver)
	if err != nil {
		return nil, fmt.Errorf("创建ID发号器失败: %v", err)
	}

	if receiver.Helper.GetDatabase() == nil {
		return nil, errors.New("数据库连接获取失败，初始化失败")
	}

	domainDao := dao.NewDomainDao(receiver.Helper)
	shortLinkDao := dao.NewShortLinkDao(receiver.Helper)

	// 获取所有活跃域名
	domains, err := domainDao.GetActiveDomains()
	if err != nil {
		return nil, fmt.Errorf("获取活跃域名失败: %v", err)
	}

	receiver.Helper.GetLogger().Info(fmt.Sprintf("开始初始化%d个域名的计数器", len(domains)))

	// 为每个域名初始化计数器
	for _, domain := range domains {
		// 查询该域名下的最大short_link ID
		maxID, err := shortLinkDao.GetMaxIDByDomain(domain.Domain)
		if err != nil {
			return nil, fmt.Errorf("查询域名%s最大ID失败: %v", domain.Domain, err)
		}

		// 初始化计数器
		if err := idGenerator.InitializeDomainCounter(domain.ID, maxID); err != nil {
			return nil, fmt.Errorf("初始化域名%s Redis计数器失败: %v", domain.Domain, err)
		}

		receiver.Helper.GetLogger().Info(fmt.Sprintf("域名%s(ID:%d)计数器初始化完成，起始值:%d", domain.Domain, domain.ID, maxID))
	}

	receiver.Helper.GetLogger().Info("所有域名计数器初始化完成")
	return idGenerator, nil
}
