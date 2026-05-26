-- +goose Up
ALTER TABLE `short_links`
  ADD COLUMN `short_code_active_key` VARCHAR(20)
    GENERATED ALWAYS AS (
      CASE
        WHEN `deleted_at` IS NULL AND `short_code` IS NOT NULL AND `short_code` <> ''
        THEN `short_code`
        ELSE NULL
      END
    ) STORED,
  ADD UNIQUE KEY `uk_short_links_domain_code` (`domain`, `short_code_active_key`);

-- +goose Down
ALTER TABLE `short_links`
  DROP INDEX `uk_short_links_domain_code`,
  DROP COLUMN `short_code_active_key`;
