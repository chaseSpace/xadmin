DROP TABLE IF EXISTS admin_user CASCADE;
CREATE TABLE IF NOT EXISTS admin_user (
  id BIGSERIAL PRIMARY KEY,
  uid INTEGER NOT NULL,
  username VARCHAR(64) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  display_name VARCHAR(64) NOT NULL DEFAULT '',
  avatar VARCHAR(500) NOT NULL DEFAULT '',
  email VARCHAR(128) NOT NULL DEFAULT '',
  phone VARCHAR(32) NOT NULL DEFAULT '',
  status SMALLINT NOT NULL DEFAULT 1,
  department_id BIGINT NOT NULL DEFAULT 0,
  position_id BIGINT NOT NULL DEFAULT 0,
  limit_single_login BOOLEAN NOT NULL DEFAULT FALSE,
  deactivated_at TIMESTAMP NULL,
  last_login_at TIMESTAMP NULL,
  last_login_ip VARCHAR(64) NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at BIGINT NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_admin_user_uid ON admin_user (uid);
CREATE UNIQUE INDEX IF NOT EXISTS uk_admin_user_username_deleted_at ON admin_user (username, deleted_at);
CREATE INDEX IF NOT EXISTS idx_admin_user_status ON admin_user (status);
CREATE INDEX IF NOT EXISTS idx_admin_user_department_id ON admin_user (department_id);
CREATE INDEX IF NOT EXISTS idx_admin_user_position_id ON admin_user (position_id);
CREATE INDEX IF NOT EXISTS idx_admin_user_deactivated_at ON admin_user (deactivated_at);
CREATE INDEX IF NOT EXISTS idx_admin_user_email ON admin_user (email);
CREATE INDEX IF NOT EXISTS idx_admin_user_phone ON admin_user (phone);
CREATE INDEX IF NOT EXISTS idx_admin_user_deleted_at ON admin_user (deleted_at);

INSERT INTO admin_user (uid, username, password_hash, display_name, avatar, email, phone, status, department_id, position_id,
                        limit_single_login, last_login_at, last_login_ip)
SELECT 10001, 'admin', '$2a$10$E7izPmKoa8FVP4fkae4QjOZ5SEs4Tg4LBf2FVbFb8GqTYZC0lEGim', 'Luso',
       'https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcTQuYOZEAwUQl5Q-LVpeodw4iBjhWa4jidhvSUY0KQ3_iQqWrjn2s95HnGQjLWt1UWO8fLH&s', '', '', 1, d.id, p.id, FALSE, NULL, ''
FROM organization_department d
JOIN organization_position p ON p.department_id = d.id AND p.code = 'POS-CEO' AND p.deleted_at = 0
WHERE d.code = 'HQ' AND d.deleted_at = 0
LIMIT 1
ON CONFLICT (uid) DO UPDATE SET
  password_hash = EXCLUDED.password_hash,
  avatar = EXCLUDED.avatar,
  display_name = EXCLUDED.display_name,
  department_id = EXCLUDED.department_id,
  position_id = EXCLUDED.position_id,
  updated_at = CURRENT_TIMESTAMP;
