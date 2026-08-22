# 权限体系与管理权设计

## 1. 文档定位

本文只描述 XAdmin **当前已经实现** 的权限体系，包括数据库模型、后端判断顺序、前后端能力契约、岗位管理层级、安全边界和测试要求。

当前实现的核心原则是：

> 业务权限决定“能不能执行这个动作”，岗位管理层级决定“能不能管理这个目标”，可委派权限决定“能不能把这些权限向下授予”。

三类判断彼此独立，任何一类都不能替代另外两类。

当前实现不包含：

- 按部门树限制管理范围。
- 存量岗位管理层级编辑接口。
- 独立的管理层级配置页面。
- 已部署数据库的增量 DDL。

## 2. 权限模型总览

### 2.1 用户权限继承链

后台用户不直接绑定角色，权限通过岗位继承：

```text
admin_user
  -> organization_position
    -> organization_position_role
      -> permission_role
        -> permission_role_menu
          -> permission_menu.permission_key
```

对应数据表：

| 表 | 作用 | 关键字段 |
|---|---|---|
| `admin_user` | 后台用户 | `uid`、`department_id`、`position_id`、`status` |
| `organization_position` | 岗位与组织管理层级 | `department_id`、`level`、`management_rank`、`status` |
| `organization_position_role` | 岗位绑定角色 | `position_id`、`role_id` |
| `permission_role` | 角色定义 | `role_code`、`role_type`、`is_protected` |
| `permission_role_menu` | 角色绑定菜单/API 权限 | `role_id`、`menu_id` |
| `permission_menu` | 菜单、页面和按钮权限 | `permission_key`、`is_delegable`、`status` |

### 2.2 三条独立判断轴

| 判断轴 | 回答的问题 | 数据来源 |
|---|---|---|
| 动作权限 | 操作者能否调用某个功能或接口 | `permission_key` |
| 目标管理权 | 操作者能否管理目标用户或岗位 | `management_rank` |
| 权限委派权 | 操作者能否把目标岗位权限授予用户 | `is_delegable` + 权限子集 |

一次写操作会按实际场景组合这些判断。

例如调整用户岗位必须同时满足：

1. 操作者拥有 `organization.users.assign_position`。
2. 操作者有权管理目标用户。
3. 操作者有权分配新岗位。
4. 用户、部门和岗位状态允许本次调整。

## 3. 动作权限

### 3.1 permission_key

每类敏感动作使用稳定的 `permission_key`，不依赖用户名、角色名称、岗位名称或前端路由文案。

组织模块示例：

| 动作 | 权限键 |
|---|---|
| 编辑用户资料 | `organization.users.edit_profile` |
| 调整用户岗位 | `organization.users.assign_position` |
| 修改用户状态 | `organization.users.change_status` |
| 重置用户密码 | `organization.users.reset_password` |
| 编辑岗位资料 | `organization.positions.edit_profile` |
| 岗位绑定角色 | `organization.positions.assign_roles` |

写接口首先通过 `RequirePermission(...)` 检查动作权限。

### 3.2 权限实时生效

权限中间件按当前 UID 从数据库查询有效权限，不依赖长期不失效的权限缓存。

有效权限必须满足：

- 用户、岗位和关联数据未删除。
- 菜单权限未删除。
- 菜单权限状态为启用。
- `permission_key` 非空。

岗位、角色或菜单关系变化后，后续请求会读取最新权限。

### 3.3 前端不是安全边界

前端通过当前用户的权限键控制菜单、路由和按钮显示，但隐藏按钮不能替代后端鉴权。

直接构造请求仍会经过：

```text
认证中间件
  -> 动作权限中间件
    -> Service 目标校验
      -> 状态与数据一致性校验
        -> Repository 写入
```

## 4. 超级管理员与受保护目标

### 4.1 超级管理员识别

超级管理员通过稳定数据库条件识别：

1. 用户绑定岗位。
2. 岗位绑定未删除角色。
3. 角色 `role_code=super_admin`。
4. 角色 `is_protected=true`。

不会使用以下信息识别超级管理员：

- 用户名。
- 显示名称。
- 可修改的角色名称。
- 岗位名称。
- 前端传入的布尔值。

