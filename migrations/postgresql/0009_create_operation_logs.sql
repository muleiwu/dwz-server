-- +goose Up
CREATE TABLE operation_logs (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT,
  username VARCHAR(50),
  operation VARCHAR(100) NOT NULL,
  resource VARCHAR(100),
  resource_id VARCHAR(100),
  method VARCHAR(10),
  path VARCHAR(255),
  request_body TEXT,
  response_code INTEGER NOT NULL DEFAULT 0,
  response_body TEXT,
  ip VARCHAR(45),
  user_agent VARCHAR(500),
  execute_time BIGINT NOT NULL DEFAULT 0,
  status SMALLINT NOT NULL DEFAULT 1,
  error_message VARCHAR(1000),
  created_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_operation_logs_user_id ON operation_logs(user_id);

-- +goose Down
DROP TABLE IF EXISTS operation_logs;
