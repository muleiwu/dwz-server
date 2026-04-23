-- +goose Up
CREATE TABLE `ab_test_variants` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `ab_test_id` BIGINT UNSIGNED NOT NULL,
  `name` VARCHAR(100) NOT NULL,
  `target_url` VARCHAR(2000) NOT NULL,
  `weight` INT NOT NULL DEFAULT 50,
  `is_control` TINYINT(1) NOT NULL DEFAULT 0,
  `description` VARCHAR(500) NULL,
  `is_active` TINYINT(1) NOT NULL DEFAULT 1,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  KEY `idx_ab_test_variants_ab_test_id` (`ab_test_id`),
  KEY `idx_ab_test_variants_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS `ab_test_variants`;
