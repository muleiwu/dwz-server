-- +goose Up
ALTER TABLE `users` ADD COLUMN `is_system_admin` BOOLEAN NOT NULL DEFAULT FALSE AFTER `status`;

UPDATE `users`
SET `is_system_admin` = TRUE
WHERE `id` = (
  SELECT `id` FROM (
    SELECT MIN(`id`) AS `id`
    FROM `users`
    WHERE `deleted_at` IS NULL AND `status` = 1
  ) AS first_user
)
AND NOT EXISTS (
  SELECT 1 FROM (
    SELECT `id`
    FROM `users`
    WHERE `is_system_admin` = TRUE AND `status` = 1
    LIMIT 1
  ) AS system_admins
);

-- +goose Down
ALTER TABLE `users` DROP COLUMN `is_system_admin`;
