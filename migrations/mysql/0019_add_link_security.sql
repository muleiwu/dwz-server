-- +goose Up
CREATE TABLE `link_security_settings` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `workspace_id` BIGINT UNSIGNED NOT NULL,
  `short_link_id` BIGINT UNSIGNED NOT NULL,
  `password_enabled` TINYINT(1) NOT NULL DEFAULT 0,
  `password_hash` VARCHAR(255) NULL,
  `access_window_start` DATETIME NULL,
  `access_window_end` DATETIME NULL,
  `max_clicks` BIGINT NULL,
  `ip_policy` VARCHAR(20) NOT NULL DEFAULT 'off',
  `bot_policy` VARCHAR(30) NOT NULL DEFAULT 'record_only',
  `report_enabled` TINYINT(1) NOT NULL DEFAULT 0,
  `url_blocked` TINYINT(1) NOT NULL DEFAULT 0,
  `url_blocked_reason` VARCHAR(500) NULL,
  `created_by` BIGINT UNSIGNED NULL,
  `updated_by` BIGINT UNSIGNED NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_link_security_short_link` (`short_link_id`),
  KEY `idx_link_security_workspace` (`workspace_id`),
  KEY `idx_link_security_created_by` (`created_by`),
  KEY `idx_link_security_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `link_security_ip_rules` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `workspace_id` BIGINT UNSIGNED NOT NULL,
  `short_link_id` BIGINT UNSIGNED NOT NULL,
  `cidr` VARCHAR(64) NOT NULL,
  `description` VARCHAR(255) NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME NULL,
  PRIMARY KEY (`id`),
  KEY `idx_link_security_ip_workspace` (`workspace_id`),
  KEY `idx_link_security_ip_short_link` (`short_link_id`),
  KEY `idx_link_security_ip_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `security_url_rules` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `workspace_id` BIGINT UNSIGNED NOT NULL,
  `rule_type` VARCHAR(20) NOT NULL,
  `action` VARCHAR(20) NOT NULL,
  `pattern` VARCHAR(500) NOT NULL,
  `enabled` TINYINT(1) NOT NULL DEFAULT 1,
  `created_by` BIGINT UNSIGNED NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME NULL,
  PRIMARY KEY (`id`),
  KEY `idx_security_url_rules_workspace` (`workspace_id`),
  KEY `idx_security_url_rules_type` (`rule_type`),
  KEY `idx_security_url_rules_action` (`action`),
  KEY `idx_security_url_rules_enabled` (`enabled`),
  KEY `idx_security_url_rules_created_by` (`created_by`),
  KEY `idx_security_url_rules_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `abuse_reports` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `workspace_id` BIGINT UNSIGNED NOT NULL,
  `short_link_id` BIGINT UNSIGNED NOT NULL,
  `report_type` VARCHAR(30) NOT NULL,
  `description` VARCHAR(1000) NULL,
  `reporter_email` VARCHAR(255) NULL,
  `reporter_ip` VARCHAR(45) NULL,
  `user_agent` VARCHAR(1024) NULL,
  `status` VARCHAR(30) NOT NULL DEFAULT 'pending',
  `resolution_note` VARCHAR(1000) NULL,
  `handled_by` BIGINT UNSIGNED NULL,
  `handled_at` DATETIME NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_abuse_reports_workspace` (`workspace_id`),
  KEY `idx_abuse_reports_short_link` (`short_link_id`),
  KEY `idx_abuse_reports_type` (`report_type`),
  KEY `idx_abuse_reports_status` (`status`),
  KEY `idx_abuse_reports_reporter_ip` (`reporter_ip`),
  KEY `idx_abuse_reports_handled_by` (`handled_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `link_security_events` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `workspace_id` BIGINT UNSIGNED NOT NULL,
  `short_link_id` BIGINT UNSIGNED NOT NULL,
  `event_type` VARCHAR(50) NOT NULL,
  `reason` VARCHAR(500) NULL,
  `client_ip` VARCHAR(45) NULL,
  `user_agent` VARCHAR(1024) NULL,
  `referer` VARCHAR(2048) NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_link_security_events_workspace` (`workspace_id`),
  KEY `idx_link_security_events_short_link` (`short_link_id`),
  KEY `idx_link_security_events_type` (`event_type`),
  KEY `idx_link_security_events_client_ip` (`client_ip`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +goose Down
DROP TABLE IF EXISTS `link_security_events`;
DROP TABLE IF EXISTS `abuse_reports`;
DROP TABLE IF EXISTS `security_url_rules`;
DROP TABLE IF EXISTS `link_security_ip_rules`;
DROP TABLE IF EXISTS `link_security_settings`;
