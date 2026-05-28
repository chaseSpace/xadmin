DROP TABLE IF EXISTS admin_user_personal_setting CASCADE;
CREATE TABLE IF NOT EXISTS admin_user_personal_setting (
  id BIGSERIAL PRIMARY KEY,
  uid INTEGER NOT NULL,
  limit_single_login BOOLEAN NOT NULL DEFAULT FALSE,
  background_image_url VARCHAR(500) NOT NULL DEFAULT '',
  locale VARCHAR(32) NOT NULL DEFAULT 'zh-CN',
  global_background_apply_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  warm_tip_interval_minutes INTEGER NOT NULL DEFAULT 1440,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_admin_user_personal_setting_uid ON admin_user_personal_setting (uid);
