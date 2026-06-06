# AGENTS.md

This file provides guidance to Codex when working with code in this repository.

## Build Commands

```bash
# Build framework + all apps
go build -ldflags="-s -w" ./cmd/xin ./framework/... ./apps/...

# Run in dev mode
go run ./cmd/xin run

# Run tests and vet
go vet ./...
```

## Architecture Overview

XinFramework is an enterprise SaaS foundation framework built with Go + Gin. Multi-module plugin architecture with framework + pluggable business apps.

### Multi-Module Structure

```
XinFramework/
├── go.work                                        # 多module仓库（ms/ framework/ apps/、、)
├── cmd/xin/                       # 进程入口，加载 config 并按 cfg.Apps 列表装车插件
├── framework/                      # gx1727.com/xin/framework （独立 Go模块)
│   ├── framework.go                # Run() / RegisterModule() / initModules()
│   ├── cmd.go                     # start / stop / restart / reload / hot-restart / status
│   ├── signal.go                   # waitForSignal() ，优雅关闭
│   ├── .env.example                # XIN_* 环境变量样例
│   ├── pkg/                        # 可被 apps/* 导入的公共包
│   │   ├── cache/                 # go-redis/v8 客户端
│   │   ├── config/                 # YAML + 环境变量
│   │   ├── context/               # XinContext / UserContext （懒加载)
│   │   ├── db/                    # pgx/v5/pgxpool + tenant RLS
│   │   ├── dict/                  # 字典数据访问 + 缓存
│   │   ├── extapi/                # 外部 API Provider 接口
│   │   ├── jwt/                   # Token 生成/校验
│   │   ├── logger/                # 按天切分日志
│   │   ├── middleware/            # AuthLite / Require 等公开中间件
│   │   ├── migrate/              # SQL 迁移运行器
│   │   ├── model/                 # 领域模型 + 通用错误
│   │   ├── permission/            # RBAC + DataScope 类型
│   │   ├── plugin/                # Module 接口 + Register / Apps 注册表
│   │   ├── resp/                  # {code, msg, data} 统一响应
│   │   ├── session/               # SessionManager (Redis / DB [备])
│   │   ├── storage/               # 文件存储（local / cos)
│   ├── internal/                # 构架内部，不对外
│   │   ├── core/
│   │   │   ├── boot/              # boot.Init / Shutdown
│   │   │   ├── ext_impl/            # extapi Provider 默认实现
│   │   │   ├── middleware/        # Recovery / RequestID / CORS / Logger / Auth / OptionalAuth
│   │   │   └── server/            # XinServer ，优雅关闭
│   │   ├── module/                # 12 个内置模块
│   │   └── service/               # permission_service / authorization_service
├── apps/                          # 外部业务插件（独立 Go模块)
│   ├── cms/                       # CMS 模板
│   └── flag/                      # 头像 / 相框社团
├── config/                        # YAML 配置（config.yaml / *.dev.yaml / *.prod.yaml / cms.yaml)
└── migrations/                   # framework.sql / cms.sql / flag.sql
```

## Key Interfaces

### Module Interface (`pkg/plugin/plugin.go`)
```go
type Module interface {
    Name() string
    Init() error
    Register(public *gin.RouterGroup, protected *gin.RouterGroup)
    Shutdown() error
}

// module is the standard implementation
type module struct {
    name       string
    register   ModuleFunc
    initFn     func() error
    shutdownFn func() error
}

// NewModule / NewModuleWithOpts creates a module
// WithInit(fn) / WithShutdown(fn) for lifecycle hooks
```

### SessionManager (`pkg/session/session.go`)
```go
type SessionManager interface {
    Create(sessionID string, userID, tenantID uint, role string, ttl time.Duration) error
    Validate(sessionID string) (bool, error)
    Revoke(sessionID string) error
}
```

### XinContext (`pkg/context/context.go`)
```go
type XinContext struct {
    TenantID  uint
    UserID    uint
    SessionID string
    Role      string
}
func New(c *gin.Context) *XinContext
xc.GetUserID() / xc.GetTenantID() / xc.GetSessionID() / xc.GetRole()
```

