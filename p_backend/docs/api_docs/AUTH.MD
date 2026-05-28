# AUTH HTTP API DOC

## POST /v1/auth/login
- 描述：后台账号登录并创建会话。
- 鉴权：否。

### Request Body
```json
{
  "username": "admin",
  "password": "******"
}
```

### Response 200
```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "access_token": "jwt-token",
    "expires_at": "2026-04-20T12:00:00+08:00",
    "uid": 10001,
    "username": "admin",
    "display_name": "系统管理员",
    "avatar": "https://cdn.example.com/avatar/admin.png",
    "session_id": "d9ff4f6e-2cec-4fbb-b4cf-443503418b72"
  }
}
```

## POST /v1/auth/logout
- 描述：当前会话登出（撤销当前 token 对应会话）。
- 鉴权：是（`Authorization: Bearer <access_token>`）。

### Response 200
```json
{
  "code": 200,
  "message": "CodeSuccess",
  "data": null
}
```

## POST /v1/auth/logout_others
- 描述：注销当前账号其他设备会话（保留当前会话）。
- 鉴权：是。

### Response 200
```json
{
  "code": 200,
  "message": "CodeSuccess",
  "data": null
}
```

## POST /v1/auth/force_logout
- 描述：管理员强制下线目标用户全部会话。
- 鉴权：是。

### Request Body
```json
{
  "target_uid": 10002
}
```

### Response 200
```json
{
  "code": 200,
  "message": "CodeSuccess",
  "data": null
}
```

## POST /v1/auth/deactivate
- 描述：注销目标账号并撤销其全部会话。
- 鉴权：是。

### Request Body
```json
{
  "target_uid": 10002
}
```

### Response 200
```json
{
  "code": 200,
  "message": "CodeSuccess",
  "data": null
}
```

## GET /v1/auth/sessions
- 描述：查看当前登录用户会话列表。
- 鉴权：是。
- Query 参数：
  - `status`（可选）：`active` / `revoked` / `expired`。
  - `page_size`（可选）：1-100，默认 20。

### Response 200
```json
{
  "code": 200,
  "message": "ok",
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

## GET /v1/account/system/settings
- 描述：获取系统设置。
- 鉴权：是。
- 说明：`server_timezone` 读取启动配置中的 `app.server.timezone`，只读返回，前端用于时间展示与自动时区切换判断。

### Response 200
```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "site_name": "XAdmin 管理后台",
    "locale": "zh-CN",
    "timezone": "Asia/Shanghai",
    "server_timezone": "Asia/Shanghai",
    "login_lock_threshold": 5,
    "password_min_length": 8,
    "session_timeout_minutes": 30,
    "password_policy": ["uppercase", "number"],
    "global_watermark_enabled": false,
    "global_watermark_font_size": 16
  }
}
```
