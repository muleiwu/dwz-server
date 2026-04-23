-- +goose Up
CREATE TABLE ab_tests (
  id BIGSERIAL PRIMARY KEY,
  short_link_id BIGINT NOT NULL,
  name VARCHAR(255) NOT NULL,
  description VARCHAR(500),
  status VARCHAR(20) NOT NULL DEFAULT 'draft',
  traffic_split VARCHAR(20) NOT NULL DEFAULT 'equal',
  start_time TIMESTAMP WITH TIME ZONE,
  end_time TIMESTAMP WITH TIME ZONE,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP WITH TIME ZONE,
  updated_at TIMESTAMP WITH TIME ZONE,
  deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_ab_tests_short_link_id ON ab_tests(short_link_id);
CREATE INDEX idx_ab_tests_deleted_at ON ab_tests(deleted_at);

-- +goose Down
DROP TABLE IF EXISTS ab_tests;
