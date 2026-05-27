-- +goose Up
CREATE TABLE IF NOT EXISTS `ee_workspace_brandings` (
  `workspace_id` BIGINT UNSIGNED NOT NULL,
  `logo_url` VARCHAR(500) NOT NULL DEFAULT '',
  `copyright_enabled` BOOLEAN NOT NULL DEFAULT TRUE,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`workspace_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ee_domain_brandings` (
  `domain_id` BIGINT UNSIGNED NOT NULL,
  `workspace_id` BIGINT UNSIGNED NOT NULL,
  `override_logo` BOOLEAN NOT NULL DEFAULT FALSE,
  `logo_url` VARCHAR(500) NOT NULL DEFAULT '',
  `override_copyright` BOOLEAN NOT NULL DEFAULT FALSE,
  `copyright_enabled` BOOLEAN NOT NULL DEFAULT TRUE,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`domain_id`),
  KEY `idx_ee_domain_brandings_workspace_id` (`workspace_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS `ee_domain_brandings`;
DROP TABLE IF EXISTS `ee_workspace_brandings`;
