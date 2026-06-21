# HTTP API 参考

> 当前共 **120+ 路由**。Base path: `/api/v1`，分 3 个语义空间：
> - `/api/v1/<resource>` — 业务消费（需登录 + ResX 权限）
> - `/api/v1/admin/<platform_resource>` — 平台管理（需 super_admin + ResX）
> - `/api/v1/public/<resource>` — 公开访问（OptionalAuth）
>
> 文档版本：2026-06（config 重构 + platform_menu/platform_tenant 后）

## 通用约定

### 1. 响应格式

所有响应都是 JSON，固定三个字段：

```json
{ "code": 0, "msg": "ok", "data": { /* 业务数据 */ } }
```

- `code = 0`：成功
- `code != 0`：业务错误（具体含义见错误码表）
- `code >= 5000`：服务端错误（走 `resp.Error`，HTTP 500）
- `code >= 4000`：权限不足（HTTP 403）
- `code >= 3000`：资源不存在（HTTP 404）
- `code >= 2000`：参数错误（HTTP 400）

### 2. 分页响应

```json
{
  "code": 0, "msg": "ok",
  "data": {
    "total": 142,
    "list": [ { /* row 1 */ }, { /* row 2 */ } ]
  }
}
```

### 3. 认证头

`protected` 路由需要：

```
Authorization: Bearer <jwt>
```

JWT 由 `POST /api/v1/auth/login` 颁发。

### 4. 多租户 Header

可选 header：`X-Tenant-ID`。如果登录用户 JWT 中 `TenantID == 0`，框架会用此 header 作为兜底（仅 public 路由生效）。

### 5. 错误码分段

| 区段 | module |
|---|---|
| 0 | 成功 |
| 1001-1999 | auth |
| 2001-2999 | user |
| 3001-3999 | tenant / platform_tenant |
| 4001-4999 | role |
| 5001-5999 | menu |
| 6001-6999 | organization |
| 7001-7999 | permission |
| 8001-8999 | resource |
| 9001-9999 | asset |
| 10001-10999 | dict |
| 11001-11999 | system |
| 12001-12999 | weixin |
| 13001-13999 | flag |
| 15001-15999 | platform_menu |
| 18001-18999 | config |

---

## 健康检查

### GET `/health`

无需认证。

```bash
curl http://localhost:8087/api/v1/health
```

**响应 200**：

```json
{ "code": 0, "msg": "ok", "data": { "status": "ok" } }
```

---

## 认证（/auth/*）

### POST `/auth/login`

无需认证。

**请求**：

```json
{
  "account": "admin",
  "password": "your-password",
  "tenant_code": "bootstrap"
}
```

**响应 200**：

```json
{
  "code": 0, "msg": "ok",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "...",
    "expires_at": 1700000000,
    "user": { "id": 1, "real_name": "管理员", "tenant_id": 1 }
  }
}
```

**错误码**：`1001` 账号不存在 / `1002` 密码错误 / `1003` 账号被禁用 / `1004` 租户不存在 / `1005` 租户被禁用 / `1006` 用户未绑定到该租户。

### POST `/auth/register`

无需认证。

**请求**：

```json
{
  "account": "newuser",
  "password": "...",
  "phone": "13900139000",
  "email": "...",
  "tenant_code": "bootstrap"
}
```

**响应 200**：

```json
{ "code": 0, "msg": "ok", "data": { "user_id": 42 } }
```

**错误码**：`1010` 账号已存在 / `1011` 密码强度不足。

### POST `/auth/refresh`

无需认证（用 refresh_token 换新 access_token）。

**请求**：

```json
{ "refresh_token": "..." }
```

**响应**：同 `/auth/login`。

### POST `/auth/logout`

需要登录。撤销当前 session。

**请求**：无 body。

**响应 200**：

```json
{ "code": 0, "msg": "ok", "data": null }
```

---

## 用户（/users/*, /user/*）

### GET `/users`

需要登录 + `user:list`。

**Query**：

