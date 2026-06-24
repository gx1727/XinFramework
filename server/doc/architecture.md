# 架构总览

> XinFramework 最关键的设计文档。第一次接触代码从这里开始。
> 最后更新：2026-06（Phase 0023 全阶段完成，19 模块终态）

## 1. Go Module 结构

仓库根目录下只有一个 Go module：

```
server/
└── go.mod         (path: gx1727.com/xin)         # cmd/ + migrations/ + framework/ + apps/
```

依赖方向（`internal/` 强制）：

```
cmd/xin ──→ framework ──→ apps
            (internal)     (apps 不可 import)
```

- `framework` 不能 import `apps`（编译就过不去）
- `apps` 不能 import `cmd/xin`
- `apps` 之间**只能**通过 `plugin.AppContext.Reader/Writer` 接口通信
- **单 Go module**（Phase 1 合并，不要回退到 multi-module / `go.work`）

## 2. 启动时序

[`cmd/xin/main.go`](../cmd/xin/main.go) 4 步显式 Build：

```go
func main() {
    cfg, err := config.Load("config/config.yaml")        // 1. 配置
    app, rt, err := framework.Boot(cfg)                  // 2. 装配 (*appx.App, *Runtime)
    modules := []plugin.Module{                          // 3. 显式模块列表（19 个）
        // alwaysOn
        auth.Module(app), tenants.Module(app), system.Module(app),
        // 平台管理域（optional）
        sysuser.Module(app), sysrole.Module(app),
        sysmenu.Module(app), syspermission.Module(app),
        // 租户域 RBAC 套件（optOut）
        menu.Module(app), organization.Module(app), permission.Module(app),
        resource.Module(app), role.Module(app), user.Module(app),
        // reference 套件
        asset.Module(app), refconfig.Module(app), dict.Module(app), weixin.Module(app),
        // external（optional）
        cms.Module(app), flag.Module(app),
    }
    framework.Serve(cfg, app, rt, modules)               // 4. 启动
}
```

[`framework.Serve`](../framework/framework.go) 内部流程：

```go
func Serve(cfg, app, rt, modules) {
    migrate.Run(app.DB, "migrations")       // → SQL 迁移（幂等）
    for _, m := range modules {             // → Init 阶段
        m.Init(ctx, w)                      //    模块写 own slot
    }
    setupRouter(cfg, app.DB, rt, modules)   // → 中间件 + 路由
    go rt.Server.Start(addr)                // → 后台监听
    waitForSignal(rt, app, modules)         // → SIGINT/SIGTERM 优雅退出
}
```

[`framework.Boot`](../framework/internal/core/boot/boot.go) 装配：

```go
func Init(cfg) (*appx.App, *server.XinServer, *plugin.AppContext, error) {
    logger.Init(...)
    pool, _      := db.Init(ctx, &cfg.Database)        // → pgxpool
    dict.Init(pool)
    cache.Init(&cfg.Redis)                             // → go-redis
    sm           := session.NewRedisSessionManager()   // → Session
    appCtx       := plugin.NewAppContext(pool, cache, cfg, sm)
    permService  := service.NewPermissionService(...)   // → RBAC
    authzService := service.NewAuthorizationService(permService)
    appCtx.SetAuthz(authzService)
    srv          := server.New(cfg)
    app          := &appx.App{Config: cfg, DB: pool}
    return app, srv, appCtx, nil
}
```

## 3. 模块生命周期：Init / Register / Shutdown

每个 module 实现 [`plugin.Module`](../framework/pkg/plugin/plugin.go) 接口：

```go
type Module interface {
    Name() string
    Init(ctx Reader, w Writer) error
    Register(ctx Reader, public, tenant, protected *gin.RouterGroup)
    Shutdown(ctx Reader) error
}
```

实际形态用 `BaseModule` struct：

```go
func Module(app *appx.App) plugin.Module {
    return &plugin.BaseModule{
        NameStr: "user",
        InitFn: func(_ plugin.Reader, w plugin.Writer) error {
            w.SetUserRepo(NewUserRepository(app.DB))
            return nil
        },
        RegFn: func(ctx plugin.Reader, _, tenant, _ *gin.RouterGroup) {
            h := NewHandler(NewService(app.DB, ...))
            Register(tenant, h)
        },
    }
}
```

### 3.1 Init 阶段：Writer 写 own slot，Reader 读别人贡献的

- **Writer** = 写自己负责的 slot：`SetAccountRepo` / `SetTenantRepo` / `SetUserRepo` ...
- **Reader** = 读别人贡献的 slot：`AccountRepo()` / `TenantRepo()` / `UserRepo()` ...
- 模块必须 nil-check Reader 返回的 Repository（producer 可能被 `cfg.Module` 关闭）
- 模块**永远不会**拿到写别人 slot 的 Writer，编译期类型保证

