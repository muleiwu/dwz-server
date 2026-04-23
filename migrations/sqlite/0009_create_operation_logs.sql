-- +goose Up
CREATE TABLE operation_logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER,
  username TEXT,
  operation TEXT NOT NULL,
  resource TEXT,
  resource_id TEXT,
  method TEXT,
  path TEXT,
  request_body TEXT,
  response_code INTEGER NOT NULL DEFAULT 0,
  response_body TEXT,
  ip TEXT,
  user_agent TEXT,
  execute_time INTEGER NOT NULL DEFAULT 0,
  status INTEGER NOT NULL DEFAULT 1,
  error_message TEXT,
  created_at DATETIME
);
CREATE INDEX idx_operation_logs_user_id ON operation_logs(user_id);

-- +goose Down
DROP TABLE IF EXISTS operation_logs;
