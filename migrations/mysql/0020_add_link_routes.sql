-- +goose Up
ALTER TABLE `short_links`
  ADD COLUMN `fallback_url` VARCHAR(2000) NULL,
  ADD COLUMN `redirect_code` INT NOT NULL DEFAULT 302;

ALTER TABLE `click_statistics`
  ADD COLUMN `route_id` BIGINT UNSIGNED NULL,
  ADD COLUMN `route_name` VARCHAR(100) NULL;

CREATE INDEX `idx_click_statistics_route_id` ON `click_statistics` (`route_id`);

CREATE TABLE `link_routes` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `workspace_id` BIGINT UNSIGNED NOT NULL,
  `short_link_id` BIGINT UNSIGNED NOT NULL,
  `name` VARCHAR(100) NOT NULL,
  `description` VARCHAR(500) NULL,
  `priority` INT NOT NULL DEFAULT 100,
  `target_url` VARCHAR(2000) NOT NULL,
  `is_active` TINYINT(1) NOT NULL DEFAULT 1,
  `created_by` BIGINT UNSIGNED NULL,
  `updated_by` BIGINT UNSIGNED NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME NULL,
  PRIMARY KEY (`id`),
  KEY `idx_link_routes_workspace` (`workspace_id`),
  KEY `idx_link_routes_short_link` (`short_link_id`),
  KEY `idx_link_routes_priority` (`priority`),
  KEY `idx_link_routes_active` (`is_active`),
  KEY `idx_link_routes_created_by` (`created_by`),
  KEY `idx_link_routes_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `link_route_condition_groups` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `route_id` BIGINT UNSIGNED NOT NULL,
  `position` INT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_route_condition_groups_route` (`route_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `link_route_conditions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `group_id` BIGINT UNSIGNED NOT NULL,
  `condition_type` VARCHAR(30) NOT NULL,
  `operator` VARCHAR(20) NOT NULL,
  `condition_key` VARCHAR(255) NULL,
  `condition_value` VARCHAR(1000) NULL,
  `position` INT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_route_conditions_group` (`group_id`),
  KEY `idx_route_conditions_type` (`condition_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +goose Down
DROP TABLE IF EXISTS `link_route_conditions`;
DROP TABLE IF EXISTS `link_route_condition_groups`;
DROP TABLE IF EXISTS `link_routes`;
DROP INDEX `idx_click_statistics_route_id` ON `click_statistics`;
ALTER TABLE `click_statistics`
  DROP COLUMN `route_name`,
  DROP COLUMN `route_id`;
ALTER TABLE `short_links`
  DROP COLUMN `redirect_code`,
  DROP COLUMN `fallback_url`;
