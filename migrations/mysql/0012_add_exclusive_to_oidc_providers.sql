-- +goose Up
ALTER TABLE `oidc_providers` ADD COLUMN `exclusive` TINYINT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE `oidc_providers` DROP COLUMN `exclusive`;
