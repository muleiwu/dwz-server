-- +goose Up
ALTER TABLE `click_statistics`
  ADD COLUMN `campaign_id` BIGINT UNSIGNED NULL AFTER `workspace_id`,
  ADD COLUMN `utm_source` VARCHAR(255) NULL AFTER `query_params`,
  ADD COLUMN `utm_medium` VARCHAR(255) NULL AFTER `utm_source`,
  ADD COLUMN `utm_campaign` VARCHAR(255) NULL AFTER `utm_medium`,
  ADD COLUMN `utm_term` VARCHAR(255) NULL AFTER `utm_campaign`,
  ADD COLUMN `utm_content` VARCHAR(255) NULL AFTER `utm_term`,
  ADD COLUMN `device_type` VARCHAR(50) NULL AFTER `utm_content`,
  ADD COLUMN `browser` VARCHAR(100) NULL AFTER `device_type`,
  ADD COLUMN `os` VARCHAR(100) NULL AFTER `browser`,
  ADD COLUMN `is_bot` TINYINT(1) NOT NULL DEFAULT 0 AFTER `os`,
  ADD COLUMN `bot_name` VARCHAR(100) NULL AFTER `is_bot`;

ALTER TABLE `ab_test_click_statistics`
  ADD COLUMN `campaign_id` BIGINT UNSIGNED NULL AFTER `workspace_id`,
  ADD COLUMN `utm_source` VARCHAR(255) NULL AFTER `query_params`,
  ADD COLUMN `utm_medium` VARCHAR(255) NULL AFTER `utm_source`,
  ADD COLUMN `utm_campaign` VARCHAR(255) NULL AFTER `utm_medium`,
  ADD COLUMN `utm_term` VARCHAR(255) NULL AFTER `utm_campaign`,
  ADD COLUMN `utm_content` VARCHAR(255) NULL AFTER `utm_term`,
  ADD COLUMN `device_type` VARCHAR(50) NULL AFTER `utm_content`,
  ADD COLUMN `browser` VARCHAR(100) NULL AFTER `device_type`,
  ADD COLUMN `os` VARCHAR(100) NULL AFTER `browser`,
  ADD COLUMN `is_bot` TINYINT(1) NOT NULL DEFAULT 0 AFTER `os`,
  ADD COLUMN `bot_name` VARCHAR(100) NULL AFTER `is_bot`;

CREATE INDEX `idx_click_statistics_campaign_id` ON `click_statistics` (`campaign_id`);
CREATE INDEX `idx_click_statistics_device_type` ON `click_statistics` (`device_type`);
CREATE INDEX `idx_click_statistics_is_bot` ON `click_statistics` (`is_bot`);
CREATE INDEX `idx_ab_test_click_statistics_campaign_id` ON `ab_test_click_statistics` (`campaign_id`);
CREATE INDEX `idx_ab_test_click_statistics_device_type` ON `ab_test_click_statistics` (`device_type`);
CREATE INDEX `idx_ab_test_click_statistics_is_bot` ON `ab_test_click_statistics` (`is_bot`);

-- +goose Down
DROP INDEX `idx_ab_test_click_statistics_is_bot` ON `ab_test_click_statistics`;
DROP INDEX `idx_ab_test_click_statistics_device_type` ON `ab_test_click_statistics`;
DROP INDEX `idx_ab_test_click_statistics_campaign_id` ON `ab_test_click_statistics`;
DROP INDEX `idx_click_statistics_is_bot` ON `click_statistics`;
DROP INDEX `idx_click_statistics_device_type` ON `click_statistics`;
DROP INDEX `idx_click_statistics_campaign_id` ON `click_statistics`;
ALTER TABLE `ab_test_click_statistics`
  DROP COLUMN `bot_name`,
  DROP COLUMN `is_bot`,
  DROP COLUMN `os`,
  DROP COLUMN `browser`,
  DROP COLUMN `device_type`,
  DROP COLUMN `utm_content`,
  DROP COLUMN `utm_term`,
  DROP COLUMN `utm_campaign`,
  DROP COLUMN `utm_medium`,
  DROP COLUMN `utm_source`,
  DROP COLUMN `campaign_id`;
ALTER TABLE `click_statistics`
  DROP COLUMN `bot_name`,
  DROP COLUMN `is_bot`,
  DROP COLUMN `os`,
  DROP COLUMN `browser`,
  DROP COLUMN `device_type`,
  DROP COLUMN `utm_content`,
  DROP COLUMN `utm_term`,
  DROP COLUMN `utm_campaign`,
  DROP COLUMN `utm_medium`,
  DROP COLUMN `utm_source`,
  DROP COLUMN `campaign_id`;
