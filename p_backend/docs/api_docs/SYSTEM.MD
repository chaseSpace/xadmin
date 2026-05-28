# SYSTEM HTTP API DOC

## API 列表
- `GET /v1/system/audit-logs`
- `GET /v1/system/audit-logs/:id`
- `GET /v1/system/audit-logs/export`
- `POST /v1/system/audit-logs/retention`
- `GET /v1/system/ip-blacklist`
- `GET /v1/system/ip-blacklist/creators`
- `POST /v1/system/ip-blacklist`
- `PUT /v1/system/ip-blacklist/:id`
- `POST /v1/system/ip-blacklist/:id/unblock`
- `DELETE /v1/system/ip-blacklist/:id`
- `POST /v1/system/ip-blacklist/unblock-batch`
- `POST /v1/system/ip-blacklist/import`
- `GET /v1/system/warm-tips`
- `POST /v1/system/warm-tips`
- `PUT /v1/system/warm-tips/:id`
- `DELETE /v1/system/warm-tips/:id`
- `GET /v1/system/alert-bots`
- `POST /v1/system/alert-bots`
- `DELETE /v1/system/alert-bots/:id`

## GET /v1/system/audit-logs
- 描述：分页查询操作审计日志。
- 鉴权：是。
- Query：
  - `page_no` `page_size`
  - `order_field`：`id|uid|actor|action|result|trace_id|source_ip|created_at`
  - `order_type`：`asc|desc`
  - `actor` `action` `result` `trace_id` `source_ip` `keyword`
  - `created_from` `created_to`：格式 `YYYY-MM-DD HH:mm:ss`
- 字段说明：`duration` 为后端记录的请求耗时，格式同 Go `time.Duration.String()`，历史数据可能为空。

### Response 200
```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": 1,
        "uid": 10001,
        "actor": "系统管理员",
        "action": "login_success",
        "result": "success",
        "trace_id": "trace-1",
        "source_ip": "127.0.0.1",
        "duration": "12.3ms",
        "user_agent": "Mozilla/5.0",
        "detail": "",
        "created_at": "2026-04-24 08:00:00"
      }
    ],
    "total": 1,
    "page": {
      "pn": 1,
      "ps": 10
    }
  }
}
```

## GET /v1/system/audit-logs/:id
- 描述：获取单条审计日志详情。
- 鉴权：是。

### Response 200
```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "id": 1,
    "uid": 10001,
    "actor": "系统管理员",
    "action": "login_success",
    "result": "success",
    "trace_id": "trace-1",
    "source_ip": "127.0.0.1",
    "duration": "12.3ms",
    "user_agent": "Mozilla/5.0",
    "detail": "",
    "created_at": "2026-04-24 08:00:00"
  }
}
```

## GET /v1/system/audit-logs/export
- 描述：导出当前筛选条件下的审计日志 CSV。
- 鉴权：是。
- Query：与 `GET /v1/system/audit-logs` 一致。

### Response 200
- `Content-Type: text/csv; charset=utf-8`
- `Content-Disposition: attachment; filename=system_audit_logs.csv`

## POST /v1/system/audit-logs/retention
- 描述：设置操作审计日志保留天数，并在确认后删除截止时间之前的过期数据。
- 鉴权：是。
- Body：
```json
{
  "retain_days": 30,
  "confirm": false
}
```
- 说明：
  - `confirm=false` 时仅返回预估的过期数据条数和剩余有效数据条数，不执行删除。
  - `confirm=true` 时若存在过期数据，将执行删除。

### Response 200
```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "success": false,
    "retain_days": 30,
    "expired_count": 12,
    "valid_count": 34,
    "cutoff_at": "2026-04-16 00:00:00"
  }
}
```

## GET /v1/system/ip-blacklist
- 描述：分页查询 IP 黑名单；返回前会根据 `status` 与 `end_at` 动态计算生效、到期失效、主动失效状态，不写入 DB。
- 鉴权：是。
- 生效机制：服务启动时将 `status=active` 的 IP 黑名单加载到内存 Store；请求中间件只查内存，不访问 DB。新增、编辑、主动失效、批量主动失效、导入成功后通过 Store `Add`/`AddMany`/`DeleteByID` 增量同步。
- 命中规则：按 Fiber `c.IP()` 获取请求 IP；同一 IP 允许存在多条记录，内存 Store 会按 IP 合并，只要任一记录当前时间处于 `start_at` 到 `end_at` 即返回业务码 `403`，并在内存中记录对应记录的命中增量。本地回环地址会同时匹配 `127.0.0.1` 与 `::1`，避免 IPv4/IPv6 回环不一致导致误判。
- 管理例外：`/v1/system/ip-blacklist` 及其子路径不受该中间件拦截，用于误封后查询和主动失效。列表返回的 `hit_count` 会合并 DB 值与当前进程内存命中增量。
- Query：
  - `page_no` `page_size`
  - `order_field`：`id|ip|ban_type|status|hit_count|updated_at`
  - `order_type`：`asc|desc`
  - `keyword`（可选）
  - `status`（可选：`active|inactive`，`inactive` 包含到期失效和主动失效）
  - `ban_type`（可选：`temp|permanent`）
  - `creator`（可选）

