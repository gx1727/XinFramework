# server/AGENTS.md

> XinFramework 后端（Go + PostgreSQL）的设计参考。读这一份，比从头扫 `cmd/` `framework/` `apps/` 三个目录快得多。
>
> **本文件只描述当前事实**。代码在哪里、做什么、用什么约定；不写 refactor 过程、phase 编号、兼容期说明。

---

## 1. 技术栈

- **语言**：Go 1.26.2（`server/go.mod`）
- **HTTP**：Gin v1.12
- **数据库**：PostgreSQL，pgx/v5 + pgxpool（无 ORM）
- **JWT**：golang-jwt/jwt v5，HS256
- **密码哈希**：argon2id（`golang.org/x/crypto/argon2`）
- **缓存**：go-redis/redis v8（可选；不可用时自动降级到 DB session）
- **对象存储**：腾讯云 COS（`tencentyun/cos-go-sdk-v5`）+ 本地文件系统二选一
- **配置**：YAML + 环境变量（`gopkg.in/yaml.v3`）
- **MIME / JSON**：`goccy/go-json`、`clbanning/mxj`（间接依赖）

---

## 2. 目录结构

```
server/
├── cmd/
│   ├── xin/                              # 服务入口（start/stop/run/reload/hot-restart/status）
│   └── rotate-admin-password/           # 独立工具：重置 admin 密码
├── config/
│   ├── config.yaml                       # 主配置（dev 默认值；prod 走 env 覆盖）
│   ├── config.dev.yaml                   # dev 覆盖（仅 app.port 等少量字段）
│   ├── config.prod.yaml                  # prod 覆盖（强制 jwt.secret=""、显式列 module）
│   └── cms.yaml                          # 业务模块私有配置（cms 走 config.LoadModule）
├── migrations/
│   ├── init_schema.sql                   # 34 张表 + 索引 + RLS policy
│   ├── init_seed.sql                     # 种子数据（admin/角色/菜单/权限码/字典/配置）
│   ├── asset.sql / cms.sql / flag.sql    # 业务模块独立 schema
│   └── message.sql                       # 站内信
├── apps/                                 # 业务模块
│   ├── boot/auth/                        # 登录 / 账号（必装）
│   ├── platform/tenants/                 # 平台租户 CRUD（必装）
│   ├── platform/sys_user|role|menu|permission/   # 平台管理域
│   ├── tenant/user|role|menu|resource|organization|permission/  # 租户域 RBAC
│   ├── tenant/message/                   # 站内信
│   ├── reference/asset|config|dict|weixin/      # 基础设施 / 集成
│   ├── system/                           # health / cache 运维（必装）
│   ├── flag/                             # 头像/相框（optional）
│   └── cms/                              # CMS（optional）
└── framework/                            # 框架内核（公开 + internal 两层）
    ├── framework.go                      # Serve / Boot 入口
    ├── runtime.go                        # Runtime{Server, AppCtx}
    ├── cmd.go                            # xin start/stop/reload 命令
    ├── internal/core/
    │   ├── boot/boot.go                  # 启动编排（构造 App + Server + AppContext）
    │   ├── server/server.go              # HTTP server（包装 gin.Engine）
    │   └── middleware/                   # Recovery / RequestID / CORS / ClientIP / Logger
    ├── pkg/
    │   ├── appx/                         # 进程级资源容器 {Config, DB}
    │   ├── config/                       # YAML 加载 + env 覆盖 + 校验
    │   ├── db/                           # pool Init + Querier + RunInTx/TenantTx/PlatformTx
    │   ├── plugin/                       # Module 契约 + AppContext（Reader/Writer）
    │   ├── authz/                        # Authorization 接口
    │   ├── permission/                   # Spec / Resource / Action / DataScope
    │   ├── jwt/                          # token 签发/校验 + Claims
    │   ├── session/                      # Redis / DB 两套 SessionManager
    │   ├── cache/                        # Redis 单例
    │   ├── auth/                         # AccountRepository 公开契约
    │   ├── tenant/                       # TenantRepository 公开契约
    │   ├── platformauth/                 # 平台域 User/Role/Menu/Permission 公开契约
    │   ├── tenant/auth/                  # 租户域 User/Role/Org/RoleResource 公开契约
    │   ├── middleware/                   # Require / RequirePlatformRole / RequireTenantContext
    │   ├── audit/                        # db_logs 写入入口
    │   ├── migrate/                      # 启动期 SQL 迁移
    │   ├── resp/                         # 统一响应 + 错误码分段
    │   ├── logger/                       # zap 包装
    │   ├── storage/{local,cos}/          # 对象存储实现
    │   ├── xincontext/                   # 请求上下文（Context / UserContext）
    │   └── identity/                     # 平台 + 租户共享的 User/Role/Menu/Permission/Org 基类
    └── ...
```

