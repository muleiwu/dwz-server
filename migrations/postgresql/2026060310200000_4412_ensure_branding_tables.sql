-- +goose Up
CREATE TABLE IF NOT EXISTS ee_system_brandings (
  id SMALLINT NOT NULL,
  brand_name VARCHAR(80) NOT NULL DEFAULT '',
  logo_url VARCHAR(500) NOT NULL DEFAULT '',
  copyright_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  copyright_text VARCHAR(200) NOT NULL DEFAULT '',
  copyright_link VARCHAR(500) NOT NULL DEFAULT '',
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS ee_workspace_brandings (
  workspace_id BIGINT NOT NULL,
  brand_name VARCHAR(80) NOT NULL DEFAULT '',
  logo_url VARCHAR(500) NOT NULL DEFAULT '',
  copyright_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  copyright_text VARCHAR(200) NOT NULL DEFAULT '',
  copyright_link VARCHAR(500) NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (workspace_id)
);

CREATE TABLE IF NOT EXISTS ee_domain_brandings (
  domain_id BIGINT NOT NULL,
  workspace_id BIGINT NOT NULL,
  override_brand_name BOOLEAN NOT NULL DEFAULT FALSE,
  brand_name VARCHAR(80) NOT NULL DEFAULT '',
  override_logo BOOLEAN NOT NULL DEFAULT FALSE,
  logo_url VARCHAR(500) NOT NULL DEFAULT '',
  override_copyright BOOLEAN NOT NULL DEFAULT FALSE,
  copyright_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  override_copyright_text BOOLEAN NOT NULL DEFAULT FALSE,
  copyright_text VARCHAR(200) NOT NULL DEFAULT '',
  copyright_link VARCHAR(500) NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (domain_id)
);

CREATE INDEX IF NOT EXISTS idx_ee_domain_brandings_workspace_id
  ON ee_domain_brandings(workspace_id);

-- +goose Down
SELECT 1;
