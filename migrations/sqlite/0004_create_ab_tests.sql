-- +goose Up
CREATE TABLE ab_tests (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  short_link_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  description TEXT,
  status TEXT NOT NULL DEFAULT 'draft',
  traffic_split TEXT NOT NULL DEFAULT 'equal',
  start_time DATETIME,
  end_time DATETIME,
  is_active INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME
);
CREATE INDEX idx_ab_tests_short_link_id ON ab_tests(short_link_id);
CREATE INDEX idx_ab_tests_deleted_at ON ab_tests(deleted_at);

-- +goose Down
DROP TABLE IF EXISTS ab_tests;
