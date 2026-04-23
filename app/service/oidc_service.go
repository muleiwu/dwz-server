package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/dao"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

const (
	oidcStateCacheKeyPrefix = "oidc:state:"
	oidcStateTTL            = 5 * time.Minute
	defaultOIDCScopes       = "openid profile email"
)

// OIDCFlowType 标记一次授权流的用途:登录 / 绑定。
type OIDCFlowType string

const (
	OIDCFlowLogin OIDCFlowType = "login"
	OIDCFlowBind  OIDCFlowType = "bind"
)

// oidcStateEntry 跨 authorize → callback 的一次性状态。
type oidcStateEntry struct {
	Flow         OIDCFlowType `json:"flow"`
	CodeVerifier string       `json:"code_verifier"`
	Nonce        string       `json:"nonce"`
	BindUserID   uint64       `json:"bind_user_id,omitempty"`
	ReturnTo     string       `json:"return_to,omitempty"`
	RedirectURI  string       `json:"redirect_uri"`
	ProviderName string       `json:"provider"`
	CreatedAt    int64        `json:"created_at"`
}

// OIDCResult HandleCallback 的结果,login 流返回 JWT,bind 流只返回用户信息。
type OIDCResult struct {
	Flow      OIDCFlowType
	User      *model.User
	Token     string
	ExpiresAt time.Time
	ReturnTo  string
}

// OIDCService 承担 OIDC 发现、授权 URL 构造、回调处理与绑定管理。
// 底层 *oidc.Provider + *oauth2.Config 按 issuer+clientID 缓存,
// 管理员修改配置时调用 Invalidate() 主动失效。
type OIDCService struct {
	helper      interfaces.HelperInterface
	providerDAO *dao.OIDCProviderDAO
	bindingDAO  *dao.OIDCBindingDAO
	userDAO     *dao.UserDAO

	mu            sync.Mutex
	cachedKey     string
	cachedOIDC    *oidc.Provider
	cachedOAuth2  *oauth2.Config
	cachedEntity  *model.OIDCProvider
	cachedAt      time.Time
	cacheValidity time.Duration
}

// NewOIDCService 构造 OIDCService。
func NewOIDCService(h interfaces.HelperInterface) *OIDCService {
	return &OIDCService{
		helper:        h,
		providerDAO:   dao.NewOIDCProviderDAO(h),
		bindingDAO:    dao.NewOIDCBindingDAO(h),
		userDAO:       dao.NewUserDAO(h),
		cacheValidity: 10 * time.Minute,
	}
}

// GetEnabledProvider 返回当前启用的 provider 实体(明文 client_secret 已解密)。
// 无启用时返回 (nil, nil)。
func (s *OIDCService) GetEnabledProvider() (*model.OIDCProvider, error) {
	p, err := s.providerDAO.GetEnabled()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	plain, err := helper.DecryptSecret(p.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("解密 client_secret 失败: %w", err)
	}
	p.ClientSecret = plain
	return p, nil
}

// GetLoginOptions 提供给登录页的公开信息。
func (s *OIDCService) GetLoginOptions() (*dto.LoginOptionsResponse, error) {
	p, err := s.GetEnabledProvider()
	if err != nil {
		return nil, err
	}
	resp := &dto.LoginOptionsResponse{OIDCEnabled: p != nil && p.IsEnabled()}
	if resp.OIDCEnabled {
		resp.OIDCProvider = p.Name
		if p.DisplayName != "" {
			resp.OIDCDisplayName = p.DisplayName
		} else {
			resp.OIDCDisplayName = p.Name
		}
		// 只有 exclusive 启用且当前没有 bypass 时才让前端隐藏密码表单。
		if p.IsExclusive() && !s.IsExclusiveBypass() {
			resp.LocalLoginDisabled = true
		}
	}
	return resp, nil
}

// IsExclusiveMode 返回 (是否处于 OIDC 唯一登录态, provider 名)。
// 仅当存在启用的 provider 且 exclusive=1 才为 true。错误只在 DB 异常时返回,
// 调用方可按"未启用"处理以保守地放行。
func (s *OIDCService) IsExclusiveMode() (bool, string, error) {
	p, err := s.GetEnabledProvider()
	if err != nil {
		return false, "", err
	}
	if p == nil || !p.IsExclusive() {
		return false, "", nil
	}
	return true, p.Name, nil
}

