-- 系统告警机器人配置表
DROP TABLE IF EXISTS system_alert_bot CASCADE;
CREATE TABLE IF NOT EXISTS system_alert_bot (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(20) NOT NULL,
  username VARCHAR(30) NOT NULL DEFAULT '',
  token VARCHAR(100) NOT NULL,
  bot_type VARCHAR(16) NOT NULL DEFAULT 'telegram',
  enabled BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at BIGINT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_system_alert_bot_bot_type ON system_alert_bot (bot_type);
CREATE INDEX IF NOT EXISTS idx_system_alert_bot_enabled ON system_alert_bot (enabled);
CREATE INDEX IF NOT EXISTS idx_system_alert_bot_deleted_at ON system_alert_bot (deleted_at);

INSERT INTO system_alert_bot (id, name, username, token, bot_type, enabled) VALUES
  (1, '运维告警Bot', 'ops_alert_bot', 'demo_token_placeholder', 'telegram', true),
  (2, '安全告警Bot', 'sec_alert_bot', 'demo_token_placeholder_2', 'telegram', true),
  (3, '飞书通知Bot', '', 'demo_feishu_webhook_token', 'feishu', true)
ON CONFLICT (id) DO NOTHING;
SELECT setval(pg_get_serial_sequence('system_alert_bot', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM system_alert_bot), 1), true);
