# 权限控制面前端接入说明

## 当前用户权限上下文

登录后通过 `GET /v1/account/me/profile` 获取：

- `permission_keys`：全部有效权限键，写操作入口统一通过 `hasPermission` 判断。
- `is_super_admin`：稳定超管身份，禁止通过用户名或角色显示名称推断。

权限上下文保存在 Auth Store；重新登录或刷新页面时以服务端返回为准。

## 用户操作

- 用户资料：`PUT /organization/users/:uid/profile`
- 用户调岗：`POST /organization/users/:uid/position`
- 用户状态：`POST /organization/users/:uid/status`
- 密码重置：`POST /organization/users/:uid/reset-password`
- 批量调岗：`POST /organization/users/transfer-position`

用户列表返回 `is_protected/can_manage`。按钮显示条件为“当前用户拥有对应权限键”且“目标 `can_manage=true`”。资料编辑请求不得夹带岗位和状态字段。

## 岗位操作

- 岗位资料：`POST/PUT /organization/positions[/:id]`，不携带 `role_ids`。
- 岗位角色：`POST /organization/positions/:id/roles`，独立超管入口。

岗位返回 `is_protected/can_manage/can_assign/can_assign_roles`。用户调岗下拉仅展示 `can_assign=true` 的岗位。

## 角色与菜单

- 角色使用 `role_code/is_protected`，超管只判断 `role_code=super_admin`。
- 角色写操作使用 `can_manage/can_assign_menus`。
- 菜单使用 `is_delegable` 展示可委派状态。
- 岗位绑角色、角色绑菜单、菜单定义三个控制面仅超管显示。

## 错误处理

- 401：沿用全局未授权退出流程，不显示业务 message。
- 403：展示后端目标等级或授权范围提示，并刷新相关列表。
- 权限变更可能撤销受影响用户会话，前端不得继续使用本地旧权限判断绕过重新登录。
