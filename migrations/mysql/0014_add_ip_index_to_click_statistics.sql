-- +goose Up
CREATE INDEX `idx_click_statistics_ip` ON `click_statistics` (`ip`);
CREATE INDEX `idx_ab_test_click_statistics_ip` ON `ab_test_click_statistics` (`ip`);

-- +goose Down
DROP INDEX `idx_click_statistics_ip` ON `click_statistics`;
DROP INDEX `idx_ab_test_click_statistics_ip` ON `ab_test_click_statistics`;
