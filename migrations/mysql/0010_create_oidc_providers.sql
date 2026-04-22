-- +goose Up
CREATE TABLE `oidc_providers` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(50) NOT NULL,
  `display_name` VARCHAR(100) NULL,
  `issuer` VARCHAR(255) NOT NULL,
  `client_id` VARCHAR(255) NOT NULL,
  `client_secret` VARCHAR(1024) NOT NULL,
  `scopes` VARCHAR(255) NULL,
  `redirect_uri` VARCHAR(255) NULL,
  `enabled` TINYINT NOT NULL DEFAULT 0,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_oidc_providers_name` (`name`),
  KEY `idx_oidc_providers_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS `oidc_providers`;