### 3.2 Register 阶段：拿到完整 Reader

到 Register 时，所有模块的 Init 已完成。`framework.go` 把 `AppContext` 作为 Reader 传给每个 module。

### 3.3 三类模块（按 `cfg.Module` 行为分）

| 类型 | 数量 | 表现 |
|---|---|---:|
| **alwaysOn** | 3（`system`, `auth`, `tenants`） | 启动必需，无法关闭，配置不列也加回去 |
| **optOut** | 13（RBAC + 字典/资产/配置 + 平台管理域） | 默认启用；用户写 `module:` 时切白名单语义（不列就关） |
| **optional** | 3（`weixin`, `cms`, `flag`） | 默认不启用；必须在 `cfg.Module` 显式列出才开 |

`optOut` 13 个：

- 租户域 RBAC：`menu` / `user` / `role` / `resource` / `organization` / `permission`
- 租户域基础设施：`dict` / `asset` / `config`
- 平台管理域：`sys_user` / `sys_role` / `sys_menu` / `sys_permission`

定义在 [`framework/pkg/config/config.go`](../framework/pkg/config/config.go) `alwaysOnModules` / `optOutModules`。

> **新约定**：除 weixin / cms / flag 这 3 个纯业务/集成模块外，其它都属于「框架默认能力」，无需在 `module:` 里列出。`module: 删一行就能关掉对应模块` 这条承诺在重新分类后变得更直观。

## 4. AppContext：唯一的依赖容器

[`framework/pkg/plugin/appcontext.go`](../framework/pkg/plugin/appcontext.go) 是整个重构的成果物。两个不变量：

1. **构造一次，终身不变**——在 `boot.Init` 中构造，后续只读
2. **Reader / Writer 接口分离**——"读别人 repo" 和 "写自己 repo" 在类型系统上不可混

### 4.1 接口定义

```go
type Reader interface {
    // 基础设施（Init 之前就填好）
    DB()       *pgxpool.Pool
    Cache()    *redis.Client           // 可能 nil
    Config()   *config.Config
    Session()  session.SessionManager

    // 跨模块贡献（Init 完成后填好）
    Authz()            authz.Authorization
    AccountRepo()      auth.AccountRepository
    AccountAuthRepo()  auth.AccountAuthRepository
    TenantRepo()       tenant.TenantRepository
    UserRepo()         pkgauth.UserRepository
    RoleRepo()         pkgauth.RoleRepository
    OrgRepo()          pkgauth.OrganizationRepository
    PermRepo()         pkgauth.RoleResourceRepository
}

type Writer interface {
    SetAuthz(authz.Authorization)
    SetAccountRepo(auth.AccountRepository)
    SetAccountAuthRepo(auth.AccountAuthRepository)
    SetTenantRepo(tenant.TenantRepository)
    SetUserRepo(pkgauth.UserRepository)
    SetRoleRepo(pkgauth.RoleRepository)
    SetOrgRepo(pkgauth.OrganizationRepository)
    SetPermRepo(pkgauth.RoleResourceRepository)
}
```

> 当前 AppContext 提供 **8 个跨模块 Repo slot**（Authz + 7 个 Repository）。新增跨模块接口的步骤：
> 1. 在 `framework/pkg/<scope>/xxx.go` 定义窄 interface + struct
> 2. 在 `appcontext.go` Reader/Writer 接口加方法
> 3. 在 `AppContext` struct 加字段 + getter/setter
> 4. provider module 在 `InitFn` 调 `w.SetXxx(myRepo)`
>
> 编译会引导你走完整链路——任何缺漏都会编译失败。

## 5. 中间件链

[`framework/framework.go`](../framework/framework.go) `setupRouter` 注册全局中间件（按顺序）：

```go
r.Use(middleware.Recovery())      // 1. panic recover
r.Use(middleware.RequestID())     // 2. 注入 X-Request-ID
r.Use(middleware.CORS(&cfg.CORS)) // 3. CORS 预检
r.Use(middleware.ClientIP())      // 4. 客户端 IP
r.Use(middleware.Logger())        // 5. access log
```

然后在 `/api/v1` 路由组里分 **三组 RouterGroup**：

```go
public := v1.Group("")                                   // /api/v1/*
public.Use(middleware.OptionalAuth(...))

tenant := v1.Group("")                                   // /api/v1/*（无 /t 前缀）
tenant.Use(middleware.Auth(...))
tenant.Use(pkgmiddleware.RequireTenantContext())

protected := v1.Group("/platform")                       // /api/v1/platform/*
protected.Use(middleware.Auth(...))
// 平台模块内部追加 RequirePlatformRole("super_admin")
```

