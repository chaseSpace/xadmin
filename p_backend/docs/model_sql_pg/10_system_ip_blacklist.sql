DROP TABLE IF EXISTS system_ip_blacklist CASCADE;
CREATE TABLE IF NOT EXISTS system_ip_blacklist (
  id BIGSERIAL PRIMARY KEY,
  ip VARCHAR(64) NOT NULL,
  ban_type VARCHAR(16) NOT NULL,
  start_at TIMESTAMP NOT NULL,
  end_at TIMESTAMP NOT NULL,
  reason VARCHAR(255) NOT NULL,
  creator VARCHAR(64) NOT NULL DEFAULT '',
  status VARCHAR(16) NOT NULL DEFAULT 'active',
  hit_count INTEGER NOT NULL DEFAULT 0,
  last_action VARCHAR(64) NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at BIGINT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_system_ip_blacklist_ip_deleted_at ON system_ip_blacklist (ip, deleted_at);
CREATE INDEX IF NOT EXISTS idx_system_ip_blacklist_status ON system_ip_blacklist (status);
CREATE INDEX IF NOT EXISTS idx_system_ip_blacklist_ban_type ON system_ip_blacklist (ban_type);
CREATE INDEX IF NOT EXISTS idx_system_ip_blacklist_creator ON system_ip_blacklist (creator);
CREATE INDEX IF NOT EXISTS idx_system_ip_blacklist_end_at ON system_ip_blacklist (end_at);
CREATE INDEX IF NOT EXISTS idx_system_ip_blacklist_updated_at ON system_ip_blacklist (updated_at);
CREATE INDEX IF NOT EXISTS idx_system_ip_blacklist_deleted_at ON system_ip_blacklist (deleted_at);

INSERT INTO system_ip_blacklist (id, ip, ban_type, start_at, end_at, reason, creator, status, hit_count, last_action) VALUES
  (1, '120.12.8.91', 'temp', '2026-04-26 08:00:00', '2026-04-29 08:00:00', '短时高频请求', 'sec_admin', 'active', 42, 'view_audit_logs'),
  (2, '10.24.33.17', 'permanent', '2026-03-01 00:00:00', '2099-12-31 00:00:00', '已确认攻击源', 'sec_admin', 'active', 118, 'auth_failed'),
  (3, '223.66.19.42', 'temp', '2026-04-01 10:00:00', '2026-04-03 10:00:00', '登录异常', 'risk_admin', 'inactive', 17, 'expired'),
  (4, '185.77.3.214', 'temp', '2026-04-28 04:20:00', '2026-04-28 22:20:00', '疑似撞库行为', 'risk_admin', 'active', 64, 'login_failed'),
  (5, '103.45.90.8', 'permanent', '2026-04-18 00:10:00', '2099-12-31 00:00:00', '恶意扫描', 'sec_admin', 'inactive', 93, 'manual_unblock')
ON CONFLICT (id) DO NOTHING;
SELECT setval(pg_get_serial_sequence('system_ip_blacklist', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM system_ip_blacklist), 1), true);
