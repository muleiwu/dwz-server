-- +goose Up
CREATE INDEX idx_click_statistics_workspace_click_date ON click_statistics(workspace_id, click_date);
CREATE INDEX idx_click_statistics_workspace_country_click_date ON click_statistics(workspace_id, country, click_date);
CREATE INDEX idx_click_statistics_workspace_country_province_click_date ON click_statistics(workspace_id, country, province, click_date);
CREATE INDEX idx_click_statistics_workspace_short_link_click_date ON click_statistics(workspace_id, short_link_id, click_date);
CREATE INDEX idx_click_statistics_workspace_campaign_click_date ON click_statistics(workspace_id, campaign_id, click_date);
CREATE INDEX idx_click_statistics_workspace_route_click_date ON click_statistics(workspace_id, route_id, click_date);

-- +goose Down
DROP INDEX IF EXISTS idx_click_statistics_workspace_route_click_date;
DROP INDEX IF EXISTS idx_click_statistics_workspace_campaign_click_date;
DROP INDEX IF EXISTS idx_click_statistics_workspace_short_link_click_date;
DROP INDEX IF EXISTS idx_click_statistics_workspace_country_province_click_date;
DROP INDEX IF EXISTS idx_click_statistics_workspace_country_click_date;
DROP INDEX IF EXISTS idx_click_statistics_workspace_click_date;
