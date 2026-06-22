# 架构总览

> XinFramework 最关键的设计文档。第一次接触代码从这里开始�?>
> 最后更新：2026-06（Phase 0022 全分离：业务/平台/公开 三域路由 + 登录入口分离�?
## 1. �?Go Module 结构

仓库根目录下只有一�?Go module�?
```
server/
└── go.mod         (path: gx1727.com/xin)         # cmd/ + migrations/ + framework/ + apps/
```

历史上有�?`framework` / `apps` / `cmd` 三模块拆分，**Phase 1 已经合并**为单一 module。原因：

- 三个 module 之间本就不应该独立发布（发布节奏强耦合�?- `internal/` 已经强制�?framework �?apps 之间的可见性边�?- �?module �?`cmd/xin` 可以显式 import `apps/<x>` 并放�?`[]plugin.Module`，无 `replace` 指令�?
依赖方向�?`internal/` 强制�?
```
cmd/xin ──�?framework ──�?apps
            (internal)     (apps 不可�?import)
```

- `framework` 不能 import `apps`（编译就过不去）
- `apps` 不能 import `cmd/xin`
- `apps` 之间**只能**通过 `plugin.AppContext.Reader/Writer` 接口通信

## 2. 启动时序

[`cmd/xin/main.go`](cmd/xin/main.go) 4 步显�?Build�?
```go
func main() {
    cfg, err := config.Load("config/config.yaml")        // 1. 配置
    app, err := framework.Boot(cfg)                      // 2. 装配 *appx.App
    modules := []plugin.Module{                          // 3. 显式模块列表�?6 个）
        // alwaysOn
        auth.Module(app), tenant.Module(app), system.Module(app),
        // optOut
        menu.Module(app), user.Module(app), role.Module(app),
        organization.Module(app), permission.Module(app), resource.Module(app),
        asset.Module(app), dict.Module(app),
        // optional
        refconfig.Module(app), platformmenu.Module(app),
        weixin.Module(app), cms.Module(app), flag.Module(app),
    }
    framework.Serve(cfg, app, modules)                   // 4. 启动
}
```

[`framework.Serve`](framework/framework.go) 内部�?
```go
func Serve(cfg *config.Config, app *appx.App, modules []plugin.Module) {
    migrate.Run(app.DB, "migrations")       // �?SQL 迁移（幂等）
    for _, m := range modules {             // �?Init 阶段
        m.Init(ctx, w)                      //    模块�?own slot
    }
    setupRouter(app, modules)               // �?中间�?+ 路由
    go app.Server.Start(addr)               // �?后台监听
    waitForSignal(app.Server, app)          // �?SIGINT/SIGTERM 优雅退�?}
```

[`framework.Boot`](framework/internal/core/boot/boot.go)（即 `boot.Init`）的 6 步装配：

```go
func Init(cfg *config.Config) (*appx.App, error) {
    logger.Init(cfg.Log.Dir, cfg.Log.Level)
    pool, _      := db.Init(ctx, &cfg.Database)        // �?pgxpool
    dict.Init(pool)
    cache.Init(&cfg.Redis)                            // �?go-redis (enabled)
    sm            := session.NewRedisSessionManager()  // �?Session
    permCache     := permission.NewRedisPermissionCache()
    appCtx        := plugin.NewAppContext(...)         // �?唯一的依赖容�?    ext_impl.InitExtApi(appCtx)
    permService   := service.NewPermissionService(...)// �?RBAC 服务
    appCtx.SetAuthz(authz.Wrap(authService))           // �?跨模块共�?    return &appx.App{Config, DB, SessionMgr, Server,
                     PermService, Authz, AppContext}, nil
}
```

## 3. 模块生命周期：Init / Register / Shutdown

每个 module 实现 [`plugin.Module`](framework/pkg/plugin/plugin.go) 接口�?
```go
type Module interface {
    Name() string
    Init(ctx Reader, w Writer) error                              // �?own slot
    Register(ctx Reader, public, protected *gin.RouterGroup)      // 路由
    Shutdown(ctx Reader) error                                   // 释放资源
}
```

实际形态用 `BaseModule` struct（避免每�?module 都写 method set）：