| 参数 | 类型 | 说明 |
|---|---|---|
| `tenant_id` | int | (super_admin 可跨租户) |
| `keyword` | string | 模糊匹配 real_name / nickname / code |
| `page` | int | 默认 1 |
| `size` | int | 默认 20，最大 200 |

### POST `/users`

需要登录 + `user:create`。

**请求**：

```json
{
  "tenant_id": 1,
  "account_id": 5,
  "code": "u001",
  "real_name": "张三",
  "nickname": "zhangsan",
  "phone": "13900139000",
  "email": "zhang@example.com",
  "org_id": 3,
  "role_ids": [1, 2]
}
```

### GET `/users/:id` / PUT `/users/:id` / PATCH `/users/:id`

需要登录 + `user:list` / `user:update`。PATCH 是部分更新。

### PUT `/users/:id/status` / PUT `/users/:id/org`

需要登录 + `user:update`。

```json
{ "status": 1 }
{ "org_id": 5 }
```

### GET `/user/profile` / PUT `/user/profile` / POST `/user/avatar`

需要登录（不带 RBAC spec）。

---

## 角色（/roles/*）

### GET `/roles` / GET `/roles/:id` / POST `/roles` / PUT `/roles/:id` / PATCH `/roles/:id` / DELETE `/roles/:id`

需要登录 + `role:list` / `role:list` / `role:create` / `role:update` / `role:update` / `role:delete`。

**POST /roles 请求**：

```json
{
  "tenant_id": 1,
  "code": "admin",
  "name": "管理员",
  "description": "...",
  "data_scope": { "type": 1, "org_ids": [] }
}
```

### GET `/roles/:id/permissions` / POST `/roles/:id/permissions`

需要登录 + `role:list` / `role:update`。返回/接受资源码列表。

```json
{ "resource_codes": ["user:create", "user:update", "flag:list"] }
```

### GET `/roles/:id/data-scopes` / PUT `/roles/:id/data-scopes`

需要登录 + `role:list` / `role:update`。

```json
{ "type": 2, "org_ids": [3, 5, 7] }
```

### GET `/roles/:id/menus` / PUT `/roles/:id/menus`

需要登录 + `role:list` / `role:update`。

```json
{ "menu_ids": [1, 2, 5, 8] }
```

---

## 平台租户（/admin/platform-tenants）  ⭐

**所有路由额外要求 `super_admin` 平台角色**（在 `adminGroup` 分组上统一拦截），并叠加资源权限码。

### GET `/admin/platform-tenants`

需要登录 + `super_admin` + `tenant:list`。

### GET `/admin/platform-tenants/:id`

需要登录 + `super_admin` + `tenant:list`。

### POST `/admin/platform-tenants`

需要登录 + `super_admin` + `tenant:create`。

**请求**：

```json
{
  "code": "acme",
  "name": "ACME 公司",
  "contact": "张三",
  "phone": "13900139000",
  "email": "contact@acme.com",
  "province": "上海市",
  "city": "上海市",
  "area": "浦东新区",
  "address": "世纪大道 100 号",
  "config": { /* 任意 JSON，存为 tenants.config JSONB */ }
}
```

### PUT `/admin/platform-tenants/:id` / PUT `/admin/platform-tenants/:id/status`

需要登录 + `super_admin` + `tenant:update`。

### DELETE `/admin/platform-tenants/:id`

需要登录 + `super_admin` + `tenant:delete`。**软删**。

### POST `/admin/platform-tenants/:id/purge`

需要登录 + `super_admin` + `tenant:delete`。**硬删**（物理 DELETE，需先软删）。

---

## 平台菜单（/admin/platform-menus）  ⭐

**所有路由额外要求 `super_admin` 平台角色**（group 级守卫，单层）。

### GET `/admin/platform-menus`

需要登录 + `super_admin`。返回平台菜单列表（`tenant_id = 0`）。

### GET `/admin/platform-menus/tree`

需要登录 + `super_admin`。返回树形。

### GET `/admin/platform-menus/:id` / POST / PUT / DELETE

需要登录 + `super_admin`。

