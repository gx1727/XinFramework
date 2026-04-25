# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Build Commands

```bash
# Build framework + all apps
go build -ldflags="-s -w" ./cmd/xin ./apps/cms/...

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
├── framework/                      # gx1727.com/xin/framework
│   ├── framework.go                # Run(), RegisterModule(), runServer()
│   ├── signal.go                   # waitForSignal(), graceful shutdown
│   ├── pkg/                        # Public packages (importable by apps)
│   │   ├── config/                 # YAML + env config system
│   │   ├── db/                    # pgx/v5/pgxpool + tenant session variable
│   │   ├── cache/                 # Redis client (go-redis/v8)
│   │   ├── logger/                # Daily rotating file logger
│   │   ├── session/               # SessionManager interface + Redis/DB impl
│   │   ├── jwt/                   # Token generation/validation
│   │   ├── migrate/               # SQL migration runner
│   │   ├── model/                 # Domain models and repository interfaces
│   │   ├── plugin/                # Module interface and registry
│   │   ├── repository/            # Repository Provider implementation
│   │   ├── resp/                  # Unified response {code, msg, data}
│   │   └── context/               # XinContext (UserID, TenantID, SessionID, Role)
│   ├── internal/core/              # Core framework components
│   │   ├── boot/                  # App initialization (boot.Init, Shutdown)
│   │   ├── server/                # XinServer with graceful shutdown
│   │   └── middleware/            # CORS, RequestID, Logger, Recovery, Tenant, Auth
│   ├── internal/module/            # Built-in modules
│   │   ├── user/                  # Login/Logout/Register/Refresh
│   │   ├── tenant/                # Tenant CRUD
│   │   ├── auth/                  # Role/permission placeholder
│   │   ├── system/                # Health check (/health)
│   │   └── weixin/               # WeChat integration (/weixin/ping)
│   └── api/v1/                   # Route registration
│       └── register.go            # Registers builtin + plugin modules
├── apps/                          # External business plugins
│   └── cms/                       # CMS plugin template
│       ├── go.mod                 # gx1727.com/xin/module/cms
│       ├── config.yaml             # Default config
│       ├── config.go              # LoadConfig()
│       ├── routes.go              # Module factory + Register()
│       ├── internal/
│       │   ├── handler/           # HTTP handlers
│       │   └── service/           # Business logic
│       └── migrations/            # App-specific SQL
├── cmd/xin/                       # Entry point
│   └── main.go                    # Config load, app registry, framework.Run()
├── config/
│   └── config.yaml                # System config
└── migrations/
    ├── 001_framework_init.sql     # Core tables
    ├── 002_cms_create_posts.sql   # CMS tables
    └── 003_search_indexes.sql     # Performance indexes
```

## Key Interfaces

### Module Interface (`pkg/plugin/plugin.go`)
```go
type Module interface {
    Name() string
    Init() error
    Register(public, protected *gin.RouterGroup)
    Shutdown() error
}
```

### SessionManager Interface (`pkg/session/session.go`)
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
func New(c *gin.Context) *XinContext  // reads from request context
xc.GetUserID() / xc.GetTenantID() / xc.GetSessionID() / xc.GetRole()
```

## Startup Flow

```
main() [cmd/xin/main.go]
  → config.Load("config/config.yaml")
  → moduleRegistry[app]() → framework.RegisterModule()
  → framework.Run(cfg)

framework.Run(cfg)
  → boot.Init(cfg)                  # logger → db → repository → cache → session
  → initModules()                   # plugin.All() → each m.Init()
  → runMigrations()                 # migrate.Run("migrations")
  → setupRouter(app)               # middleware chain + route registration
  → srv.Start(addr)                # HTTP server start
  → waitForSignal(srv, app)       # signal.Notify → srv.Shutdown() → boot.Shutdown(app)
