# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Build Commands

```bash
# Build using the build script (outputs to ./out/)
./build.ps1          # Windows
./build.sh           # Linux

# Manual build
go build -ldflags="-s -w" -o ./out/xin ./cmd/xin

# Run in dev mode
go run ./cmd/xin run

# Run tests
go vet ./...
```

## Architecture Overview

XinFramework is an enterprise SaaS foundation framework built with Go + Gin. Multi-module plugin architecture with framework + pluggable business apps.

### Multi-Module Structure

```
XinFramework/
├── framework/                   # Framework core (gx1727.com/xin/framework)
│   ├── framework.go             # Run(), cmdStart/Stop, waitForSignal
│   ├── cmd.go                   # Daemon commands (start/stop/restart/status/reload)
│   ├── signal.go                # Signal handling for graceful shutdown
│   ├── pkg/                     # Public packages
│   │   ├── config/              # YAML + env config system
│   │   ├── db/                  # GORM + tenant session variable
│   │   ├── cache/               # Redis client
│   │   ├── logger/             # Daily rotating file logger
│   │   ├── session/            # Session management (Redis/DB)
│   │   ├── jwt/                 # Token generation
│   │   ├── migrate/            # SQL migration runner
│   │   ├── plugin/              # Module interface and registry
│   │   └── resp/                # Unified response {code, msg, data}
│   ├── internal/core/           # Core framework components
│   │   ├── boot/                # Initialization (logger→db→cache→server)
│   │   ├── server/              # HTTP server wrapper with graceful shutdown
│   │   ├── middleware/         # RequestID, Logger, Recovery, Tenant, Auth, RateLimit
│   │   └── context/             # XinContext (TenantID, UserID wrapper)
│   ├── internal/module/         # Built-in modules (enabled via config)
│   │   ├── auth/                # Role/permission placeholder (not implemented)
│   │   ├── user/                # Login/Logout/Register
│   │   ├── system/              # Health check (/health)
│   │   └── weixin/             # WeChat integration (/weixin/ping)
│   └── api/v1/                  # Route registration
│       └── register.go          # Registers builtin + plugin modules
├── apps/                         # External business plugins
│   └── cms/                     # CMS plugin example
│       ├── cms.go               # Plugin with Init/Migrate/Register
│       └── migrations/          # CMS-specific migrations
├── cmd/xin/                     # Entry point
│   └── main.go                  # Loads config, registers plugins, calls framework.Run()
├── config/
│   └── config.yaml              # System config (app, db, redis, jwt, saas, log, module, apps, user)
└── migrations/
    └── framework/               # Framework-level SQL migrations
```

### Plugin System

**Two types of modules:**

1. **Built-in modules** (`framework/internal/module/*`) - Always loaded, enabled/disabled via config
2. **External plugins** (`apps/*`) - Must call `framework.RegisterModule()` in main.go, enabled via `cfg.AppEnabled()`

**Module interface:**
```go
type Module interface {
    Name() string
    Init() error                                   // Load config, setup
    Migrate() error                                // Run SQL migrations
    Register(public, protected *gin.RouterGroup)  // Register routes
}
```

**Registration flow:**
```go
// cmd/xin/main.go
cfg, _ := config.Load("config/config.yaml")
if cfg.AppEnabled("cms") {
    framework.RegisterModule(cms.Module())  // External plugins register themselves
}
framework.Run(cfg)

// framework/framework.go - Run() handles the rest
// initModules() → runFrameworkMigrations() → migrateModules() → setupRouter() → srv.Start()
```

### Startup Flow

```
main() [cmd/xin/main.go]
  → config.Load("config/config.yaml")
  → framework.RegisterModule(cms) (if AppEnabled)
  → framework.Run(cfg)

framework.Run(cfg)
  → boot.Init(cfg)                    # logger → db → cache
  → initModules()                     # each module's Init()
  → runFrameworkMigrations()         # migrations/framework/*.sql
  → migrateModules()                 # each module's Migrate()
  → setupRouter(srv, cfg)            # middleware chain + route registration
  → srv.Start(addr)                  # HTTP server start
```

### Middleware Chain (order matters)

1. `RequestID()` - X-Request-ID header generation/propagation
2. `Logger()` - request logging
3. `Recovery()` - panic recovery
4. `Tenant(cfg.Saas.Mode)` - tenant isolation via `SET app.tenant_id = ?`
5. Route handlers → `Auth()` group for protected routes

### Multi-Tenant Isolation

Tenant context via PostgreSQL session variable (`SET app.tenant_id = ?`):
- `Tenant` middleware sets this per-request from `X-Tenant-ID` header
- `db.SetTenantID()` / `db.ClearTenantID()` manage the lifecycle
- `Saas.Mode`: `shared` (default) | `schema` | `database`

### Module Enabling

**Built-in modules** (controlled by `module:` list in config):
```yaml
module:
  - weixin  # system and user are always enabled
```

**External apps** (controlled by `apps:` list in config):
```yaml
apps:
  - cms
```
Enabled check: `cfg.AppEnabled("cms")` in main.go.

### Configuration System

Config loads from `config/config.yaml`, overrides with `XIN_*` env vars.

| Component | Env Prefix | Example |
|-----------|------------|---------|
| App | `XIN_APP_*` | `XIN_APP_PORT=9999` |
| Database | `XIN_DB_*` | `XIN_DB_HOST=localhost` |
| Redis | `XIN_REDIS_*` | `XIN_REDIS_HOST=127.0.0.1` |
| JWT | `XIN_JWT_*` | `XIN_JWT_SECRET=xxx` |
| SaaS | `XIN_SAAS_*` | `XIN_SAAS_MODE=shared` |
| Log | `XIN_LOG_*` | `XIN_LOG_LEVEL=debug` |
| Module Config | `XIN_<NAME>_*` | `XIN_USER_MAX_LOGIN_ATTEMPTS=10` |

### API Response Format

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

Business errors use `resp.BizError` with predefined error dictionaries per module. Handler layer uses `resp.HandleError(c, err)`.

### Database Conventions

See `doc/database-conventions.md`:
- Tables use `BIGSERIAL PRIMARY KEY` / `BIGINT GENERATED ALWAYS AS IDENTITY`
- All tables have `created_at`, `updated_at`, `is_deleted`
- Tenant tables include `tenant_id` + partial index `WHERE is_deleted = FALSE`
- Use `TIMESTAMPTZ` (not `TIMESTAMP`)
- Migrations tracked in `_schema_migrations` table

### Developer Guide

See `doc/developer-guide.md`:
- How to add config parameters
- How to use logger
- How to add new module (Handler/Service/Repo)
- How to develop middleware
- Database and cache operations
- JWT usage