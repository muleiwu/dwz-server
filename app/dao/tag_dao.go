package dao

import (
	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
	"gorm.io/gorm"
)

type TagDao struct {
	helper interfaces.HelperInterface
}

func NewTagDao(helper interfaces.HelperInterface) *TagDao {
	return &TagDao{helper: helper}
}

func (d *TagDao) Create(tag *model.Tag) error {
	return d.helper.GetDatabase().Create(tag).Error
}

func (d *TagDao) Update(tag *model.Tag) error {
	return d.helper.GetDatabase().Save(tag).Error
}

func (d *TagDao) Delete(id, workspaceID uint64) error {
	return d.helper.GetDatabase().Where("id = ? AND workspace_id = ?", id, workspaceID).Delete(&model.Tag{}).Error
}

func (d *TagDao) FindByID(id, workspaceID uint64) (*model.Tag, error) {
	var tag model.Tag
	err := d.helper.GetDatabase().
		Where("id = ? AND workspace_id = ? AND deleted_at IS NULL", id, workspaceID).
		First(&tag).Error
	return &tag, err
}

func (d *TagDao) List(workspaceID uint64, offset, limit int, keyword string) ([]model.Tag, int64, error) {
	var tags []model.Tag
	var total int64
	query := d.helper.GetDatabase().Model(&model.Tag{}).
		Where("workspace_id = ? AND deleted_at IS NULL", workspaceID)
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&tags).Error
	return tags, total, err
}

func (d *TagDao) FindMany(ids []uint64, workspaceID uint64) ([]model.Tag, error) {
	var tags []model.Tag
	if len(ids) == 0 {
		return tags, nil
	}
	err := d.helper.GetDatabase().
		Where("id IN ? AND workspace_id = ? AND deleted_at IS NULL", ids, workspaceID).
		Find(&tags).Error
	return tags, err
}

func (d *TagDao) ReplaceShortLinkTags(shortLinkID uint64, tagIDs []uint64) error {
	db := d.helper.GetDatabase()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("short_link_id = ?", shortLinkID).Delete(&model.ShortLinkTag{}).Error; err != nil {
			return err
		}
		for _, tagID := range tagIDs {
			if err := tx.Create(&model.ShortLinkTag{
				ShortLinkID: shortLinkID,
				TagID:       tagID,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (d *TagDao) GetTagsByShortLinkID(shortLinkID uint64) ([]model.Tag, error) {
	var tags []model.Tag
	err := d.helper.GetDatabase().
		Joins("JOIN short_link_tags slt ON slt.tag_id = tags.id").
		Where("slt.short_link_id = ? AND tags.deleted_at IS NULL", shortLinkID).
		Order("tags.name ASC").
		Find(&tags).Error
	return tags, err
}
