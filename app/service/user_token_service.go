package service

import (
	"errors"

	"cnb.cool/mliev/open/dwz-server/app/dao"
	"cnb.cool/mliev/open/dwz-server/app/dto"
	"cnb.cool/mliev/open/dwz-server/app/model"
	"cnb.cool/mliev/open/dwz-server/internal/helper"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"gorm.io/gorm"
)

type UserTokenService struct {
	tokenDAO        *dao.UserTokenDAO
	userDAO         *dao.UserDAO
	signatureHelper *helper.SignatureHelper
}

func NewUserTokenService(helperInterface interfaces.HelperInterface) *UserTokenService {
	return &UserTokenService{
		tokenDAO:        dao.NewUserTokenDAO(helperInterface),
		userDAO:         dao.NewUserDAO(helperInterface),
		signatureHelper: helper.GetSignatureHelper(),
	}
}

// CreateToken 创建Token
// 支持两种类型：signature（默认）和 bearer
// - signature 类型：生成 AppID 和 AppSecret，用于 HMAC-SHA256 签名认证
// - bearer 类型：生成传统 Token，用于 Bearer Token 认证
func (s *UserTokenService) CreateToken(userID uint64, req *dto.CreateUserTokenRequest) (*dto.CreateUserTokenResponse, error) {
	// 检查用户是否存在
	_, err := s.userDAO.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 确定 Token 类型，默认为 signature
	tokenType := req.TokenType
	if tokenType == "" {
		tokenType = model.TokenTypeSignature
	}

	// 创建Token
	token := &model.UserToken{
		UserID:    userID,
		TokenName: req.TokenName,
		TokenType: tokenType,
		ExpireAt:  req.ExpireAt,
		Status:    1, // 默认启用
	}

	// 用于返回给用户的明文 AppSecret（仅创建时返回一次）
	var plainAppSecret string

	if tokenType == model.TokenTypeSignature {
		// 签名类型：生成 AppID 和 AppSecret
		if err := token.GenerateAppID(); err != nil {
			return nil, errors.New("AppID生成失败")
		}

		// 生成明文 AppSecret
		plainAppSecret, err = token.GenerateAppSecret()
		if err != nil {
			return nil, errors.New("AppSecret生成失败")
		}

		// 加密 AppSecret 后存储
		encryptedSecret, err := s.signatureHelper.EncryptAppSecret(plainAppSecret)
		if err != nil {
			return nil, errors.New("AppSecret加密失败")
		}
		token.AppSecret = encryptedSecret
	} else {
		// Bearer 类型：生成传统 Token
		if err := token.GenerateToken(); err != nil {
			return nil, errors.New("Token生成失败")
		}
	}

	// 保存到数据库
	if err := s.tokenDAO.Create(token); err != nil {
		return nil, err
	}

	// 构建响应
	response := &dto.CreateUserTokenResponse{
		ID:        token.ID,
		TokenName: token.TokenName,
		TokenType: token.TokenType,
		ExpireAt:  token.ExpireAt,
		CreatedAt: token.CreatedAt,
	}

	// 根据类型设置不同的返回字段
	if tokenType == model.TokenTypeSignature {
		if token.AppID != nil {
			response.AppID = *token.AppID
		}
		response.AppSecret = plainAppSecret // 明文 AppSecret，仅创建时返回
	} else {
		if token.Token != nil {
			response.Token = *token.Token
		}
	}

	return response, nil
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
	info := dto.UserTokenInfo{
		ID:         token.ID,
		TokenName:  token.TokenName,
		TokenType:  token.TokenType,
		LastUsedAt: token.LastUsedAt,
		ExpireAt:   token.ExpireAt,
		Status:     token.Status,
		CreatedAt:  token.CreatedAt,
	}

	// 根据类型设置不同的字段
	if token.TokenType == model.TokenTypeSignature {
		// 签名类型：显示 AppID（不显示 AppSecret）
		if token.AppID != nil {
			info.AppID = *token.AppID
		}
	} else {
		// Bearer 类型：只显示 Token 的前8位
		if token.Token != nil {
			maskedToken := *token.Token
			if len(maskedToken) > 8 {
				maskedToken = maskedToken[:8] + "..."
			}
			info.Token = maskedToken
		}
	}

	return info
}
