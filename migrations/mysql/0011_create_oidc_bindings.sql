-- +goose Up
CREATE TABLE `oidc_bindings` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `provider` VARCHAR(50) NOT NULL,
  `sub` VARCHAR(255) NOT NULL,
  `email` VARCHAR(255) NULL,
  `last_login_at` DATETIME(3) NULL,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_oidc_bindings_provider_sub` (`provider`, `sub`),
  KEY `idx_oidc_bindings_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS `oidc_bindings`;