---

## 3. 模块契约（plugin.Module）

```go
type Module interface {
    Name() string
    Init(ctx Reader, w Writer) error
    Register(ctx Reader, public, tenant, protected *gin.RouterGroup)
    Shutdown(ctx Reader) error
}
```

字段填充用 `plugin.BaseModule`（`NameStr` / `InitFn` / `RegFn` / `StopFn`），不需的方法留 nil。

### 3.1 三组 RouterGroup 语义

| 组 | URL 前缀 | 中间件（`framework.go:registerModules`） | 用途 |
|---|---|---|---|
| `public` | `/api/v1/*` | `OptionalAuth` | 登录、公开读；需要隔离的子资源挂 `/public/<x>` |
| `tenant` | `/api/v1/*` | `Auth` + `RequireTenantContext` | 业务域，模块直接挂资源路径（**无 `/t` 前缀**） |
| `protected` | `/api/v1/platform/*` | `Auth` | 平台域，模块内部追加 `RequirePlatformRole` |

### 3.2 AppContext（Reader/Writer）

跨模块共享资源走 `plugin.AppContext`，不通过包级全局：

| 槽 | Reader | Writer（写方） |
|---|---|---|
| `DB()` / `Cache()` / `Config()` / `Session()` | 读 | `framework/boot` 启动期填充 |
| `Authz()` | 读 | `framework/boot`（AuthorizationService） |
| `AccountRepo()` / `AccountAuthRepo()` | 读 | `apps/boot/auth` |
| `TenantRepo()` | 读 | `apps/platform/tenants` |
| `UserRepo()` / `RoleRepo()` / `OrgRepo()` / `PermRepo()` | 读 | `apps/tenant/{user,role,organization,permission}` |

Reader 方法返回 nil 表示对应模块未启用，调用方必须 nil-check。

---

## 4. 数据域

### 4.1 分层

| 域 | 标识 | RLS | API 守卫 |
|---|---|---|---|
| **平台域** `sys_*` | 无 `tenant_id` | ❌ 不启用 | `RequirePlatformRole("super_admin")` + `db.RunInPlatformTx` |
| **租户域** `tenant_*` | 带 `tenant_id` | ✅ 全部启用 | `RequireTenantContext`（tenant_id > 0） |
| **共享层** | `accounts` / `tenants` / `auth_sessions` | ❌ | 由对应业务模块自己路由 |

### 4.2 表与模块对照

| 域 | 表 | 业务模块 |
|---|---|---|
| 平台域 | `sys_users` / `sys_orgs` / `sys_roles` / `sys_menus` / `sys_permissions` / `sys_user_roles` / `sys_role_menus` / `sys_role_permissions` | `apps/platform/sys_*` |
| 租户域 | `tenant_organizations` / `tenant_users` / `tenant_roles` / `tenant_role_data_scopes` / `tenant_user_roles` / `tenant_menus` / `tenant_permissions` / `tenant_role_menus` / `tenant_role_resources` / `tenant_user_seq` | `apps/tenant/*` |
| 共享 | `tenants` / `accounts` / `account_auths` / `auth_sessions` | `apps/platform/tenants` + `apps/boot/auth` |
| 字典/配置 | `dicts` / `dict_items` / `dict_visibility` / `config_categories` / `config_items` / `config_visibility` | `apps/reference/dict` + `apps/reference/config` |
| 业务 | `subscriptions` / `usage_records` / `db_logs` / `routes` / `plans` | 对应模块 |
| 业务模块 | `assets` / `cms_*` / `flag_*` / `messages` | `apps/reference/asset` / `apps/cms` / `apps/flag` / `apps/tenant/message` |