```go
// Phase 5 之后统一形态（Phase 0022 后稳定）
func Module(app *appx.App) plugin.Module {
    return &plugin.BaseModule{
        NameStr: "platform_tenant",
        InitFn: func(_ plugin.Reader, w plugin.Writer) error {
            // �?own slot
            w.SetTenantRepo(&tenantRepoAdapter{repo: NewTenantRepository(app.DB)})
            return nil
        },
        RegFn: func(_ plugin.Reader, _, protected *gin.RouterGroup) {
            // 注册路由
            h := NewHandler(NewService(app.DB, NewTenantRepository(app.DB)))
            Register(protected, h)
        },
    }
}
```

### 3.1 Init 阶段：Writer �?own slot，Reader 是别人贡献的

- **Writer** = 写自己负责的 slot：`SetAccountRepo` / `SetTenantRepo` / `SetUserRepo` ...
- **Reader** = 读别人贡献的 slot：`AccountRepo()` / `TenantRepo()` / `UserRepo()` ...
- 模块必须 nil-check Reader 返回�?Repository（producer 可能�?`cfg.Module` 关闭�?- 模块**永远不会**拿到写别�?slot �?Writer，编译期类型保证

### 3.2 Register 阶段：拿到完�?Reader

�?Register 时，所有模块的 Init 已完成。`framework.go::setupRouter` �?`app.AppContext` 作为 Reader 传给每个 module�?
### 3.3 三类模块（按 `cfg.Module` 行为�?
| 类型 | 数量 | 表现 |
|---|---:|---|
| **alwaysOn** | 3（`system`, `auth`, `platform_tenant`�?| 启动必需，无法关闭，配置不列也加回去 |
| **optOut** | 8（`menu`, `user`, `role`, `resource`, `organization`, `dict`, `asset`, `permission`�?| 默认启用；用户写 `module:` 时切白名单语义（不列就关�?|
| **optional** | 5（`config`, `weixin`, `platform_menu`, `cms`, `flag`�?| 默认不启用；必须�?`cfg.Module` 显式列出才加�?|

定义�?[`framework/pkg/config/config.go`](framework/pkg/config/config.go) `alwaysOnModules` / `optOutModules`�?
## 4. AppContext：唯一的依赖容器（Phase 0022 不变�?
[`framework/pkg/plugin/appcontext.go`](framework/pkg/plugin/appcontext.go) 是整个重构的成果物�?*两件不变�?*�?
1. **构造一次，终身不变**——在 `boot.Init` 中构造，后续只读
2. **Reader / Writer 接口分离**—�?读别�?repo" �?"写别�?repo" 在类型系统上不可�?
### 4.1 接口定义（当前）

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
    UserRepo()         rbac.UserRepository
    RoleRepo()         rbac.RoleRepository
    OrgRepo()          rbac.OrganizationRepository
    PermRepo()         rbac.RoleResourceRepository
}

