-- +goose Up
CREATE TABLE domains (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  protocol TEXT NOT NULL DEFAULT 'https',
  domain TEXT NOT NULL,
  site_name TEXT NOT NULL DEFAULT '',
  icp_number TEXT NOT NULL DEFAULT '',
  police_number TEXT NOT NULL DEFAULT '',
  pass_query_params INTEGER NOT NULL DEFAULT 0,
  description TEXT,
  is_active INTEGER NOT NULL DEFAULT 1,
  random_suffix_length INTEGER DEFAULT 2,
  enable_checksum INTEGER DEFAULT 1,
  enable_xor_obfuscation INTEGER DEFAULT 0,
  enable_anti_red INTEGER DEFAULT 0,
  xor_secret INTEGER,
  xor_rot INTEGER,
  default_start_number INTEGER DEFAULT 0,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME
);
CREATE UNIQUE INDEX uk_domains_domain ON domains(domain);
CREATE INDEX idx_domains_deleted_at ON domains(deleted_at);

-- +goose Down
DROP TABLE IF EXISTS domains;
