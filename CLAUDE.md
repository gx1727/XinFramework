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
XinFramework/                    # Root (gx1727.com/xin-project)
├── framework/                   # Framework core (gx1727.com/xin/framework)
│   ├── framework.go             # Run(), RegisterModule()
│   ├── pkg/                     # Public packages (accessible by external modules)
│   │   ├── plugin/              # Module interface definition
│   │   ├── migrate/             # Auto SQL migration engine
│   │   ├── config/              # YAML + env config system
│   │   ├── db/                  # GORM + tenant session variable
│   │   ├── cache/               # Redis client
│   │   ├── logger/              # Daily rotating file logger
│   │   ├── session/             # Session management (Redis/DB)
│   │   ├── jwt/                 # Token generation
│   │   └── resp/                # Unified response {code, msg, data}
│   ├── internal/module/         # Built-in modules
│   │   ├── auth/                # Authentication (login/register/logout)
│   │   ├── user/                # User models and queries
│   │   ├── system/              # Health check
│   │   └── weixin/              # WeChat module
│   └── api/v1/register.go       # Dynamic route registration
├── apps/                         # External business modules (source code)
│   └── cms/                     # CMS module (gx1727.com/xin/module/cms)
│       ├── cms.go               # Implements plugin.Module
│       ├── config.yaml          # Dev config
│       └── migrations/          # Dev migrations
├── cmd/xin/main.go              # Entry point (registers modules → framework.Run())
├── config/config.yaml           # System config (app, db, redis, jwt, auth, domain)
└── migrations/                   # Framework-level SQL migrations
```

### Plugin System

External modules implement `plugin.Module` interface:
```go
type Module interface {
    Name() string
    Init() error                                    // Load config, setup
    Migrate() error                                 // Run SQL migrations
    RegisterV1(public, protected *gin.RouterGroup)  // Register routes
}
```

Entry point registers modules before `framework.Run()`:
```go
func main() {
    framework.RegisterModule(cms.Module())
    framework.Run()
}
```

### Startup Flow

```
main() → RegisterModule() → Run()
  → config.Load("config/config.yaml")
  → boot.Init() (logger → db → cache)
  → initModules()        (each module's Init())
  → runFrameworkMigrations()  (migrations/*.sql)
  → migrateModules()     (each module's Migrate())
  → setupRouter()        (middleware + builtin + external routes)
  → srv.Start()
```

### Middleware Chain (order matters)

1. `RequestID()` - X-Request-ID header
2. `Logger()` - request logging
3. `Recovery()` - panic recovery
4. `Tenant()` - tenant isolation (reads `X-Tenant-ID` header)
5. Route handlers → `Auth()` group for protected routes

### Multi-Tenant Isolation

Tenant context via PostgreSQL session variable (`SET app.tenant_id = ?`). The `Tenant` middleware sets this per-request. `db.SetTenantID()`/`db.ClearTenantID()` manage the lifecycle.

### Dev/Publish Dual Path

External modules use dual-path config/migration loading:
- Dev: `apps/{app}/config.yaml`, `apps/{app}/migrations/`
- Published: `config/{app}/config.yaml`, `migrations/{app}/`
- Logic: check dev path first, fallback to publish path
- Override: `XIN_{APP}_CONFIG` env var

### Build Output (out/)

```
out/
├── xin.exe
├── .env.example
├── config/
│   ├── config.yaml           # System config
│   └── cms/config.yaml       # App configs (copied from apps/cms/)
└── migrations/
    ├── 001_init.sql          # Framework migrations
    └── cms/                  # App migrations (copied from apps/cms/migrations/)
```

### Configuration System

Config loads from `config/config.yaml`, overrides with `XIN_*` env vars. Auth config is inline (not separate file). Environment variable names: `XIN_APP_*`, `XIN_DB_*`, `XIN_REDIS_*`, `XIN_JWT_*`, `XIN_SAAS_*`, `XIN_LOG_*`, `XIN_AUTH_*`, `XIN_DOMAIN`.

### API Response Format

Unified response: `{"code": 0, "msg": "ok", "data": {...}}`

Business errors use `resp.BizError` with predefined error dictionaries per module. Handler layer uses `resp.HandleError(c, err)`.

### Database Conventions

See `doc/database-conventions.md` for full details:
- Tables use `BIGSERIAL PRIMARY KEY`
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
