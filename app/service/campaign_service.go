package service

import (
	"errors"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/dao"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
	"gorm.io/gorm"
)

type CampaignService struct {
	campaignDao *dao.CampaignDao
}

func NewCampaignService(helper interfaces.HelperInterface) *CampaignService {
	return &CampaignService{
		campaignDao: dao.NewCampaignDao(helper),
	}
}

func (s *CampaignService) Create(workspaceID, userID uint64, req *dto.CampaignRequest) (*dto.CampaignResponse, error) {
	status := req.Status
	if status == "" {
		status = model.CampaignStatusActive
	}
	campaign := &model.Campaign{
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Description: req.Description,
		StartAt:     req.StartAt,
		EndAt:       req.EndAt,
		Status:      status,
		CreatedBy:   &userID,
	}
	if err := s.campaignDao.Create(campaign); err != nil {
		return nil, err
	}
	resp := s.modelToResponse(campaign)
	return &resp, nil
}

func (s *CampaignService) Update(id, workspaceID uint64, req *dto.CampaignRequest) (*dto.CampaignResponse, error) {
	campaign, err := s.campaignDao.FindByID(id, workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("活动不存在")
		}
		return nil, err
	}
	campaign.Name = req.Name
	campaign.Description = req.Description
	campaign.StartAt = req.StartAt
	campaign.EndAt = req.EndAt
	if req.Status != "" {
		campaign.Status = req.Status
	}
	if err := s.campaignDao.Update(campaign); err != nil {
		return nil, err
	}
	resp := s.modelToResponse(campaign)
	return &resp, nil
}

func (s *CampaignService) Delete(id, workspaceID uint64) error {
	return s.campaignDao.Delete(id, workspaceID)
}

func (s *CampaignService) Get(id, workspaceID uint64) (*dto.CampaignResponse, error) {
	campaign, err := s.campaignDao.FindByID(id, workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("活动不存在")
		}
		return nil, err
	}
	resp := s.modelToResponse(campaign)
	return &resp, nil
}

func (s *CampaignService) List(workspaceID uint64, req *dto.CampaignListRequest) (*dto.CampaignListResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}
	offset := (req.Page - 1) * req.PageSize
	campaigns, total, err := s.campaignDao.List(workspaceID, offset, req.PageSize, req.Keyword, req.Status)
	if err != nil {
		return nil, err
	}
	list := make([]dto.CampaignResponse, 0, len(campaigns))
	for _, campaign := range campaigns {
		list = append(list, s.modelToResponse(&campaign))
	}
	return &dto.CampaignListResponse{
		List:  list,
		Total: total,
		Page:  req.Page,
		Size:  req.PageSize,
	}, nil
}

func (s *CampaignService) Reports(workspaceID uint64, req *dto.CampaignReportRequest) (*dto.CampaignReportResponse, error) {
	var startAt, endAt = req.StartDate, req.EndDate
	if endAt != nil {
		end := endAt.AddDate(0, 0, 1)
		endAt = &end
	}
	items, err := s.campaignDao.Reports(workspaceID, req.CampaignID, startAt, endAt)
	if err != nil {
		return nil, err
	}
	return &dto.CampaignReportResponse{List: items}, nil
}

func (s *CampaignService) modelToResponse(campaign *model.Campaign) dto.CampaignResponse {
	return dto.CampaignResponse{
		ID:          campaign.ID,
		WorkspaceID: campaign.WorkspaceID,
		Name:        campaign.Name,
		Description: campaign.Description,
		StartAt:     campaign.StartAt,
		EndAt:       campaign.EndAt,
		Status:      campaign.Status,
		CreatedBy:   campaign.CreatedBy,
		CreatedAt:   campaign.CreatedAt,
		UpdatedAt:   campaign.UpdatedAt,
	}
}
