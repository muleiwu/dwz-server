package dao

import (
	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
)

type WorkspaceDao struct {
	helper interfaces.HelperInterface
}

func NewWorkspaceDao(helper interfaces.HelperInterface) *WorkspaceDao {
	return &WorkspaceDao{helper: helper}
}

func (d *WorkspaceDao) Create(workspace *model.Workspace) error {
	return d.helper.GetDatabase().Create(workspace).Error
}

func (d *WorkspaceDao) Update(workspace *model.Workspace) error {
	return d.helper.GetDatabase().Save(workspace).Error
}

func (d *WorkspaceDao) FindByID(id uint64) (*model.Workspace, error) {
	var workspace model.Workspace
	err := d.helper.GetDatabase().Where("id = ? AND deleted_at IS NULL", id).First(&workspace).Error
	return &workspace, err
}

func (d *WorkspaceDao) FindBySlug(slug string) (*model.Workspace, error) {
	var workspace model.Workspace
	err := d.helper.GetDatabase().Where("slug = ? AND deleted_at IS NULL", slug).First(&workspace).Error
	return &workspace, err
}

func (d *WorkspaceDao) ListByUser(userID uint64) ([]model.WorkspaceMember, error) {
	var members []model.WorkspaceMember
	err := d.helper.GetDatabase().
		Preload("Workspace").
		Where("user_id = ? AND status = ? AND deleted_at IS NULL", userID, 1).
		Order("workspace_id ASC").
		Find(&members).Error
	return members, err
}

func (d *WorkspaceDao) GetMember(workspaceID, userID uint64) (*model.WorkspaceMember, error) {
	var member model.WorkspaceMember
	err := d.helper.GetDatabase().
		Preload("User").
		Where("workspace_id = ? AND user_id = ? AND status = ? AND deleted_at IS NULL", workspaceID, userID, 1).
		First(&member).Error
	return &member, err
}

func (d *WorkspaceDao) ListMembers(workspaceID uint64) ([]model.WorkspaceMember, error) {
	var members []model.WorkspaceMember
	err := d.helper.GetDatabase().
		Preload("User").
		Where("workspace_id = ? AND deleted_at IS NULL", workspaceID).
		Order("created_at ASC").
		Find(&members).Error
	return members, err
}

func (d *WorkspaceDao) CreateMember(member *model.WorkspaceMember) error {
	return d.helper.GetDatabase().Create(member).Error
}

func (d *WorkspaceDao) UpdateMember(member *model.WorkspaceMember) error {
	return d.helper.GetDatabase().Save(member).Error
}

func (d *WorkspaceDao) DeleteMember(workspaceID, userID uint64) error {
	return d.helper.GetDatabase().
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Delete(&model.WorkspaceMember{}).Error
}

func (d *WorkspaceDao) EnsureDefaultWorkspaceForUser(userID uint64, role string) error {
	db := d.helper.GetDatabase()
	return db.FirstOrCreate(&model.WorkspaceMember{}, model.WorkspaceMember{
		WorkspaceID: 1,
		UserID:      userID,
	}).Updates(map[string]any{
		"role":   role,
		"status": 1,
	}).Error
}
