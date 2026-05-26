package service

import (
	"errors"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/dao"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
	"gorm.io/gorm"
)

type WorkspaceService struct {
	workspaceDao *dao.WorkspaceDao
	userDao      *dao.UserDAO
}

func NewWorkspaceService(helper interfaces.HelperInterface) *WorkspaceService {
	return &WorkspaceService{
		workspaceDao: dao.NewWorkspaceDao(helper),
		userDao:      dao.NewUserDAO(helper),
	}
}

func (s *WorkspaceService) ResolveWorkspace(userID uint64, requestedWorkspaceID uint64) (*model.WorkspaceMember, error) {
	if requestedWorkspaceID > 0 {
		return s.workspaceDao.GetMember(requestedWorkspaceID, userID)
	}
	members, err := s.workspaceDao.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return nil, errors.New("用户没有可用工作区")
	}
	return &members[0], nil
}

func (s *WorkspaceService) ListWorkspaces(userID uint64) (*dto.WorkspaceListResponse, error) {
	members, err := s.workspaceDao.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	list := make([]dto.WorkspaceResponse, 0, len(members))
	for _, member := range members {
		if member.Workspace.ID == 0 || !member.Workspace.IsActive() {
			continue
		}
		resp := s.workspaceToResponse(&member.Workspace)
		resp.Role = member.Role
		list = append(list, resp)
	}
	return &dto.WorkspaceListResponse{List: list}, nil
}

func (s *WorkspaceService) CreateWorkspace(userID uint64, req *dto.CreateWorkspaceRequest) (*dto.WorkspaceResponse, error) {
	if _, err := s.workspaceDao.FindBySlug(req.Slug); err == nil {
		return nil, errors.New("工作区标识 slug 已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	workspace := &model.Workspace{
		Slug:        req.Slug,
		Name:        req.Name,
		Description: req.Description,
		OwnerUserID: &userID,
		Status:      1,
	}
	if err := s.workspaceDao.Create(workspace); err != nil {
		return nil, err
	}
	if err := s.workspaceDao.CreateMember(&model.WorkspaceMember{
		WorkspaceID: workspace.ID,
		UserID:      userID,
		Role:        model.WorkspaceRoleOwner,
		Status:      1,
	}); err != nil {
		return nil, err
	}
	resp := s.workspaceToResponse(workspace)
	resp.Role = model.WorkspaceRoleOwner
	return &resp, nil
}

func (s *WorkspaceService) UpdateWorkspace(workspaceID uint64, req *dto.UpdateWorkspaceRequest) (*dto.WorkspaceResponse, error) {
	workspace, err := s.workspaceDao.FindByID(workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("工作区不存在")
		}
		return nil, err
	}
	workspace.Name = req.Name
	workspace.Description = req.Description
	if err := s.workspaceDao.Update(workspace); err != nil {
		return nil, err
	}
	resp := s.workspaceToResponse(workspace)
	return &resp, nil
}

func (s *WorkspaceService) ListMembers(workspaceID uint64) (*dto.WorkspaceMemberListResponse, error) {
	members, err := s.workspaceDao.ListMembers(workspaceID)
	if err != nil {
		return nil, err
	}
	list := make([]dto.WorkspaceMemberResponse, 0, len(members))
	for _, member := range members {
		list = append(list, s.memberToResponse(&member))
	}
	return &dto.WorkspaceMemberListResponse{List: list}, nil
}

func (s *WorkspaceService) AddMember(workspaceID uint64, req *dto.AddWorkspaceMemberRequest) (*dto.WorkspaceMemberResponse, error) {
	if _, err := s.userDao.GetByID(req.UserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	if _, err := s.workspaceDao.GetMember(workspaceID, req.UserID); err == nil {
		return nil, errors.New("用户已在工作区中")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	member := &model.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      req.UserID,
		Role:        req.Role,
		Status:      1,
	}
	if err := s.workspaceDao.CreateMember(member); err != nil {
		return nil, err
	}
	member, _ = s.workspaceDao.GetMember(workspaceID, req.UserID)
	resp := s.memberToResponse(member)
	return &resp, nil
}

func (s *WorkspaceService) UpdateMember(workspaceID, userID uint64, req *dto.UpdateWorkspaceMemberRequest) (*dto.WorkspaceMemberResponse, error) {
	member, err := s.workspaceDao.GetMember(workspaceID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("成员不存在")
		}
		return nil, err
	}
	if member.Role == model.WorkspaceRoleOwner {
		return nil, errors.New("不能修改所有者 owner 成员")
	}
	member.Role = req.Role
	if req.Status != nil {
		member.Status = *req.Status
	}
	if err := s.workspaceDao.UpdateMember(member); err != nil {
		return nil, err
	}
	resp := s.memberToResponse(member)
	return &resp, nil
}

func (s *WorkspaceService) RemoveMember(workspaceID, userID uint64) error {
	member, err := s.workspaceDao.GetMember(workspaceID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("成员不存在")
		}
		return err
	}
	if member.Role == model.WorkspaceRoleOwner {
		return errors.New("不能移除所有者 owner 成员")
	}
	return s.workspaceDao.DeleteMember(workspaceID, userID)
}

func (s *WorkspaceService) workspaceToResponse(workspace *model.Workspace) dto.WorkspaceResponse {
	return dto.WorkspaceResponse{
		ID:          workspace.ID,
		Slug:        workspace.Slug,
		Name:        workspace.Name,
		Description: workspace.Description,
		OwnerUserID: workspace.OwnerUserID,
		Status:      workspace.Status,
		CreatedAt:   workspace.CreatedAt,
		UpdatedAt:   workspace.UpdatedAt,
	}
}

func (s *WorkspaceService) memberToResponse(member *model.WorkspaceMember) dto.WorkspaceMemberResponse {
	return dto.WorkspaceMemberResponse{
		ID:          member.ID,
		WorkspaceID: member.WorkspaceID,
		UserID:      member.UserID,
		Username:    member.User.Username,
		RealName:    member.User.RealName,
		Role:        member.Role,
		Status:      member.Status,
		CreatedAt:   member.CreatedAt,
		UpdatedAt:   member.UpdatedAt,
	}
}