### 4.3 RLS policy 模板

租户域表（无 `tenant_id=0` 语义）：

```sql
CREATE POLICY tenant_isolation_policy ON xxx USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);
```

字典/配置类表（`tenant_id=0` 表示平台级）多一条 `tenant_id = 0` 短路。

---

## 5. 启动流程

```
cmd/xin/main.go
  ├─ config.Load("config/config.yaml")          # YAML + .env + 环境变量，prod 强校验 jwt
  ├─ framework.Boot(cfg)                        # 返回 (app, rt, err)
  │    └─ boot.Init(cfg)
  │         ├─ logger.Init
  │         ├─ db.Init  → *pgxpool.Pool
  │         ├─ dict.Init(pool)
  │         ├─ cache.Init  → Redis（可降级）
  │         ├─ session.Init（Redis 或 DB 实现）
  │         ├─ plugin.NewAppContext(...)
  │         └─ service.NewAuthorizationService + appCtx.SetAuthz
  ├─ modules := []plugin.Module{ ... }          # main.go 显式列
  └─ framework.Serve(cfg, app, rt, modules)
       ├─ migrate.Run(app.DB, "migrations")
       ├─ for m in modules:  m.Init(ctx, w)    # 各模块写自己的 slot
       ├─ setupRouter：注册全局中间件 + 三组 RouterGroup
       ├─ for m in modules:  m.Register(ctx, public, tenant, protected)
       └─ 阻塞，收到信号后按顺序调 m.Shutdown
```

`cmd/xin/main.go` 同时是 `xin start/stop/reload/hot-restart/status` 命令的入口（`framework/cmd.go`）。

### 5.1 配置加载

`framework/pkg/config/config.go`：

- 主文件 `config/config.yaml`，`config.dev.yaml` / `config.prod.yaml` 按 `app.env` 叠加
- `.env` 优先于默认值（仅当对应环境变量未设）
- 环境变量 `XIN_*` 覆盖 YAML（详见 `overrideWithEnv`）
- 模块私有配置走 `config.LoadModule("cms", &cfg)`，从 `config/cms.yaml` 或主文件 `cms:` 段读

### 5.2 模块开关

`cfg.module` 现在的语义是**累加 optional**，不是白名单：

| 档 | 模块 | cfg.module 写法 |
|---|---|---|
| `alwaysOn`（3 个，必装） | `system` / `auth` / `platform_tenant` | 不可关 |
| `optOut`（13 个，默认全开） | RBAC（menu/user/role/resource/organization/permission）+ dict/asset/config + 平台管理（sys_user/sys_role/sys_menu/sys_permission） | 框架默认启用，不写即为开 |
| `optional`（3 个，默认关） | `weixin` / `cms` / `flag` | 必须显式列出 |

prod 环境必须显式声明 `module:` 列表（`config.prod.yaml` 强制）。

### 5.3 JWT 强校验（prod）

`config.validateJWTSecret` 在 `app.env=prod` 时检查：

1. secret 非空、非占位符（`your-secret-key` / `changeme` / `please-change-me` / `secret` / `12345678`）
2. 长度 ≥ 32 字节

失败直接 `FATAL` 退出。生产部署通过 `XIN_JWT_SECRET=$(openssl rand -base64 48)` 注入。

---

## 6. 鉴权与会话

### 6.1 Token 形态

`framework/pkg/jwt/jwt.go:Claims`：

```go
type Claims struct {
    UserID        uint     // tenant login: tenant_users.id；platform login: account_id
    TenantID      uint     // 0 = platform 域
    Role          string   // tenant role code / "_platform"（platform 域占位）
    SessionID     string   // 服务端 session 标识（Redis key / auth_sessions PK）
    TokenType     string   // "access" | "refresh"
    PlatformRoles []string // super_admin 等
    jwt.RegisteredClaims
}
```

