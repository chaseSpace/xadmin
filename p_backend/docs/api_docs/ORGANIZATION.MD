# ORGANIZATION HTTP API DOC

## 部门管理 API 列表
- `GET /v1/organization/departments/tree`
- `GET /v1/organization/departments/:id`
- `POST /v1/organization/departments`
- `PUT /v1/organization/departments/:id`
- `POST /v1/organization/departments/:id/status`
- `DELETE /v1/organization/departments/:id`

## 岗位管理 API 列表
- `GET /v1/organization/positions`
- `GET /v1/organization/positions/:id`
- `POST /v1/organization/positions`
- `PUT /v1/organization/positions/:id`
- `POST /v1/organization/positions/:id/roles`
- `POST /v1/organization/positions/:id/status`
- `DELETE /v1/organization/positions/:id`

## GET /v1/organization/departments/tree
- 描述：获取组织部门树（用于左侧树结构展示）。
- 鉴权：是。
- 统计口径：
  - `member_count`：该部门及全部下级部门的未删除账号数量（不区分账号状态）。
  - `position_count`：该部门及全部下级部门的未删除岗位数量（不区分岗位状态）。

### Response 200
```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "1",
        "parent_id": "0",
        "name": "集团总部",
        "code": "HQ",
        "status": "enabled",
        "member_count": 16,
        "updated_at": "2026-04-22 23:30:00",
        "children": []
      }
    ]
  }
}
```

## GET /v1/organization/departments/:id
- 描述：获取部门详情。
- 鉴权：是。
- 统计口径同 `/v1/organization/departments/tree`。

## POST /v1/organization/departments
- 描述：创建部门（支持一级部门和下级部门）。
- 鉴权：是。
- 约束：
  - 最多只支持 4 级部门；当父部门已是 4 级时，禁止继续新增下级部门。

### Request Body
```json
{
  "parent_id": "0",
  "name": "产品与技术中心",
  "code": "DEPT-TECH"
}
```

## PUT /v1/organization/departments/:id
- 描述：更新部门基础信息（名称/编码）。
- 鉴权：是。

## POST /v1/organization/departments/:id/status
- 描述：启用/停用部门。
- 鉴权：是。
- 约束：
  - 不允许停用自己所在部门。

### Request Body
```json
{
  "enabled": true
}
```

## DELETE /v1/organization/departments/:id
- 描述：删除部门。
- 鉴权：是。
- Query：
  - `force`（可选）：当部门存在下级部门或岗位但无成员时，首次删除会返回确认提示；确认后传 `force=true` 继续删除。
- 约束：
  - 若当前部门下仍有成员，禁止删除并返回明确提示。

## GET /v1/organization/positions
- 描述：岗位列表（支持分页、筛选与排序）。
- 鉴权：是。
- Query 参数：
  - `page_no`（可选）：>=1，默认 1。
  - `page_size`（可选）：1-100，默认 10。
  - `order_field`（可选）：`id` / `name` / `level` / `status` / `updated_at`。
  - `order_type`（可选）：`asc` / `desc`，默认 `desc`。
  - `keyword`（可选）：岗位名称/岗位编码模糊匹配。
  - `department_id`（可选）：部门 ID。
  - `inherit_parent`（可选）：`true/false`，默认 `false`；为 `true` 且所选部门没有直属岗位时，返回最近一个配置了岗位的上级部门岗位，供新增用户/调岗级联选择。
  - `level`（可选）：岗位职级精确匹配。
  - `status`（可选）：`enabled` / `disabled`。

## GET /v1/organization/positions/:id
- 描述：岗位详情。
- 鉴权：是。
- 列表和详情均返回只读 `management_rank`，以及 `is_protected/can_manage/can_assign/can_assign_roles`；能力字段由后端按当前操作者实时计算。新增岗位仅允许超管设置非零 `management_rank`，编辑岗位请求不接收该字段。

## POST /v1/organization/positions
- 描述：创建岗位基础资料，不绑定角色。
- 鉴权：是。
- 权限：`organization.positions.edit_profile`。
- 约束：请求包含非空 `role_ids` 时拒绝，角色绑定必须使用专用接口。

### Request Body
```json
{
  "name": "前端工程师",
  "code": "POS-FE",
  "department_id": "1",
  "level": "P5",
  "management_rank": 60,
  "hc": 5,
  "staffed": 3
}
```
- `management_rank`：可选，范围 `0~1000`，数值越高管理层级越高；只有超级管理员可以创建非零管理层级岗位，普通管理员省略或传 `0`。

## PUT /v1/organization/positions/:id
- 描述：更新岗位基础信息，不修改角色。
- 鉴权：是。
- 权限：`organization.positions.edit_profile`。
- 约束：请求包含非空 `role_ids` 时拒绝；非超管只能操作 `management_rank` 严格低于自己的非受保护岗位，任何人都不能通过普通岗位资料接口操作自己所在岗位。

