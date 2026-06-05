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
