-- +goose Up
ALTER TABLE users ADD COLUMN is_system_admin BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE users
SET is_system_admin = TRUE
WHERE id = (
  SELECT MIN(id)
  FROM users
  WHERE deleted_at IS NULL AND status = 1
)
AND NOT EXISTS (
  SELECT 1
  FROM users
  WHERE is_system_admin = TRUE AND status = 1
);

-- +goose Down
ALTER TABLE users DROP COLUMN is_system_admin;