### Permission Interfaces (`pkg/permission/interfaces.go`)
```go
type UserPermissionRepository interface {
    GetUserPermissions(ctx context.Context, userID uint) (map[string]bool, error)
    GetUserRoles(ctx context.Context, userID uint) ([]string, error)
    GetUserIDsByRole(ctx context.Context, roleID uint) ([]uint, error)
    GetUserIDsByResource(ctx context.Context, resourceID uint) ([]uint, error)
}

// pkg/permission/interfaces.go 变化：PermissionRepository 用于角色→资源映射
type PermissionRepository interface {
    GetByRoleID(ctx context.Context, roleID uint) ([]Permission, error)
    DeleteByRoleID(ctx context.Context, roleID uint) error
    Create(ctx context.Context, tenantID, roleID uint, p Permission) error
}

type DataScopeRepository interface {
    GetDataScope(ctx context.Context, userID uint) (*DataScope, error)
    GetUserOrgID(ctx context.Context, userID uint) (int64, error)
    GetByRoleID(ctx context.Context, roleID uint) ([]uint, error)
    SetForRole(ctx context.Context, roleID uint, orgIDs []uint) error
}
```

## Built-in Modules (12 total)

| Module | Path | Purpose | Key Routes |
|--------|------|---------|------------|
| asset | internal/module/asset | File storage (local/cos) | GET/POST /assets/* |
| auth | internal/module/auth | Login/Logout/Register/Refresh | POST /login, /register, /refresh |
| tenant | internal/module/tenant | Tenant CRUD | GET/POST/PUT/DELETE /tenants |
| user | internal/module/user | User queries + storage config | GET /users, /users/:id, PUT /users/:id/status |
| menu | internal/module/menu | Menu hierarchy (ltree) | GET /menus, /menus/tree |
| dict | internal/module/dict | Dictionary data | GET/POST/PUT/DELETE /dicts |
| role | internal/module/role | Role CRUD + data scopes | CRUD /roles, GET/PUT /roles/:id/data-scopes |
| resource | internal/module/resource | Button permissions | CRUD /resources, GET /resources/by-menu/:menu_id |
| organization | internal/module/organization | Org tree | CRUD /organizations, GET /organizations/tree |
| permission | internal/module/permission | Role-permission assignment | GET/POST/PUT /roles/:id/permissions |
| system | internal/module/system | Health check | GET /health |
| weixin | internal/module/weixin | WeChat stub | GET /weixin/ping |

## Startup Flow

```
main() [cmd/xin/main.go]
  → config.Load("config/config.yaml")
  → moduleRegistry[app]() → framework.RegisterModule()
  → framework.Run(cfg)

framework.Run(cfg)
  → boot.Init(cfg)                  # logger → db → cache → session → PermService
  → initModules(cfg)                 # builtinMap[name].Init() + plugin.Apps()[].Init()
  → runMigrations()                 # migrate.Run("migrations")
  → setupRouter(app)               # middleware chain + route registration
  → srv.Start(addr)                # HTTP server start
  → waitForSignal(srv, app)        # signal.Notify → srv.Shutdown() → boot.Shutdown(app)

initModules(cfg):
  → for _, name := range cfg.Module { builtinMap[name].Init(); builtinMap[name].Register() }
  → for _, m := range plugin.Apps() { m.Init(); m.Register() }
```

## Dependency Injection

### boot.App (`internal/core/boot/boot.go`)
```go
type App struct {
    Config      *config.Config
    DB          *pgxpool.Pool
    SessionMgr  session.SessionManager
    Server      *server.XinServer
    PermService *service.PermissionService
}
```

### Module Pattern (current)

Each module defines a `Module()` function returning `plugin.Module`. Dependencies are wired in the closure:

```go
// internal/module/xxx/module.go
func Module() plugin.Module {
    return plugin.NewModule("xxx", func(public, protected *gin.RouterGroup) {
        repos := xxx.NewXxxRepository(db.Get())
        svc := xxx.NewService(repos)
        h := xxx.NewHandler(svc)
        Register(protected, h)
    })
}
```

### Builtin Module Registration (`framework/framework.go`)
```go
var builtinMap = map[string]plugin.Module{
    "asset":        assetModule.Module(),
    "auth":         authModule.Module(),
    "tenant":       tenantModule.Module(),
    "user":         userModule.Module(),
    "menu":         menuModule.Module(),
    "dict":         dictModule.Module(),
    "role":         roleModule.Module(),
    "resource":     resourceModule.Module(),
    "organization": orgModule.Module(),
    "permission":   permModule.Module(),
    "system":       systemModule.Module(),
    "weixin":       weixinModule.Module(),
}
```

Modules are initialized and registered via `cfg.Module` list.

### External Plugins
Apps in `apps/*` call `plugin.Register(xxx.Module())` in their init, registered via `plugin.Apps()`.

## Middleware Chain (order matters)

```
# 全局（pkg/framework/setupRouter 中按此顺序注册)
1. Recovery()     — panic recovery，必须最先
2. RequestID()    — X-Request-ID generation/propagation
3. CORS(&cfg.CORS)   — Cross-origin + OPTIONS preflight
4. Logger()      — 请求日志（依赖 RequestID)

# /api/v1 分组（pkg/framework/registerModules 中的 registerModules）
public := v1.Group("")
public.Use(middleware.OptionalAuth(&cfg.JWT, sm, permSvc))
protected := v1.Group("")
protected.Use(middleware.Auth(&cfg.JWT, sm, permSvc))
```

## Permission System

**Permission format**: `"resource_code:action"` (e.g., `"user:create"`, `"*:*"` for super admin)

**RBAC flow**: `users → user_roles → roles → permissions → resources`

**Data scope types** (`pkg/permission/types.go`):
| Value | Name | Description |
|-------|------|-------------|
| 1 | DataScopeAll | All data in tenant |
| 2 | DataScopeCustom | Only specified org_ids (from role_data_scopes) |
| 3 | DataScopeDept | Only user's department (org_id) |
| 4 | DataScopeDeptAndBelow | User's dept + all descendant depts |
| 5 | DataScopeSelf | Only own records |

**Auth middleware** (`internal/core/middleware/auth.go`) 会验

```go
// pkg/context.go:XinContext ，没有 Set* method
xc := &xinContext.XinContext{
    TenantID:  claims.TenantID,
    UserID:    claims.UserID,
    SessionID: claims.SessionID,
    Role:      claims.Role,
}
ctx = xinContext.WithXinContext(ctx, xc)
ctx = xinContext.WithTenantID(ctx, claims.TenantID)

// 懒加载:权限/数据范围只在使用MustNewUserContext 时才去�查
ctx = xinContext.WithUserContextLoader(ctx, func() *xinContext.UserContext {
    // permSvc.LoadUserSecurityContext() -> perms, roles, ds, orgID
    return &xinContext.UserContext{XinContext: xc, ...}
})
c.Request = c.Request.WithContext(ctx)
```## Tenant Isolation Modes

| Mode | Behavior |
|------|----------|
| `strict` | RLS enforces tenant_id = current_setting('app.tenant_id') |

**RLS policies** use `current_setting('app.tenant_id')`.

## Configuration System

Config loads from `config/config.yaml`, overrides with `XIN_*` env vars.

| Component | Env Prefix | Example |
|-----------|------------|---------|
| App | `XIN_APP_*` | `XIN_APP_PORT=9999` |
| Database | `XIN_DB_*` | `XIN_DB_HOST=localhost` |
| Redis | `XIN_REDIS_*` | `XIN_REDIS_HOST=127.0.0.1` |
| JWT | `XIN_JWT_*` | `XIN_JWT_SECRET=xxx` |
| Module Config | `XIN_<NAME>_*` | `XIN_CMS_POST_PER_PAGE=20` |

## API Response Format

Unified response: `{"code": 0, "msg": "ok", "data": {...}}`

| Function | HTTP Status | Business Code |
|----------|-------------|---------------|
| `Success` | 200 | 0 |
| `Error` | 200 | custom |
| `Unauthorized` | 401 | 401 |
| `Forbidden` | 403 | 403 |
| `BadRequest` | 400 | 400 |
| `NotFound` | 404 | 404 |
| `ServerError` | 500 | 500 |

## Database Conventions

- Tables use `BIGSERIAL PRIMARY KEY` or `BIGINT GENERATED ALWAYS AS IDENTITY`
- All tables have `created_at`, `updated_at`, `is_deleted`
- Tenant tables include `tenant_id` + partial index `WHERE is_deleted = FALSE`
- Use `TIMESTAMPTZ` (not `TIMESTAMP`)
- Migrations tracked in `_schema_migrations` table
- `ltree` extension for hierarchical data (menus, organizations)
- `pg_trgm` extension for ILIKE fuzzy search

## Plugin System

**Two types of modules:**

1. **Built-in modules** (`framework/internal/module/*`) - Loaded via `cfg.Module` list, each has `Module()` returning `plugin.Module`
2. **External apps** (`apps/*`) - Call `plugin.Register(xxx.Module())` in init, enabled via `apps:` config

**CMS app structure** (`apps/cms/`):
```
module.go      → plugin.Register(xxx.Module())
routes.go     → Register() wires routes
handler.go    → thin HTTP handlers
```

**Flag app structure** (`apps/flag/`):
```
module.go      → plugin.Register(xxx.Module())
routes.go      → Register() wires routes
handler.go     → HTTP handlers
```

## Tables (21 in 001_framework_init.sql)

| Table | Purpose |
|-------|---------|
| tenants | SaaS multi-tenant core |
| accounts | Global cross-tenant account (phone/email/password) |
| account_auths | Third-party OAuth (WeChat, QQ, Weibo) |
| organizations | Org tree structure (ltree) |
| users | Tenant-scoped user, links account_id to tenant |
| roles | RBAC role with data_scope |
| role_data_scopes | Custom org IDs per role |
| user_roles | Many-to-many user-role |
| menus | Navigation menu items with ltree |
| resources | Button/operation permissions |
| routes | API route permissions |
| permissions | Role-to-resource assignment |
| dicts | System data dictionaries |
| dict_items | Dictionary items |
| db_logs | Audit log |
| subscriptions | Tenant subscription plans |
| plans | SaaS billing plans |
| usage_records | Tenant resource usage tracking |
| ai_documents | AI knowledge base |
| auth_sessions | Session persistence fallback |
| tenant_user_seq | Auto-increment user_code per tenant |
| account_roles | Platform-level roles (super_admin) |


---

## 踩坑与决策日志（新增项）

下面记录的坑处和决策在开发中明显遭过，在新模块上线前必顾。

### 常规决策

- **租户 RLS**：新增带 `tenant_id` 的表必须：CREATE TABLE → ALTER TABLE ENABLE ROW LEVEL SECURITY → CREATE POLICY，三处同步添加。
- **软删**：`is_deleted = TRUE` + `updated_at = NOW()`，禁止 `DELETE FROM`（除硬删定时任务外）。
- **审计写入**：`audit.Log(ctx, audit.Entry{...})`，失败 warn 不抛；IP 留空自动从 ctx 读。
- **删除前必检**：用 `fmt.Errorf("%w (xxx数=%d)", ErrXxxHasYyy, n)` 封装计数信息；organization.Delete（子组织 + 子树用户）和 dict.Delete（字典项）都有。
- **业务错误码**：`resp.Err(code, msg)` 构造；分段见下表。

### 错误码分段

| 段 | 区间 | 模块 |
|---|---|---|
| CodeUser | 2001-2999 | user |
| CodeTenant | 3001-3999 | tenant |
| CodeRole | 4001-4999 | role |
| CodeMenu | 5001-5999 | menu |
| CodeOrganization | 6001-6999 | organization |
| CodePermission | 7001-7999 | permission |
| CodeResource | 8001-8999 | resource |
| CodeAsset | 9001-9999 | asset |
| CodeDict | 10001-10999 | dict |
| CodeSystem | 11001-11999 | system |

### 调用方向约束（import cycle 实测结论）

- `user` → `organization`：允许（user 注入 `organization.OrganizationRepository`）。
- `organization` → `user`：**禁止**（import cycle）。需查 user 表时，org 仓库内部用 SQL 跨表查询（参考 `CountUsersInOrgTree` / `CountChildren`）。
- `dict` 模块 → `pkg/dict` 包：可，但 import 必须用别名 `dictpkg`，否则与当前包名同名编译失败。

### 经典踩坑

- `pgx.PgError` 在 `github.com/jackc/pgx/v5/pgconn`，不在 `pgx`。`isUniqueViolation` 必须 `import "github.com/jackc/pgx/v5/pgconn"`。
- `db.GetQuerier` 返回接口 `pgx.Tx`/pool，方法 `QueryRow` 只给 `pgx.Row`，**不是 `*sql.Row`**；不要混用 `database/sql` 类型。
- `dicts.code` 和 `dict_items.code` 是唯一索引组合（`uk_dict_code` / `uk_dict_items_code`）；插入 `23505` → 抛 `ErrDictCodeExists` / `ErrDictItemCodeExists`（409）。
- `db_logs` 未启用 RLS（参见 migrations），但 `audit.Log` 内显式 `NULLIF($2, 0)`，user_id 可空；不依赖 `app.tenant_id`。
- 所有审计写**在 `db.RunInTenantTx` 内**调用，让 `app.tenant_id` 与 `db_logs.tenant_id` 对齐。
- 沙箱无网络，`go build` 下不到依赖；用 `gofmt -e ./path/to/file.go > $null 2>&1; if ($LASTEXITCODE -eq 0) {"OK"} else {"ERROR"}` 代替。
- 新模块 `Module()`：`plugin.NewModule(name, func(public, protected) { h := NewHandler(NewService(NewPostgresXxxRepository(db.Get()))); Register(protected, h) })`。dict 之前是 `NewDictRepository` 旧签名，已替换为 `NewPostgresDictRepository`。

### 审计与 ClientIP

- `framework/internal/core/middleware/client_ip.go` 在 CORS 之后、Logger 之前调用 `audit.WithIP(c.Request.Context(), c.ClientIP())`。
- `framework/pkg/audit/context.go` 提供 `WithIP` / `IPFrom`。

### 软删缓存一致性

- 字典项写操作完成后调 `dictpkg.RefreshDict(ctx, tenantID, code)` 重建缓存；ctx 走 `context.WithoutCancel` detach，避免请求取消中断缓存重建。

### 用户-组织关系（users.org_id 单外键）

- "换组" = UPDATE `users.org_id`，无 `user_organizations` 关联表。
- "0 = 移出" 三层一致：handler 透传 `*uint`、repository 写 NULL（`arg = nil`）、前端 `UserItem.org_id: number | null`。
- 跨租户校验在 service 层 `validateOrg`：`orgRepo.GetByIDScoped` + `TenantID` 比对 + `Status==1`。
- 新端点 `PUT /users/:id/org`（权限 `ResUser/ActUpdate`）专用于换主组织；PATCH 仍可顺带改 `org_id`。

### 子树筛选语义

- 前端做：用 `ancestors` 字符串前缀 + `collectOrgSubtreeIds` 递归；不依赖后端 SQL。
- 如后端做：repository 加 `?org_subtree=1` + `org_id IN (SELECT id FROM organizations WHERE ancestors LIKE $1)`。

### 数据初始化约定

- admin 角色绑定**所有 menu** + **超级资源**（`code='*'` + `action='*'`）；新增 menu/resource/dict 用 `INSERT ... OVERRIDING SYSTEM VALUE` + `SELECT setval(...)` 手动推 sequence。
- 新增 menu 后需要确保 admin 角色绑定；用 `role_menus` 走全量重新分配（`SELECT 1, 1, id FROM menus WHERE is_deleted = FALSE`）。


### 字典缓存预热（潜在 bug，需注意）

- `dict.Init(pool)` 确实在 `framework/internal/core/boot/boot.go::Init()` 里被调用，但**只设置 `dbPool`，不预热任何租户的字典数据**。
- `pkg/dict.LoadTenant(ctx, tenantID)` 在整个 server 业务代码里**没有任何调用方**（只有 `dict.go` 内部定义 + `RefreshDict` 用）。
- 后果：业务模块如果调 `dictpkg.Get(tenantID, "gender")` 读字典值，会拿到 `nil, false`（缓存空），不会有任何 fallback 报错。
- 现状安全：目前业务代码没人读 `dictpkg.Get`（只用 dict module 自己的 `RefreshDict` 在写时重建），所以不爆。
- 修复方式（如需）：在 `boot.Init` 末尾遍历 `tenants` 表，对每个 tenant 调 `dictpkg.LoadTenant(ctx, tenant.ID)`；或在 `auth.Login` 成功后按需懒加载。

### PowerShell 中文写入彻底方案（GBK 管道污染）

**症状**：在 Windows PowerShell 5.1 下，用 `Set-Content` / `@'...'@ | Out-File` / `>` 重定向 写 Go / TS 文件时，文件里的中文全部变成 `?`（或 `?????`）。CI / sandbox 也常见。

**根因**：PowerShell 5.1 默认走 OEM 代码页（中文 Windows = GBK / CP936）。管道里 UTF-8 字节被 GBK 解读，每个中文字节 0xE4/0xE5 等都被替换成 `?` (0x3F)，写入磁盘后变成纯 `?` 串。即使文件后缀是 `.go`，IDE 看似能打开，但读出来全是乱码，编译能过、运行中文全错。

**错误做法（不要用）**：
- `Set-Content -Encoding UTF8` —— PowerShell 5.1 会写 BOM，且会把不含 BOM 的 stdin 字节重编码
- `Out-File -Encoding UTF8` —— 同样 BOM 问题
- `Get-Content ... | Set-Content ...` —— 管道会重编码
- `[Console]::OutputEncoding = [System.Text.Encoding]::UTF8` —— 只影响控制台，不影响 Out-File
- 直接 `"$content" | Out-File` —— 走 stdin，被 GBK 解读

**彻底方案**（已验证）：

1. PowerShell 写一个 Python 脚本到 `$env:TEMPix_xxx.py`，**用 .NET API 强写 UTF-8 无 BOM**：

   ```powershell
   $py = @\'
   # -*- coding: utf-8 -*-
   content = "..."   # 这里是目标文件全文
   \'@
   [System.IO.File]::WriteAllText(
       "$env:TEMP\fix_xxx.py", $py,
       [System.Text.UTF8Encoding]::new($false)   # false = 不写 BOM
   )
   ```

2. 跑这个 Python 脚本（Python 自身 UTF-8 兼容中文）：

   ```powershell
   python "$env:TEMP\fix_xxx.py"
   ```

3. Python 脚本里把目标 Go / TS 内容用三引号字符串直接写：

   ```python
   target = r"D:\work\xin\XinFramework\server\framework\internal\module\dict\handler.go"
   with open(target, "w", encoding="utf-8") as f:
       f.write(content)
   ```

4. 防御性剥 BOM（如果旧文件已有 BOM）：

   ```python
   with open(p, "rb") as f: data = f.read()
   if data.startswith(b"\xef\xbb\xbf"): data = data[3:]
   with open(p, "wb") as f: f.write(data)
   ```

**为什么这样能行**：
- `[System.IO.File]::WriteAllText(..., [System.Text.UTF8Encoding]::new($false))` 走 .NET BCL，不经过 PowerShell 编码层
- Python 进程内 `open(..., 'w', encoding='utf-8')` 由 Python 运行时管编码，UTF-8 字节原样落盘
- 三引号字符串里所有 `中文` / `特殊符号` 全部保留

**快速自检命令**（Python 一行）：

```python
python -c "import re; t=open(r'<file>','r',encoding='utf-8').read(); print('CN=', len(re.findall(r'[一-鿿]', t)), 'QM=', len(re.findall(r'\?{2,}', t)))"
```

- CN 应该是合理数（Go 注释 + log 几十个）；QM 必须是 0
- 头 3 字节应该是 `2f 2f 20` (`// `) 而不是 `ef bb bf`（BOM）

**已应用此方案的文件**（2026-06，dict 模块编码大修）：
- `server/framework/internal/module/dict/errors.go`
- `server/framework/internal/module/dict/types.go`
- `server/framework/internal/module/dict/model.go`
- `server/framework/internal/module/dict/routes.go`
- `server/framework/internal/module/dict/module.go`
- `server/framework/internal/module/dict/repository.go`（同步修复 `pgx.PgError` → `pgconn.PgError`）
- `server/framework/internal/module/dict/service.go`
- `server/framework/internal/module/dict/handler.go`

