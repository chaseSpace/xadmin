# p_frontend AGENTS.MD

本文件定义前端 AI/人工协作的统一开发范式。目标是让每次改动可追踪、可验证，并与后端 API/表结构文档保持一致。

核心链路：
`需求 -> API 契约(OpenAPI/Proto/后端文档) -> services/api -> 页面与组件 -> 测试 -> 构建校验 -> 交付说明`

## 1. 适用范围
- 项目：`p_frontend`
- 栈：React + TypeScript + Vite + Ant Design + TanStack Query + TanStack Router + Zod
- 适用任务：页面开发、接口联调、交互优化、缺陷修复、前端文档维护

## 2. 当前项目结构（以仓库现状为准）
- 应用入口：`src/main.tsx`、`src/app/App.tsx`
- 路由：`src/app/router.tsx`
- 页面：`src/pages/*`
- 布局：`src/app/layouts/AdminLayout.tsx`
- API 层：`src/services/api/*`
- 类型：`src/services/types/openapi.generated.ts`、`src/services/types/*`
- Schema：`src/services/schemas/*`
- UI 封装：`src/components/ui/*`
- 全局样式/主题：`src/styles/*`
- 状态：`src/store/*`
- OpenAPI 文件：`openapi/openapi.json`
- UI 设计规范：`UI_DESIGN.MD`
- 前端执行规范补充：`docs/AI_EXECUTION_RULES.md`

## 3. 开发分层约束

### 3.1 页面层（`src/pages`）
- 负责渲染、交互编排、调用 query/mutation。
- 禁止在页面中直接拼接底层请求细节（URL、错误归一化、鉴权处理）。

### 3.2 API 层（`src/services/api`）
- 统一通过 `apiClient` 发请求。
- 响应成功与否以响应体 `code` 为准（项目后端约定），不能只看 HTTP 状态码。
- 参数映射、响应适配、错误归一化在 API 层完成。

### 3.3 类型与校验层
- 接口类型优先使用 `openapi.generated.ts`。
- 业务表单与关键写操作必须走 zod schema 校验。

### 3.4 UI 组件层
- 基础组件优先复用 `src/components/ui` 二次封装。
- 禁止在页面中重复造通用轮子（表格、弹窗、空态、错误态、抽屉）。

## 4. API 开发范式（前端视角）

### Step 1: 先对齐后端契约
- 优先依据：
  1) `openapi/openapi.json`
  2) 后端 `docs/api_docs/*.MD`
  3) 已生成类型 `src/services/types/openapi.generated.ts`
- 若契约不一致，先确认契约再写页面逻辑。

### Step 2: 更新/生成类型
- OpenAPI 更新后必须执行：
  - `cd p_frontend && pnpm gen:types`
- 禁止在业务代码中滥用 `any` 绕过类型。

### Step 3: 在 `services/api` 实现接口函数
- 使用 `apiClient`，保留统一拦截器能力（token、trace id、全局错误处理）。
- 列表接口统一使用 `page_no/page_size/order_field/order_type` 参数。
- 返回值在 API 层做结构归一，页面拿到可直接渲染的数据结构。

### Step 4: 页面接入 Query
- 查询统一通过 `useQuery`；写操作通过 `useMutation`。
- Query Key 必须稳定、可复用，避免魔法字符串散落。
- 列表页必须服务端分页，不做本地假分页。

### Step 5: 错误与鉴权处理
- 接口错误统一走 `normalizeApiError`。
- 401 场景统一走未授权处理链路（`registerUnauthorizedHandler`）。

## 5. TABLE/DDL 变更联动要求（前端必须配合）
- 当后端 `docs/model_sql` 导致字段/枚举/约束变化时，前端必须同步：
  - 类型定义（OpenAPI 生成或手写类型）
  - 表单 schema（zod）
  - 列表列与详情展示
  - 导入导出模板/提示文案（若受影响）
- 未确认后端 SQL 与 API 已发布前，不得臆测字段上线状态。

## 6. DOCS 规范（前端）
- 若改动影响接口使用约定、错误码认知、联调方式，需要同步前端文档（当前在 `docs/`）。
- 若仅样式调整且无行为变更，可不额外加文档，但要在交付说明中标明影响范围。

## 7. UI/交互规范（当前项目约束）
- 基础组件库以 Ant Design 体系为主，优先复用 `src/components/ui`。
- 新增或调整页面 UI 前，必须遵守 `UI_DESIGN.MD`；其中关于表格页滚动模型、页面底部安全间距、分页到底部距离、树表容器等高等规则优先作为同类页面的设计约束。
- 页面禁止硬编码大量样式常量；优先使用 `src/styles/tokens.css` 与主题变量。
- 表格列表页需控制滚动区域，避免内容撑破布局。
- 动效可用但应克制，优先保证信息密度与操作效率。

## 8. 测试与校验（必须执行）
在 `p_frontend` 目录执行：
- `pnpm lint`
- `pnpm typecheck`
- `pnpm test`
- `pnpm build`

或执行一键命令：
- `pnpm check:all`

规则：
- 任一失败必须先修复再交付。
- 不允许跳过失败项直接宣称完成。

## 9. 典型任务模板

### 9.1 新增 CRUD 页面
1. 对齐 API 契约与类型（必要时 `pnpm gen:types`）
2. 在 `src/services/api` 增加 list/detail/create/update/delete
3. 定义或复用 query keys
4. 页面接入查询、表单校验、提交与刷新
5. 补充对应测试（至少 API 层/关键页面交互）
6. 跑 `pnpm check:all`

### 9.2 接口字段变更
1. 更新类型与 schema
2. 修正 API 适配函数
3. 修正页面渲染与交互
4. 回归相关测试并跑全量检查

## 10. 完成定义（DoD）
一次前端任务完成必须同时满足：
- 功能实现与需求一致。
- API 层、页面层、类型层保持一致。
- 校验命令通过（至少 `lint + typecheck + test + build`）。
- 输出包含：变更文件、命令结果、剩余风险（如有）。

## 11. 禁止事项
- 页面直接 `fetch/axios` 绕过 `services/api/client.ts`。
- 仅改页面不改类型/校验，造成运行时漂移。
- 将后端业务错误当 HTTP 成功直接吞掉。
- 未执行校验命令就提交“已完成”。

## 12. 每次交付输出格式
每次任务结束必须给出：
- 变更文件清单
- 关键实现点
- 校验命令与结果
- 已知风险与后续建议（如有）
