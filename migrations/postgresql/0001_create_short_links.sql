-- +goose Up
CREATE TABLE short_links (
  id BIGSERIAL PRIMARY KEY,
  issuer_number BIGINT,
  domain_id BIGINT NOT NULL,
  protocol VARCHAR(10) NOT NULL DEFAULT 'https',
  domain VARCHAR(100) NOT NULL,
  original_url VARCHAR(2000) NOT NULL,
  title VARCHAR(255),
  is_custom_code BOOLEAN NOT NULL DEFAULT FALSE,
  short_code VARCHAR(20),
  click_count BIGINT NOT NULL DEFAULT 0,
  creator_ip VARCHAR(45),
  description VARCHAR(500),
  expire_at TIMESTAMP WITH TIME ZONE,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP WITH TIME ZONE,
  updated_at TIMESTAMP WITH TIME ZONE,
  deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_short_links_issuer_number ON short_links(issuer_number);
CREATE INDEX idx_short_links_domain_id ON short_links(domain_id);
CREATE INDEX idx_short_links_domain ON short_links(domain);
CREATE INDEX idx_short_links_short_code ON short_links(short_code);
CREATE INDEX idx_short_links_deleted_at ON short_links(deleted_at);

-- +goose Down
DROP TABLE IF EXISTS short_links;
