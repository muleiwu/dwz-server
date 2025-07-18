package controller

import (
	"cnb.cool/mliev/open/dwz-server/app/service"
	"cnb.cool/mliev/open/dwz-server/helper/env"
	"github.com/gin-gonic/gin"
	"net/http"
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

func (receiver IndexController) GetIndex(c *gin.Context) {
	// 获取当前访问的域名
	host := c.Request.Host

	// 获取域名信息
	domainService := service.NewDomainService()
	domain, err := domainService.GetDomainByName(host)

	siteName := env.EnvString("website.name", "短网址服务")
	copyright := env.EnvString("website.name", "")
	// 默认数据
	pageData := IndexPageData{
		SiteName:     "",
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
