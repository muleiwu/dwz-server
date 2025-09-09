package service

import (
	"errors"
	"strconv"
	"time"

	"cnb.cool/mliev/open/dwz-server/app/dao"
	"cnb.cool/mliev/open/dwz-server/app/dto"
	"cnb.cool/mliev/open/dwz-server/app/model"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"gorm.io/gorm"
)

type UserService struct {
	userDAO *dao.UserDAO
	helper  interfaces.HelperInterface
}

func NewUserService(helper interfaces.HelperInterface) *UserService {
	return &UserService{
		helper:  helper,
		userDAO: dao.NewUserDAO(helper),
	}
}

// Login 用户登录
func (s *UserService) Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// 根据用户名查找用户
	user, err := s.userDAO.GetByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}

	// 检查用户状态
	if !user.IsActive() {
		return nil, errors.New("用户已被禁用")
	}

	// 验证密码
	if !user.CheckPassword(req.Password) {
		return nil, errors.New("用户名或密码错误")
	}

	// 更新最后登录时间
	if err := s.userDAO.UpdateLastLogin(user.ID); err != nil {
		// 记录日志但不中断登录流程
	}

	// 生成Token（这里简化处理，实际应该使用JWT）
	token := "user_" + strconv.FormatUint(user.ID, 10) + "_" + strconv.FormatInt(time.Now().Unix(), 10)
	expiresAt := time.Now().Add(24 * time.Hour) // 24小时过期

	return &dto.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      s.convertToUserInfo(user),
	}, nil
}

// CreateUser 创建用户
func (s *UserService) CreateUser(req *dto.CreateUserRequest) (*dto.UserInfo, error) {
	// 检查用户名是否存在
	exists, err := s.userDAO.CheckUsernameExists(req.Username, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否存在
	if req.Email != "" {
		exists, err := s.userDAO.CheckEmailExists(req.Email, 0)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("邮箱已存在")
		}
	}

	// 创建用户
	user := &model.User{
		Username: req.Username,
		RealName: req.RealName,
		Email:    req.Email,
		Phone:    req.Phone,
		Status:   1, // 默认启用
	}

	// 设置密码
	if err := user.SetPassword(req.Password); err != nil {
		return nil, errors.New("密码加密失败")
	}

	// 保存到数据库
	if err := s.userDAO.Create(user); err != nil {
		return nil, err
	}

	userInfo := s.convertToUserInfo(user)
	return &userInfo, nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(id uint64, req *dto.UpdateUserRequest) (*dto.UserInfo, error) {
	// 获取用户
	user, err := s.userDAO.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 检查邮箱是否存在
	if req.Email != "" && req.Email != user.Email {
		exists, err := s.userDAO.CheckEmailExists(req.Email, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("邮箱已存在")
		}
	}

	// 更新字段
	if req.RealName != "" {
		user.RealName = req.RealName
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	// 保存更新
	if err := s.userDAO.Update(user); err != nil {
		return nil, err
	}

	userInfo := s.convertToUserInfo(user)
	return &userInfo, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id uint64) error {
	// 检查用户是否存在
	_, err := s.userDAO.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	// 删除用户的所有Token
	tokenDAO := dao.NewUserTokenDAO(s.helper)
	if err := tokenDAO.DeleteByUserID(id); err != nil {
		// 记录日志但不中断删除流程
	}

	// 删除用户
	return s.userDAO.Delete(id)
}

// GetUser 获取用户详情
func (s *UserService) GetUser(id uint64) (*dto.UserInfo, error) {
	user, err := s.userDAO.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	userInfo := s.convertToUserInfo(user)
	return &userInfo, nil
}

// GetUserList 获取用户列表
func (s *UserService) GetUserList(req *dto.UserListRequest) (*dto.UserListResponse, error) {
	offset := (req.Page - 1) * req.PageSize
	users, total, err := s.userDAO.GetList(offset, req.PageSize, req.Username, req.RealName, req.Status)
	if err != nil {
		return nil, err
	}

	var userInfos []dto.UserInfo
	for _, user := range users {
		userInfos = append(userInfos, s.convertToUserInfo(&user))
	}

	return &dto.UserListResponse{
		List:       userInfos,
		Pagination: dto.NewPagination(total, req.Page, req.PageSize),
	}, nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(id uint64, req *dto.ChangePasswordRequest) error {
	// 获取用户
	user, err := s.userDAO.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	// 验证旧密码
	if !user.CheckPassword(req.OldPassword) {
		return errors.New("原密码错误")
	}

	// 设置新密码
	if err := user.SetPassword(req.NewPassword); err != nil {
		return errors.New("密码加密失败")
	}

	// 保存更新
	return s.userDAO.Update(user)
}

// ResetPassword 重置密码
func (s *UserService) ResetPassword(id uint64, req *dto.ResetPasswordRequest) error {
	// 获取用户
	user, err := s.userDAO.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	// 设置新密码
	if err := user.SetPassword(req.NewPassword); err != nil {
		return errors.New("密码加密失败")
	}

	// 保存更新
	return s.userDAO.Update(user)
}

// ConvertToUserInfo 转换为UserInfo（公开方法）
func (s *UserService) ConvertToUserInfo(user *model.User) dto.UserInfo {
	return s.convertToUserInfo(user)
}

// convertToUserInfo 转换为UserInfo
func (s *UserService) convertToUserInfo(user *model.User) dto.UserInfo {
	return dto.UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		RealName:  user.RealName,
		Email:     user.Email,
		Phone:     user.Phone,
		Status:    user.Status,
		LastLogin: user.LastLogin,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
