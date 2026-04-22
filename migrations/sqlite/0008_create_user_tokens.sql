-- +goose Up
CREATE TABLE user_tokens (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  token_name TEXT NOT NULL,
  token_type TEXT NOT NULL DEFAULT 'bearer',
  token TEXT,
  app_id TEXT,
  app_secret TEXT,
  last_used_at DATETIME,
  expire_at DATETIME,
  status INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME
);
CREATE UNIQUE INDEX uk_user_tokens_token ON user_tokens(token);
CREATE UNIQUE INDEX uk_user_tokens_app_id ON user_tokens(app_id);
CREATE INDEX idx_user_tokens_user_id ON user_tokens(user_id);
CREATE INDEX idx_user_tokens_deleted_at ON user_tokens(deleted_at);

-- +goose Down
DROP TABLE IF EXISTS user_tokens;
