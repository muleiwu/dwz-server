-- +goose Up
ALTER TABLE click_statistics
  ADD COLUMN province VARCHAR(100),
  ADD COLUMN isp VARCHAR(100);

ALTER TABLE ab_test_click_statistics
  ADD COLUMN province VARCHAR(100),
  ADD COLUMN isp VARCHAR(100);

-- +goose Down
ALTER TABLE click_statistics
  DROP COLUMN IF EXISTS isp,
  DROP COLUMN IF EXISTS province;

ALTER TABLE ab_test_click_statistics
  DROP COLUMN IF EXISTS isp,
  DROP COLUMN IF EXISTS province;
