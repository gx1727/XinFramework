# Flag App API 文档

> 头像框生成器 / 活动头像工具 API

## 基础信息

- **Base URL**: `/api/v1`
- **认证方式**: JWT Bearer Token（公开端点除外）
- **返回格式**: JSON `{ code, msg, data }`（`code=0` 表示成功）

---

## 头像框（Frames）

完整路由列表见 [../../doc/modules.md#flag](../../doc/modules.md#flag)。

### GET `/flag/frames`

无需认证（公开浏览）。支持 query：

| 参数 | 类型 | 说明 |
|---|---|---|
| `category_id` | uint | 按分类过滤 |
| `page` | int | 默认 1 |
| `size` | int | 默认 20 |

**响应**：

```json
{
  "code": 0, "msg": "ok",
  "data": {
    "list": [
      {
        "id": 1,
        "category_id": 1,
        "name": "程序员专属",
        "description": "适合程序员的简约头像框",
        "preview_url": "/frames/preview/programmer.png",
        "template_url": "/frames/template/programmer.png",
        "template_config": {
          "avatar_x": 100, "avatar_y": 100,
          "avatar_width": 300, "avatar_height": 300
        },
        "type": "public",
        "sort": 1,
        "status": 1,
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-01-01T00:00:00Z"
      }
    ],
    "total": 1
  }
}
```

### GET `/flag/frames/:id`

无需认证。

### POST `/flag/frames`

需要登录 + `flag:create`。

**Request Body**：

```json
{
  "category_id": 1,
  "name": "程序员专属",
  "description": "适合程序员的简约头像框",
  "preview_url": "/frames/preview/programmer.png",
  "template_url": "/frames/template/programmer.png",
  "template_config": "{\"avatar_x\":100,\"avatar_y\":100,\"avatar_width\":300,\"avatar_height\":300}",
  "type": "public",
  "sort": 1
}
```

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `name` | string | ✅ | 模板名称 |
| `category_id` | uint | ❌ | 分类 ID |
| `description` | string | ❌ | 模板描述 |
| `preview_url` | string | ❌ | 预览图 URL |
| `template_url` | string | ❌ | 模板底图 URL |
| `template_config` | string | ❌ | **JSON 字符串**，存为 `template_config JSONB`（SQL 显式 `::jsonb` cast） |
| `type` | string | ❌ | 类型：`public` / `private` / `space` |
| `sort` | int | ❌ | 排序号 |

### PUT `/flag/frames/:id`

需要登录 + `flag:update`。

```json
{
  "name": "更新后的名称",
  "description": "更新后的描述",
  "category_id": 2,
  "template_config": "{\"avatar_x\":120,...}",
  "type": "public",
  "sort": 10,
  "status": 1
}
```

### DELETE `/flag/frames/:id`

需要登录 + `flag:delete`。

---

## 头像框分类（Frame Categories）

### GET `/flag/frame-categories`

无需认证。

**响应**：

```json
{
  "code": 0, "msg": "ok",
  "data": [
    {
      "id": 1,
      "code": "emotion",
      "name": "情绪类",
      "sort": 1,
      "status": 1
    }
  ]
}
```

### POST `/flag/frame-categories`

需要登录 + `flag:create`。

```json
{
  "code": "emotion",
  "name": "情绪类",
  "sort": 1
}
```

### PUT `/flag/frame-categories/:id` / DELETE `/flag/frame-categories/:id`

需要登录 + `flag:update` / `flag:delete`。

---

## 活动空间（Spaces）

### GET `/flag/spaces/:code`

无需认证。按**邀请码**查公开活动。

**响应**：

```json
{
  "code": 0, "msg": "ok",
  "data": {
    "id": 1,
    "name": "测试活动",
    "description": "...",
    "frame_id": 1,
    "access_type": "public",
    "invite_code": "test",
    "max_usage": 100,
    "usage_count": 10,
    "status": 1
  }
}
```

### POST `/flag/spaces`

需要登录 + `flag:create`。

```json
{
  "name": "校庆100周年活动",
  "description": "庆祝学校建校100周年",
  "frame_id": 1,
  "access_type": "public",
  "start_at": "2026-05-01T00:00:00Z",
  "end_at": "2026-05-31T23:59:59Z"
}
```

### PUT `/flag/spaces/:id` / DELETE `/flag/spaces/:id`

需要登录 + `flag:update` / `flag:delete`。

### GET `/flag/spaces`

需要登录 + `flag:list`。返回当前用户可见的 spaces。

---

## 头像生成

### POST `/flag/generate`

需要登录 + `flag:create`。触发头像生成流程（可能耗时）。

```json
{
  "frame_id": 1,
  "space_id": 1,
  "source_image": "https://example.com/my-photo.jpg",
  "field_values": {
    "grade": "2024届",
    "college": "计算机学院"
  }
}
```

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `frame_id` | uint | ✅ | 头像框 ID |
| `space_id` | uint | ❌ | 关联的 Space ID |
| `source_image` | string | ✅ | 用户上传的原图 URL |
| `field_values` | object | ❌ | 动态字段值 |

**响应**：

```json
{
  "code": 0, "msg": "ok",
  "data": {
    "id": 1,
    "result_url": "https://img.gx1727.com/flag/1/abc123.png",
    "share_text": "我正在参加活动，快来一起玩！"
  }
}
```

### GET `/flag/my-avatars`

需要登录 + `flag:list`。**自动应用 DataScopeSelf**：只返回 `creator_id = 当前 userID` 的记录。

---

## 头像分类（Avatar Categories）

### GET `/flag/avatar-categories`

无需认证。

**Query**：

| 参数 | 类型 | 说明 |
|---|---|---|
| `type` | string | `public` / `custom` |

**响应**：

```json
{
  "code": 0, "msg": "ok",
  "data": [
    {
      "id": 1,
      "code": "selfie",
      "name": "自拍头像",
      "sort": 1,
      "status": 1
    }
  ]
}
```

### POST `/flag/avatar-categories`

需要登录 + `flag:create`。

```json
{
  "code": "my-category",
  "name": "我的分类",
  "sort": 10
}
```

### PUT `/flag/avatar-categories/:id` / DELETE `/flag/avatar-categories/:id`

需要登录 + `flag:update` / `flag:delete`。

---

## 头像管理（Avatars）

### GET `/flag/avatars`

无需认证（公开浏览）。

**Query**：

| 参数 | 类型 | 说明 |
|---|---|---|
| `category_id` | uint | 分类 ID |
| `user_id` | uint | 用户 ID |
| `type` | string | `custom` / `system` |
| `page` | int | 默认 1 |
| `size` | int | 默认 20 |

**响应**：

```json
{
  "code": 0, "msg": "ok",
  "data": {
    "list": [
      {
        "id": 1,
        "user_id": 1,
        "category_id": 1,
        "name": "我的头像1",
        "source_url": "/avatars/source/1.png",
        "thumbnail_url": "/avatars/thumb/1.png",
        "type": "custom",
        "is_public": true,
        "status": 1,
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-01-01T00:00:00Z"
      }
    ],
    "total": 1
  }
}
```

### GET `/flag/avatars/:id`

无需认证。

### POST `/flag/avatars`

需要登录 + `flag:create`。

```json
{
  "category_id": 1,
  "name": "我的头像",
  "source_url": "https://cos.example.com/avatars/123.jpg",
  "thumbnail_url": "https://cos.example.com/avatars/123_thumb.jpg",
  "is_public": true
}
```

### PUT `/flag/avatars/:id` / DELETE `/flag/avatars/:id`

需要登录 + `flag:update` / `flag:delete`。

---

## 错误码

| 错误码 | 说明 |
|---|---|
| 13001 | 头像框不存在 |
| 13002 | 活动空间不存在 |
| 13003 | 头像生成失败 |
| 13004 | 头像不存在 |

---

## 数据范围（DataScope）

`/flag/my-avatars` 自动应用 `DataScopeSelf`——只返回 `creator_id = 当前 userID` 的记录。详见 [../../doc/permissions.md](../../doc/permissions.md#3-数据范围-datascope)。

## JSONB 字段

`flag_frames.template_config` 是 JSONB。Go 端字段是 `string`，写入时 SQL 显式 `::jsonb` cast：

```go
// repository.go 实际写法
configStr := nullStr(frame.TemplateConfig)   // string 或 nil
q.Exec(ctx, `UPDATE flag_frames SET ... template_config = $7::jsonb ...`, ..., configStr, ...)
```

如果传了非法 JSON（非 `text/json` 字符串），PG 会报 `22P02 invalid input syntax for type json`。

## 调试

```bash
# 公开浏览相框
curl http://localhost:8087/api/v1/flag/frames

# 登录后创建相框
TOKEN=$(curl -s -X POST http://localhost:8087/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"account":"admin","password":"...","tenant_code":"default"}' \
  | jq -r '.data.token')

curl -X POST http://localhost:8087/api/v1/flag/frames \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "测试相框",
    "category_id": 1,
    "template_url": "/frames/template/test.png",
    "template_config": "{\"avatar_x\":100,\"avatar_y\":100,\"avatar_width\":300,\"avatar_height\":300}"
  }'
```
