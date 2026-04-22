package dao

import (
	"time"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
)

// OIDCBindingDAO 管理 oidc_bindings 表(本地用户 ↔ 远端 OIDC 身份映射)。
type OIDCBindingDAO struct {
	helper interfaces.HelperInterface
}

func NewOIDCBindingDAO(helper interfaces.HelperInterface) *OIDCBindingDAO {
	return &OIDCBindingDAO{helper: helper}
}

// GetByProviderSub 按 (provider, sub) 唯一索引查询。
func (dao *OIDCBindingDAO) GetByProviderSub(provider, sub string) (*model.OIDCBinding, error) {
	var b model.OIDCBinding
	err := dao.helper.GetDatabase().Where("provider = ? AND sub = ?", provider, sub).First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// GetByUserID 列出某用户的所有绑定。
func (dao *OIDCBindingDAO) GetByUserID(userID uint64) ([]model.OIDCBinding, error) {
	var list []model.OIDCBinding
	err := dao.helper.GetDatabase().Where("user_id = ?", userID).Find(&list).Error
	return list, err
}

// GetByUserAndProvider 返回用户在指定 provider 下的绑定(若存在)。
func (dao *OIDCBindingDAO) GetByUserAndProvider(userID uint64, provider string) (*model.OIDCBinding, error) {
	var b model.OIDCBinding
	err := dao.helper.GetDatabase().Where("user_id = ? AND provider = ?", userID, provider).First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// Create 新增绑定。
func (dao *OIDCBindingDAO) Create(b *model.OIDCBinding) error {
	return dao.helper.GetDatabase().Create(b).Error
}

// UpdateLastLogin 刷新最后登录时间与 email 快照。
func (dao *OIDCBindingDAO) UpdateLastLogin(id uint64, email string) error {
	now := time.Now()
	updates := map[string]any{"last_login_at": now}
	if email != "" {
		updates["email"] = email
	}
	return dao.helper.GetDatabase().Model(&model.OIDCBinding{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteByUserAndProvider 按用户和 provider 解绑。
func (dao *OIDCBindingDAO) DeleteByUserAndProvider(userID uint64, provider string) error {
	return dao.helper.GetDatabase().
		Where("user_id = ? AND provider = ?", userID, provider).
		Delete(&model.OIDCBinding{}).Error
}
