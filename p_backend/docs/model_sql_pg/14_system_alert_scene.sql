-- 系统告警场景配置表
DROP TABLE IF EXISTS system_alert_scene CASCADE;
CREATE TABLE IF NOT EXISTS system_alert_scene (
  id BIGSERIAL PRIMARY KEY,
  scene_key VARCHAR(20) NOT NULL,
  bot_id BIGINT NOT NULL DEFAULT 0,
  parse_mode VARCHAR(16) NOT NULL DEFAULT '',
  group_name VARCHAR(20) NOT NULL DEFAULT '',
  group_id VARCHAR(20) NOT NULL DEFAULT '',
  notify_template VARCHAR(1000) NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at BIGINT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_system_alert_scene_scene_key ON system_alert_scene (scene_key);
CREATE INDEX IF NOT EXISTS idx_system_alert_scene_bot_id ON system_alert_scene (bot_id);
CREATE INDEX IF NOT EXISTS idx_system_alert_scene_deleted_at ON system_alert_scene (deleted_at);

INSERT INTO system_alert_scene (id, scene_key, bot_id, group_name, group_id, notify_template) VALUES
  (1, 'server_down', 1, '运维群', 'ops_001', '服务器 {host} 已宕机，请立即处理'),
  (2, 'login_failed', 3, '安全群', '', '账号 {user} 连续登录失败 {count} 次')
ON CONFLICT (id) DO NOTHING;
SELECT setval(pg_get_serial_sequence('system_alert_scene', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM system_alert_scene), 1), true);