### 4.2 受保护角色、岗位和用户

`permission_role.is_protected` 是硬保护开关。

- 岗位绑定任意受保护角色，则该岗位是受保护岗位。
- 用户所在岗位绑定任意受保护角色，则该用户是受保护用户。
- 普通管理员不能管理受保护角色、岗位或用户。
- 超级管理员可以管理其他受保护目标，但仍不能通过敏感管理接口操作自己或自己所在岗位。

当前初始化数据中只有 `super_admin` 是受保护角色。

`role_type=system` 与 `is_protected=true` 含义不同：

- `role_type=system` 表示系统内置角色，不能按普通自定义角色编辑或删除。
- `is_protected=true` 表示其关联目标只能由超级管理员管理。

### 4.3 超管控制面

以下能力同时要求专用权限键和超级管理员身份：

| 控制面 | 权限键 | 后端保护 |
|---|---|---|
| 岗位绑定角色 | `organization.positions.assign_roles` | `RequireSuperAdmin` + Service 二次校验 |
| 角色绑定菜单 | `permission.roles.assign_menus` | `RequireSuperAdmin` + Service 二次校验 |
| 菜单权限定义维护 | `permission.menus.manage_schema` | `RequireSuperAdmin` + Service 二次校验 |

菜单权限定义维护包括新增、修改、启停、删除和同步。

超级管理员也不能修改自己所在岗位的角色关系。岗位列表对自己岗位返回 `can_assign_roles=false`，写接口会再次拒绝。

## 5. 岗位管理层级

### 5.1 字段定义

`organization_position.management_rank` 专门表达组织管理顺序：

```sql
management_rank SMALLINT NOT NULL DEFAULT 0
  CHECK (management_rank >= 0 AND management_rank <= 1000)
```

数值越大，管理层级越高。

管理层级不从以下字段推导：

- 岗位名称。
- 岗位编码。
- 职级字符串，例如 `P3`、`M2`。
- 岗位绑定的角色数量。
- 岗位拥有的权限数量。

### 5.2 为什么层级属于岗位

- 角色表示业务能力，例如组织管理、审计或普通员工。
- 岗位表示组织身份，例如总经理、主管、经理或专员。
- 同一角色可以绑定到不同管理层级岗位。
- 修改角色权限不应意外改变组织管理关系。
- 不同业务权限域之间不需要形成权限包含关系。

### 5.3 当前初始化层级

| 岗位 | `management_rank` |
|---|---:|
| 总经理 | 100 |
| 组织主管 | 80 |
| 研发经理 | 60 |
| 产品运营经理 | 60 |
| 审计经理 | 60 |
| 后端工程师 | 20 |
| 用户运营 | 20 |
| 风控专员 | 20 |
| 无岗位/新建普通岗位默认值 | 0 |

数字只表达比较顺序，不在代码中针对岗位名称写特殊分支。

## 6. OperatorScope

列表能力计算使用当前操作者快照：

```go
type OperatorScope struct {
    UID                  int32
    PositionID           int64
    ManagementRank       int32
    SuperAdmin           bool
    Permissions          []string
    DelegablePermissions []string
}
```

字段含义：

- `UID`：当前操作者 UID。
- `PositionID`：当前岗位，用于阻止操作自己所在岗位。
- `ManagementRank`：当前岗位管理层级。
- `SuperAdmin`：稳定条件计算出的超管身份。
- `Permissions`：全部有效动作权限。
- `DelegablePermissions`：操作者拥有且 `is_delegable=true` 的权限。

列表使用 `OperatorScope` 计算能力布尔值；写接口不会信任旧列表结果，而是重新查询数据库并执行同一规则。

## 7. 用户目标管理规则

### 7.1 判断公式

```go
func CanManageUser(operator OperatorScope, target UserScope) bool {
    if operator.UID <= 0 || target.UID <= 0 {
        return false
    }
    if operator.UID == target.UID {
        return false
    }
    if operator.SuperAdmin {
        return true
    }
    if target.Protected {
        return false
    }
    return operator.ManagementRank > target.ManagementRank
}
```

### 7.2 规则解释

