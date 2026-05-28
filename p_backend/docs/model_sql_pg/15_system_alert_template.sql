-- 系统告警模板配置表
DROP TABLE IF EXISTS system_alert_template CASCADE;
CREATE TABLE IF NOT EXISTS system_alert_template (
  id BIGSERIAL PRIMARY KEY,
  bot_type VARCHAR(16) NOT NULL DEFAULT 'telegram',
  name VARCHAR(30) NOT NULL,
  parse_mode VARCHAR(16) NOT NULL DEFAULT '',
  content VARCHAR(1000) NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at BIGINT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_system_alert_template_bot_type ON system_alert_template (bot_type);
CREATE INDEX IF NOT EXISTS idx_system_alert_template_deleted_at ON system_alert_template (deleted_at);

INSERT INTO system_alert_template (id, bot_type, name, parse_mode, content) VALUES
  (1, 'telegram', '纯文本通知', '', '【{level}】{title}\n{message}\n时间：{time}'),
  (2, 'telegram', 'Markdown简单', 'Markdown', '*{title}*\n_{message}_\n`{time}`'),
  (3, 'telegram', 'MarkdownV2告警', 'MarkdownV2', '*【{level}】{title}*\n\n_{message}_\n\n`时间：{time}`\n`来源：{source}`'),
  (4, 'telegram', 'MarkdownV2详细', 'MarkdownV2', '*🚨 {title}*\n\n*级别：* `{level}`\n*来源：* `{source}`\n*详情：*\n_{message}_\n\n||影响范围：{scope}||\n\n>处理建议：{suggestion}'),
  (5, 'telegram', 'HTML简单', 'HTML', '<b>【{level}】{title}</b>\n<i>{message}</i>\n<code>{time}</code>'),
  (6, 'telegram', 'HTML详细', 'HTML', '<b>🚨 {title}</b>\n\n<b>级别：</b><code>{level}</code>\n<b>来源：</b><code>{source}</code>\n<b>详情：</b>\n<i>{message}</i>\n\n<pre>{detail}</pre>'),
  (7, 'feishu', '文本-服务异常', '', '⚠️ 服务异常通知\n\n服务 {service} 于 {time} 出现异常，当前状态：{status}\n请相关同学及时关注处理。'),
  (8, 'feishu', '文本-登录提醒', '', '🔐 安全提醒\n\n账号 {user} 于 {time} 在 {location} 登录失败（连续 {count} 次），请确认是否本人操作。'),
  (9, 'feishu', '富文本-部署通知', 'post', '{"zh_cn":{"title":"🚀 部署完成通知","content":[[{"tag":"text","text":"应用：{app}"}],[{"tag":"text","text":"环境：{env}"}],[{"tag":"text","text":"版本：{version}"}],[{"tag":"text","text":"部署人：{operator}"}],[{"tag":"text","text":"时间：{time}"}],[{"tag":"a","text":"📋 查看发布记录","href":"{url}"}]]}}'),
  (10, 'feishu', '卡片-服务告警', 'interactive', '{"header":{"title":{"content":"🚨 服务告警","tag":"plain_text"}},"elements":[{"tag":"div","text":{"content":"服务：{service}  级别：{level}  详情：{message}  时间：{time}","tag":"lark_md"}}]}'),
  (11, 'feishu', '卡片-任务完成', 'interactive', '{"header":{"title":{"content":"✅ 任务执行完成","tag":"plain_text"}},"elements":[{"tag":"div","text":{"content":"任务：{task}  耗时：{duration}  结果：{result}","tag":"lark_md"}}]}')
ON CONFLICT (id) DO NOTHING;
SELECT setval(pg_get_serial_sequence('system_alert_template', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM system_alert_template), 1), true);