签发：`GenerateWithPlatformRoles`（带 platform roles）。`HasPlatformRole(role)` 便捷判断。

### 6.2 三种登录入口

| 端点 | scope | tenant_id | 适用 |
|---|---|---|---|
| `POST /auth/tenant-login` | `tenant` | 必填 | 业务用户登录 |
| `POST /auth/platform-login` | `platform` | 0 | super_admin 登录 |
| `POST /auth/login-precheck` | — | — | 列出账号可用身份（多身份账号） |
| `POST /auth/select-tenant` | `tenant` | 必填 | 选完身份后签 token（等价 tenant-login） |
| `POST /auth/refresh` | 同原 token | 可选 | 切租户时传 `tenant_id`；platform token 不允许切租户 |
| `POST /auth/logout` | — | — | 撤销 session |
| `POST /auth/register` | `tenant` | 必填 | 注册新用户 |
| `POST /auth/login` | `tenant` | 必填 | 旧入口，**等价转发到 tenant-login** |

### 6.3 中间件

`framework/internal/core/middleware/auth.go`（`internal/`） + `framework/pkg/middleware/auth.go`（`pkg/`）：

| 中间件 | 行为 |
|---|---|
| `Auth` | 必登录，注入 `Context`（轻量身份） + 注册 `UserContext` 懒加载器（RBAC + DataScope） |
| `AuthLite` | 必登录，只注入 `Context`；不挂权限加载器，杜绝 `MustNewUserContext` 误调用 |
| `OptionalAuth` | 有 token 走 `Auth`；无 token 时从 `X-Tenant-ID` header 兜底注入 `Context.TenantID`，用于公共接口租户识别 |
| `Require(spec)` | 全部 spec 校验（`MatchAll`），缺一个 403 |
| `RequireAny(specs...)` | 任一通过即可 |
| `RequireAll(specs...)` | 同 Require，签名版 |
| `RequireAuthenticated()` | 只校验登录 |
| `RequirePlatformRole(roles...)` | 校验 `Context.PlatformRoles` 包含任一 role |
| `RequireTenantContext()` | `Context.TenantID > 0`；挡住 platform token |
| `RequirePlatformScope()` | `Context.TenantID == 0`；挡住 tenant token |

**短路**：拥有 `*:*` 资源权限或 `super_admin` 平台角色时，`Require*` 全部放行（`requireWithSpecs` L119）。

### 6.4 密码与 Session

- **密码**：`apps/boot/auth/password.go` 用 argon2id（m=64MiB, t=3, p=4），编码格式 `$argon2id$v=19$m=...,t=...,p=...$<salt>$<hash>`
- **Session**：`framework/pkg/session` 提供 `RedisSessionManager` / `DBSessionManager` 两套实现；Redis 可用时优先；DB 模式自动 `CREATE TABLE IF NOT EXISTS auth_sessions`

---

## 7. 数据库访问

### 7.1 包契约

`framework/pkg/db`：

- `Init(ctx, *DatabaseConfig) (*pgxpool.Pool, error)` — 由 `boot.Init` 调用
- `Querier` — 抽象 `Pool` / `Tx` 共同接口（`Exec` / `Query` / `QueryRow`）
- `WithTx(ctx, tx)` / `GetQuerier(ctx, pool)` — 事务注入与提取
- `RunInTx(ctx, pool, fn)` — 自动 `Begin/Commit/Rollback`，**嵌套安全**（已存在事务则复用）
- `RunInTenantTx(ctx, pool, tenantID, fn)` — 套 `app.tenant_id = $tenantID`
- `RunInPlatformTx(ctx, pool, fn)` — 设 `app.bypass_rls = 'on'` + `app.tenant_id = 0`

### 7.2 Repository 写法约定

```go
// 入口拿 Querier（自动 join 当前事务或 fallback pool）
q, err := db.GetQuerier(ctx, r.db)
if err != nil { return err }

// 用 q.Exec / q.Query / q.QueryRow 走事务
err = q.QueryRow(ctx, `SELECT ...`, id).Scan(&v)
```

