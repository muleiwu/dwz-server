-- +goose Up
ALTER TABLE short_links ADD COLUMN fallback_url VARCHAR(2000);
ALTER TABLE short_links ADD COLUMN redirect_code INT NOT NULL DEFAULT 302;

ALTER TABLE click_statistics ADD COLUMN route_id BIGINT;
ALTER TABLE click_statistics ADD COLUMN route_name VARCHAR(100);
CREATE INDEX idx_click_statistics_route_id ON click_statistics(route_id);

CREATE TABLE link_routes (
  id BIGSERIAL PRIMARY KEY,
  workspace_id BIGINT NOT NULL,
  short_link_id BIGINT NOT NULL,
  name VARCHAR(100) NOT NULL,
  description VARCHAR(500),
  priority INT NOT NULL DEFAULT 100,
  target_url VARCHAR(2000) NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_by BIGINT,
  updated_by BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_link_routes_workspace ON link_routes(workspace_id);
CREATE INDEX idx_link_routes_short_link ON link_routes(short_link_id);
CREATE INDEX idx_link_routes_priority ON link_routes(priority);
CREATE INDEX idx_link_routes_active ON link_routes(is_active);
CREATE INDEX idx_link_routes_created_by ON link_routes(created_by);
CREATE INDEX idx_link_routes_deleted_at ON link_routes(deleted_at);

CREATE TABLE link_route_condition_groups (
  id BIGSERIAL PRIMARY KEY,
  route_id BIGINT NOT NULL,
  position INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_route_condition_groups_route ON link_route_condition_groups(route_id);

CREATE TABLE link_route_conditions (
  id BIGSERIAL PRIMARY KEY,
  group_id BIGINT NOT NULL,
  condition_type VARCHAR(30) NOT NULL,
  operator VARCHAR(20) NOT NULL,
  condition_key VARCHAR(255),
  condition_value VARCHAR(1000),
  position INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
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
