DROP TABLE IF EXISTS admin_user_login_audit CASCADE;
DROP TABLE IF EXISTS admin_user_session CASCADE;

CREATE TABLE IF NOT EXISTS admin_user_session (
  id BIGSERIAL PRIMARY KEY,
  session_id VARCHAR(64) NOT NULL,
  uid INTEGER NOT NULL,
  token_hash CHAR(64) NOT NULL,
  status VARCHAR(16) NOT NULL DEFAULT 'active',
  login_ip VARCHAR(64) NOT NULL DEFAULT '',
  user_agent VARCHAR(255) NOT NULL DEFAULT '',
  last_seen_at TIMESTAMP NULL,
  expired_at TIMESTAMP NOT NULL,
  revoked_at TIMESTAMP NULL,
  revoked_reason VARCHAR(64) NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_admin_user_session_session_id ON admin_user_session (session_id);
CREATE UNIQUE INDEX IF NOT EXISTS uk_admin_user_session_token_hash ON admin_user_session (token_hash);
CREATE INDEX IF NOT EXISTS idx_admin_user_session_uid_status ON admin_user_session (uid, status);
CREATE INDEX IF NOT EXISTS idx_admin_user_session_expired_at ON admin_user_session (expired_at);
CREATE INDEX IF NOT EXISTS idx_admin_user_session_created_at ON admin_user_session (created_at);

CREATE TABLE IF NOT EXISTS admin_user_login_audit (
  id BIGSERIAL PRIMARY KEY,
  uid INTEGER NOT NULL DEFAULT 0,
  action VARCHAR(32) NOT NULL,
  result VARCHAR(16) NOT NULL DEFAULT 'success',
  trace_id VARCHAR(64) NOT NULL DEFAULT '',
  request_id VARCHAR(64) NOT NULL DEFAULT '',
  source_ip VARCHAR(64) NOT NULL DEFAULT '',
  duration VARCHAR(32) NOT NULL DEFAULT '',
  user_agent VARCHAR(255) NOT NULL DEFAULT '',
  detail VARCHAR(255) NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_admin_user_login_audit_uid_action_created_at ON admin_user_login_audit (uid, action, created_at);
CREATE INDEX IF NOT EXISTS idx_admin_user_login_audit_action_created_at ON admin_user_login_audit (action, created_at);
CREATE INDEX IF NOT EXISTS idx_admin_user_login_audit_result_created_at ON admin_user_login_audit (result, created_at);
CREATE INDEX IF NOT EXISTS idx_admin_user_login_audit_trace_id ON admin_user_login_audit (trace_id);
CREATE INDEX IF NOT EXISTS idx_admin_user_login_audit_source_ip_created_at ON admin_user_login_audit (source_ip, created_at);
