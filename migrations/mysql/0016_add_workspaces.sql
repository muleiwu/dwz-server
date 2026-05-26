-- +goose Up
CREATE TABLE `workspaces` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `slug` VARCHAR(100) NOT NULL,
  `name` VARCHAR(100) NOT NULL,
  `description` VARCHAR(500) NULL,
  `owner_user_id` BIGINT UNSIGNED NULL,
  `status` TINYINT NOT NULL DEFAULT 1,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_workspaces_slug` (`slug`),
  KEY `idx_workspaces_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `workspace_members` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `workspace_id` BIGINT UNSIGNED NOT NULL,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `role` VARCHAR(20) NOT NULL,
  `status` TINYINT NOT NULL DEFAULT 1,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_workspace_members_workspace_user` (`workspace_id`, `user_id`),
  KEY `idx_workspace_members_user_id` (`user_id`),
  KEY `idx_workspace_members_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE `short_links`
  ADD COLUMN `workspace_id` BIGINT UNSIGNED NOT NULL DEFAULT 1 AFTER `id`,
  ADD COLUMN `created_by` BIGINT UNSIGNED NULL AFTER `creator_ip`,
  ADD COLUMN `updated_by` BIGINT UNSIGNED NULL AFTER `created_by`;
ALTER TABLE `domains` ADD COLUMN `workspace_id` BIGINT UNSIGNED NOT NULL DEFAULT 1 AFTER `id`;
ALTER TABLE `ab_tests` ADD COLUMN `workspace_id` BIGINT UNSIGNED NOT NULL DEFAULT 1 AFTER `id`;
ALTER TABLE `user_tokens` ADD COLUMN `workspace_id` BIGINT UNSIGNED NOT NULL DEFAULT 1 AFTER `id`;
ALTER TABLE `operation_logs` ADD COLUMN `workspace_id` BIGINT UNSIGNED NOT NULL DEFAULT 1 AFTER `id`;
ALTER TABLE `click_statistics` ADD COLUMN `workspace_id` BIGINT UNSIGNED NOT NULL DEFAULT 1 AFTER `id`;
ALTER TABLE `ab_test_click_statistics` ADD COLUMN `workspace_id` BIGINT UNSIGNED NOT NULL DEFAULT 1 AFTER `id`;

CREATE INDEX `idx_short_links_workspace_id` ON `short_links` (`workspace_id`);
CREATE INDEX `idx_short_links_created_by` ON `short_links` (`created_by`);
CREATE INDEX `idx_domains_workspace_id` ON `domains` (`workspace_id`);
CREATE INDEX `idx_ab_tests_workspace_id` ON `ab_tests` (`workspace_id`);
CREATE INDEX `idx_user_tokens_workspace_id` ON `user_tokens` (`workspace_id`);
CREATE INDEX `idx_operation_logs_workspace_id` ON `operation_logs` (`workspace_id`);
CREATE INDEX `idx_click_statistics_workspace_id` ON `click_statistics` (`workspace_id`);
CREATE INDEX `idx_ab_test_click_statistics_workspace_id` ON `ab_test_click_statistics` (`workspace_id`);

INSERT INTO `workspaces` (`id`, `slug`, `name`, `owner_user_id`, `status`, `created_at`, `updated_at`)
SELECT 1, 'default', '默认工作区', (SELECT MIN(`id`) FROM `users` WHERE `deleted_at` IS NULL), 1, NOW(3), NOW(3)
WHERE NOT EXISTS (SELECT 1 FROM `workspaces` WHERE `id` = 1);

INSERT INTO `workspace_members` (`workspace_id`, `user_id`, `role`, `status`, `created_at`, `updated_at`)
SELECT 1, `id`, CASE WHEN `id` = (SELECT MIN(`id`) FROM `users` WHERE `deleted_at` IS NULL) THEN 'owner' ELSE 'admin' END, 1, NOW(3), NOW(3)
FROM `users`
WHERE `deleted_at` IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM `workspace_members` wm
    WHERE wm.`workspace_id` = 1 AND wm.`user_id` = `users`.`id`
  );

-- +goose Down
DELETE FROM `workspace_members` WHERE `workspace_id` = 1;
DELETE FROM `workspaces` WHERE `id` = 1;
DROP INDEX `idx_ab_test_click_statistics_workspace_id` ON `ab_test_click_statistics`;
DROP INDEX `idx_click_statistics_workspace_id` ON `click_statistics`;
DROP INDEX `idx_operation_logs_workspace_id` ON `operation_logs`;
DROP INDEX `idx_user_tokens_workspace_id` ON `user_tokens`;
DROP INDEX `idx_ab_tests_workspace_id` ON `ab_tests`;
DROP INDEX `idx_domains_workspace_id` ON `domains`;
DROP INDEX `idx_short_links_created_by` ON `short_links`;
DROP INDEX `idx_short_links_workspace_id` ON `short_links`;
ALTER TABLE `ab_test_click_statistics` DROP COLUMN `workspace_id`;
ALTER TABLE `click_statistics` DROP COLUMN `workspace_id`;
ALTER TABLE `operation_logs` DROP COLUMN `workspace_id`;
ALTER TABLE `user_tokens` DROP COLUMN `workspace_id`;
ALTER TABLE `ab_tests` DROP COLUMN `workspace_id`;
ALTER TABLE `domains` DROP COLUMN `workspace_id`;
ALTER TABLE `short_links`
  DROP COLUMN `updated_by`,
  DROP COLUMN `created_by`,
  DROP COLUMN `workspace_id`;
DROP TABLE IF EXISTS `workspace_members`;
DROP TABLE IF EXISTS `workspaces`;
