package dao

import (
	"cnb.cool/mliev/open/dwz-server/app/model"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"gorm.io/gorm"
)

type UserTokenDAO struct {
	helper interfaces.HelperInterface
}

func NewUserTokenDAO(helper interfaces.HelperInterface) *UserTokenDAO {
	return &UserTokenDAO{
		helper: helper,
	}
}

// Create 创建Token
func (dao *UserTokenDAO) Create(token *model.UserToken) error {
	return dao.helper.GetDatabase().Create(token).Error
}

// GetByID 根据ID获取Token
func (dao *UserTokenDAO) GetByID(id uint64) (*model.UserToken, error) {
	var token model.UserToken
	err := dao.helper.GetDatabase().Preload("User").Where("id = ?", id).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// GetByToken 根据Token值获取Token信息
func (dao *UserTokenDAO) GetByToken(token string) (*model.UserToken, error) {
	var userToken model.UserToken
	err := dao.helper.GetDatabase().Preload("User").Where("token = ? AND status = 1", token).First(&userToken).Error
	if err != nil {
		return nil, err
	}
	return &userToken, nil
}

// Update 更新Token
func (dao *UserTokenDAO) Update(token *model.UserToken) error {
	return dao.helper.GetDatabase().Save(token).Error
}

// Delete 删除Token（软删除）
func (dao *UserTokenDAO) Delete(id uint64) error {
	return dao.helper.GetDatabase().Where("id = ?", id).Delete(&model.UserToken{}).Error
}

// GetListByUserID 根据用户ID获取Token列表
func (dao *UserTokenDAO) GetListByUserID(userID uint64, offset, limit int, tokenName string, status *int8) ([]model.UserToken, int64, error) {
	var tokens []model.UserToken
	var total int64

	query := dao.helper.GetDatabase().Model(&model.UserToken{}).Where("user_id = ?", userID)

	if tokenName != "" {
		query = query.Where("token_name LIKE ?", "%"+tokenName+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取列表
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&tokens).Error
	return tokens, total, err
}

// UpdateLastUsed 更新最后使用时间
func (dao *UserTokenDAO) UpdateLastUsed(id uint64) error {
	return dao.helper.GetDatabase().Model(&model.UserToken{}).Where("id = ?", id).Update("last_used_at", gorm.Expr("NOW()")).Error
}

// DeleteByUserID 删除用户的所有Token
func (dao *UserTokenDAO) DeleteByUserID(userID uint64) error {
	return dao.helper.GetDatabase().Where("user_id = ?", userID).Delete(&model.UserToken{}).Error
}
