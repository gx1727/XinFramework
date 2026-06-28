# API 契约

> 本文件描述 XinFramework 的 HTTP API 规范：响应格式、错误码、鉴权、跨域、关键端点。
> 完整路由清单见 [modules.md](./modules.md)。

---

## 1. 统一响应

所有 API 返回 **JSON** 结构：

```json
{
  "code": 0,
  "msg": "ok",
  "data": { /* 业务数据 */ }
}
```

| 字段 | 类型 | 含义 |
|---|---|---|
| `code` | int | 业务码；`0` = 成功；其它按分段映射 HTTP 状态 |
| `msg` | string | 提示信息（中文） |
| `data` | any | 业务数据；`null` 表示无返回数据 |

**分页接口**的 `data` 结构：

```json
{
  "total": 100,
  "list": [ /* ... */ ],
  "page": 1,
  "size": 20
}
```

---

## 2. 错误码分段 → HTTP 状态码

`framework/pkg/resp/resp.go:CodeToHTTPStatus` 单点决定：

| 段 | HTTP | 用途 | 段常量 |
|---|---|---|---|
| `1xxx` | 200 | 鉴权 / 账号 / 通用业务 | `CodeAuth = 1000` |
| `2xxx` | 400 | 参数校验 / 业务规则 | `CodeUser = 2000` |
| `3xxx` | 404 | 资源不存在 / 租户级 | `CodeTenant = 3000` |
| `4xxx` | 403 | 权限不足 / 角色冲突 | `CodeRole = 4000` |
| `5xxx+` | 500 | 服务端故障 / 系统异常 | `CodeMenu = 5000` |

**模块码段分配**（`framework/pkg/resp/errors.go`）：

| 段 | 模块 | 段 | 模块 |
|---|---|---|---|
| 1000 | auth | 8000 | resource |
| 2000 | user | 9000 | asset |
| 3000 | tenant | 10000 | dict |
| 4000 | role | 11000 | system |
| 5000 | menu | 12000 | weixin |
| 6000 | organization | 13000 | flag |
| 7000 | permission | 14000 | cms |

**注意**：menu（5xxx）/ organization（6xxx）等模块的"资源不存在"类错误目前共用 5xxx+ 段，按现有规则会被映射为 HTTP 500。这是已知 gap（修复路径见各模块重新分配段位）。

---

## 3. 错误响应

```json
{
  "code": 4001,
  "msg": "角色不存在",
  "data": null
}
```

`msg` 用于直接展示给最终用户，**禁止**包含堆栈、SQL、内部路径等技术细节。

---

## 4. 鉴权

### 4.1 通用流程

```
1. 客户端 POST /auth/tenant-login 或 /auth/platform-login
2. 服务端验证账号/密码 → 签 access_token (1h) + refresh_token (24h)
3. 客户端用 access_token 调业务 API：
     Authorization: Bearer <access_token>
4. 401 时客户端用 refresh_token 调 /auth/refresh 换新 token
```

### 4.2 三种登录入口

| 端点 | scope | tenant_id | 适用 |
|---|---|---|---|
| `POST /auth/tenant-login` | `tenant` | 必填 | 业务用户登录 |
| `POST /auth/platform-login` | `platform` | 0 | super_admin 登录 |
| `POST /auth/login-precheck` | — | — | 多身份账号列出身份 |
| `POST /auth/select-tenant` | `tenant` | 必填 | precheck 后选身份签 token |
| `POST /auth/refresh` | 同原 | 可选 | 切租户时传 `tenant_id` |
| `POST /auth/logout` | — | — | 撤销 session |

### 4.3 多身份账号（路径 B）

**场景**：一个 `accounts` 记录对应多个 `tenant_users`（一个 admin 在多个租户都有身份）。

**登录流程**：

```
1. 客户端 POST /auth/login-precheck { account, password }
   → {
       account_id: 1,
       platform_available: true,
       platform_roles: ["super_admin"],
       tenant_identities: [
         { tenant_id: 1, tenant_code: "acme", tenant_name: "ACME",
           user_id: 10, user_code: "admin", role: "admin" },
         { tenant_id: 2, tenant_code: "globex", tenant_name: "Globex",
           user_id: 20, user_code: "admin", role: "admin" },
       ]
     }

2. 用户选择身份 → POST /auth/select-tenant
   { account, password, tenant_id: 1 }
   → 完整登录响应（同 /auth/tenant-login）

3. 登录后可调 /auth/refresh?tenant_id=2 无密码切到 Globex
```

