-- +goose Up
CREATE TABLE ab_test_feedbacks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  workspace_id INTEGER NOT NULL DEFAULT 1,
  ab_test_id INTEGER NOT NULL,
  variant_id INTEGER NOT NULL,
  short_link_id INTEGER NOT NULL,
  session_id TEXT NOT NULL,
  event_id TEXT NOT NULL,
  value REAL,
  currency TEXT,
  metadata TEXT,
  ip TEXT,
  user_agent TEXT,
  referer TEXT,
  occurred_at DATETIME NOT NULL,
  created_at DATETIME
);

CREATE UNIQUE INDEX uk_ab_test_feedback_event ON ab_test_feedbacks(ab_test_id, event_id);
CREATE INDEX idx_ab_test_feedbacks_workspace_id ON ab_test_feedbacks(workspace_id);
CREATE INDEX idx_ab_test_feedbacks_variant_id ON ab_test_feedbacks(variant_id);
CREATE INDEX idx_ab_test_feedbacks_short_link_id ON ab_test_feedbacks(short_link_id);
CREATE INDEX idx_ab_test_feedbacks_session_id ON ab_test_feedbacks(session_id);
CREATE INDEX idx_ab_test_feedbacks_occurred_at ON ab_test_feedbacks(occurred_at);

-- +goose Down
DROP TABLE IF EXISTS ab_test_feedbacks;
