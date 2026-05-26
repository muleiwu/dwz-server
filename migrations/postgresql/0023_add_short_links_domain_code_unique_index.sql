-- +goose Up
CREATE UNIQUE INDEX uk_short_links_domain_code
  ON short_links(domain, short_code)
  WHERE deleted_at IS NULL AND short_code IS NOT NULL AND short_code <> '';

-- +goose Down
DROP INDEX IF EXISTS uk_short_links_domain_code;
