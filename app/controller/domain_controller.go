package controller

import (
	"strconv"
	"strings"

	"cnb.cool/mliev/open/dwz-server/app/constants"
	"cnb.cool/mliev/open/dwz-server/app/dto"
	"cnb.cool/mliev/open/dwz-server/app/service"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"github.com/gin-gonic/gin"
)

type DomainController struct {
	BaseResponse
}

// CreateDomain 创建域名
func (ctrl DomainController) CreateDomain(c *gin.Context, helper interfaces.HelperInterface) {
	var req dto.DomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	domainService := service.NewDomainService(helper)
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
func (ctrl DomainController) GetDomainList(c *gin.Context, helper interfaces.HelperInterface) {
	response, err := service.NewDomainService(helper).GetDomainList()
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}

// UpdateDomain 更新域名
func (ctrl DomainController) UpdateDomain(c *gin.Context, helper interfaces.HelperInterface) {
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

	// 获取原始域名记录，保护 random_suffix_length 和 enable_checksum 字段
	domainService := service.NewDomainService(helper)
	originalDomain, err := domainService.GetDomainByID(id)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeNotFound, "域名不存在")
		return
	}

	// 使用原始值覆盖请求中的配置字段，防止修改
	req.RandomSuffixLength = originalDomain.RandomSuffixLength
	req.EnableChecksum = originalDomain.EnableChecksum
	req.EnableXorObfuscation = originalDomain.EnableXorObfuscation

	// XorSecret 需要从 uint64 转换为 string
	if originalDomain.XorSecret != nil {
		secretStr := strconv.FormatUint(*originalDomain.XorSecret, 10)
		req.XorSecret = &secretStr
	} else {
		req.XorSecret = nil
	}

	req.XorRot = originalDomain.XorRot

	// 保护 DefaultStartNumber 字段，创建后不允许修改
	if originalDomain.DefaultStartNumber != nil {
		req.DefaultStartNumber = *originalDomain.DefaultStartNumber
	} else {
		req.DefaultStartNumber = 0
	}

	response, err := domainService.UpdateDomain(id, &req)
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

// UpdateStatusDomain UpdateDomain 更新域名
func (ctrl DomainController) UpdateStatusDomain(c *gin.Context, helper interfaces.HelperInterface) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的ID格式")
		return
	}

	var req dto.UpdateStatusDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	response, err := service.NewDomainService(helper).UpdateStatusDomain(id, &req)
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
func (ctrl DomainController) DeleteDomain(c *gin.Context, helper interfaces.HelperInterface) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的ID格式")
		return
	}

	err = service.NewDomainService(helper).DeleteDomain(id)
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
func (ctrl DomainController) GetActiveDomains(c *gin.Context, helper interfaces.HelperInterface) {
	response, err := service.NewDomainService(helper).GetActiveDomains()
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}
