-- +goose Up
CREATE TABLE ab_test_variants (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  ab_test_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  target_url TEXT NOT NULL,
  weight INTEGER NOT NULL DEFAULT 50,
  is_control INTEGER NOT NULL DEFAULT 0,
  description TEXT,
  is_active INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME
);
CREATE INDEX idx_ab_test_variants_ab_test_id ON ab_test_variants(ab_test_id);
CREATE INDEX idx_ab_test_variants_deleted_at ON ab_test_variants(deleted_at);

-- +goose Down
DROP TABLE IF EXISTS ab_test_variants;