业务模块不直接 import `pgxpool.Pool`，只接 `*pgxpool.Pool`（如 `apps/boot/auth/module.go` 注入的 `pool`）；Repository 方法第一个参数是 `ctx`，不要在内部新建 context。

### 7.3 错误约定

- **DB 层**：必须用命名 sentinel（`Err*DB` / `err*DB`）表示"未找到"等业务预期错误
- **Service 层**：在 `mapRepoError` 把 DB sentinel 翻译为 `resp.Err(code, msg)`
- **Handler 层**：未知 error 走 `resp.HandleError(c, err)`，由 `CodeToHTTPStatus` 决定 HTTP 状态

---

## 8. 错误与响应

`framework/pkg/resp/resp.go` 单一入口，**禁止业务模块手写 JSON**。

```go
type Response struct {
    Code int    `json:"code"`
    Msg  string `json:"msg"`
    Data any    `json:"data"`
}
type BizError struct { Code int; Msg string }
```

### 8.1 Code 分段 → HTTP 状态码

| 段 | HTTP | 用途 |
|---|---|---|
| `1xxx` | 200 | 鉴权 / 账号 / 通用业务（body Code 表达真实结果） |
| `2xxx` | 400 | 参数校验 / 业务规则 |
| `3xxx` | 404 | 资源不存在 / 租户级 |
| `4xxx` | 403 | 权限不足 / 角色冲突 |
| `5xxx+` | 500 | 服务端故障 / 系统异常 |

### 8.2 分段常量（`errors.go`）

```go
const (
    CodeAuth         = 1000  // 1001-1999
    CodeUser         = 2000  // 2001-2999
    CodeTenant       = 3000  // 3001-3999
    CodeRole         = 4000  // 4001-4999
    CodeMenu         = 5000  // 5001-5999
    CodeOrganization = 6000  // 6001-6999
    CodePermission   = 7000  // 7001-7999
    CodeResource     = 8000  // 8001-8999
    CodeAsset        = 9000  // 9001-9999
    CodeDict         = 10000 // 10001-10999
    CodeSystem       = 11000 // 11001-11999
    CodeWeixin       = 12000 // 12001-12999
    CodeFlag         = 13000 // 13001-13999
    CodeCMS          = 14000 // 14001-14999
)
```

新模块申请连续区间。`resp.Err(code, msg)` 构造，`resp.HandleError(c, err)` 出口。

### 8.3 Handler 模板

```go
func (h *Handler) Get(c *gin.Context) {
    id, err := parseIDParam(c, "id")
    if err != nil { resp.BadRequest(c, "无效的ID参数"); return }
    out, err := h.svc.GetByID(c.Request.Context(), id)
    if err != nil { resp.HandleError(c, err); return }
    resp.Success(c, out)
}

func (h *Handler) List(c *gin.Context) {
    var req ListReq
    if err := c.ShouldBindQuery(&req); err != nil { resp.BadRequest(c, "请求参数格式错误"); return }
    list, total, err := h.svc.List(c.Request.Context(), req)
    if err != nil { resp.HandleError(c, err); return }
    resp.Paginate(c, total, list)
}
```

---

## 9. 权限模型

### 9.1 资源 × 操作

`framework/pkg/permission/constants.go`：

```go
ResSystem / ResAsset / ResDict / ResTenant / ResOrganization /
ResResource / ResMenu / ResRole / ResUser / ResPermission /
ResWeixin / ResAuth / ResFlag / ResConfig

ActList / ActGet / ActCreate / ActUpdate / ActDelete / ActTree
```

### 9.2 Spec 与中间件

```go
spec := permission.P(permission.ResUser, permission.ActDelete)  // = "user:delete"
g.DELETE("/users/:id", middleware.Require(spec), h.Delete)
```

通配符 `*:*` = 全资源全操作（admin role 默认绑定）。`requireWithSpecs` 在 `Context.HasPlatformRole(super_admin)` 时短路放行。

### 9.3 DataScope

