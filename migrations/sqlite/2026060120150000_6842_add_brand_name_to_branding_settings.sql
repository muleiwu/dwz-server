-- +goose Up
CREATE TABLE IF NOT EXISTS ee_system_brandings (
  id INTEGER NOT NULL,
  logo_url TEXT NOT NULL DEFAULT '',
  copyright_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME,
  updated_at DATETIME,
  PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS ee_workspace_brandings (
  workspace_id INTEGER NOT NULL,
  logo_url VARCHAR(500) NOT NULL DEFAULT '',
  copyright_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (workspace_id)
);

CREATE TABLE IF NOT EXISTS ee_domain_brandings (
  domain_id INTEGER NOT NULL,
  workspace_id INTEGER NOT NULL,
  override_logo BOOLEAN NOT NULL DEFAULT FALSE,
  logo_url VARCHAR(500) NOT NULL DEFAULT '',
  override_copyright BOOLEAN NOT NULL DEFAULT FALSE,
  copyright_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (domain_id)
);

CREATE INDEX IF NOT EXISTS idx_ee_domain_brandings_workspace_id
  ON ee_domain_brandings(workspace_id);

ALTER TABLE ee_system_brandings ADD COLUMN brand_name TEXT NOT NULL DEFAULT '';
ALTER TABLE ee_workspace_brandings ADD COLUMN brand_name TEXT NOT NULL DEFAULT '';
ALTER TABLE ee_domain_brandings ADD COLUMN override_brand_name BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE ee_domain_brandings ADD COLUMN brand_name TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE ee_domain_brandings DROP COLUMN brand_name;
ALTER TABLE ee_domain_brandings DROP COLUMN override_brand_name;
ALTER TABLE ee_workspace_brandings DROP COLUMN brand_name;
ALTER TABLE ee_system_brandings DROP COLUMN brand_name;
