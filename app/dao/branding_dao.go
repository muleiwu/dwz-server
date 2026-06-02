package dao

import (
	"errors"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
	"gorm.io/gorm"
)

type BrandingDao struct {
	helper interfaces.HelperInterface
}

func NewBrandingDao(helper interfaces.HelperInterface) *BrandingDao {
	return &BrandingDao{helper: helper}
}

func (d *BrandingDao) FindSystem() (*model.SystemBranding, error) {
	var branding model.SystemBranding
	err := d.helper.GetDatabase().Where("id = ?", 1).First(&branding).Error
	if err != nil {
		return nil, err
	}
	return &branding, nil
}

func (d *BrandingDao) UpsertSystemBase(branding *model.SystemBranding) error {
	var existing model.SystemBranding
	err := d.helper.GetDatabase().Where("id = ?", branding.ID).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return d.helper.GetDatabase().Create(branding).Error
		}
		return err
	}
	existing.BrandName = branding.BrandName
	existing.LogoURL = branding.LogoURL
	return d.helper.GetDatabase().Save(&existing).Error
}
