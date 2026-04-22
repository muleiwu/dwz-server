-- +goose Up
CREATE TABLE `click_statistics` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `short_link_id` BIGINT UNSIGNED NOT NULL,
  `ip` VARCHAR(45) NULL,
  `user_agent` VARCHAR(1024) NULL,
  `referer` VARCHAR(2048) NULL,
  `query_params` VARCHAR(2048) NULL,
  `country` VARCHAR(100) NULL,
  `city` VARCHAR(100) NULL,
  `click_date` DATETIME(3) NULL,
  `created_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  KEY `idx_short_link_date` (`short_link_id`, `click_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS `click_statistics`;