### 5.1 Auth 中间件做了什么

[`framework/internal/core/middleware/auth.go`](../framework/internal/core/middleware/auth.go)：

1. 从 `Authorization: Bearer <jwt>` 提取 token
2. JWT 验证（HS256 + `cfg.JWT.Secret`）
3. Session 验证（去 Redis 或 DB 查 SessionID）
4. 把 `XinContext` 注入到 `c.Request.Context()`
5. **懒加载** `UserContextLoader`（第一次有模块调 `MustNewUserContext(c)` 才查 DB）

### 5.2 RBAC 中间件

[`framework/pkg/middleware/auth.go`](../framework/pkg/middleware/auth.go)：

| 函数 | 行为 |
|---|---|
| `Require(spec)` | 一个 spec 必须满足 |
| `RequireAny(specs...)` | 任一满足即可 |
| `RequireAll(specs...)` | 所有 spec 都必须满足 |
| `RequireAuthenticated()` | 登录即可，不看 RBAC |
| `RequirePlatformRole(roles...)` | 必须持有平台角色（跨租户用） |
| `RequireTenantContext()` | tenant_id > 0 才放行 |

`super_admin` 平台角色自动 bypass 所有 RBAC（spec 不需要写通配）。

## 6. 路由空间分布

所有路由分成 **3 个语义空间**：

| 空间 | URL 前缀 | 中间件 | 说明 |
|---|---|---|---|
| **业务消费** | `/api/v1/<resource>` | Auth + RequireTenantContext + Require(ResX) | 租户域业务接口 |
| **平台管理** | `/api/v1/platform/<resource>` | Auth + RequirePlatformRole(super_admin) + Require(ResX) | super_admin 跨租户接口 |
| **公开访问** | `/api/v1/public/<resource>` 或 `/api/v1/<auth>` | OptionalAuth | 不需登录 |

**完整路由示例**：

```
# 业务域（需登录 + tenant_id）
POST   /api/v1/users
GET    /api/v1/menus/tree
GET    /api/v1/configs/resolve?code=site
GET    /api/v1/dicts/resolve
GET    /api/v1/roles/:id/menus

# 平台域（强制 super_admin）
GET    /api/v1/platform/tenants
GET    /api/v1/platform/sys-menus/tree
GET    /api/v1/platform/configs
GET    /api/v1/platform/dicts

# 公开域（可选登录）
POST   /api/v1/auth/tenant-login       (业务用户登录，需要 tenant_id)
POST   /api/v1/auth/platform-login     (平台管理员登录，无需 tenant_id)
POST   /api/v1/auth/register
POST   /api/v1/auth/refresh
GET    /api/v1/health
GET    /api/v1/public/configs          (公开配置读)
POST   /api/v1/weixin/login            (微信小程序登录)
```

### 6.1 三组 RouterGroup 在代码中的对应

```go
// framework.go::registerModules
v1 := r.Group("/api/v1")

public := v1.Group("")                                 // /api/v1/*
public.Use(middleware.OptionalAuth(...))

tenant := v1.Group("")                                 // /api/v1/*（无 /t 前缀）
tenant.Use(middleware.Auth(...))
tenant.Use(pkgmiddleware.RequireTenantContext())

protected := v1.Group("/platform")                     // /api/v1/platform/*
protected.Use(middleware.Auth(...))
// 平台模块内部追加 RequirePlatformRole("super_admin")
```

业务模块接收三个 RouterGroup 引用，由模块决定把路由挂在哪一组：

```go
RegFn: func(_ plugin.Reader, public, tenant, protected *gin.RouterGroup) {
    // public:    /api/v1/public/*            （OptionalAuth）
    // tenant:    /api/v1/*                   （Auth + RequireTenantContext；无 /t 前缀）
    // protected: /api/v1/platform/*           （Auth；模块内部追加 RequirePlatformRole）
}
```

## 7. 响应协议

[`framework/pkg/resp/resp.go`](../framework/pkg/resp/resp.go)：

```json
// 成功
{ "code": 0, "msg": "ok", "data": { ... } }

// 业务错误
{ "code": 2001, "msg": "用户不存在", "data": null }

// 分页
{ "code": 0, "msg": "ok", "data": { "total": 100, "list": [ ... ] } }
```

### 7.1 错误码分段管理

[`framework/pkg/resp/errors.go`](../framework/pkg/resp/errors.go) 集中定义，每个 module 一段 1000：