- 所有人都不能通过敏感管理接口操作自己。
- 超级管理员可以管理非自己的目标。
- 普通管理员不能管理受保护用户。
- 普通管理员只能管理岗位层级严格更低的用户。
- 同级用户不能互相管理。
- 无岗位用户按层级 0 处理。
- 目标用户的业务权限集合不参与管理层级比较。

### 7.3 使用该规则的操作

- 编辑用户资料。
- 调整岗位。
- 修改账号状态。
- 重置密码。
- 强制下线。
- 注销账号。
- 删除账号。
- 批量转移岗位中的每一个目标用户。

上述操作仍需各自的动作权限和账号状态校验。

## 8. 岗位目标管理规则

### 8.1 岗位资料管理

普通管理员管理岗位资料时：

```go
canManagePosition :=
    operator.PositionID != targetPosition.ID &&
    !targetPosition.Protected &&
    operator.ManagementRank > targetPosition.ManagementRank
```

判断顺序：

1. 岗位 ID 和操作者有效。
2. 目标不是操作者自己的岗位。
3. 超管可跳过受保护和层级比较。
4. 普通管理员不能管理受保护岗位。
5. 普通管理员层级必须严格高于目标岗位。

自己岗位的资料、状态、删除和角色关系均不能通过普通管理流程修改。

### 8.2 岗位分配

普通管理员把岗位分配给用户时：

```go
canAssignPosition :=
    !targetPosition.Protected &&
    operator.ManagementRank > targetPosition.ManagementRank &&
    IsSubset(
        targetPosition.Permissions,
        operator.DelegablePermissions,
    )
```

三个条件必须同时满足：

1. 目标岗位不是受保护岗位。
2. 目标岗位管理层级严格低于操作者。
3. 目标岗位全部有效权限都位于操作者可委派权限集合内。

该规则防止普通管理员：

- 把用户提升到与自己同级或更高岗位。
- 授予自己没有的权限。
- 授予自己拥有但不可继续委派的敏感权限。

超级管理员可以给其他用户分配任意有效岗位。给其他用户分配自己的岗位不等于修改自己的岗位，因此不受自己岗位资料保护规则限制。

### 8.3 管理岗位与分配岗位的区别

`can_manage` 和 `can_assign` 是两个独立结果：

- `can_manage`：能否编辑、启停或删除岗位资料。
- `can_assign`：能否把该岗位分配给用户。

一个岗位可能允许管理但不允许分配，例如岗位层级较低但包含不可委派权限。

## 9. 可委派权限

`permission_menu.is_delegable` 决定权限是否允许普通管理员继续向下授予。

普通管理员的可委派权限集合必须同时满足：

1. 操作者实际拥有该权限。
2. 权限未删除且处于启用状态。
3. `permission_key` 非空。
4. `is_delegable=true`。

权限控制面、岗位角色配置、重置密码和系统设置等敏感权限默认不可委派。

`is_delegable` 只控制向下授予，不影响操作者本人能否使用该权限。

## 10. 角色管理规则

角色管理仍使用角色类型、保护标记和权限范围，不使用岗位 `management_rank`。

当前规则：

- 系统角色不能按普通自定义角色修改或删除。
- 普通管理员只能查看后端允许其查看的非受保护角色。
- 普通管理员只能管理权限范围严格低于自己的非保护自定义角色。
- 角色菜单分配仅限超级管理员。
- 超级管理员角色不能通过普通角色配置流程被降级或替换。

角色权限集合的严格子集比较只保留在角色授权控制面，不再用于判断用户或岗位的组织管理关系。

## 11. 岗位新增与管理层级写入

### 11.1 新增岗位

`OrganizationCreatePositionReq.management_rank` 的约束为 `0~1000`。

Service 规则：

- 省略字段按 0 处理。
- 普通管理员可以创建层级 0 岗位。
- 请求值非 0 时，Service 强制执行 `EnsureSuperAdmin`。
- 超级管理员可以创建 `0~1000` 的岗位。
- 新增成功会记录 `create_position` 审计日志，并包含岗位编码和管理层级。

前端只向超级管理员显示管理层级输入框，但 Service 校验才是最终安全边界。

### 11.2 编辑存量岗位

`OrganizationUpdatePositionReq` 不包含 `management_rank`。

