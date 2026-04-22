-- +goose Up
CREATE TABLE oidc_bindings (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  provider VARCHAR(50) NOT NULL,
  sub VARCHAR(255) NOT NULL,
  email VARCHAR(255),
  last_login_at TIMESTAMP WITH TIME ZONE,
  created_at TIMESTAMP WITH TIME ZONE,
  updated_at TIMESTAMP WITH TIME ZONE
);
CREATE UNIQUE INDEX uk_oidc_bindings_provider_sub ON oidc_bindings(provider, sub);
CREATE INDEX idx_oidc_bindings_user_id ON oidc_bindings(user_id);

-- +goose Down
DROP TABLE IF EXISTS oidc_bindings;
