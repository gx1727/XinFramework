# HTTP API 参考

> 当前共 **100+ 路由**。Base path: `/api/v1`,分 `public`(游客可访问) 和 `protected`(必须登录)。

## 通用约定

### 1. 响应格式

所有响应都是 JSON,固定三个字段:

```json
{ "code": 0, "msg": "ok", "data": { /* 业务数据 */ } }
```

- `code = 0`:成功
- `code != 0`:业务错误(具体含义见错误码表)
- `code >= 5000`:服务端错误(走 `resp.Error` 走 HTTP 500)
- `code >= 4000`:权限不足(HTTP 403)
- `code >= 3000`:资源不存在(HTTP 404)
- `code >= 2000`:参数错误(HTTP 400)

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

`protected` 路由需要:

```
Authorization: Bearer <jwt>
```

JWT 由 `POST /api/v1/auth/login` 颁发。

### 4. 多租户 Header

可选 header:`X-Tenant-ID`。如果登录用户 JWT 中 `TenantID == 0`,框架会用此 header 作为兜底(仅 public 路由生效)。

### 5. 错误码分段

| 区段 | module |
|---|---|
| 0 | 成功 |
| 1001-1999 | auth |
| 2001-2999 | user |
| 3001-3999 | tenant |
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

---

## 健康检查

### GET `/health`

无需认证。

```bash
curl http://localhost:8087/api/v1/health
```

**响应 200**:

```json
{ "code": 0, "msg": "ok", "data": { "status": "ok" } }
```

---

## 认证 (/auth/*)

### POST `/auth/login`

无需认证。

**请求**:

```json
{
  "account": "admin",       // username / phone / email 任意一种
  "password": "your-password",
  "tenant_code": "default"   // 必填
}
```

**响应 200**:

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

**错误码**:`1001` 账号不存在 / `1002` 密码错误 / `1003` 账号被禁用 / `1004` 租户不存在 / `1005` 租户被禁用 / `1006` 用户未绑定到该租户。

### POST `/auth/register`

无需认证。

**请求**:

```json
{
  "account": "newuser",   // username / phone / email
  "password": "...",
  "phone": "13900139000",  // 任一即可
  "email": "...",
  "tenant_code": "default"
}
```

**响应 200**:

```json
{ "code": 0, "msg": "ok", "data": { "user_id": 42 } }
```

**错误码**:`1010` 账号已存在 / `1011` 密码强度不足。

### POST `/auth/refresh`

无需认证(用 refresh_token 换新 access_token)。

**请求**:

```json
{ "refresh_token": "..." }
```

**响应**:同 `/auth/login`。

### POST `/auth/logout`

需要登录。撤销当前 session,后续 token 立即失效。

**请求**:无 body。

**响应 200**:

```json
{ "code": 0, "msg": "ok", "data": null }
```

---

## 用户 (/users/*)

### GET `/users`

需要登录 + `user:list` 权限。

**Query**:

| 参数 | 类型 | 说明 |
|---|---|---|
| `tenant_id` | int | (super_admin 可跨租户查询) |
| `keyword` | string | 模糊匹配 real_name / nickname / code |
| `page` | int | 默认 1 |
| `size` | int | 默认 20,最大 200 |

**响应**:分页。

### POST `/users`

需要登录 + `user:create` 权限。

**请求**:

```json
{
  "tenant_id": 1,
  "account_id": 5,            // 已存在的 account.id
  "code": "u001",
  "real_name": "张三",
  "nickname": "zhangsan",
  "phone": "13900139000",
  "email": "zhang@example.com",
  "org_id": 3,
  "role_ids": [1, 2]
}
```

### GET `/users/:id`

需要登录 + `user:list` 权限。

### PUT `/users/:id`

需要登录 + `user:update` 权限。**整体替换**。

### PATCH `/users/:id`

需要登录 + `user:update` 权限。**部分更新**(只更新提供的字段)。

### PUT `/users/:id/status`

需要登录 + `user:update` 权限。

**请求**:

```json
{ "status": 1 }   // 1=启用, 0=禁用
```

### PUT `/users/:id/org`

需要登录 + `user:update` 权限。**调岗**。

**请求**:

```json
{ "org_id": 5 }
```

### GET `/user/profile`

需要登录。返回**当前**用户的 profile。

### POST `/user/upload-avatar`

需要登录。multipart/form-data,字段名 `file`。

### PUT `/user/profile`

需要登录。更新当前用户的 profile(real_name / nickname / email / phone / avatar)。

---

## 角色 (/roles/*)

### GET `/roles`

需要登录 + `role:list`。

### POST `/roles`

需要登录 + `role:create`。

