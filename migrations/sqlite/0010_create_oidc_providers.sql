-- +goose Up
CREATE TABLE oidc_providers (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  display_name TEXT,
  issuer TEXT NOT NULL,
  client_id TEXT NOT NULL,
  client_secret TEXT NOT NULL,
  scopes TEXT,
  redirect_uri TEXT,
  enabled INTEGER NOT NULL DEFAULT 0,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME
);
CREATE UNIQUE INDEX uk_oidc_providers_name ON oidc_providers(name);
CREATE INDEX idx_oidc_providers_deleted_at ON oidc_providers(deleted_at);

-- +goose Down
DROP TABLE IF EXISTS oidc_providers;