**单身份账号**可直接调 `/auth/tenant-login`，跳过 precheck。

---

## 5. 关键端点示例

### 5.1 用户管理（租户域）

```http
GET /api/v1/users?page=1&size=20&keyword=admin&org_id=5
Authorization: Bearer <token>

→ 200 {
    code: 0,
    msg: "ok",
    data: {
      total: 35,
      list: [
        { id: 1, code: "admin", real_name: "系统管理员", ... },
        ...
      ],
      page: 1,
      size: 20
    }
  }
```

### 5.2 创建用户

```http
POST /api/v1/users
Authorization: Bearer <token>
Content-Type: application/json

{
  "username": "u001",
  "password": "secret123",
  "real_name": "张三",
  "phone": "13800000001",
  "email": "zhang@example.com",
  "org_id": 5,
  "status": 1
}

→ 201 { code: 0, msg: "ok", data: { id: 100, code: "u001", ... } }

# 重复 username
→ 409 { code: 2009, msg: "用户名已存在", data: null }
```

### 5.3 字典查询

```http
# 平台域（公开）
GET /api/v1/dicts/resolve?code=gender

→ 200 {
    code: 0, msg: "ok",
    data: {
      id: 1, code: "gender", name: "性别",
      items: [
        { id: 1, code: "male", name: "男" },
        { id: 2, code: "female", name: "女" }
      ]
    }
  }

# 平台域 CRUD（仅 super_admin）
GET /api/v1/platform/dicts
```

### 5.4 配置中心

```http
# 公开读
GET /api/v1/public/configs?category=site

→ 200 {
    code: 0, msg: "ok",
    data: {
      items: [
        { key: "site_name", value: "XinFramework", type: "string", is_public: true },
        { key: "site_logo", value: "/uploads/logo.png", type: "image", is_public: true }
      ]
    }
  }
```

### 5.5 平台租户管理（仅 super_admin）

```http
GET /api/v1/platform/tenants
Authorization: Bearer <super_admin_token>

→ 200 { code: 0, msg: "ok", data: { total: 10, list: [...] } }

# 触发新租户"首装"（从 bootstrap 租户复制菜单/字典/配置）
POST /api/v1/platform/tenants
{
  "code": "acme",
  "name": "ACME Corp",
  "contact": "张总",
  "phone": "13800000001",
  "email": "admin@acme.com"
}

→ 201 { code: 0, msg: "ok", data: { id: 5, code: "acme", ... } }
# 后台自动跑 first_install.go 复制 menu/permission/dict/config
```

### 5.6 文件上传

```http
POST /api/v1/asset/upload
Authorization: Bearer <token>
Content-Type: multipart/form-data

file: <binary>

→ 200 { code: 0, msg: "ok", data: { id: 100, url: "/uploads/2026/06/24/xxx.png" } }
```

---

## 6. 公共请求头

| Header | 必填 | 用途 |
|---|---|---|
| `Authorization: Bearer <token>` | 业务 API 必填 | 访问令牌 |
| `Content-Type: application/json` | POST/PUT 必填 | 请求体格式 |
| `X-Tenant-ID: <id>` | 公开域可选 | 当无 token 时，用 `X-Tenant-ID` 兜底注入 `Context.TenantID` |
| `X-Request-ID` | 自动 | 由 RequestID 中间件生成 |
| `X-Requested-With` | 可选 | 跨域兼容性 |

---

## 7. 公共响应头

| Header | 用途 |
|---|---|
| `X-Request-ID` | 请求追踪 ID（与请求 X-Request-ID 对齐） |
| `Access-Control-Allow-Origin` | CORS 跨域头（按 cfg.cors.allow_origins） |

---

## 8. CORS

`config/config.yaml`：

```yaml
cors:
  enabled: true
  allow_origins:
    - "http://localhost:5241"
    - "http://127.0.0.1:5241"
  allow_methods: "GET,POST,PUT,DELETE,PATCH,OPTIONS"
  allow_headers: "Content-Type,Authorization,X-Requested-With,X-Request-ID,X-Tenant-ID"
  allow_credentials: true
  max_age: 86400
```

