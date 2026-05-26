-- +goose Up
CREATE TABLE `ab_test_feedbacks` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `workspace_id` BIGINT UNSIGNED NOT NULL DEFAULT 1,
  `ab_test_id` BIGINT UNSIGNED NOT NULL,
  `variant_id` BIGINT UNSIGNED NOT NULL,
  `short_link_id` BIGINT UNSIGNED NOT NULL,
  `session_id` VARCHAR(128) NOT NULL,
  `event_id` VARCHAR(128) NOT NULL,
  `value` DECIMAL(18,4) NULL,
  `currency` VARCHAR(16) NULL,
  `metadata` TEXT NULL,
  `ip` VARCHAR(45) NULL,
  `user_agent` VARCHAR(1024) NULL,
  `referer` VARCHAR(2048) NULL,
  `occurred_at` DATETIME(3) NOT NULL,
  `created_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_ab_test_feedback_event` (`ab_test_id`, `event_id`),
  KEY `idx_ab_test_feedbacks_workspace_id` (`workspace_id`),
  KEY `idx_ab_test_feedbacks_variant_id` (`variant_id`),
  KEY `idx_ab_test_feedbacks_short_link_id` (`short_link_id`),
  KEY `idx_ab_test_feedbacks_session_id` (`session_id`),
  KEY `idx_ab_test_feedbacks_occurred_at` (`occurred_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS `ab_test_feedbacks`;
