-- +goose Up
ALTER TABLE `click_statistics`
  ADD COLUMN `province` VARCHAR(100) NULL AFTER `country`,
  ADD COLUMN `isp` VARCHAR(100) NULL AFTER `city`;

ALTER TABLE `ab_test_click_statistics`
  ADD COLUMN `province` VARCHAR(100) NULL AFTER `country`,
  ADD COLUMN `isp` VARCHAR(100) NULL AFTER `city`;

-- +goose Down
ALTER TABLE `click_statistics`
  DROP COLUMN `isp`,
  DROP COLUMN `province`;

ALTER TABLE `ab_test_click_statistics`
  DROP COLUMN `isp`,
  DROP COLUMN `province`;
