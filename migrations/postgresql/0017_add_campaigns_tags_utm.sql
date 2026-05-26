-- +goose Up
CREATE TABLE campaigns (
  id BIGSERIAL PRIMARY KEY,
  workspace_id BIGINT NOT NULL,
  name VARCHAR(150) NOT NULL,
  description VARCHAR(500),
  start_at TIMESTAMP WITH TIME ZONE,
  end_at TIMESTAMP WITH TIME ZONE,
  status VARCHAR(20) NOT NULL DEFAULT 'active',
  created_by BIGINT,
  created_at TIMESTAMP WITH TIME ZONE,
  updated_at TIMESTAMP WITH TIME ZONE,
  deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_campaigns_workspace_id ON campaigns(workspace_id);
CREATE INDEX idx_campaigns_status ON campaigns(status);
CREATE INDEX idx_campaigns_deleted_at ON campaigns(deleted_at);

CREATE TABLE tags (
  id BIGSERIAL PRIMARY KEY,
  workspace_id BIGINT NOT NULL,
  name VARCHAR(100) NOT NULL,
  color VARCHAR(20),
  created_at TIMESTAMP WITH TIME ZONE,
  updated_at TIMESTAMP WITH TIME ZONE,
  deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE UNIQUE INDEX uk_tags_workspace_name ON tags(workspace_id, name);
CREATE INDEX idx_tags_deleted_at ON tags(deleted_at);

CREATE TABLE short_link_tags (
  short_link_id BIGINT NOT NULL,
  tag_id BIGINT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE,
  PRIMARY KEY (short_link_id, tag_id)
);
CREATE INDEX idx_short_link_tags_tag_id ON short_link_tags(tag_id);

ALTER TABLE short_links ADD COLUMN campaign_id BIGINT;
ALTER TABLE short_links ADD COLUMN utm_source VARCHAR(255);
ALTER TABLE short_links ADD COLUMN utm_medium VARCHAR(255);
ALTER TABLE short_links ADD COLUMN utm_campaign VARCHAR(255);
ALTER TABLE short_links ADD COLUMN utm_term VARCHAR(255);
ALTER TABLE short_links ADD COLUMN utm_content VARCHAR(255);
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
