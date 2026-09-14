# XAdmin AI 协作指南

开始任何任务前，先阅读 [AI_TODO.md](AI_TODO.md) 中与当前任务有关的条目。任务清单与推进规则以该文件为准。

按改动范围继续阅读对应的专项说明：

- 前端：[`p_frontend/AGENTS.MD`](p_frontend/AGENTS.MD)
- 后端：[`p_backend/AGENTS.MD`](p_backend/AGENTS.MD)

涉及前后端联动时，两个专项说明都必须阅读（读过即可），并同步实现、测试和文档。

## 编码规则

- 以全栈视角完成前后端功能：满足核心需求，提供清晰的 UI 与交互，并保持 API 精简。
- 改动旧模块时，仔细核对影响范围；完成后同步相关前后端文档，并在交付说明中列出变更。
- 修改文件超过 10 个时，可以按逻辑拆分为多次英文说明的提交。
- 涉及数据表变更时，先编写临时 DDL SQL，执行后合并到原表定义，再删除临时 SQL 文件。
- 除非必要，不得使用 `make execsql` 执行全量 SQL；它会 DROP 会话表导致登录失效。仅可用 `psql` 执行单条 SQL。
- 前端常规开发不默认执行 `pnpm prettier`、`pnpm lint` 或 `pnpm build`；仅在用户明确要求时运行。
- 新页面包含表格时，必须先参考 [p_frontend/UI_DESIGN.MD](p_frontend/UI_DESIGN.MD)。
