-- +goose Up
CREATE TABLE click_statistics (
  id BIGSERIAL PRIMARY KEY,
  short_link_id BIGINT NOT NULL,
  ip VARCHAR(45),
  user_agent VARCHAR(1024),
  referer VARCHAR(2048),
  query_params VARCHAR(2048),
  country VARCHAR(100),
  city VARCHAR(100),
  click_date TIMESTAMP WITH TIME ZONE,
  created_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_short_link_date ON click_statistics(short_link_id, click_date);

-- +goose Down
DROP TABLE IF EXISTS click_statistics;
