-- +goose Up
CREATE TABLE `operation_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NULL,
  `username` VARCHAR(50) NULL,
  `operation` VARCHAR(100) NOT NULL,
  `resource` VARCHAR(100) NULL,
  `resource_id` VARCHAR(100) NULL,
  `method` VARCHAR(10) NULL,
  `path` VARCHAR(255) NULL,
  `request_body` TEXT NULL,
  `response_code` INT NOT NULL DEFAULT 0,
  `response_body` TEXT NULL,
  `ip` VARCHAR(45) NULL,
  `user_agent` VARCHAR(500) NULL,
  `execute_time` BIGINT NOT NULL DEFAULT 0,
  `status` TINYINT NOT NULL DEFAULT 1,
  `error_message` VARCHAR(1000) NULL,
  `created_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  KEY `idx_operation_logs_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS `operation_logs`;