type Writer interface {
    SetAuthz(authz.Authorization)
    SetAccountRepo(auth.AccountRepository)
    SetAccountAuthRepo(auth.AccountAuthRepository)
    SetTenantRepo(tenant.TenantRepository)
    SetUserRepo(rbac.UserRepository)
    SetRoleRepo(rbac.RoleRepository)
    SetOrgRepo(rbac.OrganizationRepository)
    SetPermRepo(rbac.RoleResourceRepository)
}
```

> 当前 AppContext 提供 **8 个跨模块 Repo slot**（Authz + 7 �?Repository）。新增跨模块接口的步骤：
>
> 1. �?`framework/pkg/<scope>/xxx.go` 定义�?interface + struct
> 2. �?appcontext.go Reader/Writer 接口加方�?> 3. �?`AppContext` struct 加字�?+ getter/setter
> 4. provider module �?`InitFn` �?`w.SetXxx(myRepo)`
>
> 编译会引导你走完整链路——任何缺漏都会编译失败�?
### 4.2 为什�?AppContext �?concrete struct 而非 interface�?
- **构造期 panic**：`NewAppContext` 校验 db / cfg �?nil，在启动期暴露配置错�?- **零运行时断言**：Reader/Writer 是接口，struct 同时实现两者，编译�?`var _ Reader = (*AppContext)(nil)`
- **测试友好**：测试可以传一�?`&AppContext{db: fakePool}` 而非 mock 整个 interface

## 5. 中间件链

[`framework/framework.go`](framework/framework.go) `setupRouter` 注册全局中间件（按顺序）�?
```go
srv.Engine.Use(
    middleware.Recovery(),       // 1. panic recover，最�?    middleware.RequestID(),      // 2. 注入 X-Request-ID
    middleware.CORS(&cfg.CORS),  // 3. CORS 预检
    middleware.ClientIP(),       // 4. 客户�?IP（供审计�?    middleware.Logger(),         // 5. access log（依�?RequestID�?)
```

然后�?`/api/v1` 路由组里�?**三组 RouterGroup**（Phase 0022 拆分）：

```go
public := v1.Group("")
public.Use(middleware.OptionalAuth(...))         // 可选登�?
tenant := v1.Group("/t")
tenant.Use(middleware.Auth(...))                  // 必须登录
tenant.Use(pkgmiddleware.RequireTenantContext())  // 必须 tenant_id > 0

protected := v1.Group("/admin")
protected.Use(middleware.Auth(...))              // 必须登录（平台域�?// 模块内部追加 RequirePlatformRole("super_admin")
```

### 5.1 Auth 中间件做了什�?
[`framework/internal/core/middleware/auth.go`](framework/internal/core/middleware/auth.go)�?
1. �?`Authorization: Bearer <jwt>` 提取 token
2. JWT 验证（HS256 + `cfg.JWT.Secret`�?3. Session 验证（去 Redis �?DB �?SessionID�?4. �?`XinContext` 注入�?`c.Request.Context()`
5. **懒加�?* `UserContextLoader`（第一次有�?`MustNewUserContext(c)` 才查 DB�?
### 5.2 RBAC 中间�?
[`framework/pkg/middleware/auth.go`](framework/pkg/middleware/auth.go) 暴露给业务模块用�?
| 函数 | 行为 |
|---|---|
| `Require(spec)` | 一�?spec 必须满足 |
| `RequireAny(specs...)` | 任一 spec 满足即可 |
| `RequireAll(specs...)` | 所�?spec 都必须满�?|
| `RequireAuthenticated()` | 登录即可，不�?RBAC |
| `RequirePlatformRole(roles...)` | 必须持有平台角色（跨租户�?|
| **`RequireTenantContext()`** | **Phase 0022 新增**：tenant_id > 0 才放�?|
| **`RequirePlatformScope()`** | **Phase 0022 新增**：tenant_id == 0 才放�?|

`Spec` �?[`framework/pkg/permission/spec.go`](framework/pkg/permission/spec.go) 定义�?
```go
spec := permission.P("user", "list")     // resource=user, action=list
spec := permission.AuthOnly()            // 仅登�?```

**`super_admin` 平台角色自动 bypass 所�?RBAC**（spec 不需要写通配）�?
## 6. 路由空间分布

Phase 0022 终极拆分：所有路由分�?**3 个语义空�?*，URL 自解�?我在哪个�?�?
| 空间 | 前缀 | 中间�?| 说明 |
|---|---|---|---|
| **业务消费** | `/api/v1/<resource>` | Auth + RequireTenantContext + Require(ResX) | 租户域业务接口（用户/角色/字典/配置等） |
| **平台管理** | `/api/v1/platform/<platform_resource>` | Auth + RequirePlatformRole(super_admin) + Require(ResX) | super_admin 跨租户接口（平台菜单/平台租户/平台配置/平台字典�?|
| **公开访问** | `/api/v1/public/<resource>` �?`/api/v1/<auth>` | OptionalAuth | 不需登录，公共读（如 `/public/configs`、`/auth/tenant-login`、`/auth/platform-login`�?|

**完整示例**�?
```
# 业务域（需登录 + tenant_id�?POST   /api/v1/users
GET    /api/v1/menus/tree
GET    /api/v1/configs/resolve?code=site
POST   /api/v1/configs/:id/items/:item_id/override
GET    /api/v1/dicts/resolve

# 平台域（强制 super_admin�?GET    /api/v1/platform/platform-menus
GET    /api/v1/platform/platform-menus/tree
POST   /api/v1/platform/platform-tenants
GET    /api/v1/platform/platform-configs
GET    /api/v1/platform/platform-dicts

# 公开域（可选登录）
POST   /api/v1/auth/tenant-login      (业务用户登录，必�?tenant_id)
POST   /api/v1/auth/platform-login    (平台管理员登录，无需 tenant_id)
POST   /api/v1/auth/register
POST   /api/v1/auth/refresh
GET    /api/v1/health
GET    /api/v1/public/configs          (公开配置�?
POST   /api/v1/weixin/login            (微信小程序登�?
```

### 6.1 三组 RouterGroup 在代码中的对�?
```go
// framework.go::registerModules
v1 := r.Group("/api/v1")

public := v1.Group("")                                 // /api/v1/*
public.Use(middleware.OptionalAuth(...))

tenant := v1.Group("/t")                                // /api/v1/*
tenant.Use(middleware.Auth(...))
tenant.Use(pkgmiddleware.RequireTenantContext())

protected := v1.Group("/admin")                         // /api/v1/platform/*
protected.Use(middleware.Auth(...))
// 平台模块内部追加 RequirePlatformRole("super_admin")
```

业务模块接收三个 RouterGroup 引用，由模块决定把路由挂在哪一组：

```go
func (m *Module) RegFn(_ plugin.Reader,
    public *gin.RouterGroup, tenant *gin.RouterGroup, protected *gin.RouterGroup) {
    // public:    /api/v1/public/*            （OptionalAuth�?    // tenant:    /api/v1/*                 （Auth + RequireTenantContext�?    // protected: /api/v1/platform/*             （Auth；模块内�?RequirePlatformRole�?}
```

每个 platform 模块�?routes.go 都遵�?`adminGroup := protected.Group("/platform/<x>", RequirePlatformRole(...))` 模式（见 `apps/platform/platform_menu/routes.go`、`apps/platform/platform_tenant/routes.go`）�?
### 6.2 兼容期路�?
过渡期保留以下兼容路径，客户�?SDK / �?curl 脚本不会立即失效�?
| 旧路�?| 重定向到 |
|---|---|
| `POST /api/v1/auth/login` | `POST /api/v1/auth/tenant-login`（handler 内部转发�?|
| `GET /api/v1/dashboard` | `GET /api/v1/app/dashboard` |
| `GET /api/v1/tenants` | `GET /api/v1/platform/tenants` |
| `GET /api/v1/menus` | `GET /api/v1/app/menus` |
| `GET /api/v1/dicts` | `GET /api/v1/app/dicts` |
| `GET /api/v1/configs` | `GET /api/v1/app/configs` |

移除时机由后�?phase 决定（建议观�?1-2 个月无流量后下线）�?
## 7. 响应协议

[`framework/pkg/resp/resp.go`](framework/pkg/resp/resp.go)�?
```json
// 成功
{ "code": 0, "msg": "ok", "data": { ... } }

// 业务错误
{ "code": 2001, "msg": "用户不存�?, "data": null }

// 分页
{ "code": 0, "msg": "ok", "data": { "total": 100, "list": [ ... ] } }
```

### 7.1 错误码分段管�?
[`framework/pkg/resp/errors.go`](framework/pkg/resp/errors.go) 集中定义，每�?module 一�?1000 段：

| 区段 | module | 文件位置 |
|---|---|---|
| 1001-1999 | auth | apps/boot/auth |
| 2001-2999 | user | apps/rbac/user |
| 3001-3999 | tenant / **platform_tenant** | apps/rbac/... �?apps/platform/platform_tenant |
| 4001-4999 | role | apps/rbac/role |
| 5001-5999 | menu | apps/rbac/menu |
| 6001-6999 | organization | apps/rbac/organization |
| 7001-7999 | permission | apps/rbac/permission |
| 8001-8999 | resource | apps/rbac/resource |
| 9001-9999 | asset | apps/reference/asset |
| 10001-10999 | dict | apps/reference/dict |
| 11001-11999 | system | apps/system |
| 12001-12999 | weixin | apps/reference/weixin |
| 13001-13999 | flag | apps/flag |
| 15001-15999 | **platform_menu** | apps/platform/platform_menu |
| 18001-18999 | **config**（Phase 0022 重构�?| apps/reference/config |

> **3001-3999 共用**：原 `tenant`（未来业务层租户管理）与 `platform_tenant`（平台管理）共享段，因为底层表相同、错误语义一致�?>
> **新增模块找段**：从 11000 段以上找空段；避开 14001-14999（旧 config 段已废弃）�?5001-15999（platform_menu）�?6001-17999（预留）�?8001-18999（config）�?
## 8. 数据层核心约�?
### 8.1 多租户隔离（RLS�?
业务表通过 `db.RunInTenantTx(ctx, pool, tenantID, fn)` 自动 SET LOCAL `app.tenant_id`，配�?PG �?Row-Level Security 策略实现强隔离�?
```go
err := db.RunInTenantTx(ctx, s.pool, uc.TenantID, func(txCtx context.Context) error {
    q, _ := db.GetQuerier(txCtx, s.pool)
    // SQL 自动�?RLS 限制
    return s.repo.GetByID(txCtx, userID)
})
```

平台管理（如 `/platform/platform-tenants`）用 `db.RunInPlatformTx(ctx, pool, fn)` 跳过 RLS。详�?[database.md](database.md)�?
### 8.2 JSONB 字段（必�?`::jsonb` cast�?
9 �?JSONB 列：`db_logs.old_data/new_data`、`config_items.value/default_value/options/validation`、`dicts.extend`、`dict_items.extend`、`flag_frames.template_config`�?
pgx 默认�?Go `string` �?`text` 发、`[]byte` �?`bytea` 发，�?JSONB 列会�?`42804`。SQL 必须显式 `::jsonb` cast�?
```sql
UPDATE t SET value = $1::jsonb WHERE id = $2
UPDATE t SET value = COALESCE($1::jsonb, value) WHERE id = $2  -- patch 场景
```

### 8.3 软删�?
所有业务表都有 `is_deleted BOOLEAN DEFAULT FALSE`，唯一索引�?partial index�?
```sql
CREATE UNIQUE INDEX uk_users_account ON users (tenant_id, account_id)
    WHERE is_deleted = FALSE;
```

## 9. 重构历程（Phase 0-0022�?
| Phase | 内容 |
|---|---|
| 0 | 摸底：找�?16 个跨模块全局�?09 处引�?|
| 1-2 | �?module / AppContext 骨架 |
| 3-4c | 删全局变量（authz/registry/ext_impl/middleware wrapper�?|
| 5 | �?module + main.go 4 步显�?Build |
| 001x | cms/flag/cms 等示例业务补�?|
| 0020 | platform_tenant �?`apps/boot/tenant` 迁到 `apps/platform/platform_tenant` |
| 0021 | 新增 platform_menu 模块（super_admin 域） |
| 0022a | **config 模块完全重构**（路�?`/config/*` �?`/configs/*`，加 Scope/Visibility/Override/Resolve 三层，错误码段迁移到 18xxx�?|
| **0022b** | **全分�?Phase C**：登录入口拆 `tenant-login` / `platform-login`；所有业务域路由 `/api/v1/<resource>` �?`/api/v1/<resource>`；config/dict 平台域迁�?`/api/v1/platform/platform-<x>`；前�?`App.tsx` �?`/app/*` + `/platform/*`；login �?TenantLogin / PlatformLogin；`RequireTenantContext` / `RequirePlatformScope` 中间�?|

### 9.1 重构�?vs 重构�?
| 维度 | 重构�?| 重构�?|
|---|---|---|
| Go modules | 3 个（cmd/framework/apps�?| **1 �?*（`gx1727.com/xin`�?|
| 跨模块全局变量 | 12 �?| 1 个（`authz.Authorization` interface�?|
| 模块�?| 15 | **16**�?platform_menu�?|
| 路由空间 | 业务 + 业务 | 业务 + 平台�?admin�? 公开�?public�?|
| 数据流传递方�?| 隐式（全局�?| 显式（AppContext�?|
| 编译期可追踪 | �?| ✓（Reader/Writer 接口�?|
| P0 单测 | �?| 36 个，3 包覆盖率 48.4% |

## 10. 延伸阅读

| 文档 | 内容 |
|---|---|
| [doc/quickstart.md](quickstart.md) | �?PG、跑 migration、首�?`xin run` |
| [doc/modules.md](modules.md) | 16 �?module 的清单和职责 |
| [doc/database.md](database.md) | 表结构、RLS、迁移机制、JSONB |
| [doc/permissions.md](permissions.md) | RBAC + 数据范围 + 平台角色 |
| [doc/developing.md](developing.md) | 新增业务模块 / 平台模块�?8 �?|
| [doc/deployment.md](deployment.md) | 编译、systemd、Docker |
| [doc/api.md](api.md) | 完整路由 API 参�?|