因此任何操作者都不能通过普通岗位编辑接口修改存量岗位层级，也不能通过前端夹带字段改变层级。

当前没有独立的层级修改接口或页面。

## 12. API 与前端能力契约

### 12.1 用户列表

用户列表返回：

| 字段 | 含义 |
|---|---|
| `is_protected` | 用户是否从岗位继承受保护角色 |
| `can_manage` | 当前操作者是否满足用户目标管理规则 |

用户列表不返回目标原始 `management_rank`，该值只在后端内部参与计算。

### 12.2 岗位列表与详情

岗位列表和详情返回：

| 字段 | 含义 |
|---|---|
| `management_rank` | 岗位管理层级，只读展示 |
| `is_protected` | 岗位是否绑定受保护角色 |
| `can_manage` | 当前操作者能否管理岗位资料 |
| `can_assign` | 当前操作者能否把岗位分配给用户 |
| `can_assign_roles` | 当前操作者能否修改岗位角色 |

`can_assign_roles` 只有超级管理员且目标不是自己所在岗位时才为 `true`。

### 12.3 角色列表

角色列表返回：

| 字段 | 含义 |
|---|---|
| `is_protected` | 角色是否受保护 |
| `can_manage` | 当前操作者能否管理角色资料 |
| `can_assign_menus` | 当前操作者能否修改角色菜单 |

### 12.4 前端使用方式

前端按钮通常同时检查：

```text
操作者动作权限
AND
后端返回的目标能力字段
AND
目标当前状态是否允许该动作
```

写接口会重新读取数据库。列表返回的 `can_manage`、`can_assign` 或 `can_assign_roles` 不能作为写接口的可信输入。

## 13. 岗位管理页面

当前不新增独立管理层级页面，而是在现有岗位管理页面接入。

### 13.1 表格

- 增加只读“管理层级”列。
- 问号说明：数值越高，可管理的岗位层级越高；同级不可互相管理。
- 岗位编码问号说明其为系统关联使用的稳定唯一标识。
- 职级问号说明职级仅用于展示，不参与权限或管理层级判断。

### 13.2 新增岗位弹窗

每个字段均提供示例占位值：

| 字段 | 示例 |
|---|---|
| 岗位名称 | 后端工程师 |
| 岗位编码 | POS-BACKEND-ENGINEER |
| 所属部门 | 技术部 |
| 职级 | P3 |
| 管理层级 | 60 |
| 编制人数 | 6 |
| 在岗人数 | 3 |
| 状态 | 启用 |

补充说明：

- 管理层级字段仅向超级管理员显示。
- 职级右侧灰色小字说明：P 为专业序列，M 为管理序列；同序列数字越大职级越高。
- 岗位编码、职级和管理层级均使用问号说明各自含义。
- 编辑岗位弹窗不显示或提交管理层级。

## 14. 状态约束与会话

权限允许不代表状态允许。

例如：

- 已注销账号不能执行普通调岗、停用和密码重置。
- 删除用户要求账号已注销并满足冷却期。
- 停用岗位会影响该岗位用户继续登录。
- 停用部门会影响该部门用户继续登录。

以下变化会按现有事务规则撤销相关会话：

- 用户停用或注销。
- 用户调岗。
- 批量转移岗位。
- 密码重置。
- 岗位停用。
- 岗位角色变化。
- 部门停用。

## 15. 错误语义

| 错误键 | 含义 |
|---|---|
| `auth.not_logged_in` | 操作者身份无效 |
| `auth.self_privilege_change_forbidden` | 禁止操作自己或自己所在岗位 |
| `auth.protected_target` | 目标受保护 |
| `auth.target_management_rank_exceeded` | 目标管理层级不低于操作者 |
| `auth.grant_scope_exceeded` | 目标岗位包含不可委派权限 |
| `auth.target_scope_exceeded` | 角色权限范围不低于操作者 |
| `auth.superadmin_required` | 操作要求超级管理员 |
| `auth.dedicated_endpoint_required` | 敏感字段必须使用专用接口 |

用户/岗位管理使用 `auth.target_management_rank_exceeded`；`auth.target_scope_exceeded` 仅保留给角色权限范围判断。

