# 权限控制面人工安全测试清单

## 1. 测试账号与数据准备

- [ ] 准备超管账号 `SA`：岗位绑定 `role_code=super_admin` 且 `is_protected=true`。
- [ ] 准备普通管理员 `A`：拥有用户资料、调岗、状态、重置密码等细粒度权限，但不是超管。
- [ ] 准备同级管理员 `B`：有效权限集合与 `A` 相同。
- [ ] 准备低权限账号 `C`：有效权限是 `A` 的真子集。
- [ ] 准备受保护账号 `P`：岗位至少绑定一个 `is_protected=true` 的角色。
- [ ] 准备无权限岗位、普通岗位、受保护岗位和超管岗位各一个。
- [ ] 为 `A/B/C/P` 各建立至少一个有效登录会话，记录会话 token。

## 2. 当前用户能力契约

- [ ] 使用 `SA` 请求 `GET /v1/account/me/profile`，确认 `is_super_admin=true`，并返回完整 `permission_keys`。
- [ ] 使用 `A` 请求同一接口，确认 `is_super_admin=false`，且权限键不包含三个超管控制面权限。
- [ ] 刷新浏览器后确认权限键仍从服务端重新加载，不依赖用户名或本地硬编码。
- [ ] 将角色名称“超级管理员”改为其他名称后，`SA` 仍被识别为超管。
- [ ] 将普通角色名称改为“超级管理员”后，对应账号仍不是超管。

## 3. 岗位角色控制面

- [ ] `SA` 调用 `POST /v1/organization/positions/:id/roles` 可以保存合法角色集合。
- [ ] `A` 在页面上看不到“配置角色”入口。
- [ ] `A` 直接调用岗位角色接口返回 403。
- [ ] `A` 在普通岗位编辑请求中夹带 `role_ids`，请求被拒绝且岗位角色不变。
- [ ] `A` 尝试给自己岗位追加超管角色，返回 403，重新登录后仍不是超管。
- [ ] `A` 尝试给其他岗位绑定受保护角色，返回 403。
- [ ] `SA` 修改岗位角色后，该岗位原有会话立即失效；旧 token 下一次请求返回 401。

## 4. 用户资料、调岗和状态拆分

- [ ] 编辑资料请求只修改 `display_name/avatar/email/phone`。
- [ ] 在资料请求中夹带 `position_id/department_id/status`，目标岗位和状态保持不变。
- [ ] `A` 调整 `C` 到权限处于可委派范围内的岗位，操作成功，`C` 的旧会话失效。
- [ ] `A` 调整自己岗位，返回 403。
- [ ] 批量调岗包含 `A` 自己时，整次请求返回 403，其他用户也不应被部分更新。
- [ ] `A` 把 `C` 调到超管岗位或受保护岗位，返回 403。
- [ ] `A` 把 `C` 调到包含不可委派权限的岗位，返回 403。
- [ ] `A` 把 `C` 调到超出自身权限范围的岗位，返回 403。
- [ ] `A` 修改自己的账号状态，返回 403。
- [ ] `A` 修改 `B` 的状态，因同级权限返回 403。
- [ ] `A` 修改 `P` 的状态，因受保护目标返回 403。
- [ ] `A` 停用 `C` 成功，`C` 的全部旧会话立即返回 401。

## 5. 新建账号防提权

- [ ] `A` 新建不带岗位或位于可委派普通岗位的账号，按业务权限正常处理。
- [ ] `A` 新建账号并指定超管岗位，返回 403，账号不得以超管权限创建成功。
- [ ] `A` 新建账号并指定包含不可委派权限的岗位，返回 403。
- [ ] 绕过前端直接提交上述请求，结果与页面操作一致。

## 6. 密码重置与账号接管

- [ ] `A` 重置 `C` 密码成功，返回临时密码，`C` 全部旧会话立即失效。
- [ ] `A` 重置自己密码的管理接口返回 403。
- [ ] `A` 重置同级 `B` 密码返回 403。
- [ ] `A` 重置受保护账号 `P` 或超管密码返回 403。
- [ ] 前端重置成功提示中的临时密码可点击复制。
- [ ] 未拥有 `organization.users.reset_password` 的账号看不到重置入口，直接调用返回 403。

## 7. 角色与菜单权限控制面

- [ ] 普通管理员看不到受保护角色；直接按已知角色 ID 修改仍返回 403 或不可修改结果。
- [ ] 角色判断使用 `role_code/is_protected`，修改角色显示名称不改变保护结果。
- [ ] `A` 调用 `POST /v1/permission/roles/:id/menus` 返回 403。
- [ ] `A` 调用菜单新增、编辑、删除、状态或同步接口均返回 403。
- [ ] `A` 尝试把已绑定菜单的 `permission_key` 改成超管权限，返回 403，数据库值不变。
- [ ] `SA` 可以配置非超管角色菜单；超管角色自身菜单不可修改。
- [ ] `SA` 修改角色菜单后，所有受该角色影响的旧会话立即失效。
- [ ] 自定义角色不能转换为系统角色；系统角色不能改名、改类型或删除。

## 8. 可委派权限

- [ ] 菜单列表正确显示 `is_delegable`。
- [ ] 所有 `permission.*` 权限均为不可委派。
- [ ] `organization.positions.assign_roles`、`organization.users.reset_password`、`system.settings.edit`、`permission.menus.manage_schema` 为不可委派。
- [ ] 超管控制面权限只绑定到超管角色。
- [ ] 普通管理员只能把目标调整到其 `DelegablePermissions` 覆盖的岗位。

## 9. 权限即时生效与多会话

- [ ] 用户权限检查每次实时读取数据库，不需要重启服务或等待本地缓存过期。
- [ ] 移除角色权限后，用户旧 token 即使尚未退出，也不能再访问对应接口。
- [ ] 调岗、角色菜单变化、密码重置、停用和注销均撤销对应活动会话。
- [ ] 同一用户同时登录两个浏览器，发生上述变化后两个会话都失效。
- [ ] 如果部署多个后端实例，分别访问两个实例，权限结果一致，不出现某个实例仍保留旧权限。
- [ ] 401 触发统一退出登录流程，前端不重复展示业务 message。

## 10. 数据库最终核验

- [ ] `permission_role.role_code` 在未删除角色中唯一，超管编码为 `super_admin`。
- [ ] 超管角色 `is_protected=true`。
- [ ] `permission_menu.is_delegable` 均有非空布尔值。
- [ ] 不存在非超管角色绑定三个控制面权限的记录。
- [ ] 不存在因历史悬空 `menu_id` 关系误绑定新权限的角色。

参考核验 SQL：

```sql
SELECT COUNT(*) AS unauthorized_control_plane_bindings
FROM permission_role_menu prm
JOIN permission_menu m ON m.id = prm.menu_id
WHERE m.permission_key IN (
  'organization.positions.assign_roles',
  'permission.roles.assign_menus',
  'permission.menus.manage_schema'
)
AND prm.role_id <> (
  SELECT id FROM permission_role
  WHERE role_code = 'super_admin' AND deleted_at = 0
);
```

预期结果：`unauthorized_control_plane_bindings = 0`。