// IsExclusiveBypass 读环境变量 OIDC_EXCLUSIVE_BYPASS(viper 里对应 key
// oidc.exclusive_bypass)判断 breakglass 是否开启。兼容 "1" / "true" / "yes"
// / "on" 这几种常见写法。
func (s *OIDCService) IsExclusiveBypass() bool {
	v := strings.ToLower(strings.TrimSpace(s.helper.GetEnv().GetString("oidc.exclusive_bypass", "")))
	switch v {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

// BuildAuthURL 生成授权 URL 并将 state 写入缓存。
// flow=login 无需 userID;flow=bind 时 userID 必填。
func (s *OIDCService) BuildAuthURL(ctx context.Context, flow OIDCFlowType, userID uint64, returnTo string) (string, error) {
	if flow != OIDCFlowLogin && flow != OIDCFlowBind {
		return "", errors.New("unsupported flow")
	}
	if flow == OIDCFlowBind && userID == 0 {
		return "", errors.New("bind flow requires authenticated user")
	}

	p, err := s.GetEnabledProvider()
	if err != nil {
		return "", err
	}
	if p == nil || !p.IsEnabled() {
		return "", errors.New("OIDC 登录未启用")
	}

	_, oauthCfg, err := s.buildClients(ctx, p)
	if err != nil {
		return "", err
	}

	state, err := randomToken(24)
	if err != nil {
		return "", err
	}
	verifier, err := randomToken(32)
	if err != nil {
		return "", err
	}
	nonce, err := randomToken(16)
	if err != nil {
		return "", err
	}

	entry := oidcStateEntry{
		Flow:         flow,
		CodeVerifier: verifier,
		Nonce:        nonce,
		BindUserID:   userID,
		ReturnTo:     returnTo,
		RedirectURI:  oauthCfg.RedirectURL,
		ProviderName: p.Name,
		CreatedAt:    time.Now().Unix(),
	}
	if err := s.helper.GetCache().Set(ctx, oidcStateCacheKeyPrefix+state, entry, oidcStateTTL); err != nil {
		return "", fmt.Errorf("保存 state 失败: %w", err)
	}

	challenge := pkceChallenge(verifier)
	authURL := oauthCfg.AuthCodeURL(
		state,
		oauth2.AccessTypeOnline,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oidc.Nonce(nonce),
	)
	return authURL, nil
}

// HandleCallback 处理 IdP 回跳:校验 state、换 token、验 id_token,然后按 flow 处理账户。
func (s *OIDCService) HandleCallback(ctx context.Context, code, state string) (*OIDCResult, error) {
	if code == "" || state == "" {
		return nil, errors.New("缺少 code 或 state")
	}
	cacheKey := oidcStateCacheKeyPrefix + state
	var entry oidcStateEntry
	if err := s.helper.GetCache().Get(ctx, cacheKey, &entry); err != nil {
		return nil, errors.New("state 无效或已过期")
	}
	// state 一次性使用,尽早清理。
	_ = s.helper.GetCache().Del(ctx, cacheKey)

	p, err := s.GetEnabledProvider()
	if err != nil {
		return nil, err
	}
	if p == nil || p.Name != entry.ProviderName {
		return nil, errors.New("OIDC 配置不可用或已变更,请重新发起登录")
	}

	oidcProv, oauthCfg, err := s.buildClients(ctx, p)
	if err != nil {
		return nil, err
	}
	// 使用 state 中快照的 redirect_uri,确保与授权请求一致。
	if entry.RedirectURI != "" {
		oauthCfg.RedirectURL = entry.RedirectURI
	}

	token, err := oauthCfg.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", entry.CodeVerifier),
	)
	if err != nil {
		return nil, fmt.Errorf("获取 token 失败: %w", err)
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return nil, errors.New("IdP 未返回 id_token")
	}
	verifier := oidcProv.Verifier(&oidc.Config{ClientID: p.ClientID})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("id_token 验证失败: %w", err)
	}
	if idToken.Nonce != entry.Nonce {
		return nil, errors.New("nonce 不匹配")
	}

	var claims struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		PreferredName string `json:"preferred_username"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("解析 claims 失败: %w", err)
	}
	if claims.Sub == "" {
		return nil, errors.New("id_token 缺少 sub")
	}

	switch entry.Flow {
	case OIDCFlowBind:
		user, err := s.applyBind(entry.BindUserID, p.Name, claims.Sub, claims.Email)
		if err != nil {
			return nil, err
		}
		return &OIDCResult{Flow: OIDCFlowBind, User: user, ReturnTo: entry.ReturnTo}, nil
	case OIDCFlowLogin:
		user, err := s.resolveLoginUser(p.Name, claims.Sub, claims.Email, claims.EmailVerified, claims.PreferredName, claims.Name)
		if err != nil {
			return nil, err
		}
		tok, expires, err := s.issueJWT(user)
		if err != nil {
			return nil, err
		}
		_ = s.userDAO.UpdateLastLogin(user.ID)
		return &OIDCResult{
			Flow:      OIDCFlowLogin,
			User:      user,
			Token:     tok,
			ExpiresAt: expires,
			ReturnTo:  entry.ReturnTo,
		}, nil
	}
	return nil, errors.New("未知的授权流")
}

// Unbind 解除当前用户与指定 provider 的绑定。
func (s *OIDCService) Unbind(userID uint64, provider string) error {
	if userID == 0 || provider == "" {
		return errors.New("参数不完整")
	}
	return s.bindingDAO.DeleteByUserAndProvider(userID, provider)
}

// ListUserBindings 返回当前用户在各 provider 的绑定情况(附 provider 显示名)。
func (s *OIDCService) ListUserBindings(userID uint64) ([]dto.OIDCBindingInfo, error) {
	bindings, err := s.bindingDAO.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	p, err := s.GetEnabledProvider()
	displayMap := map[string]string{}
	if err == nil && p != nil {
		name := p.DisplayName
		if name == "" {
			name = p.Name
		}
		displayMap[p.Name] = name
	}
	out := make([]dto.OIDCBindingInfo, 0, len(bindings))
	for _, b := range bindings {
		name := displayMap[b.Provider]
		if name == "" {
			name = b.Provider
		}
		out = append(out, dto.OIDCBindingInfo{
			Provider:    b.Provider,
			DisplayName: name,
			Sub:         b.Sub,
			Email:       b.Email,
			LastLoginAt: b.LastLoginAt,
			CreatedAt:   b.CreatedAt,
		})
	}
	return out, nil
}

// Invalidate 清空内存缓存,用于配置保存后强制下次重建。
func (s *OIDCService) Invalidate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cachedKey = ""
	s.cachedOIDC = nil
	s.cachedOAuth2 = nil
	s.cachedEntity = nil
}

// --- 内部工具 ---

func (s *OIDCService) buildClients(ctx context.Context, p *model.OIDCProvider) (*oidc.Provider, *oauth2.Config, error) {
	key := fmt.Sprintf("%s|%s|%d", p.Issuer, p.ClientID, p.UpdatedAt.UnixNano())
	s.mu.Lock()
	if s.cachedKey == key && s.cachedOIDC != nil && s.cachedOAuth2 != nil && time.Since(s.cachedAt) < s.cacheValidity {
		prov, cfg := s.cachedOIDC, *s.cachedOAuth2
		s.mu.Unlock()
		return prov, &cfg, nil
	}
	s.mu.Unlock()

	prov, err := oidc.NewProvider(ctx, p.Issuer)
	if err != nil {
		return nil, nil, fmt.Errorf("发现 OIDC 端点失败: %w", err)
	}
	scopes := splitScopes(p.Scopes)
	redirect := p.RedirectURI
	if redirect == "" {
		redirect = s.defaultRedirectURI()
	}
	cfg := &oauth2.Config{
		ClientID:     p.ClientID,
		ClientSecret: p.ClientSecret,
		Endpoint:     prov.Endpoint(),
		RedirectURL:  redirect,
		Scopes:       scopes,
	}

	s.mu.Lock()
	s.cachedKey = key
	s.cachedOIDC = prov
	s.cachedOAuth2 = cfg
	s.cachedEntity = p
	s.cachedAt = time.Now()
	s.mu.Unlock()

	cfgCopy := *cfg
	return prov, &cfgCopy, nil
}

// defaultRedirectURI 从全局配置中拼出默认回调地址。
func (s *OIDCService) defaultRedirectURI() string {
	cfg := s.helper.GetConfig()
	if v := cfg.GetString("oidc.default_redirect_uri", ""); v != "" {
		return v
	}
	base := strings.TrimRight(cfg.GetString("shortlink.domain", ""), "/")
	if base == "" {
		base = strings.TrimRight(cfg.GetString("http.base_url", ""), "/")
	}
	if base == "" {
		addr := cfg.GetString("http.addr", ":8080")
		if strings.HasPrefix(addr, ":") {
			addr = "http://localhost" + addr
		}
		base = strings.TrimRight(addr, "/")
	}
	return base + "/api/v1/auth/oidc/callback"
}

// resolveLoginUser 按规则定位/创建本地用户:
// 1. 已有 (provider, sub) 绑定 → 直接返回绑定用户
// 2. email_verified && 本地有相同 email 用户 → 建立绑定
// 3. 否则自动创建新用户并绑定
func (s *OIDCService) resolveLoginUser(provider, sub, email string, emailVerified bool, preferredName, fullName string) (*model.User, error) {
	if binding, err := s.bindingDAO.GetByProviderSub(provider, sub); err == nil {
		user, err := s.userDAO.GetByID(binding.UserID)
		if err != nil {
			return nil, fmt.Errorf("绑定的用户已不存在: %w", err)
		}
		if !user.IsActive() {
			return nil, errors.New("用户已被禁用")
		}
		_ = s.bindingDAO.UpdateLastLogin(binding.ID, email)
		return user, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if email != "" && emailVerified {
		if user, err := s.userDAO.GetByEmail(email); err == nil {
			if !user.IsActive() {
				return nil, errors.New("用户已被禁用")
			}
			if err := s.createBinding(user.ID, provider, sub, email); err != nil {
				return nil, err
			}
			return user, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	user, err := s.createUserFromClaims(provider, sub, email, preferredName, fullName)
	if err != nil {
		return nil, err
	}
	if err := s.createBinding(user.ID, provider, sub, email); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *OIDCService) createUserFromClaims(provider, sub, email, preferredName, fullName string) (*model.User, error) {
	username := sanitizeUsername(preferredName)
	if username == "" {
		username = fmt.Sprintf("oidc_%s_%s", provider, shortSub(sub))
	}
	username, err := s.ensureUsernameUnique(username)
	if err != nil {
		return nil, err
	}
	realName := fullName
	if realName == "" {
		realName = preferredName
	}
	user := &model.User{
		Username: username,
		RealName: realName,
		Email:    email,
		Status:   1,
	}
	randomPwd, err := randomToken(24)
	if err != nil {
		return nil, err
	}
	if err := user.SetPassword(randomPwd); err != nil {
		return nil, fmt.Errorf("初始化密码失败: %w", err)
	}
	if err := s.userDAO.Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}
	return user, nil
}

func (s *OIDCService) ensureUsernameUnique(base string) (string, error) {
	candidate := base
	for i := 0; i < 5; i++ {
		exists, err := s.userDAO.CheckUsernameExists(candidate, 0)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
		suffix, err := randomToken(3)
		if err != nil {
			return "", err
		}
		candidate = fmt.Sprintf("%s_%s", base, suffix[:6])
	}
	return "", errors.New("无法生成唯一用户名")
}

func (s *OIDCService) createBinding(userID uint64, provider, sub, email string) error {
	now := time.Now()
	b := &model.OIDCBinding{
		UserID:      userID,
		Provider:    provider,
		Sub:         sub,
		Email:       email,
		LastLoginAt: &now,
	}
	return s.bindingDAO.Create(b)
}

// applyBind 把当前已登录用户与远端身份建立绑定。
func (s *OIDCService) applyBind(userID uint64, provider, sub, email string) (*model.User, error) {
	if userID == 0 {
		return nil, errors.New("用户未登录")
	}
	user, err := s.userDAO.GetByID(userID)
	if err != nil {
		return nil, err
	}

	existing, err := s.bindingDAO.GetByProviderSub(provider, sub)
	if err == nil {
		if existing.UserID == userID {
			_ = s.bindingDAO.UpdateLastLogin(existing.ID, email)
			return user, nil
		}
		return nil, errors.New("该 OIDC 账户已绑定到其他用户")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 若该用户在此 provider 下已有绑定(对应另一个 sub),先删后建。
	if prior, err := s.bindingDAO.GetByUserAndProvider(userID, provider); err == nil && prior != nil {
		_ = s.bindingDAO.DeleteByUserAndProvider(userID, provider)
	}

	if err := s.createBinding(userID, provider, sub, email); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *OIDCService) issueJWT(user *model.User) (string, time.Time, error) {
	cfg := s.helper.GetConfig()
	secret := cfg.GetString("jwt.secret", "")
	if secret == "" {
		return "", time.Time{}, errors.New("JWT secret 未配置")
	}
	helper.InitJWTHelper(secret, cfg.GetInt("jwt.expire_hours", 24))
	return helper.GetJWTHelper().GenerateToken(user.ID, user.Username)
}

// SaveConfig 后台保存 OIDC 配置。若 ClientSecret 留空则保留原值。
// Enabled=true 时自动关闭其他 provider(当前仅 1 个,保留接口以便未来扩展)。
func (s *OIDCService) SaveConfig(req *dto.SaveOIDCConfigRequest) (*dto.OIDCConfigResponse, error) {
	if req == nil {
		return nil, errors.New("请求为空")
	}
	scopes := strings.TrimSpace(req.Scopes)
	if scopes == "" {
		scopes = defaultOIDCScopes
	}

	existing, err := s.providerDAO.GetByName(req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 开启 OIDC 唯一模式前,先校验至少已有一个绑定,防止管理员把自己锁在外面。
	if req.Enabled && req.Exclusive {
		count, err := s.bindingDAO.CountByProvider(req.Name)
		if err != nil {
			return nil, fmt.Errorf("校验已有绑定失败: %w", err)
		}
		if count == 0 {
			return nil, errors.New("开启 OIDC 唯一登录前,需要至少有一位管理员先完成 OIDC 绑定")
		}
	}

	entity := &model.OIDCProvider{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Issuer:      strings.TrimSpace(req.Issuer),
		ClientID:    strings.TrimSpace(req.ClientID),
		Scopes:      scopes,
		RedirectURI: strings.TrimSpace(req.RedirectURI),
	}
	if req.Enabled {
		entity.Enabled = 1
	}
	// exclusive 仅在 enabled 时有意义;如果调用方要求 disable 自然也会关掉 exclusive。
	if req.Enabled && req.Exclusive {
		entity.Exclusive = 1
	}

	if req.ClientSecret != "" {
		encrypted, err := helper.EncryptSecret(req.ClientSecret)
		if err != nil {
			return nil, fmt.Errorf("加密 client_secret 失败: %w", err)
		}
		entity.ClientSecret = encrypted
	} else if existing != nil {
		entity.ClientSecret = existing.ClientSecret
	} else {
		return nil, errors.New("首次保存时必须提供 client_secret")
	}

	if err := s.providerDAO.Upsert(entity); err != nil {
		return nil, err
	}
	if entity.Enabled == 1 {
		if err := s.providerDAO.DisableAllExcept(entity.ID); err != nil {
			return nil, err
		}
	}
	s.Invalidate()

	saved, err := s.providerDAO.GetByName(entity.Name)
	if err != nil {
		return nil, err
	}
	return providerToResponse(saved), nil
}

// GetConfigForAdmin 返回当前(第一条)provider 的视图。无数据时返回 nil, nil。
func (s *OIDCService) GetConfigForAdmin() (*dto.OIDCConfigResponse, error) {
	p, err := s.providerDAO.GetFirst()
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, nil
	}
	return providerToResponse(p), nil
}

// TestConnection 尝试对指定的 issuer 做 Discovery,不触及 token 端点。
func (s *OIDCService) TestConnection(ctx context.Context, req *dto.TestOIDCConnectionRequest) error {
	if req == nil || req.Issuer == "" {
		return errors.New("issuer 必填")
	}
	if _, err := oidc.NewProvider(ctx, req.Issuer); err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	return nil
}

// --- helpers ---

func providerToResponse(p *model.OIDCProvider) *dto.OIDCConfigResponse {
	if p == nil {
		return nil
	}
	return &dto.OIDCConfigResponse{
		ID:          p.ID,
		Name:        p.Name,
		DisplayName: p.DisplayName,
		Issuer:      p.Issuer,
		ClientID:    p.ClientID,
		SecretSet:   p.ClientSecret != "",
		Scopes:      p.Scopes,
		RedirectURI: p.RedirectURI,
		Enabled:     p.IsEnabled(),
		Exclusive:   p.Enabled == 1 && p.Exclusive == 1,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func splitScopes(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		raw = defaultOIDCScopes
	}
	parts := strings.Fields(raw)
	seen := make(map[string]struct{}, len(parts))
	out := make([]string, 0, len(parts))
	for _, s := range parts {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	if len(out) == 0 {
		out = []string{oidc.ScopeOpenID}
	}
	return out
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func shortSub(sub string) string {
	if len(sub) <= 8 {
		return sub
	}
	return sub[:8]
}

// sanitizeUsername 把 preferred_username 转成符合本地 username 字段约束的串。
// 仅保留 [a-zA-Z0-9_.-],长度限制 50;不符合则返回空串交由调用方兜底。
func sanitizeUsername(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '.' || r == '-' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if len(out) > 50 {
		out = out[:50]
	}
	if len(out) < 3 {
		return ""
	}
	return out
}

