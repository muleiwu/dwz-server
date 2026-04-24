-- +goose Up
ALTER TABLE click_statistics ADD COLUMN province TEXT;
ALTER TABLE click_statistics ADD COLUMN isp TEXT;
ALTER TABLE ab_test_click_statistics ADD COLUMN province TEXT;
ALTER TABLE ab_test_click_statistics ADD COLUMN isp TEXT;

-- +goose Down
ALTER TABLE click_statistics DROP COLUMN isp;
ALTER TABLE click_statistics DROP COLUMN province;
ALTER TABLE ab_test_click_statistics DROP COLUMN isp;
ALTER TABLE ab_test_click_statistics DROP COLUMN province;
