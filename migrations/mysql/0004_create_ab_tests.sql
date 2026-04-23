-- +goose Up
CREATE TABLE `ab_tests` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `short_link_id` BIGINT UNSIGNED NOT NULL,
  `name` VARCHAR(255) NOT NULL,
  `description` VARCHAR(500) NULL,
  `status` VARCHAR(20) NOT NULL DEFAULT 'draft',
  `traffic_split` VARCHAR(20) NOT NULL DEFAULT 'equal',
  `start_time` DATETIME(3) NULL,
  `end_time` DATETIME(3) NULL,
  `is_active` TINYINT(1) NOT NULL DEFAULT 1,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  KEY `idx_ab_tests_short_link_id` (`short_link_id`),
  KEY `idx_ab_tests_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS `ab_tests`;
