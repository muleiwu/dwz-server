-- +goose Up
ALTER TABLE ee_system_brandings ADD COLUMN brand_name TEXT NOT NULL DEFAULT '';
ALTER TABLE ee_workspace_brandings ADD COLUMN brand_name TEXT NOT NULL DEFAULT '';
ALTER TABLE ee_domain_brandings ADD COLUMN override_brand_name BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE ee_domain_brandings ADD COLUMN brand_name TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE ee_domain_brandings DROP COLUMN brand_name;
ALTER TABLE ee_domain_brandings DROP COLUMN override_brand_name;
ALTER TABLE ee_workspace_brandings DROP COLUMN brand_name;
ALTER TABLE ee_system_brandings DROP COLUMN brand_name;