### Response Item
- `status`：`active` 生效，`expired` 到期失效，`manual_inactive` 主动失效。

## GET /v1/system/ip-blacklist/creators
- 描述：返回 IP 黑名单中所有不重复的创建者名称列表，用于前端筛选下拉框。
- 鉴权：是。
- Query：无。
- Response：
```json
{
  "code": 0,
  "message": "",
  "data": {
    "creators": ["admin", "operator"]
  }
}
```

## POST /v1/system/ip-blacklist
- 描述：新增 IP 黑名单记录。
- 鉴权：是。
- 校验：`ip` 必须为合法 IPv4 或 IPv6 地址，不支持 CIDR 网段。
- Body：
```json
{
  "ip": "1.2.3.4",
  "ban_type": 1,
  "start_at": "2026-04-28 12:00:00",
  "end_at": "2026-04-29 12:00:00",
  "reason": "manual ban",
  "creator": "10001"
}
```

## PUT /v1/system/ip-blacklist/:id
- 描述：编辑 IP 黑名单记录。
- 鉴权：是。
- Body：
```json
{
  "ban_type": 2,
  "end_at": "2099-12-31 00:00:00",
  "reason": "updated reason"
}
```

## POST /v1/system/ip-blacklist/:id/unblock
- 描述：单条主动失效 IP 黑名单记录。
- 鉴权：是。

## DELETE /v1/system/ip-blacklist/:id
- 描述：删除 IP 黑名单记录（软删除）。同时从内存 Store 移除。
- 鉴权：是。

## POST /v1/system/ip-blacklist/unblock-batch
- 描述：批量主动失效 IP 黑名单记录。
- 鉴权：是。
- Body：
```json
{
  "ids": [1, 2, 3]
}
```

## POST /v1/system/ip-blacklist/import
- 描述：批量导入 IP 黑名单。
- 鉴权：是。
- 校验：`ips` 每一项必须为合法 IPv4 或 IPv6 地址，不支持 CIDR 网段。
- Body：
```json
{
  "ips": ["1.2.3.4", "5.6.7.8"],
  "ban_type": 1,
  "duration_hours": 24,
  "end_at": "2026-05-01 18:30:00"
}
```
说明：
- `duration_hours` 与 `end_at` 二选一即可（临时封禁场景）。
- `ban_type = 2`（永久）时可不传两者。

## GET /v1/system/warm-tips
- 描述：分页查询右上角关怀提示文案。
- 鉴权：是。
- Query：
  - `page_no` `page_size`
  - `order_field`：`id|tip_type|sort|status|updated_at`
  - `order_type`：`asc|desc`
  - `keyword`：匹配中英文文案
  - `tip_type`：`rest|positive|quote|line`
  - `status`：`enabled|disabled`

### Response 200
```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": 1,
        "tip_type": "rest",
        "content_zh": "喝水伸展，眼睛休息一下",
        "content_en": "Drink water and rest your eyes",
        "sort": 10,
        "status": 1,
        "updated_at": "2026-05-12 14:00:00"
      }
    ],
    "total": 1,
    "page": {
      "pn": 1,
      "ps": 10
    }
  }
}
```

## POST /v1/system/warm-tips
- 描述：新增关怀提示。
- 鉴权：是。
- 校验：中英文文案均需控制在 3 到 20 个词内；无空格文本按字符校验，最多 40 个字符。
- Body：
```json
{
  "tip_type": "positive",
  "content_zh": "今天也把重要的事推进一点",
  "content_en": "Move one important thing forward today",
  "sort": 100,
  "status": 1
}
```

## PUT /v1/system/warm-tips/:id
- 描述：编辑关怀提示。
- 鉴权：是。
- Body：同 `POST /v1/system/warm-tips`。

## DELETE /v1/system/warm-tips/:id
- 描述：软删除关怀提示。
- 鉴权：是。

## GET /v1/system/alert-bots
- 描述：分页查询告警机器人列表。
- 鉴权：是。
- Query：
  - `page_no` `page_size`
  - `order_field`：`id|name|bot_type|enabled|created_at|updated_at`
  - `order_type`：`asc|desc`
  - `keyword`（可选，匹配 name/username）
  - `bot_type`（可选：`telegram|feishu`）
- Response Item：
  - `id` `name` `username` `token` `bot_type` `enabled` `scene_keys` `created_at` `updated_at`

## POST /v1/system/alert-bots
- 描述：新增或编辑告警机器人。id > 0 为编辑，否则新增。
- 鉴权：是。
- Body：
```json
{
  "id": 0,
  "name": "运维告警Bot",
  "username": "ops_bot",
  "token": "xxx",
  "bot_type": "telegram",
  "enabled": true,
  "scene_keys": ["server_down", "high_cpu"]
}
```

## DELETE /v1/system/alert-bots/:id
- 描述：软删除告警机器人。
- 鉴权：是。
