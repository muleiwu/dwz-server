package dao

import (
	"time"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
)

type CampaignDao struct {
	helper interfaces.HelperInterface
}

func NewCampaignDao(helper interfaces.HelperInterface) *CampaignDao {
	return &CampaignDao{helper: helper}
}

func (d *CampaignDao) Create(campaign *model.Campaign) error {
	return d.helper.GetDatabase().Create(campaign).Error
}

func (d *CampaignDao) Update(campaign *model.Campaign) error {
	return d.helper.GetDatabase().Save(campaign).Error
}

func (d *CampaignDao) Delete(id, workspaceID uint64) error {
	return d.helper.GetDatabase().Where("id = ? AND workspace_id = ?", id, workspaceID).Delete(&model.Campaign{}).Error
}

func (d *CampaignDao) FindByID(id, workspaceID uint64) (*model.Campaign, error) {
	var campaign model.Campaign
	err := d.helper.GetDatabase().
		Where("id = ? AND workspace_id = ? AND deleted_at IS NULL", id, workspaceID).
		First(&campaign).Error
	return &campaign, err
}

func (d *CampaignDao) List(workspaceID uint64, offset, limit int, keyword, status string) ([]model.Campaign, int64, error) {
	var campaigns []model.Campaign
	var total int64

	query := d.helper.GetDatabase().Model(&model.Campaign{}).
		Where("workspace_id = ? AND deleted_at IS NULL", workspaceID)
	if keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&campaigns).Error
	return campaigns, total, err
}

func (d *CampaignDao) Reports(workspaceID uint64, campaignID uint64, startDate, endDate *time.Time) ([]dto.CampaignReportItem, error) {
	var results []dto.CampaignReportItem
	db := d.helper.GetDatabase()

	query := db.Table("campaigns c").
		Select("c.id AS campaign_id, c.name AS campaign_name, COUNT(DISTINCT sl.id) AS short_link_count, COUNT(cs.id) AS click_count, COUNT(DISTINCT cs.ip) AS unique_ips").
		Joins("LEFT JOIN short_links sl ON sl.campaign_id = c.id AND sl.deleted_at IS NULL").
		Joins("LEFT JOIN click_statistics cs ON cs.campaign_id = c.id").
		Where("c.workspace_id = ? AND c.deleted_at IS NULL", workspaceID).
		Group("c.id, c.name").
		Order("click_count DESC")
	if campaignID > 0 {
		query = query.Where("c.id = ?", campaignID)
	}
	if startDate != nil {
		query = query.Where("(cs.id IS NULL OR cs.click_date >= ?)", *startDate)
	}
	if endDate != nil {
		query = query.Where("(cs.id IS NULL OR cs.click_date <= ?)", *endDate)
	}
	return results, query.Find(&results).Error
}
