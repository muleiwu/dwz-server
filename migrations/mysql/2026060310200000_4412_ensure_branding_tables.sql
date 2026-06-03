-- +goose Up
CREATE TABLE IF NOT EXISTS `ee_system_brandings` (
  `id` TINYINT UNSIGNED NOT NULL,
  `brand_name` VARCHAR(80) NOT NULL DEFAULT '',
  `logo_url` VARCHAR(500) NOT NULL DEFAULT '',
  `copyright_enabled` BOOLEAN NOT NULL DEFAULT TRUE,
  `copyright_text` VARCHAR(200) NOT NULL DEFAULT '',
  `copyright_link` VARCHAR(500) NOT NULL DEFAULT '',
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`)
);

CREATE TABLE IF NOT EXISTS `ee_workspace_brandings` (
  `workspace_id` BIGINT UNSIGNED NOT NULL,
  `brand_name` VARCHAR(80) NOT NULL DEFAULT '',
  `logo_url` VARCHAR(500) NOT NULL DEFAULT '',
  `copyright_enabled` BOOLEAN NOT NULL DEFAULT TRUE,
  `copyright_text` VARCHAR(200) NOT NULL DEFAULT '',
  `copyright_link` VARCHAR(500) NOT NULL DEFAULT '',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`workspace_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ee_domain_brandings` (
  `domain_id` BIGINT UNSIGNED NOT NULL,
  `workspace_id` BIGINT UNSIGNED NOT NULL,
  `override_brand_name` BOOLEAN NOT NULL DEFAULT FALSE,
  `brand_name` VARCHAR(80) NOT NULL DEFAULT '',
  `override_logo` BOOLEAN NOT NULL DEFAULT FALSE,
  `logo_url` VARCHAR(500) NOT NULL DEFAULT '',
  `override_copyright` BOOLEAN NOT NULL DEFAULT FALSE,
  `copyright_enabled` BOOLEAN NOT NULL DEFAULT TRUE,
  `override_copyright_text` BOOLEAN NOT NULL DEFAULT FALSE,
  `copyright_text` VARCHAR(200) NOT NULL DEFAULT '',
  `copyright_link` VARCHAR(500) NOT NULL DEFAULT '',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`domain_id`),
  KEY `idx_ee_domain_brandings_workspace_id` (`workspace_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
SELECT 1;
