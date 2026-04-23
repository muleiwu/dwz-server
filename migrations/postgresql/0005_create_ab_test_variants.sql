-- +goose Up
CREATE TABLE ab_test_variants (
  id BIGSERIAL PRIMARY KEY,
  ab_test_id BIGINT NOT NULL,
  name VARCHAR(100) NOT NULL,
  target_url VARCHAR(2000) NOT NULL,
  weight INTEGER NOT NULL DEFAULT 50,
  is_control BOOLEAN NOT NULL DEFAULT FALSE,
  description VARCHAR(500),
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP WITH TIME ZONE,
  updated_at TIMESTAMP WITH TIME ZONE,
  deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_ab_test_variants_ab_test_id ON ab_test_variants(ab_test_id);
CREATE INDEX idx_ab_test_variants_deleted_at ON ab_test_variants(deleted_at);

-- +goose Down
DROP TABLE IF EXISTS ab_test_variants;
