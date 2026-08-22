DROP TABLE IF EXISTS permission_role_user CASCADE;
DROP TABLE IF EXISTS permission_role_menu CASCADE;
DROP TABLE IF EXISTS permission_role CASCADE;
DROP TABLE IF EXISTS permission_menu CASCADE;

CREATE TABLE IF NOT EXISTS permission_menu (
  id BIGSERIAL PRIMARY KEY,
  parent_id BIGINT NOT NULL DEFAULT 0,
  name VARCHAR(64) NOT NULL,
  route_path VARCHAR(255) NOT NULL DEFAULT '',
  component_path VARCHAR(255) NOT NULL DEFAULT '',
  menu_type SMALLINT NOT NULL,
  permission_key VARCHAR(128) NOT NULL DEFAULT '',
  is_delegable BOOLEAN NOT NULL DEFAULT TRUE,
  sort INTEGER NOT NULL DEFAULT 0,
  status SMALLINT NOT NULL DEFAULT 1,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at BIGINT NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_permission_menu_permission_key_deleted_at ON permission_menu (permission_key, deleted_at);
CREATE INDEX IF NOT EXISTS idx_permission_menu_parent_id ON permission_menu (parent_id);
CREATE INDEX IF NOT EXISTS idx_permission_menu_status ON permission_menu (status);
CREATE INDEX IF NOT EXISTS idx_permission_menu_deleted_at ON permission_menu (deleted_at);

CREATE TABLE IF NOT EXISTS permission_role (
  id BIGSERIAL PRIMARY KEY,
  role_name VARCHAR(64) NOT NULL,
  role_code VARCHAR(64) NOT NULL DEFAULT '',
  role_type SMALLINT NOT NULL,
  is_protected BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at BIGINT NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_permission_role_role_name_deleted_at ON permission_role (role_name, deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_permission_role_active_role_code ON permission_role (role_code) WHERE deleted_at = 0 AND role_code <> '';
CREATE INDEX IF NOT EXISTS idx_permission_role_deleted_at ON permission_role (deleted_at);

CREATE TABLE IF NOT EXISTS permission_role_menu (
  id BIGSERIAL PRIMARY KEY,
  role_id BIGINT NOT NULL,
  menu_id BIGINT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_permission_role_menu ON permission_role_menu (role_id, menu_id);
CREATE INDEX IF NOT EXISTS idx_permission_role_menu_role_id ON permission_role_menu (role_id);
CREATE INDEX IF NOT EXISTS idx_permission_role_menu_menu_id ON permission_role_menu (menu_id);

TRUNCATE permission_menu;
WITH permission_menu_seed (id, parent_id, name, route_path, component_path, menu_type, permission_key, sort, status) AS (
VALUES
(1, 0, '组织管理', '', '', 1, 'organization.root', 10, 1),
(2, 1, '部门管理', '/organization/departments', 'pages/organization/OrganizationStructurePage', 2, 'organization.departments.view', 20, 1),
(3, 1, '岗位管理', '/organization/positions', 'pages/organization/OrganizationPositionsPage', 2, 'organization.positions.view', 30, 1),
(4, 1, '用户列表', '/organization/users', 'pages/organization/OrganizationMembersPage', 2, 'organization.users.view', 40, 1),
(5, 0, '业务管理', '', '', 1, 'business.root', 50, 1),
(6, 5, '用户列表（demo）', '/business/users', 'pages/business/BusinessUsersPage', 2, 'business.users.view', 60, 1),
(7, 5, '用户惩罚（demo）', '/business/user-punishments', 'pages/business/BusinessUserPunishmentsPage', 2, 'business.user_punishments.view', 70, 1),
(15, 0, '资源管理', '', '', 1, 'resource.root', 80, 1),
(16, 15, '文件管理', '/resource/files', 'pages/resource/ResourceFilesPage', 2, 'resource.files.view', 90, 1),
(8, 0, '权限管理', '', '', 1, 'permission.root', 100, 1),
(9, 8, '角色权限', '/permission/role-permissions', 'pages/permission/PermissionRolesPage', 2, 'permission.roles.view', 110, 1),
(10, 8, '菜单权限', '/permission/menu-permissions', 'pages/permission/PermissionPoliciesPage', 2, 'permission.menus.view', 120, 1),
(11, 0, '系统管理', '', '', 1, 'system.root', 130, 1),
(12, 11, '系统设置', '/system/settings', 'pages/system/SystemSettingsPage', 2, 'system.settings.view', 140, 1),
(13, 11, '操作审计', '/system/audit-logs', 'pages/system/SystemAuditLogsPage', 2, 'system.audit_logs.view', 150, 1),
(14, 11, 'IP黑名单', '/system/ip-blacklist', 'pages/system/SystemIpBlacklistPage', 2, 'system.ip_blacklist.view', 160, 1),
(17, 11, '关怀提示', '/system/warm-tips', 'pages/system/SystemWarmTipsPage', 2, 'system.warm_tips.view', 170, 1),
(18, 11, '告警通知', '/system/alert-bots', 'pages/system/SystemAlertBotsPage', 2, 'system.alert_bots.view', 180, 1),
-- 操作权限（menu_type=3，不显示为菜单）
(19, 2, '部门管理-编辑', '', '', 3, 'organization.departments.edit', 21, 1),
(24, 14, 'IP黑名单-编辑', '', '', 3, 'system.ip_blacklist.edit', 161, 1),
(25, 17, '关怀提示-编辑', '', '', 3, 'system.warm_tips.edit', 171, 1),
(26, 18, '机器人配置-编辑', '', '', 3, 'system.alert_bots.edit', 181, 1),
(27, 12, '系统设置-编辑', '', '', 3, 'system.settings.edit', 141, 1),
(28, 2, '部门管理-删除', '', '', 3, 'organization.departments.delete', 22, 1),
(29, 3, '岗位管理-删除', '', '', 3, 'organization.positions.delete', 32, 1),
(30, 4, '用户列表-删除', '', '', 3, 'organization.users.delete', 42, 1),
(31, 9, '角色权限-删除', '', '', 3, 'permission.roles.delete', 112, 1),
(33, 14, 'IP黑名单-删除', '', '', 3, 'system.ip_blacklist.delete', 162, 1),
(34, 17, '关怀提示-删除', '', '', 3, 'system.warm_tips.delete', 172, 1),
(35, 18, '机器人配置-删除', '', '', 3, 'system.alert_bots.delete', 182, 1),
(36, 15, '文件管理-上传', '', '', 3, 'resource.files.upload', 91, 1),
(37, 15, '文件管理-删除', '', '', 3, 'resource.files.delete', 92, 1),
(38, 3, '岗位管理-分配角色', '', '', 3, 'organization.positions.assign_roles', 33, 1),
(39, 4, '用户列表-编辑资料', '', '', 3, 'organization.users.edit_profile', 43, 1),
(40, 4, '用户列表-调岗', '', '', 3, 'organization.users.assign_position', 44, 1),
(41, 4, '用户列表-修改状态', '', '', 3, 'organization.users.change_status', 45, 1),
(42, 4, '用户列表-重置密码', '', '', 3, 'organization.users.reset_password', 46, 1),
(43, 9, '角色权限-编辑资料', '', '', 3, 'permission.roles.edit_profile', 113, 1),
(44, 9, '角色权限-分配菜单', '', '', 3, 'permission.roles.assign_menus', 114, 1),
(45, 10, '菜单权限-维护定义', '', '', 3, 'permission.menus.manage_schema', 123, 1),
(46, 3, '岗位管理-编辑资料', '', '', 3, 'organization.positions.edit_profile', 34, 1)
)
INSERT INTO permission_menu (
  id,
  parent_id,
  name,
  route_path,
  component_path,
  menu_type,
  permission_key,
  is_delegable,
  sort,
  status
)
SELECT
  id,
  parent_id,
  name,
  route_path,
  component_path,
  menu_type,
  permission_key,
  CASE
    WHEN permission_key LIKE 'permission.%'
      OR permission_key IN ('organization.positions.assign_roles', 'organization.users.reset_password', 'system.settings.edit', 'permission.menus.manage_schema')
    THEN FALSE
    ELSE TRUE
  END,
  sort,
  status
FROM permission_menu_seed
ON CONFLICT (permission_key, deleted_at) DO UPDATE SET
  parent_id = EXCLUDED.parent_id,
  name = EXCLUDED.name,
  route_path = EXCLUDED.route_path,
  component_path = EXCLUDED.component_path,
  menu_type = EXCLUDED.menu_type,
  is_delegable = EXCLUDED.is_delegable,
  sort = EXCLUDED.sort,
  status = EXCLUDED.status,
  updated_at = CURRENT_TIMESTAMP;
SELECT setval(pg_get_serial_sequence('permission_menu', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM permission_menu), 1), true);

INSERT INTO permission_role (id, role_name, role_code, role_type, is_protected) VALUES
(1, '超级管理员', 'super_admin', 1, TRUE),
(2, '组织管理员', 'organization_admin', 2, FALSE),
(3, '审计员', 'auditor', 1, TRUE),
(4, '系统运维管理员', 'system_ops_admin', 1, TRUE),
(7, '部门主管', 'department_manager', 1, TRUE),
(8, '只读观察员', 'readonly_observer', 1, TRUE),
(9, '普通员工', 'employee', 2, FALSE)
ON CONFLICT (id) DO UPDATE SET
  role_name = EXCLUDED.role_name,
  role_code = EXCLUDED.role_code,
  role_type = EXCLUDED.role_type,
  is_protected = EXCLUDED.is_protected,
  updated_at = CURRENT_TIMESTAMP;
SELECT setval(pg_get_serial_sequence('permission_role', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM permission_role), 1), true);

INSERT INTO permission_role_menu (role_id, menu_id) VALUES
(1, 1),(1,2),(1,3),(1,4),(1,5),(1,6),(1,7),(1,15),(1,16),(1,8),(1,9),(1,10),(1,11),(1,12),(1,13),(1,14),(1,17),(1,18),(1,19),(1,24),(1,25),(1,26),(1,27),(1,28),(1,29),(1,30),(1,31),(1,33),(1,34),(1,35),(1,36),(1,37),
(2, 1),(2,2),(2,3),(2,4),(2,19),(2,28),(2,29),(2,30),
(3, 8),(3,10),(3,11),(3,13),
(4, 11),(4,12),(4,13),(4,14),(4,17),(4,18),(4,24),(4,25),(4,26),(4,27),(4,33),(4,34),(4,35),
(7, 1),(7,2),(7,3),(7,4),
(8, 11),(8,13),
(1,38),(1,39),(1,40),(1,41),(1,42),(1,43),(1,44),(1,45),(1,46),
(2,39),(2,40),(2,41),(2,42),(2,46)
ON CONFLICT (role_id, menu_id) DO NOTHING;

-- 普通员工（role_code=employee）默认不授予后台菜单权限，作为岗位分配的最低权限基线。