## POST /v1/organization/positions/:id/roles
- 描述：替换岗位绑定的角色。
- 鉴权：是。
- 权限：`organization.positions.assign_roles`，并强制要求稳定角色编码 `super_admin` 对应的超管身份；Service 层会再次校验。
- 约束：即使是超级管理员，也不能修改自己所在岗位的角色关系；列表对自己岗位返回 `can_assign_roles=false`。
- 会话：岗位角色与受影响用户会话撤销在同一事务完成，避免旧会话继续使用变更前权限。

### Request Body
```json
{
  "role_ids": [2, 3]
}
```

## POST /v1/organization/positions/:id/status
- 描述：启用/停用岗位。
- 鉴权：是。
- 约束：
  - 不允许停用自己所在岗位。

### Request Body
```json
{
  "enabled": true
}
```

## DELETE /v1/organization/positions/:id
- 描述：删除岗位（软删除）。
- 鉴权：是。

## GET /v1/organization/users
- 描述：组织用户列表（含账号状态、在线状态、活跃会话数、最近登录信息）。
- 鉴权：是。
- Query 参数：
  - `page_no`（可选）：>=1，默认 1。
  - `page_size`（可选）：1-100，默认 10。
  - `order_field`（可选）：支持后端白名单字段（如 `uid`、`username`、`display_name`、`created_at`、`updated_at`、`active_session_count`）。
  - `order_type`（可选）：`asc` / `desc`，默认 `desc`。
  - `keyword`（可选）：用户名/姓名模糊匹配。
  - `phone`（可选）：手机号模糊匹配。
  - `status`（可选）：`active` / `disabled` / `deactivated`。
  - `department_id`（可选）：部门 ID（精确匹配）。
  - `position_id`（可选）：岗位 ID（精确匹配）。
  - `created_from`（可选）：`YYYY-MM-DD HH:mm:ss`。
  - `created_to`（可选）：`YYYY-MM-DD HH:mm:ss`。

### Response 200
```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "total": "30",
    "page_no": 1,
    "page_size": 10,
    "items": [
      {
        "uid": 10001,
        "username": "admin",
        "display_name": "系统管理员",
        "avatar": "https://cdn.example.com/avatar/admin.png",
        "email": "admin@example.com",
        "phone": "13800000000",
        "account_status": "active",
        "online_status": "online",
        "active_session_count": 2,
        "last_login_ip": "127.0.0.1",
        "last_login_at": "2026-04-20 12:00:00",
        "department_id": "1",
        "department_name": "集团总部",
        "position_id": "1",
        "position_name": "总经理",
        "role_names": ["超级管理员"]
      }
    ]
  }
}
```
- `role_names`：用户岗位关联角色名列表，仅关联查询展示，不在用户创建/编辑接口中直接修改。
- `is_protected`：目标用户是否继承受保护角色。
- `can_manage`：当前操作者是否满足非自己、非受保护且目标岗位 `management_rank` 严格更低等 Service 规则；前端所有用户写操作应同时结合自身权限键判断。

## POST /v1/organization/users/import
- 描述：导入组织用户（前端解析 CSV 后提交结构化数据）。
- 鉴权：是。
- 约束：`username` 必须以英文字母开头，且只能包含英文字母、数字、下划线和减号，长度 1~64。

### Request Body
```json
{
  "items": [
    {
      "username": "user_demo_001",
      "password": "Reset@123456",
      "display_name": "演示用户A",
      "email": "user_demo_001@example.com",
      "phone": "13800000001",
      "status": 1
    }
  ]
}
```

### Response 200
```json
{
  "code": 200,
  "message": "CodeSuccess",
  "data": {
    "success": true,
    "action": "import_users"
  }
}
```

## GET /v1/organization/users/export
- 描述：导出组织用户（CSV 文件下载）。
- 鉴权：是。
- Query 参数同 `/v1/organization/users`。
- Response：`text/csv` 文件流。

## GET /v1/organization/users/:uid/sessions
- 描述：查看指定用户会话列表（最新若干条）。
- 鉴权：是。
- Query 参数：
  - `page_size`（可选）：1-100，默认 10。
  - `status`（可选）：`active` / `revoked` / `expired`。

### Response 200
```json
{
  "code": 200,
  "message": "CodeSuccess",
  "data": {
    "items": [
      {
        "session_id": "d9ff4f6e-2cec-4fbb-b4cf-443503418b72",
        "status": "active",
        "login_ip": "127.0.0.1",
        "user_agent": "Mozilla/5.0",
        "last_seen_at": "2026-04-19T10:00:00+08:00",
        "expired_at": "2026-04-20T10:00:00+08:00",
        "revoked_at": "",
        "revoked_reason": ""
      }
    ]
  }
}
```

## POST /v1/organization/users
- 描述：创建组织用户。
- 鉴权：是。
- 权限：`organization.users.edit_profile`。
- Request Body 增补：
  - `username`：必须以英文字母开头，且只能包含英文字母、数字、下划线和减号，长度 1~64。
  - `department_id`（可选）：关联部门 ID。
  - `position_id`（可选）：关联岗位 ID（传入时要求岗位属于部门；无直属岗位时可属于最近配置了岗位的上级部门，且岗位必须为启用状态）。
