-- +goose Up
CREATE TABLE ab_test_click_statistics (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  ab_test_id INTEGER NOT NULL,
  variant_id INTEGER NOT NULL,
  short_link_id INTEGER NOT NULL,
  ip TEXT,
  user_agent TEXT,
  referer TEXT,
  query_params TEXT,
  country TEXT,
  city TEXT,
  session_id TEXT,
  click_date DATETIME,
  created_at DATETIME
);
CREATE INDEX idx_ab_test_click ON ab_test_click_statistics(ab_test_id, click_date);
CREATE INDEX idx_variant_click ON ab_test_click_statistics(variant_id, click_date);
CREATE INDEX idx_ab_test_click_statistics_short_link_id ON ab_test_click_statistics(short_link_id);
CREATE INDEX idx_ab_test_click_statistics_session_id ON ab_test_click_statistics(session_id);

-- +goose Down
DROP TABLE IF EXISTS ab_test_click_statistics;
