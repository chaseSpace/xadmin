# ACCOUNT HTTP API DOC

## GET /v1/account/me/profile
- 描述：获取当前登录账号简要信息（头像、昵称等）。
- 鉴权：是。

### Response 200
```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "uid": 10001,
    "username": "admin",
    "display_name": "系统管理员",
    "avatar": "https://example.com/avatar.png",
    "email": "admin@example.com",
    "phone": "13800000000",
    "menu_routes": [
      "/business/users",
      "/business/user-punishments"
    ],
    "menu_items": [
      {
        "id": 5,
        "parent_id": 0,
        "name": "业务管理",
        "route_path": "",
        "permission_key": "business.root",
        "sort": 50,
        "icon": "database",
        "children": [
          {
            "id": 6,
            "parent_id": 5,
            "name": "用户列表（demo）",
            "route_path": "/business/users",
            "permission_key": "business.users.view",
            "sort": 60,
            "icon": "usergroup-add",
            "children": []
          }
        ]
      }
    ],
    "warm_tip": {
      "id": 3,
      "tip_type": "positive",
      "content_zh": "今天也把重要的事推进一点",
      "content_en": "Move one important thing forward today"
    }
  }
}
```

## GET /v1/account/me/settings
- 描述：获取当前登录账号个人设置。
- 鉴权：是。

### Response 200
```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "limit_single_login": true,
    "background_image_url": "https://example.com/bg.avif",
    "locale": "zh-CN",
    "global_background_apply_enabled": true,
    "avatar": "https://example.com/avatar.png",
    "warm_tip_interval_minutes": 1440
  }
}
```

## POST /v1/account/me/settings
- 描述：更新当前登录账号个人设置。
- 鉴权：是。
- 行为：
  - `limit_single_login` 打开后，仅影响当前账号，并立即强制下线该账号其他会话（保留当前会话）。
  - 登录时也会按该开关执行单点登录约束。
  - `warm_tip_interval_minutes` 控制右上角关怀提示自动切换时间，允许值：`10|60|360|720|1440`。

### Request Body
```json
{
  "limit_single_login": true,
  "background_image_url": "https://example.com/bg.avif",
  "locale": "zh-CN",
  "global_background_apply_enabled": true,
  "avatar": "https://example.com/avatar.png",
  "warm_tip_interval_minutes": 1440
}
```

### Response 200
```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "limit_single_login": true,
    "background_image_url": "https://example.com/bg.avif",
    "locale": "zh-CN",
    "global_background_apply_enabled": true,
    "avatar": "https://example.com/avatar.png",
    "warm_tip_interval_minutes": 1440
  }
}
```