- 约束：非超管不能把新用户放入受保护或同级/更高层级岗位；目标岗位权限还必须在操作者的可委派权限范围内。

## POST /v1/organization/users/transfer-position
- 描述：批量转移用户岗位。
- 鉴权：是。
- 权限：`organization.users.assign_position`。

### Request Body
```json
{
  "uids": [21001, 21002],
  "department_id": "1",
  "position_id": "2"
}
```
- 约束：岗位必须属于所选部门；无直属岗位时可属于最近配置了岗位的上级部门，且岗位必须为启用状态。非超管不得包含自己、受保护账号或同级/更高管理层级账号；目标岗位管理层级必须更低，且权限必须处于可委派范围。
- 会话：调岗和所有目标用户的会话撤销在同一事务完成。

## PUT /v1/organization/users/:uid/profile
- 描述：仅编辑组织用户资料；兼容路径 `PUT /v1/organization/users/:uid` 使用同一窄请求模型。
- 鉴权：是。
- 权限：`organization.users.edit_profile`。
- Request Body：只接受 `display_name/avatar/email/phone`；夹带的 `status/department_id/position_id` 不会进入 Service。
- 约束：
  - 用户已停用/已注销后，不允许编辑。
  - 非超管只能操作岗位管理层级严格低于自己的非受保护账号，且不能通过管理接口操作自己。

## POST /v1/organization/users/:uid/position
- 描述：单用户调岗。
- 鉴权：是。
- 权限：`organization.users.assign_position`。
- Request Body：`department_id/position_id`。
- 约束：同批量调岗；调岗和目标会话撤销在同一事务完成。

## POST /v1/organization/users/:uid/status
- 描述：修改用户状态。
- 鉴权：是。
- 权限：`organization.users.change_status`。
- Request Body：`status`（`0` 停用、`1` 启用、`2` 注销）。
- 约束：非超管不得操作自己、受保护账号或同级/更高管理层级账号；停用/注销时在同一事务撤销目标会话。

## POST /v1/organization/users/:uid/reset-password
- 描述：重置目标用户密码；兼容路径 `/reset_password`。
- 鉴权：是。
- 权限：`organization.users.reset_password`（不可委派）。
- 约束：非超管不得操作自己、受保护账号或同级/更高管理层级账号；业务权限域不再参与目标用户层级比较。
- 会话：密码更新与目标会话撤销在同一事务完成。

## DELETE /v1/organization/users/:uid
- 描述：软删除组织用户。
- 鉴权：是。
- 约束：
  - 仅允许删除状态为“已注销”且已注销满 3 个月的用户。
  - 非超管不得删除自己、受保护账号或同级/更高管理层级账号；保护判断不依赖用户名。

### Response 200
```json
{
  "code": 200,
  "message": "CodeSuccess",
  "data": {
    "success": true,
    "action": "delete_user"
  }
}
```

## 岗位角色关联
- 位置：仅 `POST /v1/organization/positions/:id/roles`。
- 字段：`role_ids`（`int64[]`）。
- 约束：角色必须存在且未删除，且该接口属于超管控制面。
- 岗位列表返回：`role_names`（与 `role_ids` 顺序一致，用于前端展示角色名）。

## 权限目标规则
- 超管身份由 `permission_role.role_code=super_admin` 且 `is_protected=true` 识别，不比较角色名或用户名。
- 用户和岗位的组织管理权由 `organization_position.management_rank` 判断；非超管只能管理严格更低层级的非受保护目标，同级之间不能互相管理。
- 业务权限只控制操作者能否执行某类动作，不再用于比较用户或岗位的组织管理层级。
- 分配岗位时，目标岗位层级必须严格更低，且岗位权限必须包含于操作者的可委派权限；`permission_menu.is_delegable=false` 的权限不能继续传播。
- 角色资料管理仍使用角色保护属性和权限范围比较；岗位绑角色、角色绑菜单、菜单定义仍是超管控制面。
- `management_rank` 是受控安全属性；岗位列表和详情只读展示，原始初始化 SQL 和超级管理员新增岗位时可以设置。普通管理员只能创建层级 0 岗位，岗位编辑请求不能修改该字段。
- 所有目标校验都在 Service 层执行，不能通过绕过路由或伪造请求直接提权。

## 停用联动登录限制
- 当用户状态被改为停用/注销时：
  - 用户在线会话会被强制下线。
  - 后续登录将被拒绝并提示：`您的账户被停用` / `您的账户已注销`。
- 当部门被停用时：
  - 该部门用户在线会话会被强制下线。
  - 后续登录将被拒绝并提示：`您所在的部门被停用`。
- 当岗位被停用时：
  - 该岗位用户在线会话会被强制下线。
  - 后续登录将被拒绝并提示：`您所在的岗位被停用`。
