-- +goose Up
CREATE TABLE `domains` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `protocol` VARCHAR(10) NOT NULL DEFAULT 'https',
  `domain` VARCHAR(100) NOT NULL,
  `site_name` VARCHAR(100) NOT NULL DEFAULT '',
  `icp_number` VARCHAR(50) NOT NULL DEFAULT '',
  `police_number` VARCHAR(50) NOT NULL DEFAULT '',
  `pass_query_params` TINYINT(1) NOT NULL DEFAULT 0,
  `description` TEXT NULL,
  `is_active` TINYINT(1) NOT NULL DEFAULT 1,
  `random_suffix_length` INT NULL DEFAULT 2,
  `enable_checksum` TINYINT(1) NULL DEFAULT 1,
  `enable_xor_obfuscation` TINYINT(1) NULL DEFAULT 0,
  `enable_anti_red` TINYINT(1) NULL DEFAULT 0,
  `xor_secret` BIGINT UNSIGNED NULL,
  `xor_rot` INT NULL,
  `default_start_number` BIGINT UNSIGNED NULL DEFAULT 0,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_domains_domain` (`domain`),
  KEY `idx_domains_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS `domains`;
