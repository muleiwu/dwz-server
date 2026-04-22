-- +goose Up
CREATE TABLE user_tokens (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  token_name VARCHAR(100) NOT NULL,
  token_type VARCHAR(20) NOT NULL DEFAULT 'bearer',
  token VARCHAR(190),
  app_id VARCHAR(64),
  app_secret VARCHAR(256),
  last_used_at TIMESTAMP WITH TIME ZONE,
  expire_at TIMESTAMP WITH TIME ZONE,
  status SMALLINT NOT NULL DEFAULT 1,
  created_at TIMESTAMP WITH TIME ZONE,
  updated_at TIMESTAMP WITH TIME ZONE,
  deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE UNIQUE INDEX uk_user_tokens_token ON user_tokens(token);
CREATE UNIQUE INDEX uk_user_tokens_app_id ON user_tokens(app_id);
CREATE INDEX idx_user_tokens_user_id ON user_tokens(user_id);
CREATE INDEX idx_user_tokens_deleted_at ON user_tokens(deleted_at);

-- +goose Down
DROP TABLE IF EXISTS user_tokens;
