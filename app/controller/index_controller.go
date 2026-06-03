package controller

import (
	"net/http"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/service"
	helperPkg "cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
)

type IndexController struct {
	BaseResponse
}

// IndexPageData 首页模板数据结构
type IndexPageData struct {
	SiteName     string
	LogoURL      string
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
	logoURL := ""
	if branding, brandingErr := service.NewBrandingService(helper).GetPublicBranding(host); brandingErr == nil {
		if branding.BrandName != "" {
			siteName = branding.BrandName
		}
		logoURL = branding.LogoURL
		copyright = branding.CopyrightText
	}
	// 默认数据
	pageData := IndexPageData{
		SiteName:     siteName,
		LogoURL:      logoURL,
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