`framework/pkg/permission/types.go:DataScopeType`：

| 值 | 含义 |
|---|---|
| 1 | 全部 |
| 2 | 自定义 org 列表 |
| 3 | 本部门 |
| 4 | 本部门及以下 |
| 5 | 本人 |

通过 `UserContext.DataScopeFilterFor(columns)` 生成 SQL WHERE 子句，注入业务查询。

---

## 10. 审计

`framework/pkg/audit`：业务模块调 `audit.Log(ctx, pool, audit.Entry{...})`，**失败仅记日志，不返回错误**——业务路径不应被审计写库失败打断。

```go
audit.Log(ctx, pool, audit.Entry{
    Action:    "tenant:create",
    TableName: "tenants",
    RecordID:  t.ID,
    OldData:   map[string]any{...},
    NewData:   map[string]any{...},
})
```

`TenantID` / `UserID` / `IP` 留 0 时从 `Context` 自动取。

---

## 11. 路由清单（与前端 [UI/AGENTS.md](../UI/AGENTS.md) 对齐）

> 完整列表见 `server/framework/framework.go:registerModules`。下表只列业务模块自己挂的端点。

### 11.1 public 域

| 路径 | 模块 |
|---|---|
| `POST /auth/tenant-login` / `platform-login` / `login-precheck` / `select-tenant` / `refresh` / `register` / `login`（兼容） | `boot/auth` |
| `GET /health` | `system` |
| `GET /public/configs` | `reference/config` |
| `GET /flag/frames*` / `flag/spaces/:code` / `flag/avatar-categories` / `flag/avatars` | `flag` |
| `POST /weixin/login` / `phone` / `GET /weixin/ping` | `reference/weixin` |
| `GET /cms/ping` | `cms` |

### 11.2 tenant 域（业务）

| 路径 | 模块 |
|---|---|
| `/users` / `/user/profile` / `/user/avatar` | `tenant/user` |
| `/organizations`（含 `/tree`） | `tenant/organization` |
| `/roles`（含 `/menus` / `/data-scopes`） | `tenant/role` |
| `/menus`（含 `/tree`） | `tenant/menu` |
| `/resources`（含 `/my` / `/by-menu/:menu_id`） | `tenant/resource` |
| `/roles/:id/permissions` / `/roles/:id/resources` | `tenant/permission` |
| `/messages` | `tenant/message` |
| `/dicts` + `/dicts/resolve` + `/dicts/:id/items` | `reference/dict` |
| `/configs` + `/configs/resolve` + `/configs/:id/items` | `reference/config` |
| `/asset/upload` / `/asset/:id` | `reference/asset` |
| `/flag/frames` / `flag/frames-categories` / `flag/spaces` / `flag/avatars` / `flag/avatar-categories` / `flag/generate` / `flag/my-avatars` | `flag` |
| `/weixin/bind-phone` | `reference/weixin` |
| `/system/server-info` / `/system/clear-cache` / `/system/cache/*` | `system` |
| `/cms/me` / `/cms/users` / `/cms/tenant` / `/cms/posts*` | `cms` |

### 11.3 platform 域（super_admin）

| 路径 | 模块 |
|---|---|
| `/platform/tenants`（CRUD + `:id/status` + `:id/purge`） | `platform/tenants` |
| `/platform/sys-users`（CRUD + `:id/status` + `:id/roles`） | `platform/sys_user` |
| `/platform/sys-roles`（CRUD + `:id/menus` + `:id/permissions`） | `platform/sys_role` |
| `/platform/menus`（CRUD + `/tree`） | `platform/sys_menu` |
| `/platform/sys-permissions`（CRUD） | `platform/sys_permission` |
| `/platform/dicts` + `/platform/dicts/:id/items` + `/platform/dicts/:id/visibility` | `reference/dict` |
| `/platform/configs` + `/platform/configs/:id/items` + `/platform/configs/:id/visibility` | `reference/config` |

---

## 12. 数据库迁移

`framework/pkg/migrate.Run(pool, "migrations")` 启动期执行：

