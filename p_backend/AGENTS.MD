# p_backend AGENTS.MD

本文件定义后端 AI/人工协作的统一开发范式。目标是让每次改动都能沿着同一条链路落地：
`需求 -> DDL(SQL) -> Proto/API -> Handler/Service/Repo(DAO) -> Docs -> Test -> Verify`。

## 1. 适用范围
- 项目：`p_backend`
- 语言与框架：Go + Fiber + GORM + Proto
- 适用任务：新功能、缺陷修复、接口变更、数据结构变更、文档补齐

## 2. 项目结构（以当前仓库为准）
- 服务入口：`cmd/server/main.go`
- 路由注册：`internal/handler/*/RegisterRoutes`
- 分层代码：
  - Handler：`internal/handler`
  - Service：`internal/service`
  - Repo(DAO)：`internal/repo`
  - Model：`internal/model`
- 协议定义：`proto/defined/*.proto`
- 生成脚本：`make pb`（`shell/gen_pb.sh`）
- API 文档：`docs/api_docs/*.MD`
- 表结构 SQL：`docs/model_sql/*.sql`
- API 测试：`test`
- 通用分页解析：`pkg/xfiber/query.go` (`ParsePageSortQuery`)

## 3. 变更类型与最小交付

### 3.1 仅 API 变更（无 DB 结构变化）
必须交付：
- `proto/defined/*.proto`（Req/Resp/枚举）
- `internal/handler/*` + `internal/service/*` + `internal/repo/*`（如涉及）
- `docs/api_docs/<MODULE>.MD`
- `test/*_test.go`

### 3.2 涉及 DB 结构变更（新增表/改表/索引）
必须交付：
- `docs/model_sql/<sql-file>.sql`（先写）
- 对已存在的表除了修改原表范式，还要增加alter 语句（如涉及），让开发者执行后手动删除
- 代码层变更（model/repo/service）
- 对应 API 文档与测试（若暴露 API）

### 3.3 仅 DAO/查询逻辑优化
必须交付：
- `internal/repo/*`
- 相关 service 调整
- 回归测试（至少覆盖核心查询或关键业务路径）

## 4. 强制开发流程（按顺序执行）

### Step 1: 明确边界与归属
- 先确认是否已有近似 API 可扩展，避免盲目新增接口。
- 若新增 API，先确认归属模块（`auth/account/organization/permission/system/assets`）。

### Step 2: 先落 SQL（只要涉及表结构）
- SQL 文件目录：`docs/model_sql`
- 推荐命名：`YYYYMMDD_HHMM_<topic>.sql`
- 文件内容至少包含：目的说明、DDL/DML、兼容性与回滚说明。
- 执行验证：
  - `cd p_backend && make execsql SQL=<sql文件名>`
  - `cd p_backend && make showtable TABLE=<表名>`

### Step 3: 先定义 Proto，再写 Handler/Service/Repo
- 新增 API 的 Req/Resp 必须先进入 `proto/defined/*.proto`。
- 参数约束优先写在 proto validator 规则，统一使用 `req.Validate()`。
- 生成代码：`cd p_backend && make pb`

### Step 4: 分层实现（禁止越层）
- Handler：只做参数解析、协议转换、调用 service。
- Service：业务编排、事务边界、幂等语义。
- Repo(DAO)：只做数据访问，不写业务策略。

### Step 5: 同步 API 文档
- 文档目录：`docs/api_docs`
- 命名必须大写模块名：`<MODULE>.MD`（如 `AUTH.MD`）
- 同模块 API 追加到已有文档，禁止重复建同类文件。

### Step 6: 测试与校验
- 新增 API 必须在 `test` 补对应用例。
- 最低校验命令：
  - `cd p_backend && make pb`
  - `cd p_backend && make verify`

## 5. DDL 规范（TABLE）
- 数据库结构以 `docs/model_sql/*.sql` 为唯一权威。
- 禁止依赖 `AutoMigrate` 或 gorm tag 自动迁移作为生产方案。
- 新增表必须明确：主键、唯一约束、必要索引、审计字段。
- 建表建议显式声明 `utf8mb4` 字符集。
- 软删除统一语义：`deleted_at BIGINT NOT NULL DEFAULT 0`；查询未删除数据使用 `deleted_at = 0`。
- 若业务字段允许“软删除后重建”，唯一索引必须包含 `deleted_at`。

## 6. API 规范
- HTTP 状态码：按项目约定仅使用 `200/500`；业务成败由 body `code/message` 表达。
- 新增接口必须定义稳定错误码与可控错误信息，不泄露内部细节。
- GET 列表接口必须复用 `pkg/xfiber.ParsePageSortQuery` 解析分页/排序参数。
- 导出接口若含枚举字段，输出必须转换为业务文案（不可直接导出内部枚举值）。
- 关键接口需具备重试安全（幂等）。

## 7. DAO（Repo）规范
- Repo 只处理 SQL/ORM 访问，不处理权限与流程判断。
- 查询条件与索引保持一致，避免全表扫描。
- 统一包装数据库错误（项目已有 `xerr` 体系）。
- 涉及跨表写入时，事务在 service 层组织，repo 提供原子操作函数。

## 8. DOCS 规范

### 8.1 API 文档（`docs/api_docs`）
每个新增/变更 API 至少补齐：
- 路径与方法
- 鉴权要求
- 请求参数（query/path/body）
- 响应示例（`code/message/data`）
- 错误场景说明

### 8.2 SQL 文档（`docs/model_sql`）
每个 SQL 文件至少包含：
- 变更目的（注释）
- DDL/DML 正文
- 回滚指引（注释）

## 9. 本仓库执行命令（直接可用）
- 生成 proto：`cd p_backend && make pb`
- 全量校验：`cd p_backend && make verify`
- 快速校验：`cd p_backend && make verify-fast`
- Auth 定向测试：`cd p_backend && make verify-auth`
- 执行 SQL：`cd p_backend && make execsql SQL=<file.sql>`
- 查看表结构：`cd p_backend && make showtable TABLE=<table>`
- 检查禁止项：`cd p_backend && rg -n "AutoMigrate" ./cmd ./internal ./pkg ./test`

## 10. 完成定义（DoD）
一次后端任务完成必须同时满足：
- 代码改动与需求一致，且分层职责清晰。
- 涉及 DB 变更时，SQL 已提交并可执行验证。
- 涉及 API 变更时，proto + handler/service/repo + docs/api_docs + test 全部同步。
- `make verify` 通过；若失败，必须说明阻塞项与风险。

## 11. 禁止事项
- 未写 SQL 就改表结构并声称完成。
- 新增 API 但不改 proto、不补 `docs/api_docs`、不补测试。
- 在 handler/repo 写跨层业务逻辑。
- 发现报错不先看日志（`p_backend/.ignore_logs`）就盲改。

## 12. 每次交付输出格式
每次任务结束必须给出：
- 变更文件清单
- 执行过的命令清单
- 每条命令结果（通过/失败）
- 未通过项与阻塞原因（如有）
