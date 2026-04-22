package dao

import (
	"errors"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
	"gorm.io/gorm"
)

// OIDCProviderDAO 负责 oidc_providers 的增改查。
// 当前策略:系统只允许一个启用态的 provider,但表结构允许多行以便后续扩展。
type OIDCProviderDAO struct {
	helper interfaces.HelperInterface
}

func NewOIDCProviderDAO(helper interfaces.HelperInterface) *OIDCProviderDAO {
	return &OIDCProviderDAO{helper: helper}
}

// GetEnabled 返回当前启用中的 provider,若无则返回 gorm.ErrRecordNotFound。
func (dao *OIDCProviderDAO) GetEnabled() (*model.OIDCProvider, error) {
	var p model.OIDCProvider
	err := dao.helper.GetDatabase().Where("enabled = ?", 1).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetByName 按 name 精确查询。
func (dao *OIDCProviderDAO) GetByName(name string) (*model.OIDCProvider, error) {
	var p model.OIDCProvider
	err := dao.helper.GetDatabase().Where("name = ?", name).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetFirst 返回数据库中第一个 provider(当前单 provider 约束下等价于"当前 provider")。
// 查不到返回 nil, nil;其它错误回传。
func (dao *OIDCProviderDAO) GetFirst() (*model.OIDCProvider, error) {
	var p model.OIDCProvider
	err := dao.helper.GetDatabase().Order("id ASC").First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// Upsert 基于 name 做创建/更新。调用方应保证敏感字段已加密。
func (dao *OIDCProviderDAO) Upsert(p *model.OIDCProvider) error {
	var existing model.OIDCProvider
	err := dao.helper.GetDatabase().Where("name = ?", p.Name).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dao.helper.GetDatabase().Create(p).Error
		}
		return err
	}
	p.ID = existing.ID
	p.CreatedAt = existing.CreatedAt
	return dao.helper.GetDatabase().Save(p).Error
}

// DisableAllExcept 关闭除 exceptID 以外所有已启用的 provider。
// 当 exceptID 为 0 时关闭全部。保证运行时仅一个 active provider。
func (dao *OIDCProviderDAO) DisableAllExcept(exceptID uint64) error {
	db := dao.helper.GetDatabase().Model(&model.OIDCProvider{}).Where("enabled = ?", 1)
	if exceptID > 0 {
		db = db.Where("id <> ?", exceptID)
	}
	return db.Update("enabled", 0).Error
}
