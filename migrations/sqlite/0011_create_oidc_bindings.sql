-- +goose Up
CREATE TABLE oidc_bindings (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  provider TEXT NOT NULL,
  sub TEXT NOT NULL,
  email TEXT,
  last_login_at DATETIME,
  created_at DATETIME,
  updated_at DATETIME
);
CREATE UNIQUE INDEX uk_oidc_bindings_provider_sub ON oidc_bindings(provider, sub);
CREATE INDEX idx_oidc_bindings_user_id ON oidc_bindings(user_id);

-- +goose Down
DROP TABLE IF EXISTS oidc_bindings;
