-- +goose Up
CREATE TABLE IF NOT EXISTS ee_system_brandings (
  id INTEGER NOT NULL,
  brand_name TEXT NOT NULL DEFAULT '',
  logo_url TEXT NOT NULL DEFAULT '',
  copyright_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  copyright_text TEXT NOT NULL DEFAULT '',
  copyright_link TEXT NOT NULL DEFAULT '',
  created_at DATETIME,
  updated_at DATETIME,
  PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS ee_workspace_brandings (
  workspace_id INTEGER NOT NULL,
  brand_name TEXT NOT NULL DEFAULT '',
  logo_url VARCHAR(500) NOT NULL DEFAULT '',
  copyright_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  copyright_text TEXT NOT NULL DEFAULT '',
  copyright_link TEXT NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (workspace_id)
);

CREATE TABLE IF NOT EXISTS ee_domain_brandings (
  domain_id INTEGER NOT NULL,
  workspace_id INTEGER NOT NULL,
  override_brand_name BOOLEAN NOT NULL DEFAULT FALSE,
  brand_name TEXT NOT NULL DEFAULT '',
  override_logo BOOLEAN NOT NULL DEFAULT FALSE,
  logo_url VARCHAR(500) NOT NULL DEFAULT '',
  override_copyright BOOLEAN NOT NULL DEFAULT FALSE,
  copyright_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  override_copyright_text BOOLEAN NOT NULL DEFAULT FALSE,
  copyright_text TEXT NOT NULL DEFAULT '',
  copyright_link TEXT NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (domain_id)
);

CREATE INDEX IF NOT EXISTS idx_ee_domain_brandings_workspace_id
  ON ee_domain_brandings(workspace_id);

-- +goose Down
SELECT 1;