```json
{
  "parent_id": 0,
  "code": "platform-config",
  "name": "平台配置",
  "path": "/admin/configs",
  "icon": "settings",
  "sort": 100,
  "type": "menu"
}
```

---

## 菜单（/menus/*）

### GET `/menus/tree`

需要登录 + `menu:list`。

### POST `/menus`

需要登录 + `menu:create`。

```json
{
  "parent_id": 0,
  "code": "system",
  "name": "系统管理",
  "path": "/system",
  "icon": "settings",
  "sort": 100
}
```

---

## 组织（/organizations/*）

### GET `/organizations/tree`

需要登录 + `organization:list`。

### POST `/organizations`

需要登录 + `organization:create`。

```json
{
  "parent_id": 1,
  "code": "tech-team",
  "name": "技术团队",
  "type": "department",
  "admin_code": "tech-admin",
  "sort": 1
}
```

---

## 资源（/resources/*）

### GET `/resources/my`

需要登录（不带 RBAC spec）。**返回当前用户能点哪些按钮**。

```json
{
  "code": 0, "msg": "ok",
  "data": [
    { "code": "user", "action": "list" },
    { "code": "flag", "action": "list" }
  ]
}
```

### GET `/resources/by-menu/:menu_id`

需要登录 + `resource:list`。

---

## 字典（/dicts/*）

### 业务路由（`/api/v1/dicts*`）

| Method | Path | Spec |
|---|---|---|
| GET | `/dicts` | `dict:list` |
| GET | `/dicts/:id` | `dict:get` |
| POST | `/dicts` | `dict:create` |
| PUT | `/dicts/:id` | `dict:update` |
| DELETE | `/dicts/:id` | `dict:delete` |
| GET | `/dicts/:id/items` | `dict:list` |
| POST | `/dicts/:id/items` | `dict:update` |
| PUT | `/dicts/:id/items/:item_id` | `dict:update` |
| DELETE | `/dicts/:id/items/:item_id` | `dict:update` |
| GET | `/dicts/resolve` | `dict:get` |
| POST | `/dicts/resolve/batch` | `dict:get` |
| PUT | `/dicts/:id/items/:item_id/override` | `dict:update` |
| DELETE | `/dicts/:id/items/:item_id/override` | `dict:update` |

### 平台路由（`/api/v1/dicts/platform*`，强制 super_admin）

| Method | Path |
|---|---|
| GET | `/dicts/platform/:id/items` |
| POST | `/dicts/platform/:id/items` |
| PUT | `/dicts/platform/:id/items/:item_id` |
| DELETE | `/dicts/platform/:id/items/:item_id` |
| GET | `/dicts/platform/:id/visibility` |
| POST | `/dicts/platform/:id/visibility` |
| DELETE | `/dicts/platform/:id/visibility/:tenant_id` |

`dict_items.extend` 是 JSONB 列，SQL 显式 `::jsonb` cast。

---

## 配置中心（/configs/*, /public/configs）  ⭐ Phase 0022 重构

### 业务消费（`/api/v1/configs*`）

| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/configs` | `config:list` | ListGroups（按 tenant_id 过滤） |
| GET | `/configs/:id` | `config:get` | GetGroup（resolve 合并） |
| GET | `/configs/:id/items` | `config:list` | ListItemsByGroup |
| POST | `/configs/:id/items/:item_id/override` | `config:update` | UpsertOverride |
| DELETE | `/configs/:id/items/:item_id/override` | `config:update` | DeleteOverride |
| GET | `/configs/resolve` | `config:list` | Resolve（`?code=site`） |
| POST | `/configs/resolve/batch` | `config:list` | ResolveBatch |

### 平台管理（`/api/v1/configs/platform*`，强制 super_admin）

| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/configs/platform` | `config:list` | ListGroups |
| GET | `/configs/platform/:id` | `config:get` | GetGroup |
| POST | `/configs/platform` | `config:create` | CreateGroup |
| PUT | `/configs/platform/:id` | `config:update` | UpdateGroup |
| DELETE | `/configs/platform/:id` | `config:delete` | DeleteGroup |
| GET | `/configs/platform/:id/items` | `config:list` | ListItems |
| POST | `/configs/platform/:id/items` | `config:create` | CreateItem |
| PUT | `/configs/platform/:id/items/:item_id` | `config:update` | UpdateItem |
| DELETE | `/configs/platform/:id/items/:item_id` | `config:delete` | DeleteItem |
| GET | `/configs/platform/:id/visibility` | `config:list` | ListVisibility |
| POST | `/configs/platform/:id/visibility` | `config:update` | UpsertVisibility |
| DELETE | `/configs/platform/:id/visibility/:tenant_id` | `config:update` | DeleteVisibility |