## 16. 数据初始化

当前本地环境不保留增量 DDL。

实现方式：

1. 在 `docs/model_sql_pg/02_organization_position.sql` 原始建表语句中定义 `management_rank`。
2. 在同一原始 SQL 中写入 8 个初始化岗位层级。
3. 使用 `make execsql ENV=dev` 全量重新初始化本地数据库。
4. 初始化后检查表结构、岗位种子层级和实际用户管理关系。

全量初始化会重建会话表，因此执行后需要重新登录。

如果未来已有生产数据需要升级，必须单独设计受控迁移，先完整回填现有岗位层级，再切换应用代码。当前仓库没有为此保留临时 SQL。

## 17. 当前管理范围边界

`management_rank` 当前是全局比较，没有部门数据范围。

因此拥有动作权限的高层级管理员，可以管理全组织中层级更低的非保护目标，即使双方不在同一部门。

当前没有实现：

```text
target.DepartmentID 位于 operator.AllowedDepartmentScope
```

如果未来需要“只能管理本部门及下级部门”，必须新增独立数据范围判断，并与管理层级同时满足，不能复用角色或岗位名称推导。

## 18. 高晨案例

当前种子数据：

```text
组织主管 林嘉
management_rank = 80

高晨 / 风控专员 / 审计员角色
management_rank = 20
```

组织主管与审计员属于不同业务权限域，权限集合互不包含。

当前实现不再比较双方业务权限集合，而是判断：

```text
80 > 20
```

因此林嘉在拥有对应动作权限时，可以管理高晨。

以下限制仍独立生效：

- 高晨如果处于不允许操作的账号状态，对应按钮和接口仍会拒绝。
- 如果调入岗位包含林嘉不可委派的权限，仍不能完成调岗。
- 如果目标用户继承受保护角色，仍只能由超级管理员操作。

## 19. 安全不变量

当前实现必须持续满足：

1. 所有敏感写操作都需要动作权限。
2. 所有人都不能通过敏感管理接口操作自己。
3. 超管不能修改自己所在岗位的资料、状态或角色关系。
4. 普通管理员不能管理同级或更高层级目标。
5. 普通管理员不能管理受保护目标。
6. 普通管理员不能创建非零管理层级岗位。
7. 普通管理员不能分配同级或更高岗位。
8. 普通管理员不能授予不可委派权限。
9. 角色菜单、岗位角色和菜单定义仍是超管控制面。
10. 普通岗位编辑请求不能修改 `management_rank`。
11. 前端能力字段和表单显隐不能替代 Service 校验。
12. Service 写操作必须重新读取最新数据库状态。
13. 不允许通过用户名、角色名、岗位名、职级或前端路由推导安全身份。
14. 管理层级与业务权限必须保持分离。

## 20. 测试矩阵

### 20.1 用户管理

- 超管可以管理非自己的受保护和非保护用户。
- 任何人都不能管理自己。
- 普通管理员可以管理权限域不同但层级更低的用户。
- 普通管理员不能管理同级或更高层级用户。
- 普通管理员不能管理受保护用户。
- 用户列表 `can_manage` 与写接口结果一致。

### 20.2 岗位管理与分配

- 普通管理员可以管理更低层级非保护岗位。
- 普通管理员不能管理自己、同级、更高或受保护岗位。
- 岗位分配同时检查层级和可委派权限。
- 低层级岗位包含不可委派权限时仍不可分配。
- 超管自己岗位的 `can_assign_roles=false`，直接调用写接口同样拒绝。

### 20.3 岗位创建

- 超管可以创建 `0~1000` 管理层级岗位。
- 普通管理员可以创建层级 0 岗位。
- 普通管理员构造非零层级请求时，Service 在访问 Repository 前拒绝。
- 岗位编辑载荷不包含管理层级。
- 岗位列表和详情正确返回只读管理层级。

### 20.4 防绕过

- 直接调用 Service 仍执行管理层级校验。
- 直接构造 API 请求不能覆盖能力字段。
- 旧列表返回值不能绕过写接口实时校验。
- 普通管理员不能通过新建岗位或调岗获得同级/更高管理权。
- 权限关系变化后后续请求读取最新数据库状态。

