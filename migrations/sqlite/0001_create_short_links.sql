-- +goose Up
CREATE TABLE short_links (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  issuer_number INTEGER,
  domain_id INTEGER NOT NULL,
  protocol TEXT NOT NULL DEFAULT 'https',
  domain TEXT NOT NULL,
  original_url TEXT NOT NULL,
  title TEXT,
  is_custom_code INTEGER NOT NULL DEFAULT 0,
  short_code TEXT,
  click_count INTEGER NOT NULL DEFAULT 0,
  creator_ip TEXT,
  description TEXT,
  expire_at DATETIME,
  is_active INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME
);
CREATE INDEX idx_short_links_issuer_number ON short_links(issuer_number);
CREATE INDEX idx_short_links_domain_id ON short_links(domain_id);
CREATE INDEX idx_short_links_domain ON short_links(domain);
CREATE INDEX idx_short_links_short_code ON short_links(short_code);
CREATE INDEX idx_short_links_deleted_at ON short_links(deleted_at);

-- +goose Down
DROP TABLE IF EXISTS short_links;
