package controller

import (
	"net/http"
	"path/filepath"
	"strings"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/constants"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/middleware"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/service"
	helperPkg "cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
)

type BrandingController struct {
	BaseResponse
}

func (ctrl BrandingController) GetPublicBranding(c httpInterfaces.RouterContextInterface) {
	branding, err := service.NewBrandingService(helperPkg.GetHelper()).GetPublicBranding(c.Host())
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}
	ctrl.Success(c, branding)
}

func (ctrl BrandingController) GetSystemBranding(c httpInterfaces.RouterContextInterface) {
	if !middleware.IsSystemAdmin(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限查看系统品牌")
		return
	}
	branding, err := service.NewBrandingService(helperPkg.GetHelper()).GetSystemBranding()
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}
	ctrl.Success(c, branding)
}

func (ctrl BrandingController) SaveSystemBranding(c httpInterfaces.RouterContextInterface) {
	if !middleware.IsSystemAdmin(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限管理系统品牌")
		return
	}
	var req dto.SystemBrandingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	branding, err := service.NewBrandingService(helperPkg.GetHelper()).SaveSystemBranding(&req)
	if err != nil {
		ctrl.writeBrandingError(c, err)
		return
	}
	ctrl.Success(c, branding)
}

func (ctrl BrandingController) UploadLogo(c httpInterfaces.RouterContextInterface) {
	if !middleware.IsSystemAdmin(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限上传品牌Logo")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请选择要上传的Logo文件")
		return
	}
	logoURL, err := service.NewBrandingService(helperPkg.GetHelper()).StoreLogo(file)
	if err != nil {
		ctrl.writeBrandingError(c, err)
		return
	}
	ctrl.Success(c, dto.LogoUploadResponse{URL: logoURL})
}

func (ctrl BrandingController) ServeLogo(c httpInterfaces.RouterContextInterface) {
	filename := c.Param("filename")
	if filename == "" || filepath.Base(filename) != filename || strings.Contains(filename, "\\") {
		c.Status(http.StatusNotFound)
		return
	}
	path := filepath.Join(service.NewBrandingService(helperPkg.GetHelper()).UploadDir(), filename)
	c.SetHeader("Cache-Control", "public, max-age=31536000, immutable")
	c.File(path)
}

func (ctrl BrandingController) writeBrandingError(c httpInterfaces.RouterContextInterface, err error) {
	message := err.Error()
	switch {
	case strings.Contains(message, "格式") || strings.Contains(message, "支持") || strings.Contains(message, "超过") || strings.Contains(message, "不能为空"):
		ctrl.Error(c, constants.ErrCodeBadRequest, message)
	default:
		ctrl.Error(c, constants.ErrCodeInternal, message)
	}
}
