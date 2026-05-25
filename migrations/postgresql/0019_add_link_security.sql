-- +goose Up
CREATE TABLE link_security_settings (
  id BIGSERIAL PRIMARY KEY,
  workspace_id BIGINT NOT NULL,
  short_link_id BIGINT NOT NULL,
  password_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  password_hash VARCHAR(255),
  access_window_start TIMESTAMPTZ,
  access_window_end TIMESTAMPTZ,
  max_clicks BIGINT,
  ip_policy VARCHAR(20) NOT NULL DEFAULT 'off',
  bot_policy VARCHAR(30) NOT NULL DEFAULT 'record_only',
  report_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  url_blocked BOOLEAN NOT NULL DEFAULT FALSE,
  url_blocked_reason VARCHAR(500),
  created_by BIGINT,
  updated_by BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX uk_link_security_short_link ON link_security_settings(short_link_id);
CREATE INDEX idx_link_security_workspace ON link_security_settings(workspace_id);
CREATE INDEX idx_link_security_created_by ON link_security_settings(created_by);
CREATE INDEX idx_link_security_deleted_at ON link_security_settings(deleted_at);

CREATE TABLE link_security_ip_rules (
  id BIGSERIAL PRIMARY KEY,
  workspace_id BIGINT NOT NULL,
  short_link_id BIGINT NOT NULL,
  cidr VARCHAR(64) NOT NULL,
  description VARCHAR(255),
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_link_security_ip_workspace ON link_security_ip_rules(workspace_id);
CREATE INDEX idx_link_security_ip_short_link ON link_security_ip_rules(short_link_id);
CREATE INDEX idx_link_security_ip_deleted_at ON link_security_ip_rules(deleted_at);

CREATE TABLE security_url_rules (
  id BIGSERIAL PRIMARY KEY,
  workspace_id BIGINT NOT NULL,
  rule_type VARCHAR(20) NOT NULL,
  action VARCHAR(20) NOT NULL,
  pattern VARCHAR(500) NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_by BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_security_url_rules_workspace ON security_url_rules(workspace_id);
CREATE INDEX idx_security_url_rules_type ON security_url_rules(rule_type);
CREATE INDEX idx_security_url_rules_action ON security_url_rules(action);
CREATE INDEX idx_security_url_rules_enabled ON security_url_rules(enabled);
CREATE INDEX idx_security_url_rules_created_by ON security_url_rules(created_by);
CREATE INDEX idx_security_url_rules_deleted_at ON security_url_rules(deleted_at);

CREATE TABLE abuse_reports (
  id BIGSERIAL PRIMARY KEY,
  workspace_id BIGINT NOT NULL,
  short_link_id BIGINT NOT NULL,
  report_type VARCHAR(30) NOT NULL,
  description VARCHAR(1000),
  reporter_email VARCHAR(255),
  reporter_ip VARCHAR(45),
  user_agent VARCHAR(1024),
  status VARCHAR(30) NOT NULL DEFAULT 'pending',
  resolution_note VARCHAR(1000),
  handled_by BIGINT,
  handled_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_abuse_reports_workspace ON abuse_reports(workspace_id);
CREATE INDEX idx_abuse_reports_short_link ON abuse_reports(short_link_id);
CREATE INDEX idx_abuse_reports_type ON abuse_reports(report_type);
CREATE INDEX idx_abuse_reports_status ON abuse_reports(status);
CREATE INDEX idx_abuse_reports_reporter_ip ON abuse_reports(reporter_ip);
CREATE INDEX idx_abuse_reports_handled_by ON abuse_reports(handled_by);

CREATE TABLE link_security_events (
  id BIGSERIAL PRIMARY KEY,
  workspace_id BIGINT NOT NULL,
  short_link_id BIGINT NOT NULL,
  event_type VARCHAR(50) NOT NULL,
  reason VARCHAR(500),
  client_ip VARCHAR(45),
  user_agent VARCHAR(1024),
  referer VARCHAR(2048),
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_link_security_events_workspace ON link_security_events(workspace_id);
CREATE INDEX idx_link_security_events_short_link ON link_security_events(short_link_id);
CREATE INDEX idx_link_security_events_type ON link_security_events(event_type);
CREATE INDEX idx_link_security_events_client_ip ON link_security_events(client_ip);

-- +goose Down
DROP TABLE IF EXISTS link_security_events;
DROP TABLE IF EXISTS abuse_reports;
DROP TABLE IF EXISTS security_url_rules;
DROP TABLE IF EXISTS link_security_ip_rules;
DROP TABLE IF EXISTS link_security_settings;
