-- +goose Up
CREATE TABLE `ab_test_click_statistics` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `ab_test_id` BIGINT UNSIGNED NOT NULL,
  `variant_id` BIGINT UNSIGNED NOT NULL,
  `short_link_id` BIGINT UNSIGNED NOT NULL,
  `ip` VARCHAR(45) NULL,
  `user_agent` VARCHAR(1024) NULL,
  `referer` VARCHAR(2048) NULL,
  `query_params` VARCHAR(2048) NULL,
  `country` VARCHAR(100) NULL,
  `city` VARCHAR(100) NULL,
  `session_id` VARCHAR(128) NULL,
  `click_date` DATETIME(3) NULL,
  `created_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  KEY `idx_ab_test_click` (`ab_test_id`, `click_date`),
  KEY `idx_variant_click` (`variant_id`, `click_date`),
  KEY `idx_ab_test_click_statistics_short_link_id` (`short_link_id`),
  KEY `idx_ab_test_click_statistics_session_id` (`session_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS `ab_test_click_statistics`;
