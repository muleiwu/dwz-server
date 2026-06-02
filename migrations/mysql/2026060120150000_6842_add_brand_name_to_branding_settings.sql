-- +goose Up
ALTER TABLE `ee_system_brandings`
  ADD COLUMN `brand_name` VARCHAR(80) NOT NULL DEFAULT '' AFTER `id`;

ALTER TABLE `ee_workspace_brandings`
  ADD COLUMN `brand_name` VARCHAR(80) NOT NULL DEFAULT '' AFTER `workspace_id`;

ALTER TABLE `ee_domain_brandings`
  ADD COLUMN `override_brand_name` BOOLEAN NOT NULL DEFAULT FALSE AFTER `workspace_id`,
  ADD COLUMN `brand_name` VARCHAR(80) NOT NULL DEFAULT '' AFTER `override_brand_name`;

-- +goose Down
ALTER TABLE `ee_domain_brandings`
  DROP COLUMN `brand_name`,
  DROP COLUMN `override_brand_name`;

ALTER TABLE `ee_workspace_brandings`
  DROP COLUMN `brand_name`;

ALTER TABLE `ee_system_brandings`
  DROP COLUMN `brand_name`;