**请求**:

```json
{
  "tenant_id": 1,
  "code": "admin",
  "name": "管理员",
  "description": "...",
  "data_scope": { "type": 1, "org_ids": [] }
}
```

`data_scope.type` 取值见 [permissions.md](permissions.md#数据范围) 的 5 种类型。

### GET `/roles/:id/permissions`

需要登录 + `role:list`。返回该角色被授权的资源码(`resource:action` 列表)。

### POST `/roles/:id/permissions`

需要登录 + `role:update`。

**请求**:

```json
{ "resource_codes": ["user:create", "user:update", "flag:list"] }
```

### PUT `/roles/:id/permissions`

需要登录 + `role:update`。幂等:完全替换角色的资源码集合。

### GET `/roles/:id/data-scopes`

需要登录 + `role:list`。

### PUT `/roles/:id/data-scopes`

需要登录 + `role:update`。

**请求**:

```json
{ "type": 2, "org_ids": [3, 5, 7] }
```

### GET `/roles/:id/menus`

需要登录 + `role:list`。返回该角色的菜单 ID 列表。

### PUT `/roles/:id/menus`

需要登录 + `role:update`。

**请求**:

```json
{ "menu_ids": [1, 2, 5, 8] }
```

---

## 租户 (/tenants/*)

**所有租户管理路由额外要求 `super_admin` 平台角色**(普通租户内 admin 无法访问)。

### GET `/tenants`

需要登录 + `super_admin` 平台角色 + `tenant:list` 权限。

### POST `/tenants`

需要登录 + `super_admin` + `tenant:create`。

**请求**:

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
  "config": { /* 任意 JSON */ }
}
```

### POST `/tenants/:id/purge`

需要登录 + `super_admin` + `tenant:delete`。**硬删除**(物理 DELETE)。

---

## 菜单 (/menus/*)

### GET `/menus/tree`

需要登录 + `menu:list`。返回完整树形结构。

### POST `/menus`

需要登录 + `menu:create`。

**请求**:

```json
{
  "parent_id": 0,            // 0 = 顶级
  "code": "system",
  "name": "系统管理",
  "path": "/system",
  "icon": "settings",
  "sort": 100
}
```

---

## 组织 (/organizations/*)

### GET `/organizations/tree`

需要登录 + `organization:list`。返回完整组织树。

### POST `/organizations`

需要登录 + `organization:create`。

**请求**:

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

## 资源 (/resources/*)

### GET `/resources/my`

需要登录(不带 RBAC spec,任何登录用户都可调)。**返回当前用户能点哪些按钮**,前端用来动态渲染 UI。

**响应**:

```json
{
  "code": 0, "msg": "ok",
  "data": [
    { "code": "user", "action": "list" },
    { "code": "user", "action": "create" },
    { "code": "flag", "action": "list" }
  ]
}
```

### GET `/resources/by-menu/:menu_id`

需要登录 + `resource:list`。返回该菜单下挂的所有资源。

---

## 字典 (/dicts/*)

### GET `/dicts/:id/items`

需要登录 + `dict:list`。返回该字典的所有字典项(扁平列表)。

### POST `/dicts/:id/items`

需要登录 + `dict:update`。

**请求**:

```json
{
  "code": "active",
  "name": "启用",
  "sort": 1,
  "parent_id": 0
}
```

---

## 资产 (/asset/*)

### POST `/asset/upload`

需要登录 + `asset:create`。**multipart/form-data**,字段名 `file`。

**响应**:

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

存储后端由 `cfg.storage.provider` 决定:`local` 写到 `./uploads/`,`cos` 写到腾讯云 COS。

---

## 系统管理 (/system/*)

### GET `/system/server-info`

需要登录 + `system:list`。

**响应**:

```json
{
  "code": 0, "msg": "ok",
  "data": {
    "go_version": "go1.25.0",
    "start_time": "2026-06-18T08:33:06Z",
    "uptime_sec": 12345,
    "goroutines": 24,
    "mem_alloc_mb": 12.4,
    "db_status": "ok",
    "cache_status": "ok"
  }
}
```

### POST `/system/clear-cache`

需要登录 + `system:update`。清空所有 Redis cache(保留 session)。

### GET `/system/cache/info`

需要登录 + `system:list`。返回 Redis 服务器 info(`INFO` 命令的结果)。

### GET `/system/cache/keys`

需要登录 + `system:list`。

**Query**:

| 参数 | 类型 | 说明 |
|---|---|---|
| `pattern` | string | glob 模式,默认 `*` |
| `count` | int | 默认 100 |

### GET `/system/cache/value/*key`

需要登录 + `system:list`。**路径参数 key 用 `/` 分隔**(gin 的 `*key` 语法)。

### DELETE `/system/cache/keys/*key`

需要登录 + `system:update`。

---

## 微信 (/weixin/*)

### POST `/weixin/login`

无需认证。

**请求**:

```json
{ "code": "wx_jscode_here" }
```

**响应 200**:

```json
{
  "code": 0, "msg": "ok",
  "data": {
    "openid": "oXyz...",
    "session_key": "...",
    "token": "eyJhbGciOiJIUzI1NiIs..."   // 与 /auth/login 同 shape
  }
}
```

### POST `/weixin/phone`

无需认证。

**请求**:

```json
{ "code": "encrypted_data", "iv": "...", "encryptedData": "..." }
```

### POST `/weixin/bind-phone`

需要登录。

---

## 示例业务 (/cms/*)

展示 extapi 模式。CMS 模块**不直接连 DB**,通过 `extapi.Provider` 调用平台。

### GET `/cms/me`

需要登录。返回**当前登录用户**(走 `extapi.UserFacade.GetByID`)。

### GET `/cms/users`

需要登录。返回用户列表(走 `extapi.UserFacade.List`)。

### GET `/cms/tenant`

需要登录。返回当前租户(走 `extapi.TenantFacade.GetByID`)。

### GET/POST/PUT/DELETE `/cms/posts`

需要登录。CMS 自有 posts 表(在 `migrations/cms.sql`)。

---

## 示例业务 (/flag/*)

完整的 RBAC + DataScope 演示。

### GET `/flag/frames`

无需认证。公开浏览相框。

### GET `/flag/frames/:id`

无需认证。

### POST `/flag/frames`

需要登录 + `flag:create`。

### GET `/flag/avatars`

无需认证。公开浏览头像。

### POST `/flag/avatars`

需要登录 + `flag:create`。

### GET `/flag/my-avatars`

需要登录 + `flag:list`。**自动应用 DataScopeSelf**:只返回 `creator_id = 当前 userID` 的记录。

### POST `/flag/generate`

需要登录 + `flag:create`。触发头像生成流程(可能耗时)。

---

## 错误码示例

| 场景 | code | msg | HTTP |
|---|---|---|---|
| 成功 | 0 | ok | 200 |
| 账号密码错误 | 1002 | 账号或密码错误 | 200 |
| 未登录访问 protected | 401 | unauthorized | 401 |
| 权限不足 | 403 | permission denied: user:create | 403 |
| 资源不存在 | 3001 | 租户不存在 | 404 |
| 参数校验失败 | 2002 | 用户名不能为空 | 400 |
| 服务端异常 | 11001 | 服务器内部错误 | 500 |

---

## 中间件行为详解

### OptionalAuth(public 路由)

- 有 `Authorization` 头 → 解析并注入 `XinContext`(不查权限)
- 无 token 或解析失败 → 继续执行(把游客状态传给 handler)
- 如果 header 有 `X-Tenant-ID`,把它注入到 `XinContext.TenantID`

### Auth(protected 路由)

- 缺 `Authorization` 头 → 401 unauthorized
- token 过期 / session 被撤销 → 401 session expired or revoked
- token 签名错 / claims 异常 → 401 invalid token
- 校验通过 → 注入 `XinContext` + 注册 `UserContextLoader`(懒加载)

### Require(spec) / RequireAny / RequireAll

- `super_admin` 平台角色 → 直接放行
- `AuthOnly()` spec → 只要登录就放行
- 不满足 → 403 permission denied: `<resource>:<action>`

### RequirePlatformRole(roles...)

- 必须挂载在 `Auth` 中间件之后
- 持有任一指定平台角色 → 放行
- 不持有 → 403 需要平台级角色 / 平台角色不足

---

## SDK 示例(cURL)

完整登录 + 调用 protected 接口:

```bash
# 1. 登录
TOKEN=$(curl -s -X POST http://localhost:8087/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"account":"admin","password":"your-password","tenant_code":"default"}' \
  | jq -r '.data.token')

# 2. 调用 protected 接口
curl -s http://localhost:8087/api/v1/users \
  -H "Authorization: Bearer $TOKEN" \
  | jq

# 3. 注销
curl -s -X POST http://localhost:8087/api/v1/auth/logout \
  -H "Authorization: Bearer $TOKEN"
```

---

## 待补充

具体请求/响应字段需要参考:

- [modules.md](modules.md) — 每个 module 的路由清单 + spec
- 各 module 自己的 `handler.go` 里的 Request/Response struct
- [permissions.md](permissions.md) — Spec / DataScope 字段定义