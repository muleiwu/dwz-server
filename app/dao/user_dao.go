package dao

import (
	"cnb.cool/mliev/open/dwz-server/app/model"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"gorm.io/gorm"
)

type UserDAO struct {
	helper interfaces.GetHelperInterface
}

func NewUserDAO(helper interfaces.GetHelperInterface) *UserDAO {
	return &UserDAO{
		helper: helper,
	}
}

// Create 创建用户
func (dao *UserDAO) Create(user *model.User) error {
	return dao.helper.GetDatabase().Create(user).Error
}

// GetByID 根据ID获取用户
func (dao *UserDAO) GetByID(id uint64) (*model.User, error) {
	var user model.User
	err := dao.helper.GetDatabase().Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (dao *UserDAO) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := dao.helper.GetDatabase().Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (dao *UserDAO) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := dao.helper.GetDatabase().Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (dao *UserDAO) Update(user *model.User) error {
	return dao.helper.GetDatabase().Save(user).Error
}

// Delete 删除用户（软删除）
func (dao *UserDAO) Delete(id uint64) error {
	return dao.helper.GetDatabase().Where("id = ?", id).Delete(&model.User{}).Error
}

// GetList 获取用户列表
func (dao *UserDAO) GetList(offset, limit int, username, realName string, status *int8) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := dao.helper.GetDatabase().Model(&model.User{})

	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if realName != "" {
		query = query.Where("real_name LIKE ?", "%"+realName+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取列表
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error
	return users, total, err
}

// UpdateLastLogin 更新最后登录时间
func (dao *UserDAO) UpdateLastLogin(id uint64) error {
	return dao.helper.GetDatabase().Model(&model.User{}).Where("id = ?", id).Update("last_login", gorm.Expr("NOW()")).Error
}

// CheckUsernameExists 检查用户名是否存在
func (dao *UserDAO) CheckUsernameExists(username string, excludeID uint64) (bool, error) {
	var count int64
	query := dao.helper.GetDatabase().Model(&model.User{}).Where("username = ?", username)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// CheckEmailExists 检查邮箱是否存在
func (dao *UserDAO) CheckEmailExists(email string, excludeID uint64) (bool, error) {
	if email == "" {
		return false, nil
	}
	var count int64
	query := dao.helper.GetDatabase().Model(&model.User{}).Where("email = ?", email)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}
