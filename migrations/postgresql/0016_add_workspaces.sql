-- +goose Up
CREATE TABLE workspaces (
  id BIGSERIAL PRIMARY KEY,
  slug VARCHAR(100) NOT NULL,
  name VARCHAR(100) NOT NULL,
  description VARCHAR(500),
  owner_user_id BIGINT,
  status SMALLINT NOT NULL DEFAULT 1,
  created_at TIMESTAMP WITH TIME ZONE,
  updated_at TIMESTAMP WITH TIME ZONE,
  deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE UNIQUE INDEX uk_workspaces_slug ON workspaces(slug);
CREATE INDEX idx_workspaces_deleted_at ON workspaces(deleted_at);

CREATE TABLE workspace_members (
  id BIGSERIAL PRIMARY KEY,
  workspace_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  role VARCHAR(20) NOT NULL,
  status SMALLINT NOT NULL DEFAULT 1,
  created_at TIMESTAMP WITH TIME ZONE,
  updated_at TIMESTAMP WITH TIME ZONE,
  deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE UNIQUE INDEX uk_workspace_members_workspace_user ON workspace_members(workspace_id, user_id);
CREATE INDEX idx_workspace_members_user_id ON workspace_members(user_id);
CREATE INDEX idx_workspace_members_deleted_at ON workspace_members(deleted_at);

ALTER TABLE short_links ADD COLUMN workspace_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE short_links ADD COLUMN created_by BIGINT;
ALTER TABLE short_links ADD COLUMN updated_by BIGINT;
ALTER TABLE domains ADD COLUMN workspace_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE ab_tests ADD COLUMN workspace_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE user_tokens ADD COLUMN workspace_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE operation_logs ADD COLUMN workspace_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE click_statistics ADD COLUMN workspace_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE ab_test_click_statistics ADD COLUMN workspace_id BIGINT NOT NULL DEFAULT 1;

CREATE INDEX idx_short_links_workspace_id ON short_links(workspace_id);
CREATE INDEX idx_short_links_created_by ON short_links(created_by);
CREATE INDEX idx_domains_workspace_id ON domains(workspace_id);
CREATE INDEX idx_ab_tests_workspace_id ON ab_tests(workspace_id);
CREATE INDEX idx_user_tokens_workspace_id ON user_tokens(workspace_id);
CREATE INDEX idx_operation_logs_workspace_id ON operation_logs(workspace_id);
CREATE INDEX idx_click_statistics_workspace_id ON click_statistics(workspace_id);
CREATE INDEX idx_ab_test_click_statistics_workspace_id ON ab_test_click_statistics(workspace_id);

INSERT INTO workspaces (id, slug, name, owner_user_id, status, created_at, updated_at)
SELECT 1, 'default', '默认工作区', (SELECT MIN(id) FROM users WHERE deleted_at IS NULL), 1, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM workspaces WHERE id = 1);

SELECT setval(pg_get_serial_sequence('workspaces', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM workspaces), 1), true);

INSERT INTO workspace_members (workspace_id, user_id, role, status, created_at, updated_at)
SELECT 1, id, CASE WHEN id = (SELECT MIN(id) FROM users WHERE deleted_at IS NULL) THEN 'owner' ELSE 'admin' END, 1, NOW(), NOW()
FROM users
WHERE deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM workspace_members wm
    WHERE wm.workspace_id = 1 AND wm.user_id = users.id
  );

-- +goose Down
DELETE FROM workspace_members WHERE workspace_id = 1;
DELETE FROM workspaces WHERE id = 1;
DROP INDEX IF EXISTS idx_ab_test_click_statistics_workspace_id;
DROP INDEX IF EXISTS idx_click_statistics_workspace_id;
DROP INDEX IF EXISTS idx_operation_logs_workspace_id;
DROP INDEX IF EXISTS idx_user_tokens_workspace_id;
DROP INDEX IF EXISTS idx_ab_tests_workspace_id;
DROP INDEX IF EXISTS idx_domains_workspace_id;
DROP INDEX IF EXISTS idx_short_links_created_by;
DROP INDEX IF EXISTS idx_short_links_workspace_id;
ALTER TABLE ab_test_click_statistics DROP COLUMN workspace_id;
ALTER TABLE click_statistics DROP COLUMN workspace_id;
ALTER TABLE operation_logs DROP COLUMN workspace_id;
ALTER TABLE user_tokens DROP COLUMN workspace_id;
ALTER TABLE ab_tests DROP COLUMN workspace_id;
ALTER TABLE domains DROP COLUMN workspace_id;
ALTER TABLE short_links DROP COLUMN updated_by;
ALTER TABLE short_links DROP COLUMN created_by;
ALTER TABLE short_links DROP COLUMN workspace_id;
DROP TABLE IF EXISTS workspace_members;
DROP TABLE IF EXISTS workspaces;