### 公开读（`/api/v1/public/configs`，无需 auth）

| Method | Path | 说明 |
|---|---|---|
| GET | `/public/configs` | 按 group + key 取公开项；公开判定基于 group `visibility = 'public'` |

### 关键请求示例

**POST /configs/platform**（创建 platform group）：

```json
{
  "code": "site",
  "name": "站点配置",
  "description": "...",
  "scope": "platform",
  "visibility": "public",
  "is_public": true
}
```

**POST /configs/platform/:id/items**（创建 platform item）：

```json
{
  "key": "site.theme",
  "value": "dark",
  "default_value": "light",
  "type": "string",
  "label": "主题",
  "description": "...",
  "options": [{"label": "深色", "value": "dark"}],
  "validation": {"min": 0, "max": 100},
  "sort": 1,
  "is_public": true,
  "is_readonly": false,
  "is_system": false
}
```

**POST /configs/:id/items/:item_id/override**（租户覆盖）：

```json
{ "value": "light" }
```

**POST /configs/platform/:id/visibility**（平台限定租户可见性）：

```json
{ "tenant_id": 5, "access": "editable" }
```

**GET /configs/resolve?code=site**（业务合并消费）：

```json
{
  "code": 0, "msg": "ok",
  "data": {
    "group": "site",
    "values": {
      "site.theme": "light",
      "site.lang": "zh-CN"
    }
  }
}
```

**错误码**（18001-18019）：

| Code | 含义 |
|---|---|
| 18001 | config_group 不存在 |
| 18002 | config_group code 已存在 |
| 18003 | config_group 下有 item，无法删除 |
| 18004 | 系统 group 不可修改 |
| 18005 | config_item 不存在 |
| 18006 | config_item key 已存在 |
| 18007 | item 为只读 |
| 18008 | 系统 item 不可修改 |
| 18009 | item type 取值非法 |
| 18010 | item value 与 type 不匹配 |
| 18011 | item value 不在 options 范围 |
| 18012 | 平台 group 不可被租户修改 |
| 18013 | item 有租户覆盖，无法删除 |
| 18014 | override 指向不存在的 platform item |
| 18015 | config_group 对当前租户不可见 |
| 18016 | config_group 为只读 |
| 18017 | access 取值非法 |
| 18018 | visibility 取值非法 |
| 18019 | resolve 失败 |

---

## 资产（/asset/*）

### POST `/asset/upload`

需要登录 + `asset:create`。**multipart/form-data**，字段名 `file`。

**响应**：

```json
{
  "code": 0, "msg": "ok",
  "data": {
    "id": 42,
    "url": "/uploads/2026/06/abc.png",
    "size": 102400,
    "mime": "image/png"
  }
}
```

存储后端由 `cfg.storage.provider` 决定。

### DELETE `/asset/:id`

需要登录 + `asset:delete`。

---

## 系统管理（/system/*）

### GET `/system/server-info`

需要登录 + `system:list`。

```json
{
  "code": 0, "msg": "ok",
  "data": {
    "go_version": "go1.25.0",
    "start_time": "2026-06-21T08:33:06Z",
    "uptime_sec": 12345,
    "goroutines": 24,
    "mem_alloc_mb": 12.4,
    "db_status": "ok",
    "cache_status": "ok"
  }
}
```

### POST `/system/clear-cache`

