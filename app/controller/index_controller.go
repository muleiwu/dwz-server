package controller

import (
	"cnb.cool/mliev/open/dwz-server/app/service"
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
}

func (receiver IndexController) GetIndex(c *gin.Context) {
	// 获取当前访问的域名
	host := c.Request.Host

	// 获取域名信息
	domainService := service.NewDomainService()
	domain, err := domainService.GetDomainByName(host)

	// 默认数据
	pageData := IndexPageData{
		SiteName:     "短网址服务",
		ICPNumber:    "",
		PoliceNumber: "",
		Domain:       host,
	}

	if err == nil {
		pageData.SiteName = domain.SiteName
		if pageData.SiteName == "" {
			pageData.SiteName = "短网址服务"
		}
		pageData.ICPNumber = domain.ICPNumber
		pageData.PoliceNumber = domain.PoliceNumber
	}

	c.HTML(http.StatusOK, "index.html", pageData)
}
