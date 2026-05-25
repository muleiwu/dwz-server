-- +goose Up
CREATE TABLE link_security_settings (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  workspace_id INTEGER NOT NULL,
  short_link_id INTEGER NOT NULL,
  password_enabled INTEGER NOT NULL DEFAULT 0,
  password_hash TEXT,
  access_window_start DATETIME,
  access_window_end DATETIME,
  max_clicks INTEGER,
  ip_policy TEXT NOT NULL DEFAULT 'off',
  bot_policy TEXT NOT NULL DEFAULT 'record_only',
  report_enabled INTEGER NOT NULL DEFAULT 0,
  url_blocked INTEGER NOT NULL DEFAULT 0,
  url_blocked_reason TEXT,
  created_by INTEGER,
  updated_by INTEGER,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);
CREATE UNIQUE INDEX uk_link_security_short_link ON link_security_settings(short_link_id);
CREATE INDEX idx_link_security_workspace ON link_security_settings(workspace_id);
CREATE INDEX idx_link_security_created_by ON link_security_settings(created_by);
CREATE INDEX idx_link_security_deleted_at ON link_security_settings(deleted_at);

CREATE TABLE link_security_ip_rules (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  workspace_id INTEGER NOT NULL,
  short_link_id INTEGER NOT NULL,
  cidr TEXT NOT NULL,
  description TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);
CREATE INDEX idx_link_security_ip_workspace ON link_security_ip_rules(workspace_id);
CREATE INDEX idx_link_security_ip_short_link ON link_security_ip_rules(short_link_id);
CREATE INDEX idx_link_security_ip_deleted_at ON link_security_ip_rules(deleted_at);

CREATE TABLE security_url_rules (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  workspace_id INTEGER NOT NULL,
  rule_type TEXT NOT NULL,
  action TEXT NOT NULL,
  pattern TEXT NOT NULL,
  enabled INTEGER NOT NULL DEFAULT 1,
  created_by INTEGER,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);
CREATE INDEX idx_security_url_rules_workspace ON security_url_rules(workspace_id);
CREATE INDEX idx_security_url_rules_type ON security_url_rules(rule_type);
CREATE INDEX idx_security_url_rules_action ON security_url_rules(action);
CREATE INDEX idx_security_url_rules_enabled ON security_url_rules(enabled);
CREATE INDEX idx_security_url_rules_created_by ON security_url_rules(created_by);
CREATE INDEX idx_security_url_rules_deleted_at ON security_url_rules(deleted_at);

CREATE TABLE abuse_reports (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  workspace_id INTEGER NOT NULL,
  short_link_id INTEGER NOT NULL,
  report_type TEXT NOT NULL,
  description TEXT,
  reporter_email TEXT,
  reporter_ip TEXT,
  user_agent TEXT,
  status TEXT NOT NULL DEFAULT 'pending',
  resolution_note TEXT,
  handled_by INTEGER,
  handled_at DATETIME,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_abuse_reports_workspace ON abuse_reports(workspace_id);
CREATE INDEX idx_abuse_reports_short_link ON abuse_reports(short_link_id);
CREATE INDEX idx_abuse_reports_type ON abuse_reports(report_type);
CREATE INDEX idx_abuse_reports_status ON abuse_reports(status);
CREATE INDEX idx_abuse_reports_reporter_ip ON abuse_reports(reporter_ip);
CREATE INDEX idx_abuse_reports_handled_by ON abuse_reports(handled_by);

CREATE TABLE link_security_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  workspace_id INTEGER NOT NULL,
  short_link_id INTEGER NOT NULL,
  event_type TEXT NOT NULL,
  reason TEXT,
  client_ip TEXT,
  user_agent TEXT,
  referer TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
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
