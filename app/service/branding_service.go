package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/dao"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
	"gorm.io/gorm"
)

const (
	defaultBrandingLogoMaxBytes = 2 * 1024 * 1024
	defaultBrandingUploadDir    = "data/uploads/branding"
	maxBrandingNameRunes        = 80
	systemBrandingID            = 1
	brandingUploadURLPrefix     = "/uploads/branding"
)

var allowedBrandingLogoMIMEs = map[string]string{
	"image/gif":  ".gif",
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type BrandingService struct {
	helper interfaces.HelperInterface
	dao    *dao.BrandingDao
}

func NewBrandingService(helper interfaces.HelperInterface) *BrandingService {
	return &BrandingService{
		helper: helper,
		dao:    dao.NewBrandingDao(helper),
	}
}

func (s *BrandingService) GetPublicBranding(host string) (*dto.BrandingResponse, error) {
	branding, err := s.resolveSystemOrDefault()
	if err != nil {
		return nil, err
	}
	host = normalizeBrandingHost(host)
	if host == "" {
		return branding, nil
	}
	if domain, err := NewDomainService(s.helper).GetDomainByName(host); err == nil && domain.SiteName != "" {
		branding.BrandName = domain.SiteName
	}
	return branding, nil
}

func (s *BrandingService) GetSystemBranding() (*dto.SystemBrandingResponse, error) {
	branding, err := s.dao.FindSystem()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &dto.SystemBrandingResponse{
				ID:               systemBrandingID,
				LogoURL:          "",
				BrandName:        "",
				CopyrightEnabled: true,
				CopyrightText:    s.defaultCopyrightText(),
				CopyrightLink:    "",
			}, nil
		}
		return nil, err
	}
	return systemBrandingToResponse(branding), nil
}

func (s *BrandingService) SaveSystemBranding(req *dto.SystemBrandingRequest) (*dto.SystemBrandingResponse, error) {
	logoURL, err := NormalizeBrandingLogoURL(req.LogoURL)
	if err != nil {
		return nil, err
	}
	brandName, err := NormalizeBrandingName(req.BrandName)
	if err != nil {
		return nil, err
	}
	branding := &model.SystemBranding{
		ID:               systemBrandingID,
		LogoURL:          logoURL,
		BrandName:        brandName,
		CopyrightEnabled: true,
		CopyrightText:    s.defaultCopyrightText(),
		CopyrightLink:    "",
	}
	if err := s.dao.UpsertSystemBase(branding); err != nil {
		return nil, err
	}
	return s.GetSystemBranding()
}

func (s *BrandingService) StoreLogo(fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader == nil {
		return "", errors.New("请选择要上传的Logo文件")
	}
	maxBytes := int64(s.helper.GetConfig().GetInt("branding.upload.max_logo_bytes", 0))
	if maxBytes <= 0 {
		maxBytes = int64(s.helper.GetConfig().GetInt("ee.upload.max_logo_bytes", defaultBrandingLogoMaxBytes))
	}
	if fileHeader.Size <= 0 {
		return "", errors.New("Logo文件不能为空")
	}
	if fileHeader.Size > maxBytes {
		return "", fmt.Errorf("Logo文件不能超过%dMB", maxBytes/1024/1024)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	ext, err := DetectBrandingLogoExtension(file)
	if err != nil {
		return "", err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	filename := hex.EncodeToString(hash.Sum(nil)) + ext
	uploadDir := s.UploadDir()
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}
	dst := filepath.Join(uploadDir, filename)
	if _, err := os.Stat(dst); errors.Is(err, os.ErrNotExist) {
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return "", err
		}
		out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(out, file); err != nil {
			_ = out.Close()
			_ = os.Remove(dst)
			return "", err
		}
		if err := out.Close(); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}
	return brandingUploadURLPrefix + "/" + filename, nil
}

func DetectBrandingLogoExtension(file multipart.File) (string, error) {
	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	contentType := http.DetectContentType(header[:n])
	ext, ok := allowedBrandingLogoMIMEs[contentType]
	if !ok {
		return "", errors.New("Logo仅支持png、jpeg、webp或gif格式")
	}
	return ext, nil
}

func NormalizeBrandingLogoURL(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}
	if strings.HasPrefix(value, "/") {
		if strings.HasPrefix(value, "//") || strings.Contains(value, "\\") {
			return "", errors.New("Logo地址格式无效")
		}
		return value, nil
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" {
		return "", errors.New("Logo地址格式无效")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("Logo地址仅支持http、https或站内绝对路径")
	}
	return value, nil
}

func NormalizeBrandingName(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if len([]rune(value)) > maxBrandingNameRunes {
		return "", fmt.Errorf("品牌文字不能超过%d个字符", maxBrandingNameRunes)
	}
	return value, nil
}

func (s *BrandingService) UploadDir() string {
	uploadDir := s.helper.GetConfig().GetString("branding.upload.dir", "")
	if uploadDir == "" {
		uploadDir = s.helper.GetConfig().GetString("ee.upload.dir", defaultBrandingUploadDir)
	}
	if uploadDir == "" {
		return defaultBrandingUploadDir
	}
	return uploadDir
}

func (s *BrandingService) resolveSystemOrDefault() (*dto.BrandingResponse, error) {
	branding := dto.BrandingResponse{
		LogoURL:          "",
		BrandName:        s.helper.GetEnv().GetString("website.name", "短网址服务"),
		CopyrightEnabled: true,
		CopyrightText:    s.defaultCopyrightText(),
		CopyrightLink:    "",
		Source:           "default",
	}
	systemBranding, err := s.dao.FindSystem()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &branding, nil
		}
		return nil, err
	}
	branding.LogoURL = systemBranding.LogoURL
	if systemBranding.BrandName != "" {
		branding.BrandName = systemBranding.BrandName
	}
	branding.Source = "system"
	return &branding, nil
}

func (s *BrandingService) defaultCopyrightText() string {
	return strings.TrimSpace(s.helper.GetEnv().GetString("website.copyright", ""))
}

func systemBrandingToResponse(branding *model.SystemBranding) *dto.SystemBrandingResponse {
	return &dto.SystemBrandingResponse{
		ID:               branding.ID,
		LogoURL:          branding.LogoURL,
		BrandName:        branding.BrandName,
		CopyrightEnabled: branding.CopyrightEnabled,
		CopyrightText:    branding.CopyrightText,
		CopyrightLink:    branding.CopyrightLink,
		CreatedAt:        branding.CreatedAt,
		UpdatedAt:        branding.UpdatedAt,
	}
}

func normalizeBrandingHost(host string) string {
	value := strings.TrimSpace(strings.ToLower(host))
	if h, _, ok := strings.Cut(value, ":"); ok && h != "" {
		return h
	}
	return value
}
