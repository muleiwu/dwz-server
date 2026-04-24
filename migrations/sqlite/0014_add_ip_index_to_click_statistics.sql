-- +goose Up
CREATE INDEX IF NOT EXISTS idx_click_statistics_ip ON click_statistics (ip);
CREATE INDEX IF NOT EXISTS idx_ab_test_click_statistics_ip ON ab_test_click_statistics (ip);

-- +goose Down
DROP INDEX IF EXISTS idx_click_statistics_ip;
DROP INDEX IF EXISTS idx_ab_test_click_statistics_ip;
