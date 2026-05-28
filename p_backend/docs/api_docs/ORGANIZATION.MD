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
  - `level`（可选）：岗位职级精确匹配。
  - `status`（可选）：`enabled` / `disabled`。

## GET /v1/organization/positions/:id
- 描述：岗位详情。
- 鉴权：是。

## POST /v1/organization/positions
- 描述：创建岗位。
- 鉴权：是。

### Request Body
```json
{
  "name": "前端工程师",
  "code": "POS-FE",
  "department_id": "1",
  "level": "P5",
  "hc": 5,
  "staffed": 3
}
```

## PUT /v1/organization/positions/:id
- 描述：更新岗位基础信息。
- 鉴权：是。

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

## POST /v1/organization/users/import
- 描述：导入组织用户（前端解析 CSV 后提交结构化数据）。
- 鉴权：是。

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
- Request Body 增补：
  - `department_id`（可选）：关联部门 ID。
  - `position_id`（可选）：关联岗位 ID（传入时要求岗位属于部门，且岗位必须为启用状态）。

## POST /v1/organization/users/transfer-position
- 描述：批量转移用户岗位。
- 鉴权：是。

### Request Body
```json
{
  "uids": [21001, 21002],
  "department_id": "1",
  "position_id": "2"
}
```
- 约束：岗位必须属于所选部门，且岗位必须为启用状态。

## PUT /v1/organization/users/:uid
- 描述：编辑组织用户。
- 鉴权：是。
- Request Body 增补：
  - `department_id`（可选）：关联部门 ID。
  - `position_id`（可选）：关联岗位 ID（传入时要求岗位属于部门；若改到新岗位，岗位必须为启用状态）。
- 约束：
  - 用户已停用/已注销后，不允许编辑。
  - 不允许将当前登录账号改为停用或已注销。

## DELETE /v1/organization/users/:uid
- 描述：软删除组织用户。
- 鉴权：是。
- 约束：
  - 仅允许删除状态为“已注销”且已注销满 3 个月的用户。
  - `admin` 用户不允许删除。

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
- 位置：岗位新增与编辑接口。
- 字段：`role_ids`（可选，`int64[]`）。
- 约束：角色必须存在且未删除。
- 岗位列表返回：`role_names`（与 `role_ids` 顺序一致，用于前端展示角色名）。

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
