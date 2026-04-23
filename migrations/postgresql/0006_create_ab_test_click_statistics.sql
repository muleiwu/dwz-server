-- +goose Up
CREATE TABLE ab_test_click_statistics (
  id BIGSERIAL PRIMARY KEY,
  ab_test_id BIGINT NOT NULL,
  variant_id BIGINT NOT NULL,
  short_link_id BIGINT NOT NULL,
  ip VARCHAR(45),
  user_agent VARCHAR(1024),
  referer VARCHAR(2048),
  query_params VARCHAR(2048),
  country VARCHAR(100),
  city VARCHAR(100),
  session_id VARCHAR(128),
  click_date TIMESTAMP WITH TIME ZONE,
  created_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_ab_test_click ON ab_test_click_statistics(ab_test_id, click_date);
CREATE INDEX idx_variant_click ON ab_test_click_statistics(variant_id, click_date);
CREATE INDEX idx_ab_test_click_statistics_short_link_id ON ab_test_click_statistics(short_link_id);
CREATE INDEX idx_ab_test_click_statistics_session_id ON ab_test_click_statistics(session_id);

-- +goose Down
DROP TABLE IF EXISTS ab_test_click_statistics;
