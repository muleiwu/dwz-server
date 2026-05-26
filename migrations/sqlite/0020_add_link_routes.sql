-- +goose Up
ALTER TABLE short_links ADD COLUMN fallback_url TEXT;
ALTER TABLE short_links ADD COLUMN redirect_code INTEGER NOT NULL DEFAULT 302;

ALTER TABLE click_statistics ADD COLUMN route_id INTEGER;
ALTER TABLE click_statistics ADD COLUMN route_name TEXT;
CREATE INDEX idx_click_statistics_route_id ON click_statistics(route_id);

CREATE TABLE link_routes (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  workspace_id INTEGER NOT NULL,
  short_link_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  description TEXT,
  priority INTEGER NOT NULL DEFAULT 100,
  target_url TEXT NOT NULL,
  is_active INTEGER NOT NULL DEFAULT 1,
  created_by INTEGER,
  updated_by INTEGER,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);
CREATE INDEX idx_link_routes_workspace ON link_routes(workspace_id);
CREATE INDEX idx_link_routes_short_link ON link_routes(short_link_id);
CREATE INDEX idx_link_routes_priority ON link_routes(priority);
CREATE INDEX idx_link_routes_active ON link_routes(is_active);
CREATE INDEX idx_link_routes_created_by ON link_routes(created_by);
CREATE INDEX idx_link_routes_deleted_at ON link_routes(deleted_at);

CREATE TABLE link_route_condition_groups (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  route_id INTEGER NOT NULL,
  position INTEGER NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_route_condition_groups_route ON link_route_condition_groups(route_id);

CREATE TABLE link_route_conditions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  group_id INTEGER NOT NULL,
  condition_type TEXT NOT NULL,
  operator TEXT NOT NULL,
  condition_key TEXT,
  condition_value TEXT,
  position INTEGER NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_route_conditions_group ON link_route_conditions(group_id);
CREATE INDEX idx_route_conditions_type ON link_route_conditions(condition_type);

-- +goose Down
DROP TABLE IF EXISTS link_route_conditions;
DROP TABLE IF EXISTS link_route_condition_groups;
DROP TABLE IF EXISTS link_routes;
DROP INDEX IF EXISTS idx_click_statistics_route_id;
ALTER TABLE click_statistics DROP COLUMN route_name;
ALTER TABLE click_statistics DROP COLUMN route_id;
ALTER TABLE short_links DROP COLUMN redirect_code;
ALTER TABLE short_links DROP COLUMN fallback_url;