需要登录 + `system:update`。清空 Redis cache（保留 session）。

### GET `/system/cache/info` / `/system/cache/keys`

需要登录 + `system:list`。keys 支持 `?pattern=&count=`。

### GET `/system/cache/value/*key` / DELETE `/system/cache/keys/*key`

需要登录 + `system:list` / `system:update`。

---

## 微信（/weixin/*）

### POST `/weixin/login`

无需认证。

```json
{ "code": "wx_jscode_here" }
```

**响应 200**：

```json
{
  "code": 0, "msg": "ok",
  "data": {
    "openid": "oXyz...",
    "session_key": "...",
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

### POST `/weixin/phone` / `/weixin/bind-phone`

phone 无需认证（前端用），bind-phone 需要登录。

---

## 示例业务（/cms/*）

展示 extapi 模式。CMS 不直接连 RBAC 表。

### GET `/cms/me` / `/cms/users` / `/cms/tenant`

需要登录。返回当前用户 / 用户列表 / 当前租户（走 extapi）。

### GET/POST/PUT/DELETE `/cms/posts`

需要登录。CMS 自有 posts 表。

---

## 示例业务（/flag/*）

完整 RBAC + DataScope 演示。详见 [apps/flag/doc/api.md](../apps/flag/doc/api.md)。

### GET `/flag/frames`

无需认证。公开浏览相框。

### POST `/flag/frames`

需要登录 + `flag:create`。

### GET `/flag/my-avatars`

需要登录 + `flag:list`。**自动应用 DataScopeSelf**：只返回 `creator_id = 当前 userID` 的记录。

### POST `/flag/generate`

需要登录 + `flag:create`。触发头像生成流程。

---

## 中间件行为详解

### OptionalAuth（public 路由）

- 有 `Authorization` 头 → 解析并注入 `XinContext`（不查权限）
- 无 token 或解析失败 → 继续执行（把游客状态传给 handler）
- 如果 header 有 `X-Tenant-ID`，把它注入到 `XinContext.TenantID`

### Auth（protected 路由）

- 缺 `Authorization` 头 → 401 unauthorized
- token 过期 / session 被撤销 → 401 session expired or revoked
- token 签名错 / claims 异常 → 401 invalid token
- 校验通过 → 注入 `XinContext` + 注册 `UserContextLoader`（懒加载）

### Require(spec) / RequireAny / RequireAll

- `super_admin` 平台角色 → 直接放行
- `AuthOnly()` spec → 只要登录就放行
- 不满足 → 403 permission denied: `<resource>:<action>`

### RequirePlatformRole(roles...)

- 必须挂载在 `Auth` 中间件之后
- 持有任一指定平台角色 → 放行
- 不持有 → 403 需要平台级角色

---

## SDK 示例（cURL）

完整登录 + 调用 protected 接口：

```bash
# 1. 登录
TOKEN=$(curl -s -X POST http://localhost:8087/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"account":"admin","password":"your-password","tenant_code":"bootstrap"}' \
  | jq -r '.data.token')

# 2. 调用业务接口
curl -s http://localhost:8087/api/v1/users \
  -H "Authorization: Bearer $TOKEN" | jq

# 3. 平台管理（需 super_admin）
curl -s http://localhost:8087/api/v1/admin/platform-tenants \
  -H "Authorization: Bearer $TOKEN" | jq

# 4. 配置合并消费
curl -s "http://localhost:8087/api/v1/configs/resolve?code=site" \
  -H "Authorization: Bearer $TOKEN" | jq

# 5. 公开配置（无需 token）
curl -s http://localhost:8087/api/v1/public/configs | jq

# 6. 注销
curl -s -X POST http://localhost:8087/api/v1/auth/logout \
  -H "Authorization: Bearer $TOKEN"
```

---

## 待补充

具体请求/响应字段需要参考：

- [modules.md](modules.md) — 每个 module 的路由清单 + spec
- 各 module 自己的 `handler.go` 里的 Request/Response struct
- [permissions.md](permissions.md) — Spec / DataScope 字段定义