- 按文件名升序扫 `./migrations/*.sql`
- 跳过 `_schema_migrations` 已记录版本
- 每条在事务里跑（开头 `SET LOCAL row_security = off`）

### 12.1 当前文件

| 文件 | 内容 |
|---|---|
| `init_schema.sql` | 34 张表 + 索引 + RLS policy + 末尾断言（缺表 / 遗留旧表都 RAISE） |
| `init_seed.sql` | 种子（admin/角色/菜单/权限码/字典/配置） |
| `asset.sql` / `cms.sql` / `flag.sql` | 业务模块独立 DDL |
| `message.sql` | 站内信 |

### 12.2 表/字段规范

- 业务表必须有 `is_deleted BOOLEAN DEFAULT FALSE`
- 索引加 `WHERE is_deleted = FALSE` 谓词
- 租户域表必须有 `tenant_id BIGINT NOT NULL` + RLS policy
- 平台域表（`sys_*`）**不**带 `tenant_id`、**不**启用 RLS

### 12.3 DDL / Seed 幂等

- DDL 用 `IF NOT EXISTS` / `ADD COLUMN IF NOT EXISTS`
- Seed 用 `ON CONFLICT DO NOTHING` / `ON CONFLICT (...) DO UPDATE`

---

## 13. 构建与运行

```bash
# 构建
cd server
go build -o bin/xin ./cmd/xin

# 直接跑（前台）
./bin/xin run              # 等价于 go run ./cmd/xin

# 守护进程
./bin/xin start            # 写 PID 到 ./xin.pid，日志到 ./xin.log
./bin/xin status           # 打印 PID + 最近 5 行日志
./bin/xin stop             # SIGTERM，30s 超时
./bin/xin restart          # stop + start
./bin/xin reload           # SIGUSR1，零停机重载
./bin/xin hot-restart      # 起新进程 + 停老进程

# 平台
# Windows: build.ps1
# Linux/Mac: build.sh
```

### 13.1 启动顺序

1. PostgreSQL（必须；`database.host:port`）
2. Redis（可选；`redis.enabled`）
3. `xin run` / `xin start`
   - 启动期跑 `migrate.Run`
   - `_schema_migrations` 没记录过 `init_schema.sql` 时会建表 + 跑 seed

### 13.2 工具

```bash
go run ./cmd/rotate-admin-password -tenant <id> -account <account> -new <password>
```

---

## 14. 新增一个业务模块（配方）

以 `apps/tenant/foo` 为例：

1. **包结构**（每个文件一个职责）

   ```
   apps/tenant/foo/
   ├── module.go         # Module(app) → *plugin.BaseModule
   ├── routes.go         # Register(tenant *gin.RouterGroup, h *Handler)
   ├── handler.go        # gin handler
   ├── service.go        # 业务编排
   ├── repository.go     # 数据访问（Querier）
   ├── model.go          # 实体 struct
   ├── types.go          # 请求/响应 DTO + 业务魔数
   └── errors.go         # 业务错误码（从 resp.CodeFoo 段申请）
   ```

2. **Module 函数**（`module.go`）

   ```go
   func Module(app *appx.App) plugin.Module {
       return &plugin.BaseModule{
           NameStr: "foo",
           InitFn: func(_ plugin.Reader, w plugin.Writer) error {
               w.SetFooRepo(NewFooRepository(app.DB))
               return nil
           },
           RegFn: func(_ plugin.Reader, _ *gin.RouterGroup, tenant *gin.RouterGroup, _ *gin.RouterGroup) {
               h := NewHandler(NewService(app.DB, NewFooRepository(app.DB)))
               Register(tenant, h)
           },
       }
   }
   ```

3. **main.go 加一行**（按 alwaysOn / optOut / optional 决定是否需要 `cfg.module`）

   ```go
   modules := []plugin.Module{
       // ... 既有
       foo.Module(app),
   }
   ```

4. **路由 + 权限**（`routes.go`）

   ```go
   func Register(tenant *gin.RouterGroup, h *Handler) {
       tenant.GET("/foos", middleware.Require(permission.P(permission.ResFoo, permission.ActList)), h.List)
       tenant.POST("/foos", middleware.Require(permission.P(permission.ResFoo, permission.ActCreate)), h.Create)
       // ...
   }
   ```

