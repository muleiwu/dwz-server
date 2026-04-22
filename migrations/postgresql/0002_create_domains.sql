-- +goose Up
CREATE TABLE domains (
  id BIGSERIAL PRIMARY KEY,
  protocol VARCHAR(10) NOT NULL DEFAULT 'https',
  domain VARCHAR(100) NOT NULL,
  site_name VARCHAR(100) NOT NULL DEFAULT '',
  icp_number VARCHAR(50) NOT NULL DEFAULT '',
  police_number VARCHAR(50) NOT NULL DEFAULT '',
  pass_query_params BOOLEAN NOT NULL DEFAULT FALSE,
  description TEXT,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  random_suffix_length INTEGER DEFAULT 2,
  enable_checksum BOOLEAN DEFAULT TRUE,
  enable_xor_obfuscation BOOLEAN DEFAULT FALSE,
  enable_anti_red BOOLEAN DEFAULT FALSE,
  xor_secret BIGINT,
  xor_rot INTEGER,
  default_start_number BIGINT DEFAULT 0,
  created_at TIMESTAMP WITH TIME ZONE,
  updated_at TIMESTAMP WITH TIME ZONE,
  deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE UNIQUE INDEX uk_domains_domain ON domains(domain);
CREATE INDEX idx_domains_deleted_at ON domains(deleted_at);

-- +goose Down
DROP TABLE IF EXISTS domains;
