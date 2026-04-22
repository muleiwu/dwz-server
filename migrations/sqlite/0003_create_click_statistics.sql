-- +goose Up
CREATE TABLE click_statistics (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  short_link_id INTEGER NOT NULL,
  ip TEXT,
  user_agent TEXT,
  referer TEXT,
  query_params TEXT,
  country TEXT,
  city TEXT,
  click_date DATETIME,
  created_at DATETIME
);
CREATE INDEX idx_short_link_date ON click_statistics(short_link_id, click_date);

-- +goose Down
DROP TABLE IF EXISTS click_statistics;