5. **错误码**（`errors.go`）

   ```go
   var ErrFooNotFound = resp.Err(resp.CodeFoo+1, "foo 不存在")
   ```

6. **审计**（写操作必加）

   ```go
   audit.Log(ctx, pool, audit.Entry{
       Action: "foo:create", TableName: "foos", RecordID: f.ID,
       NewData: map[string]any{"id": f.ID, "name": f.Name},
   })
   ```

7. **前端联动**：见 [UI/AGENTS.md](../UI/AGENTS.md) §6

---

## 15. 关键文件索引

| 关注点 | 路径 |
|---|---|
| 启动入口 | `server/cmd/xin/main.go` |
| 启动编排 | `server/framework/internal/core/boot/boot.go` |
| 路由装配 + 模块过滤 | `server/framework/framework.go` |
| 三组 RouterGroup 定义 | `server/framework/framework.go:registerModules` |
| 全局中间件 | `server/framework/framework.go:setupRouter` |
| 进程级资源 | `server/framework/pkg/appx/appx.go` |
| 跨模块容器 | `server/framework/pkg/plugin/appcontext.go` |
| 鉴权中间件（internal） | `server/framework/internal/core/middleware/auth.go` |
| 鉴权守卫（公开） | `server/framework/pkg/middleware/auth.go` |
| 登录 / 账号 | `server/apps/boot/auth/` |
| 平台租户 | `server/apps/platform/tenants/` |
| 平台域 sys_* | `server/apps/platform/sys_{user,role,menu,permission}/` |
| 租户域 RBAC | `server/apps/tenant/{user,role,menu,resource,organization,permission}/` |
| 字典 / 配置 | `server/apps/reference/{dict,config}/` |
| 响应 / 错误码 | `server/framework/pkg/resp/` |
| 权限 Spec | `server/framework/pkg/permission/spec.go` |
| 权限常量 | `server/framework/pkg/permission/constants.go` |
| JWT | `server/framework/pkg/jwt/jwt.go` |
| Session | `server/framework/pkg/session/session.go` |
| 事务工具 | `server/framework/pkg/db/db.go` |
| 审计 | `server/framework/pkg/audit/audit.go` |
| 迁移 | `server/framework/pkg/migrate/migrate.go` |
| Schema | `server/migrations/init_schema.sql` |
| Seed | `server/migrations/init_seed.sql` |
| 配置 | `server/config/config.yaml` |
| 请求上下文 | `server/framework/pkg/xincontext/context.go` |

---

## 16. 调用方约束（业务模块照做即可）

1. **不要**用 `init()` 自动注册模块——main.go 显式 `[]plugin.Module{...}`
2. **不要**手写 `c.JSON(...)`——走 `resp.Success` / `resp.HandleError` / `resp.Paginate`
3. **不要**绕过 `db.RunInTenantTx` 写租户域表——会触发 RLS
4. **不要**在 Repository 里 `errors.New("xxx not found")`——必须 `var ErrXxxNotFoundDB = errors.New(...)` 让 service 层 `errors.Is` 区分
5. **不要**直接 import `pgxpool.Pool` 之外的 DB 类型——统一 `*pgxpool.Pool`，Repository 方法用 `db.GetQuerier(ctx, r.db)` 拿 Querier
6. **不要**在 `migrations/init_schema.sql` 加业务表——在自己的 `<module>.sql` 加
7. **不要**在 `init_seed.sql` 改自己模块的 seed——在自己的 `<module>.sql` 末尾追加 `ON CONFLICT DO NOTHING`
8. **不要**把 `*appx.App` 传给 framework 内部代码（`runtime` / `middleware`）——它只给业务模块用
9. **不要**自己判断 HTTP 状态码——`resp.HandleError` 调 `CodeToHTTPStatus` 单点决定
10. **不要**用 untyped string 表示权限——`permission.P(ResFoo, ActDelete)` 编译期保证