**注意**：`allow_credentials=true` 时 `allow_origins` 不能为 `*`，必须列具体源。

---

## 9. 分页约定

Query 参数：

| 参数 | 必填 | 默认 | 范围 |
|---|---|---|---|
| `page` | 否 | 1 | ≥ 1 |
| `size` | 否 | 20 | 1-200 |
| `keyword` | 否 | — | 模糊搜索 |
| `xxx_id` | 否 | — | 外键过滤 |

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "total": 100,
    "list": [/* ... */],
    "page": 1,
    "size": 20
  }
}
```

---

## 10. 资源权限码

每个受保护路由都附一个 `permission.P(ResXxx, ActYyy)` spec：

```go
g.GET("/users", middleware.Require(permission.P(permission.ResUser, permission.ActList)), h.List)
g.POST("/users", middleware.Require(permission.P(permission.ResUser, permission.ActCreate)), h.Create)
g.DELETE("/users/:id", middleware.Require(permission.P(permission.ResUser, permission.ActDelete)), h.Delete)
```

**资源常量**（`framework/pkg/permission/constants.go`）：

```go
ResSystem / ResAsset / ResDict / ResTenant / ResOrganization /
ResResource / ResMenu / ResRole / ResUser / ResPermission /
ResWeixin / ResAuth / ResFlag / ResConfig
```

**操作常量**：

```go
ActList / ActGet / ActCreate / ActUpdate / ActDelete / ActTree
```

**通配符**：

- `*:list` / `user:*` / `*:*` — 通配
- `*:*` = 全资源全操作（admin role 默认绑定）

---

## 11. 平台 vs 租户

| 域 | URL 前缀 | 中间件 | Guard |
|---|---|---|---|
| public | `/api/v1/*` | `OptionalAuth` | 无 |
| tenant | `/api/v1/*`（无 `/t` 前缀） | `Auth` + `RequireTenantContext` | `tenant_id > 0` |
| platform | `/api/v1/platform/*` | `Auth` + `RequirePlatformRole("super_admin")` | `PlatformRoles` 包含 `super_admin` |

**JWT Claims 区分**：
- `scope=tenant`：`tenant_id > 0`，业务域 token
- `scope=platform`：`tenant_id == 0`，平台域 token

`RequireTenantContext` 拒 platform token；`RequirePlatformScope` 拒 tenant token；`RequirePlatformRole` 跨域检查（`PlatformRoles` 字段）。

---

## 12. 错误处理示例

### 12.1 参数错误

```json
{ "code": 2000, "msg": "请求参数格式错误", "data": null }
```
HTTP 400

### 12.2 未登录

```json
{ "code": 401, "msg": "未登录", "data": null }
```
HTTP 401

### 12.3 权限不足

```json
{ "code": 4001, "msg": "无权限访问", "data": null }
```
HTTP 403

### 12.4 资源不存在

```json
{ "code": 3001, "msg": "租户不存在", "data": null }
```
HTTP 404

### 12.5 系统错误

```json
{ "code": 5000, "msg": "服务器内部错误", "data": null }
```
HTTP 500

---

## 13. SDK / 客户端调用模板

### 13.1 原生 fetch（前端 `api/common.ts`）

```ts
import { api } from "@/api"

const data = await api<{ list: UserItem[]; total: number }>("/users", {
  params: { page: 1, size: 20, keyword: "admin" }
})
```

`api()` 自动：
- 加 `Authorization: Bearer <token>`（从 `localStorage.token` 取）
- `Content-Type: application/json`（默认）
- 401 自动 refresh（带单例队列避免并发）
- 失败抛 `ApiError(status, code, message, data)`

### 13.2 curl 调试

```bash
# 登录拿 token
TOKEN=$(curl -s -X POST http://localhost:8087/api/v1/auth/tenant-login \
  -H "Content-Type: application/json" \
  -d '{"account":"admin","password":"admin123","tenant_id":1}' \
  | jq -r '.data.token')

# 调业务 API
curl -s http://localhost:8087/api/v1/users \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.data.list | length'
```
