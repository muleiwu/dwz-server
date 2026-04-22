-- +goose Up
CREATE TABLE oidc_providers (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(50) NOT NULL,
  display_name VARCHAR(100),
  issuer VARCHAR(255) NOT NULL,
  client_id VARCHAR(255) NOT NULL,
  client_secret VARCHAR(1024) NOT NULL,
  scopes VARCHAR(255),
  redirect_uri VARCHAR(255),
  enabled SMALLINT NOT NULL DEFAULT 0,
  created_at TIMESTAMP WITH TIME ZONE,
  updated_at TIMESTAMP WITH TIME ZONE,
  deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE UNIQUE INDEX uk_oidc_providers_name ON oidc_providers(name);
CREATE INDEX idx_oidc_providers_deleted_at ON oidc_providers(deleted_at);

-- +goose Down
DROP TABLE IF EXISTS oidc_providers;
