-- +goose Up
CREATE TABLE `short_links` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `issuer_number` BIGINT UNSIGNED NULL,
  `domain_id` BIGINT UNSIGNED NOT NULL,
  `protocol` VARCHAR(10) NOT NULL DEFAULT 'https',
  `domain` VARCHAR(100) NOT NULL,
  `original_url` VARCHAR(2000) NOT NULL,
  `title` VARCHAR(255) NULL,
  `is_custom_code` TINYINT(1) NOT NULL DEFAULT 0,
  `short_code` VARCHAR(20) NULL,
  `click_count` BIGINT NOT NULL DEFAULT 0,
  `creator_ip` VARCHAR(45) NULL,
  `description` VARCHAR(500) NULL,
  `expire_at` DATETIME(3) NULL,
  `is_active` TINYINT(1) NOT NULL DEFAULT 1,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  KEY `idx_short_links_issuer_number` (`issuer_number`),
  KEY `idx_short_links_domain_id` (`domain_id`),
  KEY `idx_short_links_domain` (`domain`),
  KEY `idx_short_links_short_code` (`short_code`),
  KEY `idx_short_links_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS `short_links`;