## 21. 主要实现位置

| 内容 | 文件 |
|---|---|
| 原始岗位 SQL | `docs/model_sql_pg/02_organization_position.sql` |
| 岗位模型 | `internal/model/organization_position.go` |
| 授权 Repository | `internal/repo/authorization/repo.go` |
| 授权规则 | `internal/service/authorization/service.go` |
| 用户/岗位列表能力映射 | `internal/service/organization/service.go` |
| 组织路由与创建审计 | `internal/handler/organization/organization.go` |
| Proto 契约 | `proto/defined/organization.proto` |
| 授权测试 | `internal/service/authorization/service_test.go` |
| 组织控制面测试 | `internal/service/organization/control_plane_test.go` |

## 22. 当前设计不足与未来计划

### 22.1 当前不足

1. **管理范围是全局的**

   当前只比较 `management_rank`，没有部门数据范围。拥有动作权限的高层级管理员可以管理全组织中层级更低的非保护目标，无法表达“仅本部门”“本部门及下级部门”或指定部门集合。

2. **存量岗位层级不能在线调整**

   当前只有初始化 SQL 和超管新增岗位可以设置层级，没有修改已有岗位层级的接口和页面。安全性较高，但组织调整时需要重新初始化或人工受控修改数据库。

3. **单一数值只能表达一条全局管理链**

   `management_rank` 适合总经理、主管、经理、专员这类纵向关系，但无法表达矩阵组织、项目负责人、临时代理、跨部门虚线汇报等复杂关系。

4. **岗位层级没有独立变更历史**

   新增高层级岗位会记录审计日志，但当前没有专门的岗位层级版本表，无法直接查看某岗位层级的完整历史、变更前后值和审批依据。

5. **层级与状态的批量治理能力有限**

   当前能力字段按实时数据计算，但没有用于发现异常配置的管理报表，例如层级为 0 却拥有组织管理动作权限、低层级岗位绑定高危角色、关键岗位未设置明确层级等。

6. **前端只能解释结果，不能展示完整拒绝原因**

   列表只返回 `can_manage/can_assign` 布尔值。用户能够看到按钮禁用，但不能总是区分是层级不足、目标受保护、不可委派权限还是账号状态导致。

7. **现有本地初始化方式不适用于已有生产数据**

   当前通过原始 SQL 全量重建数据库，没有提交生产增量迁移。正式环境升级前必须另行设计可回滚、可核对的迁移流程。

### 22.2 未来计划

建议按以下顺序演进：

1. **增加部门数据范围**

   为操作者增加 `self_department`、`department_and_descendants`、`specified_departments`、`all` 等范围，并要求目标同时满足管理层级和部门范围。

2. **增加独立超管层级配置接口**

   使用专用请求模型修改存量岗位层级，不复用普通岗位编辑接口；接口必须要求超管、禁止修改自己岗位、记录变更前后值并撤销受影响会话。

3. **增加层级配置 UI**

   在岗位管理页提供仅超管可见的独立操作或抽屉，不新增普通管理员入口。提交前展示影响用户数、同级关系变化和潜在权限风险。

4. **增加岗位层级审计历史**

   记录岗位 ID、旧层级、新层级、操作者、原因、请求追踪标识和时间，支持按岗位查询完整变更链。

5. **增加拒绝原因枚举**

   在不泄露敏感权限明细的前提下，为列表能力增加可选原因，例如 `protected_target`、`rank_exceeded`、`grant_scope_exceeded`、`invalid_status`，便于前端给出准确说明。

6. **增加安全配置巡检**

   定期检查关键岗位层级缺失、组织管理权限与层级不匹配、受保护角色异常绑定、不可委派权限传播等问题，并输出审计告警。

7. **设计生产迁移方案**

   在进入已有数据环境前，准备独立增量 DDL、岗位层级回填表、上线前校验 SQL、回滚方案和灰度切换步骤；必须先完成数据回填，再启用层级判断。

8. **评估复杂组织关系模型**

   如果出现项目制或矩阵组织需求，再引入显式管理关系或授权委托表；不要继续叠加岗位名称规则或用更多魔法数扩展 `management_rank`。
