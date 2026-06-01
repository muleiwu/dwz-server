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

CREATE TABLE IF NOT EXISTS ee_system_brandings (
  id SMALLINT NOT NULL,
  logo_url VARCHAR(500) NOT NULL DEFAULT '',
  copyright_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  PRIMARY KEY (id)
);

-- +goose Down
DROP TABLE IF EXISTS ee_system_brandings;
ALTER TABLE users DROP COLUMN is_system_admin;
