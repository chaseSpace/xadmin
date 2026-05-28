# PERMISSION HTTP API DOC

## 菜单权限 API 列表
- `GET /v1/permission/menus/tree`
- `GET /v1/permission/menus`
- `GET /v1/permission/menus/:id`
- `POST /v1/permission/menus`
- `PUT /v1/permission/menus/:id`
- `POST /v1/permission/menus/:id/status`
- `DELETE /v1/permission/menus/:id`
- `POST /v1/permission/menus/sync`

## 角色权限 API 列表
- `GET /v1/permission/roles`
- `GET /v1/permission/roles/:id`
- `POST /v1/permission/roles`
- `PUT /v1/permission/roles/:id`
- `DELETE /v1/permission/roles/:id`
- `GET /v1/permission/roles/:id/menus`
- `POST /v1/permission/roles/:id/menus`

## GET /v1/permission/menus
- 描述：菜单权限列表（分页、筛选、排序）。
- 鉴权：是。
- 默认排序：`sort ASC, id ASC`。
- Query:
  - `page_no` `page_size`
  - `order_field` `order_type`，支持 `id/name/menu_type/sort/status/updated_at`
  - `keyword`（可选）
  - `status`（可选：`enabled` / `disabled`）
  - `menu_type`（可选：`directory` / `menu` / `button`）
  - `deleted`（可选：`yes` / `no`，默认 `no`）
  - `tree_node_id`（可选：菜单树节点ID；传入后返回该节点及全部子孙节点的菜单列表）

## GET /v1/permission/menus/tree
- 描述：菜单树结构。
- 鉴权：是。

## POST /v1/permission/menus
- 描述：创建菜单/按钮权限。
- 鉴权：是。

## PUT /v1/permission/menus/:id
- 描述：更新菜单/按钮权限。
- 鉴权：是。

## POST /v1/permission/menus/:id/status
- 描述：启用/停用菜单权限。
- 鉴权：是。

## DELETE /v1/permission/menus/:id
- 描述：删除菜单权限。未删除记录执行软删除；已删除记录执行彻底删除。
- 鉴权：是。
- 约束：
  - 软删除前需先删除下级菜单。
  - 彻底删除只允许对已删除记录执行，且必须距离软删除时间超过 1 小时。

## POST /v1/permission/menus/sync
- 描述：同步前端路由到菜单权限。
- 鉴权：是。

## GET /v1/permission/roles
- 描述：角色权限列表（分页、筛选、排序）。
- 鉴权：是。
- 说明：
  - 用户与角色的关联仅通过岗位继承生效。
  - `绑定用户数` 统计口径为当前通过岗位继承到该角色的有效用户数。

## POST /v1/permission/roles
- 描述：创建角色。
- 鉴权：是。
- 约束：
  - 角色名称不可重复（重复时返回中文提示：`角色名称已存在`）。

## PUT /v1/permission/roles/:id
- 描述：更新角色。
- 鉴权：是。
- 约束：
  - 系统角色（`role_type=system`）不可改名、不可修改类型。

## DELETE /v1/permission/roles/:id
- 描述：删除角色（软删除）。
- 鉴权：是。
- 约束：
  - 系统角色（`role_type=system`）不可删除。
  - 若当前登录账号通过岗位继承了该角色，不允许删除。

## GET /v1/permission/roles/:id/menus
- 描述：查询角色已绑定菜单权限ID列表。
- 鉴权：是。

## POST /v1/permission/roles/:id/menus
- 描述：更新角色绑定菜单权限。
- 鉴权：是。
- 约束：
  - 超级管理员角色菜单权限不可修改；其他系统角色允许配置菜单。
  - 保存时会自动补齐所选菜单的上级目录/菜单，避免用户重新登录后子菜单因缺少父级节点而无法展示。

### Request Body
```json
{
  "menu_ids": [1, 2, 3]
}
```