```

## Dependency Injection

### boot.App (`internal/core/boot/boot.go`)
```go
type App struct {
    Config     *config.Config
    DB         *pgxpool.Pool
    Repository *repository.Provider  // all repos
    SessionMgr session.SessionManager
    Server     *server.XinServer
}
```

### Repository Provider (`pkg/repository/repository.go`)
- Created in `boot.Init()` via `repository.NewProvider(db.Get())`
- Stored globally via `repository.Init(provider)` for backward compatibility
- CMS app accesses via `repository.User()` / `repository.Tenant()`

### SessionManager (`pkg/session/session.go`)
- Redis impl: `NewRedisSessionManager()` when `cache.Get() != nil`
- DB impl: `NewDBSessionManager(db.Get())` otherwise
- Stored globally via `session.Init(sm)`

### Builtin Module Handlers (`framework/framework.go`)
```go
var builtinHandlers = map[string]builtinHandlerBuilder{
    "user": func(app *boot.App) interface{} {
        repos := user.Repositories{
            Account: app.Repository.Account(),
            Tenant:  app.Repository.Tenant(),
            Role:    app.Repository.Role(),
            User:    app.Repository.User(),
        }
        deps := user.DefaultDependencies(app.Config, app.DB, repos)
        return user.NewHandler(user.NewService(deps))
    },
    "tenant": func(app *boot.App) interface{} {
        return tenant.NewHandler(tenant.NewService(app.Repository.Tenant()))
    },
}
```

## Middleware Chain (order matters)

```
1. Recovery()     — panic recovery, must be first to catch all downstream panics
2. RequestID()    — X-Request-ID generation/propagation, runs early
3. CORS()        — Cross-origin resource sharing + OPTIONS preflight
4. Logger()      — Request logging (after RequestID set)
5. Tenant()      — Tenant isolation via SET app.tenant_id = ?
6. [protected routes] → Auth(cfg, sm) — JWT validation + session check
```

## Context System

**Auth middleware sets XinContext** (`middleware/middleware.go`):
```go
xc := context.New(c)
xc.SetUserID(claims.UserID)
xc.SetTenantID(claims.TenantID)
xc.SetSessionID(claims.SessionID)
xc.SetRole(claims.Role)
c.Request = c.Request.WithContext(context.WithXinContext(c.Request.Context(), xc))
c.Set("user_id", claims.UserID)   // also set on gin.Context for compatibility
```

**Handlers read via XinContext** (`pkg/context/context.go`):
```go
xc := context.New(c)
userID := xc.GetUserID()
```

## Plugin System

**Two types of modules:**

1. **Built-in modules** (`framework/internal/module/*`) - Always loaded (system, auth, user), others via `module:` config
2. **External apps** (`apps/*`) - Registered in `main.go` moduleRegistry, enabled via `apps:` config

**Registration:**
```go
// cmd/xin/main.go
var moduleRegistry = map[string]func() plugin.Module{
    "cms": cms.Module,
}

for _, app := range cfg.Apps {
    if factory, ok := moduleRegistry[app]; ok {
        framework.RegisterModule(factory())
    }
}
framework.Run(cfg)
```

**CMS app structure** (`apps/cms/`):
```
config.go       → LoadConfig() + Config struct
routes.go       → Module() returns plugin.Module, Register() wires routes
internal/
  handler/      → thin HTTP handlers, delegates to service
  service/      → business logic, uses repository.User()/Tenant()
migrations/     → app-specific SQL
```

## Configuration System

Config loads from `config/config.yaml`, overrides with `XIN_*` env vars.

| Component | Env Prefix | Example |
|-----------|------------|---------|
| App | `XIN_APP_*` | `XIN_APP_PORT=9999` |
| Database | `XIN_DB_*` | `XIN_DB_HOST=localhost` |
| Redis | `XIN_REDIS_*` | `XIN_REDIS_HOST=127.0.0.1` |
| JWT | `XIN_JWT_*` | `XIN_JWT_SECRET=xxx` |
| SaaS | `XIN_SAAS_*` | `XIN_SAAS_MODE=saas` |
| Module Config | `XIN_<NAME>_*` | `XIN_CMS_POST_PER_PAGE=20` |

Module config: `config.LoadModule("cms", &cfg)` looks for `config/cms.yaml` or `cms:` section in `config.yaml`.

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
| `Paginate` | 200 | 0 |

## Database Conventions

- Tables use `BIGSERIAL PRIMARY KEY` / `BIGINT GENERATED ALWAYS AS IDENTITY`
- All tables have `created_at`, `updated_at`, `is_deleted`
- Tenant tables include `tenant_id` + partial index `WHERE is_deleted = FALSE`
- Use `TIMESTAMPTZ` (not `TIMESTAMP`)
- Migrations tracked in `_schema_migrations` table
- `users`/`tenants` table indexes created with privilege check (skip if not owner)
