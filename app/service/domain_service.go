package service

import (
	"errors"

	"cnb.cool/mliev/open/dwz-server/app/dao"
	"cnb.cool/mliev/open/dwz-server/app/dto"
	"cnb.cool/mliev/open/dwz-server/app/model"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"cnb.cool/mliev/open/dwz-server/pkg/domain_validate"
	"gorm.io/gorm"
)

type DomainService struct {
	domainDao *dao.DomainDao
}

func NewDomainService(helper interfaces.HelperInterface) *DomainService {
	return &DomainService{
		domainDao: dao.NewDomainDao(helper),
	}
}

// CreateDomain 创建域名配置
func (s *DomainService) CreateDomain(req *dto.DomainRequest) (*dto.DomainResponse, error) {
	// 验证域名格式
	if err := domain_validate.ValidateDomain(req.Domain); err != nil {
		return nil, errors.New("无效的域名格式")
	}

	// 检查域名是否已存在
	exists, err := s.domainDao.ExistsByDomain(req.Domain)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("域名已存在")
	}

	// 创建域名记录
	// 注意：直接使用请求中的值，不做默认值回退
	// 默认值由数据库迁移时设置，确保老数据兼容
	domain := &model.Domain{
		Domain:             req.Domain,
		Protocol:           req.Protocol,
		SiteName:           req.SiteName,
		ICPNumber:          req.ICPNumber,
		PoliceNumber:       req.PoliceNumber,
		Description:        req.Description,
		IsActive:           req.IsActive,
		PassQueryParams:    req.PassQueryParams,
		RandomSuffixLength: req.RandomSuffixLength,
		EnableChecksum:     req.EnableChecksum,
	}

	if err := s.domainDao.Create(domain); err != nil {
		return nil, err
	}

	return s.modelToResponse(domain), nil
}

// GetDomainList 获取域名列表
func (s *DomainService) GetDomainList() (*dto.DomainListResponse, error) {
	domains, err := s.domainDao.List()
	if err != nil {
		return nil, err
	}

	responses := make([]dto.DomainResponse, 0, len(domains))
	for _, domain := range domains {
		responses = append(responses, *s.modelToResponse(&domain))
	}

	return &dto.DomainListResponse{
		List: responses,
	}, nil
}

func (s *DomainService) UpdateStatusDomain(id uint64, req *dto.UpdateStatusDomainRequest) (bool, error) {
	where := map[string]any{
		"is_active": req.IsActive,
	}

	if err := s.domainDao.IdToUpdate(id, where); err != nil {
		return false, err
	}

	return true, nil

}

// UpdateDomain 更新域名
func (s *DomainService) UpdateDomain(id uint64, req *dto.DomainRequest) (*dto.DomainResponse, error) {

	if err := domain_validate.ValidateDomain(req.Domain); err != nil {
		return nil, errors.New("无效的域名格式")
	}

	domain, err := s.domainDao.FindByDomain(req.Domain)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("域名不存在")
		}
		return nil, err
	}

	// 如果修改了域名，需要检查新域名是否已存在
	if domain.Domain != req.Domain {
		exists, err := s.domainDao.ExistsByDomain(req.Domain)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("新域名已存在")
		}
		domain.Domain = req.Domain
	}

	domain.IsActive = req.IsActive
	domain.PassQueryParams = req.PassQueryParams
	domain.Description = req.Description
	domain.PoliceNumber = req.PoliceNumber
	domain.ICPNumber = req.ICPNumber
	domain.Protocol = req.Protocol
	domain.SiteName = req.SiteName
	domain.RandomSuffixLength = req.RandomSuffixLength
	domain.EnableChecksum = req.EnableChecksum

	if err := s.domainDao.Update(domain); err != nil {
		return nil, err
	}

	return s.modelToResponse(domain), nil
}

// DeleteDomain 删除域名
func (s *DomainService) DeleteDomain(id uint64) error {
	// 可以添加检查是否有短网址使用此域名的逻辑

	return s.domainDao.Delete(id)
}

// GetActiveDomains 获取活跃域名列表
func (s *DomainService) GetActiveDomains() ([]dto.DomainResponse, error) {
	domains, err := s.domainDao.GetActiveDomains()
	if err != nil {
		return nil, err
	}

	responses := make([]dto.DomainResponse, 0, len(domains))
	for _, domain := range domains {
		responses = append(responses, *s.modelToResponse(&domain))
	}

	return responses, nil
}

// GetDomainByName 查询指定的domain
func (s *DomainService) GetDomainByName(domainName string) (*model.Domain, error) {
	domain, err := s.domainDao.FindByDomain(domainName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("域名不存在")
		}
		return nil, err
	}

	return domain, nil
}

// GetDomainByID 根据ID查询域名
func (s *DomainService) GetDomainByID(id uint64) (*model.Domain, error) {
	domain, err := s.domainDao.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("域名不存在")
		}
		return nil, err
	}

	return domain, nil
}

// 私有方法

// modelToResponse 将模型转换为响应格式
func (s *DomainService) modelToResponse(domain *model.Domain) *dto.DomainResponse {
	// 处理指针类型，提供默认值
	randomSuffixLength := 2
	if domain.RandomSuffixLength != nil {
		randomSuffixLength = *domain.RandomSuffixLength
	}
	enableChecksum := true
	if domain.EnableChecksum != nil {
		enableChecksum = *domain.EnableChecksum
	}

	return &dto.DomainResponse{
		ID:                 domain.ID,
		Domain:             domain.Domain,
		Protocol:           domain.Protocol,
		SiteName:           domain.SiteName,
		ICPNumber:          domain.ICPNumber,
		PoliceNumber:       domain.PoliceNumber,
		IsActive:           domain.IsActive,
		PassQueryParams:    domain.PassQueryParams,
		RandomSuffixLength: randomSuffixLength,
		EnableChecksum:     enableChecksum,
		Description:        domain.Description,
		CreatedAt:          domain.CreatedAt,
		UpdatedAt:          domain.UpdatedAt,
	}
}