| 区段 | module | 文件位置 |
|---|---|---|
| 1001-1999 | auth | apps/boot/auth |
| 2001-2999 | user | apps/tenant/user |
| 3001-3999 | tenant / platform_tenant | apps/platform/tenants |
| 4001-4999 | role | apps/tenant/role |
| 5001-5999 | menu | apps/tenant/menu |
| 6001-6999 | organization | apps/tenant/organization |
| 7001-7999 | permission | apps/tenant/permission |
| 8001-8999 | resource | apps/tenant/resource |
| 9001-9999 | asset | apps/reference/asset |
| 10001-10999 | dict | apps/reference/dict |
| 11001-11999 | system | apps/system |
| 12001-12999 | weixin | apps/reference/weixin |
| 13001-13999 | flag | apps/flag |
| 14001-14999 | cms | apps/cms |
| 15001-15999 | sys_menu | apps/platform/sys_menu |
| 18001-18999 | config | apps/reference/config |

> **新增模块找段**：从 16000+ 或 19000+ 找空段。

## 8. 数据层核心约定

### 8.1 多租户隔离（RLS）

业务表通过 `db.RunInTenantTx(ctx, pool, tenantID, fn)` 自动 `SET LOCAL app.tenant_id`，配合 PG 的 Row-Level Security 策略实现强隔离。

```go
err := db.RunInTenantTx(ctx, s.pool, uc.TenantID, func(txCtx context.Context) error {
    q, _ := db.GetQuerier(txCtx, s.pool)
    return s.repo.GetByID(txCtx, userID)
})
```

平台管理用 `db.RunInPlatformTx(ctx, pool, fn)` 跳过 RLS。

### 8.2 JSONB 字段（必须 `::jsonb` cast）

11 列 JSONB：`db_logs.old_data/new_data`、`config_items.value/default_value/options/validation`、`dicts.extend`、`dict_items.extend`、`tenant_roles.extend`、`sys_roles.extend`、`tenants.config`。

pgx 默认把 Go `string` 当 `text` 发、`[]byte` 当 `bytea` 发，写 JSONB 列会报 `42804`。SQL 必须显式 `::jsonb` cast：

```sql
UPDATE t SET value = $1::jsonb WHERE id = $2
UPDATE t SET value = COALESCE($1::jsonb, value) WHERE id = $2
```

### 8.3 软删除

所有业务表都有 `is_deleted BOOLEAN DEFAULT FALSE`，唯一索引都是 partial index：

```sql
CREATE UNIQUE INDEX uk_tenant_users_account_tenant ON tenant_users (account_id, tenant_id)
    WHERE is_deleted = FALSE;
```

## 9. 重构历程（Phase 0-0023）

| Phase | 内容 |
|---|---|
| 0 | 摸底：找到 16 个跨模块全局、109 处引用 |
| 1-2 | 建 module / AppContext 骨架 |
| 3-4c | 删全局变量（authz/registry/ext_impl/middleware wrapper） |
| 5 | 新 module + main.go 4 步显式 Build |
| 001x | cms/flag 等示例业务补齐 |
| 0020 | platform_tenant 从 `apps/boot/tenant` 迁到 `apps/platform/tenants` |
| 0021 | 新增 sys_menu 模块 |
| 0022 | config 完全重构 + 全分离 Phase（登录入口拆 tenant-login / platform-login；三域路由） |
| **0023** | **平台/租户域物理拆分**：9 张表 rename、`account_roles` drop、Go 包 rename（`apps/rbac→apps/tenant`、`framework/pkg/rbac→framework/pkg/tenant/auth`）、新增 `sys_user` / `sys_role` / `sys_permission`、登录路径切到 `sys_user_roles` |

### 9.1 重构前 vs 重构后

| 维度 | 重构前 | 重构后 |
|---|---|---|
| Go modules | 3 个 | **1 个**（`gx1727.com/xin`） |
| 跨模块全局变量 | 12 个 | 1 个（`authz.Authorization` interface） |
| 模块数 | 15 | **19**（3 alwaysOn + 8 optOut + 8 optional） |
| 路由空间 | 业务 + 管理 | 业务 + 平台（`/platform/*`）+ 公开（`/public/*`） |
| 数据流传递方式 | 隐式（全局包） | 显式（AppContext） |
| 编译期可追踪 | 否 | ✓（Reader/Writer 接口） |
| P0 单测 | 无 | 36+ 个 |

## 10. 延伸阅读

| 文档 | 内容 |
|---|---|
| [doc/quickstart.md](quickstart.md) | 装 PG、跑 migration、首次跑 `xin run` |
| [doc/modules.md](modules.md) | 19 个 module 的清单和职责 |
| [doc/database.md](database.md) | 表结构、RLS、迁移机制、JSONB |
| [doc/permissions.md](permissions.md) | RBAC + 数据范围 + 平台角色 |
| [doc/developing.md](developing.md) | 新增模块的 8 步流程 |
| [doc/deployment.md](deployment.md) | 编译、systemd、Docker |
| [doc/api.md](api.md) | 完整路由 API 参考 |
