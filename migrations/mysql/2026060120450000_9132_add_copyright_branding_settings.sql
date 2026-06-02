-- +goose Up
ALTER TABLE `ee_system_brandings`
  ADD COLUMN `copyright_text` VARCHAR(200) NOT NULL DEFAULT '' AFTER `copyright_enabled`,
  ADD COLUMN `copyright_link` VARCHAR(500) NOT NULL DEFAULT '' AFTER `copyright_text`;

ALTER TABLE `ee_workspace_brandings`
  ADD COLUMN `copyright_text` VARCHAR(200) NOT NULL DEFAULT '' AFTER `copyright_enabled`,
  ADD COLUMN `copyright_link` VARCHAR(500) NOT NULL DEFAULT '' AFTER `copyright_text`;

ALTER TABLE `ee_domain_brandings`
  ADD COLUMN `override_copyright_text` BOOLEAN NOT NULL DEFAULT FALSE AFTER `copyright_enabled`,
  ADD COLUMN `copyright_text` VARCHAR(200) NOT NULL DEFAULT '' AFTER `override_copyright_text`,
  ADD COLUMN `copyright_link` VARCHAR(500) NOT NULL DEFAULT '' AFTER `copyright_text`;

-- +goose Down
ALTER TABLE `ee_domain_brandings`
  DROP COLUMN `copyright_link`,
  DROP COLUMN `copyright_text`,
  DROP COLUMN `override_copyright_text`;

ALTER TABLE `ee_workspace_brandings`
  DROP COLUMN `copyright_link`,
  DROP COLUMN `copyright_text`;

ALTER TABLE `ee_system_brandings`
  DROP COLUMN `copyright_link`,
  DROP COLUMN `copyright_text`;
