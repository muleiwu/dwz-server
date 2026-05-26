-- +goose Up
CREATE TABLE campaigns (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  workspace_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  description TEXT,
  start_at DATETIME,
  end_at DATETIME,
  status TEXT NOT NULL DEFAULT 'active',
  created_by INTEGER,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME
);
CREATE INDEX idx_campaigns_workspace_id ON campaigns(workspace_id);
CREATE INDEX idx_campaigns_status ON campaigns(status);
CREATE INDEX idx_campaigns_deleted_at ON campaigns(deleted_at);

CREATE TABLE tags (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  workspace_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  color TEXT,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME
);
CREATE UNIQUE INDEX uk_tags_workspace_name ON tags(workspace_id, name);
CREATE INDEX idx_tags_deleted_at ON tags(deleted_at);

CREATE TABLE short_link_tags (
  short_link_id INTEGER NOT NULL,
  tag_id INTEGER NOT NULL,
  created_at DATETIME,
  PRIMARY KEY (short_link_id, tag_id)
);
CREATE INDEX idx_short_link_tags_tag_id ON short_link_tags(tag_id);

ALTER TABLE short_links ADD COLUMN campaign_id INTEGER;
ALTER TABLE short_links ADD COLUMN utm_source TEXT;
ALTER TABLE short_links ADD COLUMN utm_medium TEXT;
ALTER TABLE short_links ADD COLUMN utm_campaign TEXT;
ALTER TABLE short_links ADD COLUMN utm_term TEXT;
ALTER TABLE short_links ADD COLUMN utm_content TEXT;
ALTER TABLE short_links ADD COLUMN notes TEXT;

CREATE INDEX idx_short_links_campaign_id ON short_links(campaign_id);

-- +goose Down
DROP INDEX IF EXISTS idx_short_links_campaign_id;
ALTER TABLE short_links DROP COLUMN notes;
ALTER TABLE short_links DROP COLUMN utm_content;
ALTER TABLE short_links DROP COLUMN utm_term;
ALTER TABLE short_links DROP COLUMN utm_campaign;
ALTER TABLE short_links DROP COLUMN utm_medium;
ALTER TABLE short_links DROP COLUMN utm_source;
ALTER TABLE short_links DROP COLUMN campaign_id;
DROP TABLE IF EXISTS short_link_tags;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS campaigns;
