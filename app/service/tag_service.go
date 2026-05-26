package service

import (
	"errors"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/dao"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
	"gorm.io/gorm"
)

type TagService struct {
	tagDao *dao.TagDao
}

func NewTagService(helper interfaces.HelperInterface) *TagService {
	return &TagService{
		tagDao: dao.NewTagDao(helper),
	}
}

func (s *TagService) Create(workspaceID uint64, req *dto.TagRequest) (*dto.TagResponse, error) {
	tag := &model.Tag{
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Color:       req.Color,
	}
	if err := s.tagDao.Create(tag); err != nil {
		return nil, err
	}
	resp := s.modelToResponse(tag)
	return &resp, nil
}

func (s *TagService) Update(id, workspaceID uint64, req *dto.TagRequest) (*dto.TagResponse, error) {
	tag, err := s.tagDao.FindByID(id, workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("标签不存在")
		}
		return nil, err
	}
	tag.Name = req.Name
	tag.Color = req.Color
	if err := s.tagDao.Update(tag); err != nil {
		return nil, err
	}
	resp := s.modelToResponse(tag)
	return &resp, nil
}

func (s *TagService) Delete(id, workspaceID uint64) error {
	return s.tagDao.Delete(id, workspaceID)
}

func (s *TagService) Get(id, workspaceID uint64) (*dto.TagResponse, error) {
	tag, err := s.tagDao.FindByID(id, workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("标签不存在")
		}
		return nil, err
	}
	resp := s.modelToResponse(tag)
	return &resp, nil
}

func (s *TagService) List(workspaceID uint64, req *dto.TagListRequest) (*dto.TagListResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}
	offset := (req.Page - 1) * req.PageSize
	tags, total, err := s.tagDao.List(workspaceID, offset, req.PageSize, req.Keyword)
	if err != nil {
		return nil, err
	}
	list := make([]dto.TagResponse, 0, len(tags))
	for _, tag := range tags {
		list = append(list, s.modelToResponse(&tag))
	}
	return &dto.TagListResponse{
		List:  list,
		Total: total,
		Page:  req.Page,
		Size:  req.PageSize,
	}, nil
}

func (s *TagService) modelToResponse(tag *model.Tag) dto.TagResponse {
	return dto.TagResponse{
		ID:          tag.ID,
		WorkspaceID: tag.WorkspaceID,
		Name:        tag.Name,
		Color:       tag.Color,
		CreatedAt:   tag.CreatedAt,
		UpdatedAt:   tag.UpdatedAt,
	}
}
