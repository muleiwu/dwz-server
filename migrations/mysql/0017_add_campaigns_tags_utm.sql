-- +goose Up
CREATE TABLE `campaigns` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `workspace_id` BIGINT UNSIGNED NOT NULL,
  `name` VARCHAR(150) NOT NULL,
  `description` VARCHAR(500) NULL,
  `start_at` DATETIME(3) NULL,
  `end_at` DATETIME(3) NULL,
  `status` VARCHAR(20) NOT NULL DEFAULT 'active',
  `created_by` BIGINT UNSIGNED NULL,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  KEY `idx_campaigns_workspace_id` (`workspace_id`),
  KEY `idx_campaigns_status` (`status`),
  KEY `idx_campaigns_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `tags` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `workspace_id` BIGINT UNSIGNED NOT NULL,
  `name` VARCHAR(100) NOT NULL,
  `color` VARCHAR(20) NULL,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tags_workspace_name` (`workspace_id`, `name`),
  KEY `idx_tags_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `short_link_tags` (
  `short_link_id` BIGINT UNSIGNED NOT NULL,
  `tag_id` BIGINT UNSIGNED NOT NULL,
  `created_at` DATETIME(3) NULL,
  PRIMARY KEY (`short_link_id`, `tag_id`),
  KEY `idx_short_link_tags_tag_id` (`tag_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE `short_links`
  ADD COLUMN `campaign_id` BIGINT UNSIGNED NULL AFTER `workspace_id`,
  ADD COLUMN `utm_source` VARCHAR(255) NULL AFTER `description`,
  ADD COLUMN `utm_medium` VARCHAR(255) NULL AFTER `utm_source`,
  ADD COLUMN `utm_campaign` VARCHAR(255) NULL AFTER `utm_medium`,
  ADD COLUMN `utm_term` VARCHAR(255) NULL AFTER `utm_campaign`,
  ADD COLUMN `utm_content` VARCHAR(255) NULL AFTER `utm_term`,
  ADD COLUMN `notes` TEXT NULL AFTER `utm_content`;

CREATE INDEX `idx_short_links_campaign_id` ON `short_links` (`campaign_id`);

-- +goose Down
DROP INDEX `idx_short_links_campaign_id` ON `short_links`;
ALTER TABLE `short_links`
  DROP COLUMN `notes`,
  DROP COLUMN `utm_content`,
  DROP COLUMN `utm_term`,
  DROP COLUMN `utm_campaign`,
  DROP COLUMN `utm_medium`,
  DROP COLUMN `utm_source`,
  DROP COLUMN `campaign_id`;
DROP TABLE IF EXISTS `short_link_tags`;
DROP TABLE IF EXISTS `tags`;
DROP TABLE IF EXISTS `campaigns`;
