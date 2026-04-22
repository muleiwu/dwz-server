-- +goose Up
CREATE TABLE `user_tokens` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `token_name` VARCHAR(100) NOT NULL,
  `token_type` VARCHAR(20) NOT NULL DEFAULT 'bearer',
  `token` VARCHAR(190) NULL,
  `app_id` VARCHAR(64) NULL,
  `app_secret` VARCHAR(256) NULL,
  `last_used_at` DATETIME(3) NULL,
  `expire_at` DATETIME(3) NULL,
  `status` TINYINT NOT NULL DEFAULT 1,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_tokens_token` (`token`),
  UNIQUE KEY `uk_user_tokens_app_id` (`app_id`),
  KEY `idx_user_tokens_user_id` (`user_id`),
  KEY `idx_user_tokens_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS `user_tokens`;
