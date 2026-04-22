package controller

import (
	"cnb.cool/mliev/dwz/dwz-server/v2/app/constants"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/service"
	helperPkg "cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
)

// OIDCAdminController 暴露 OIDC 配置的后台管理接口。挂载在受认证的 /api/v1/admin/oidc 下。
type OIDCAdminController struct {
	BaseResponse
}

// GetConfig 返回当前 OIDC 配置;尚未配置时返回 data=null。
func (ctrl OIDCAdminController) GetConfig(c httpInterfaces.RouterContextInterface) {
	svc := service.NewOIDCService(helperPkg.GetHelper())
	cfg, err := svc.GetConfigForAdmin()
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}
	ctrl.Success(c, cfg)
}

// SaveConfig 新增或更新 OIDC 配置。
func (ctrl OIDCAdminController) SaveConfig(c httpInterfaces.RouterContextInterface) {
	var req dto.SaveOIDCConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	svc := service.NewOIDCService(helperPkg.GetHelper())
	resp, err := svc.SaveConfig(&req)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, err.Error())
		return
	}
	ctrl.Success(c, resp)
}

// TestConnection 对用户填写的 issuer 做一次 Discovery 探测,验证配置可用。
func (ctrl OIDCAdminController) TestConnection(c httpInterfaces.RouterContextInterface) {
	var req dto.TestOIDCConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	svc := service.NewOIDCService(helperPkg.GetHelper())
	if err := svc.TestConnection(c.Request().Context(), &req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, err.Error())
		return
	}
	ctrl.SuccessWithMessage(c, "连接测试成功", nil)
}
