package service

import (
	"errors"

	"cnb.cool/mliev/open/dwz-server/app/dao"
	"cnb.cool/mliev/open/dwz-server/app/dto"
	"cnb.cool/mliev/open/dwz-server/app/model"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"gorm.io/gorm"
)

type UserTokenService struct {
	tokenDAO *dao.UserTokenDAO
	userDAO  *dao.UserDAO
}

func NewUserTokenService(helper interfaces.GetHelperInterface) *UserTokenService {
	return &UserTokenService{
		tokenDAO: dao.NewUserTokenDAO(helper),
		userDAO:  dao.NewUserDAO(helper),
	}
}

// CreateToken 创建Token
func (s *UserTokenService) CreateToken(userID uint64, req *dto.CreateUserTokenRequest) (*dto.CreateUserTokenResponse, error) {
	// 检查用户是否存在
	_, err := s.userDAO.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 创建Token
	token := &model.UserToken{
		UserID:    userID,
		TokenName: req.TokenName,
		ExpireAt:  req.ExpireAt,
		Status:    1, // 默认启用
	}

	// 生成Token值
	if err := token.GenerateToken(); err != nil {
		return nil, errors.New("Token生成失败")
	}

	// 保存到数据库
	if err := s.tokenDAO.Create(token); err != nil {
		return nil, err
	}

	return &dto.CreateUserTokenResponse{
		ID:        token.ID,
		TokenName: token.TokenName,
		Token:     token.Token,
		ExpireAt:  token.ExpireAt,
		CreatedAt: token.CreatedAt,
	}, nil
}

// GetTokenList 获取Token列表
func (s *UserTokenService) GetTokenList(userID uint64, req *dto.UserTokenListRequest) (*dto.UserTokenListResponse, error) {
	offset := (req.Page - 1) * req.PageSize
	tokens, total, err := s.tokenDAO.GetListByUserID(userID, offset, req.PageSize, req.TokenName, req.Status)
	if err != nil {
		return nil, err
	}

	var tokenInfos []dto.UserTokenInfo
	for _, token := range tokens {
		tokenInfos = append(tokenInfos, s.convertToTokenInfo(&token))
	}

	return &dto.UserTokenListResponse{
		List:       tokenInfos,
		Pagination: dto.NewPagination(total, req.Page, req.PageSize),
	}, nil
}

// DeleteToken 删除Token
func (s *UserTokenService) DeleteToken(userID, tokenID uint64) error {
	// 获取Token
	token, err := s.tokenDAO.GetByID(tokenID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Token不存在")
		}
		return err
	}

	// 检查Token是否属于该用户
	if token.UserID != userID {
		return errors.New("无权限删除该Token")
	}

	return s.tokenDAO.Delete(tokenID)
}

// ValidateToken 验证Token
func (s *UserTokenService) ValidateToken(tokenStr string) (*model.User, error) {
	// 根据Token获取Token信息
	token, err := s.tokenDAO.GetByToken(tokenStr)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("无效的Token")
		}
		return nil, err
	}

	// 检查Token状态
	if !token.IsActive() {
		return nil, errors.New("Token已失效")
	}

	// 检查用户状态
	if !token.User.IsActive() {
		return nil, errors.New("用户已被禁用")
	}

	// 更新最后使用时间
	token.UpdateLastUsed()
	if err := s.tokenDAO.Update(token); err != nil {
		// 记录日志但不中断验证流程
	}

	return &token.User, nil
}

// convertToTokenInfo 转换为TokenInfo
func (s *UserTokenService) convertToTokenInfo(token *model.UserToken) dto.UserTokenInfo {
	// 只显示Token的前8位
	maskedToken := token.Token
	if len(maskedToken) > 8 {
		maskedToken = maskedToken[:8] + "..."
	}

	return dto.UserTokenInfo{
		ID:         token.ID,
		TokenName:  token.TokenName,
		Token:      maskedToken,
		LastUsedAt: token.LastUsedAt,
		ExpireAt:   token.ExpireAt,
		Status:     token.Status,
		CreatedAt:  token.CreatedAt,
	}
}
