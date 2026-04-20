# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
# Build using the build script (outputs to ./out/xin)
./build.sh

# Manual build
go build -ldflags="-s -w" -o ./out/xin ./cmd/server/
```

## Architecture Overview

XinFramework is an enterprise SaaS foundation framework built with Go + Gin. It follows a **modular monolith** architecture that can evolve into microservices.

### Core Patterns

**Multi-Tenant Isolation**: Tenant context is propagated via PostgreSQL session variable (`SET app.tenant_id = ?`). The `Tenant` middleware sets this per-request, and `db.SetTenantID()`/`db.ClearTenantID()` manage the lifecycle.

**Context Propagation**: `internal/core/context/XinContext` wraps `gin.Context` to carry `TenantID` and `UserID` through the request lifecycle.

**Startup Flow**: `main.go` → `config.Load()` → `boot.Init()` → `setupRouter()` → middleware chain

**Middleware Chain** (order matters):
1. `Logger()` - request logging
2. `Recovery()` - panic recovery
3. `Tenant()` - tenant isolation (reads `X-Tenant-ID` header)
4. Route handlers → `Auth()` group for protected routes

### Directory Structure

```
cmd/server/main.go        # Entry point
config/config.yaml        # Base config (committed)
.env.example              # Template for local config

api/v1/register.go        # Public routes registered here
internal/
  core/                   # Framework internals
    boot/boot.go          # Init sequence (logger → db → cache → server)
    server/server.go      # Gin engine wrapper
    middleware/           # Auth, Tenant, Logger, Recovery, RateLimit
    context/              # XinContext (TenantID, UserID)
  module/                 # Business modules (user/, auth/, saas/, etc.)
    user/model.go         # User, Role, Permission, Tenant models
  infra/                  # Infrastructure
    db/db.go              # GORM + tenant session variable
    cache/cache.go        # Redis client
    logger/logger.go      # Daily rotating file logger
pkg/
  config/config.go        # YAML load + .env override
  jwt/jwt.go              # Token generation
  resp/resp.go            # Unified response {code, msg, data}
```

### Configuration System

Config loads from `config/config.yaml`, then overrides with `.env` values (env vars take precedence). Environment variable names match YAML path: `DB_HOST`, `DB_PORT`, `REDIS_HOST`, `JWT_SECRET`, `SAAS_MODE`, etc.

### API Response Format

统一响应格式：`{"code": 0, "msg": "ok", "data": {...}}`

详见 `doc/handbook.md`：
- 错误处理分层（HTTP 状态码 + 业务码）
- 业务码范围定义
- 响应函数说明

### Multi-Tenant Modes

`saas.mode` supports:
- `""` or absent - single tenant (no isolation)
- `shared` - shared database, `tenant_id` column filter
- `schema` - PostgreSQL schema per tenant
- `database` - separate database per tenant

### Database Conventions

See `doc/database-conventions.md` for full details:
- Tables use `BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY`
- All tables have `created_at`, `updated_at`, `is_deleted`
- Tenant tables include `tenant_id` + partial index `WHERE is_deleted = FALSE`
- Use `TIMESTAMPTZ` (not `TIMESTAMP`)
- JSONB for extensible data, GIN index for querying

### Developer Guide

详见 `doc/developer-guide.md`：
- 如何添加配置参数
- 如何使用日志
- 如何添加新模块（Handler/Service/Repo）
- 如何开发中间件
- 数据库和缓存操作
- JWT 使用
