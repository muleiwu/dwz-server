package controller

import (
	"net/http"

	"cnb.cool/mliev/dwz/dwz-server/app/service"
	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
	helperPkg "cnb.cool/mliev/dwz/dwz-server/pkg/helper"
)

type IndexController struct {
	BaseResponse
}

// IndexPageData 首页模板数据结构
type IndexPageData struct {
	SiteName     string
	ICPNumber    string
	PoliceNumber string
	Domain       string
	Copyright    string
}

func (receiver IndexController) GetIndex(c httpInterfaces.RouterContextInterface) {
	helper := helperPkg.GetHelper()
	_ = helper
	// 获取当前访问的域名
	host := c.Host()

	// 获取域名信息
	domainService := service.NewDomainService(helper)
	domain, err := domainService.GetDomainByName(host)

	siteName := helper.GetEnv().GetString("website.name", "短网址服务")
	copyright := helper.GetEnv().GetString("website.copyright", "")
	// 默认数据
	pageData := IndexPageData{
		SiteName:     "木雷短网址",
		ICPNumber:    "",
		PoliceNumber: "",
		Domain:       host,
		Copyright:    copyright,
	}

	if err == nil {
		pageData.SiteName = domain.SiteName
		if pageData.SiteName == "" {
			pageData.SiteName = siteName
		}
		pageData.ICPNumber = domain.ICPNumber
		pageData.PoliceNumber = domain.PoliceNumber
	}

	c.HTML(http.StatusOK, "index.html", pageData)
}
