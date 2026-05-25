-- +goose Up
CREATE TABLE ab_test_feedbacks (
  id BIGSERIAL PRIMARY KEY,
  workspace_id BIGINT NOT NULL DEFAULT 1,
  ab_test_id BIGINT NOT NULL,
  variant_id BIGINT NOT NULL,
  short_link_id BIGINT NOT NULL,
  session_id VARCHAR(128) NOT NULL,
  event_id VARCHAR(128) NOT NULL,
  value NUMERIC(18,4),
  currency VARCHAR(16),
  metadata TEXT,
  ip VARCHAR(45),
  user_agent VARCHAR(1024),
  referer VARCHAR(2048),
  occurred_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP
);

CREATE UNIQUE INDEX uk_ab_test_feedback_event ON ab_test_feedbacks(ab_test_id, event_id);
CREATE INDEX idx_ab_test_feedbacks_workspace_id ON ab_test_feedbacks(workspace_id);
CREATE INDEX idx_ab_test_feedbacks_variant_id ON ab_test_feedbacks(variant_id);
CREATE INDEX idx_ab_test_feedbacks_short_link_id ON ab_test_feedbacks(short_link_id);
CREATE INDEX idx_ab_test_feedbacks_session_id ON ab_test_feedbacks(session_id);
CREATE INDEX idx_ab_test_feedbacks_occurred_at ON ab_test_feedbacks(occurred_at);

-- +goose Down
DROP TABLE IF EXISTS ab_test_feedbacks;
