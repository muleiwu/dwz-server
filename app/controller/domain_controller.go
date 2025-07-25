package controller

import (
	"strconv"
	"strings"

	"cnb.cool/mliev/open/dwz-server/app/dto"
	"cnb.cool/mliev/open/dwz-server/app/service"
	"cnb.cool/mliev/open/dwz-server/constants"
	"github.com/gin-gonic/gin"
)

type DomainController struct {
	BaseResponse
}

// CreateDomain 创建域名
func (ctrl DomainController) CreateDomain(c *gin.Context) {
	var req dto.DomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	domainService := service.NewDomainService()
	response, err := domainService.CreateDomain(&req)
	if err != nil {
		if strings.Contains(err.Error(), "已存在") {
			ctrl.Error(c, constants.ErrCodeConflict, err.Error())
		} else {
			ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		}
		return
	}

	ctrl.Success(c, response)
}

// GetDomainList 获取域名列表
func (ctrl DomainController) GetDomainList(c *gin.Context) {
	response, err := service.NewDomainService().GetDomainList()
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}

// UpdateDomain 更新域名
func (ctrl DomainController) UpdateDomain(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的ID格式")
		return
	}

	var req dto.DomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	response, err := service.NewDomainService().UpdateDomain(id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			ctrl.Error(c, constants.ErrCodeNotFound, err.Error())
		} else if strings.Contains(err.Error(), "已存在") {
			ctrl.Error(c, constants.ErrCodeConflict, err.Error())
		} else {
			ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		}
		return
	}

	ctrl.Success(c, response)
}

// DeleteDomain 删除域名
func (ctrl DomainController) DeleteDomain(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的ID格式")
		return
	}

	err = service.NewDomainService().DeleteDomain(id)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			ctrl.Error(c, constants.ErrCodeNotFound, err.Error())
		} else {
			ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		}
		return
	}

	ctrl.SuccessWithMessage(c, "删除成功", nil)
}

// GetActiveDomains 获取活跃域名列表
func (ctrl DomainController) GetActiveDomains(c *gin.Context) {
	response, err := service.NewDomainService().GetActiveDomains()
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}
