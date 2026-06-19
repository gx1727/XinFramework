# 架构总览

> XinFramework 最关键的设计文档。第一次接触代码从这里开始。

## 1. 单 Go Module 结构

仓库根目录下只有一个 Go module：

```
server/
└── go.mod         (path: gx1727.com/xin)         # cmd/ + migrations/ + framework/ + apps/
```

历史上有过 `framework` / `apps` / `cmd` 三模块拆分，**Phase 1 已经合并**为单一 module。原因：

- 三个 module 之间本就不应该独立发布（发布节奏强耦合）
- `internal/` 已经强制了 framework 与 apps 之间的可见性边界
- 单 module 让 `cmd/xin` 可以显式 import `apps/<x>` 并放进 `[]plugin.Module`，无 `replace` 指令坑

依赖方向由 `internal/` 强制：

```
cmd/xin ──→ framework ──→ apps
            (internal)     (apps 不可被 import)
```

- `framework` 不能 import `apps`（编译就过不去）
- `apps` 不能 import `cmd/xin`
- `apps` 之间**只能**通过 `plugin.AppContext.Reader/Writer` 接口通信

## 2. 启动时序

[`cmd/xin/main.go`](cmd/xin/main.go) 4 步显式 Build：

```go
func main() {
    cfg, _ := config.Load("config/config.yaml")              // 1. 配置
    app, _ := framework.Boot(cfg)                            // 2. 装配 *appx.App
    modules := []plugin.Module{                              // 3. 显式模块列表
        auth.Module(app), tenant.Module(app),
        user.Module(app), /* ... 共 15 个 */
    }
    framework.Serve(cfg, app, modules)                       // 4. 启动
}
```

[`framework.Serve`](framework/framework.go) 内部：

```go
func Serve(cfg *config.Config, app *appx.App, modules []plugin.Module) {
    migrate.Run(app.DB, "migrations")       // ① SQL 迁移（幂等）
    for _, m := range modules {             // ② Init 阶段
        m.Init(ctx, w)                      //    模块写 own slot
    }
    setupRouter(app, modules)               // ③ 中间件 + 路由
    go app.Server.Start(addr)               // ④ 后台监听
    waitForSignal(app.Server, app)          // ⑤ SIGINT/SIGTERM 优雅退出
}
```

[`framework.Boot`](framework/internal/core/boot/boot.go)（即 `boot.Init`）的 6 步装配：

```go
func Init(cfg *config.Config) (*appx.App, error) {
    logger.Init(cfg.Log.Dir, cfg.Log.Level)
    pool, _      := db.Init(ctx, &cfg.Database)        // ① pgxpool
    dict.Init(pool)
    cache.Init(&cfg.Redis)                            // ② go-redis (enabled)
    sm            := session.NewRedisSessionManager()  // ③ Session
    permCache     := permission.NewRedisPermissionCache()
    appCtx        := plugin.NewAppContext(...)         // ④ 唯一的依赖容器
    ext_impl.InitExtApi(appCtx)
    permService   := service.NewPermissionService(...)// ⑤ RBAC 服务
    appCtx.SetAuthz(authz.Wrap(authService))           // ⑥ 跨模块共享
    return &appx.App{Config, DB, SessionMgr, Server,
                     PermService, Authz, AppContext}, nil
}
```

## 3. 模块生命周期：Init / Register / Shutdown

每个 module 实现 [`plugin.Module`](framework/pkg/plugin/plugin.go) 接口：

```go
type Module interface {
    Name() string
    Init(ctx Reader, w Writer) error                              // 写 own slot
    Register(ctx Reader, public, protected *gin.RouterGroup)      // 路由
    Shutdown(ctx Reader) error                                   // 释放资源
}
```

实际形态用 `BaseModule` struct（避免每个 module 都写 method set）：

```go
// Phase 5 之后统一形态
func Module(app *appx.App) plugin.Module {
    return &plugin.BaseModule{
        NameStr: "tenant",
        InitFn: func(_ plugin.Reader, w plugin.Writer) error {
            w.SetTenantRepo(&tenantRepoAdapter{repo: NewTenantRepository(app.DB)})
            return nil
        },
        RegFn: func(_ plugin.Reader, _, protected *gin.RouterGroup) {
            h := NewHandler(NewService(app.DB, NewTenantRepository(app.DB)))
            Register(protected, h)
        },
    }
}
```

### 3.1 Init 阶段：Writer 写 own slot，Reader 是别人贡献的

- **Writer** = 写自己负责的 slot：`SetAccountRepo` / `SetTenantRepo` / `SetUserRepo` ...
- **Reader** = 读别人贡献的 slot：`AccountRepo()` / `TenantRepo()` / `UserRepo()` ...
- 模块必须 nil-check Reader 返回的 Repository（producer 可能被 `cfg.Module` 关闭）
- 模块**永远不会**拿到写别人 slot 的 Writer，编译期类型保证

### 3.2 Register 阶段：拿到完整 Reader

到 Register 时，所有模块的 Init 已完成。`framework.go::setupRouter` 把 `app.AppContext` 作为 Reader 传给每个 module。

## 4. AppContext：唯一的依赖容器

[`framework/pkg/plugin/appcontext.go`](framework/pkg/plugin/appcontext.go) 是整个重构的成果物。**两件不变量**：

1. **构造一次，终身不变**——在 `boot.Init` 中构造，后续只读
2. **Reader / Writer 接口分离**——"读别人 repo" 和 "写别人 repo" 在类型系统上不可能

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

### 4.2 为什么 AppContext 是 concrete struct 而非 interface？

- **构造期 panic**：`NewAppContext` 校验 db / cfg 非 nil，在启动期暴露配置错误
- **零运行时断言**：Reader/Writer 是接口，struct 同时实现两者，编译期 `var _ Reader = (*AppContext)(nil)`
- **测试友好**：测试可以传一个 `&AppContext{db: fakePool}` 而非 mock 整个 interface

## 5. 中间件链

[`framework/framework.go`](framework/framework.go) `setupRouter` 注册全局中间件（按顺序）：

```go
srv.Engine.Use(
    middleware.Recovery(),       // 1. panic recover，最先
    middleware.RequestID(),      // 2. 注入 X-Request-ID
    middleware.CORS(&cfg.CORS),  // 3. CORS 预检
    middleware.ClientIP(),       // 4. 客户端 IP（供审计）
    middleware.Logger(),         // 5. access log（依赖 RequestID）
)
```

然后在 `/api/v1` 路由组里挂两个分组：

```go
public := v1.Group("")
public.Use(middleware.OptionalAuth(...))   // 可选登录

protected := v1.Group("")
protected.Use(middleware.Auth(...))        // 必须登录
```

### 5.1 Auth 中间件做了什么

[`framework/internal/core/middleware/auth.go`](framework/internal/core/middleware/auth.go)：

1. 从 `Authorization: Bearer <jwt>` 提取 token
2. JWT 验证（HS256 + `cfg.JWT.Secret`）
3. Session 验证（去 Redis 或 DB 查 SessionID）
4. 把 `XinContext` 注入到 `c.Request.Context()`
5. **懒加载** `UserContextLoader`（第一次有人 `MustNewUserContext(c)` 才查 DB）

为什么不立即查权限？

- `List` 路由可能只查元数据，不需要权限校验
- `GetCurrentUser` 路由只需要身份，不需要数据范围
- 懒加载 + `sync.Once` 保证单请求只查一次

### 5.2 RBAC 中间件

[`framework/pkg/middleware/auth.go`](framework/pkg/middleware/auth.go) 暴露给业务模块用：

| 函数 | 行为 |
|---|---|
| `Require(spec)` | 一个 spec 必须满足 |
| `RequireAny(specs...)` | 任一 spec 满足即可 |
| `RequireAll(specs...)` | 所有 spec 都必须满足 |
| `RequireAuthenticated()` | 登录即可，不查 RBAC |
| `RequirePlatformRole(roles...)` | 必须持有平台角色（跨租户） |

`Spec` 由 [`framework/pkg/permission/spec.go`](framework/pkg/permission/spec.go) 定义：

```go
spec := permission.P("user", "list")     // resource=user, action=list
spec := permission.AuthOnly()            // 仅登录
```

**`super_admin` 平台角色自动 bypass 所有 RBAC**（spec 不需要写通配）。

## 6. 响应协议

[`framework/pkg/resp/resp.go`](framework/pkg/resp/resp.go)：

```json
// 成功
{ "code": 0, "msg": "ok", "data": { ... } }

// 业务错误
{ "code": 2001, "msg": "用户不存在", "data": null }

// 分页
{ "code": 0, "msg": "ok", "data": { "total": 100, "list": [ ... ] } }
```

**错误码分段管理**（每个 module 一个区段，[`resp/errors.go`](framework/pkg/resp/errors.go)）：

| 区段 | module |
|---|---|
| 1001-1999 | auth |
| 2001-2999 | user |
| 3001-3999 | tenant |
| 4001-4999 | role |
| 5001-5999 | menu |
| 6001-6999 | organization |
| 7001-7999 | permission |
| 8001-8999 | resource |
| 9001-9999 | asset |
| 10001-10999 | dict |
| 11001-11999 | system |
| 12001-12999 | weixin |
| 13001-13999 | flag |
| 14001-14999 | config |

## 7. 重构历程（Phase 0-5）

Phase 0-3b 完成跨模块全局变量迁移到 AppContext；Phase 4-5 完成多 module 合并 + main.go 显式 Build。

### 7.1 重构前 vs 重构后

| 维度 | 重构前 | 重构后 |
|---|---|---|
| Go modules | 3 个（cmd/framework/apps） | **1 个**（`gx1727.com/xin`） |
| 跨模块全局变量 | 12 个 | 1 个（`authz.Authorization` interface） |
| 数据流传递方式 | 隐式（全局） | 显式（AppContext） |
| 编译期可追踪 | ✗ | ✓（Reader/Writer 接口） |
| Test mocking | 难（全局副作用） | 易（注入 fake Reader） |
| 删 dead code | — | 525 行 |
| P0 单测 | — | 36 个，3 包覆盖率 48.4% |

### 7.2 Phase 时间线

| Phase | 内容 | 关键改动 |
|---|---|---|
| **0** | 摸底 | 找到 16 个跨模块全局，409 处引用 |
| **1** | go.mod 修复 | 拆出 framework / apps 两个独立 module |
| **2** | AppContext 骨架 | 定义 Reader / Writer 接口，BaseModule 引入 |
| **3** | auth + tenant | 删 `framework/pkg/auth/registry.go` 和 `tenant/registry.go` |
| **3b** | rbac 4 件套 | user / role / organization / permission 走 AppContext |
| **4** | authz | 删 `authz.global`、`service.globalAuthorizationService`、`boot.globalApp + AppInstance()`；8 处 apps cache 失效切到 `ctx.Authz()` |
| **4b** | ext_impl | 删 `ext_impl/registry.go`（189 行死代码） |
| **4c** | middleware | 删 `internal/middleware/auth.go` 5 个 wrapper 死函数（53 行） |
| **5** | 单 module + 显式 Build | 合并 go.mod，main.go 4 步显式 |

### 7.3 为什么这次重构值得做

代码现在的状态：

```go
// apps/rbac/role/module.go（当前）
func (m *Module) RegFn(ctx plugin.Reader, _, protected *gin.RouterGroup) {
    // 显式声明依赖：authz 必须存在
    if authz := ctx.Authz(); authz != nil {
        // ... 用 authz.InvalidateRole(roleID) 失效缓存
    }
    // 显式取 own repo
    roleRepo := ctx.RoleRepo()
    svc := NewService(roleRepo, ctx.Authz())
    Register(protected, NewHandler(svc))
}
```

对比之前：

```go
// apps/rbac/role/service.go（Phase 5 之前）
func (s *Service) Update(...) error {
    if err := s.repo.Update(...); err != nil { return err }
    // 隐式查全局
    if a := authz.Get(); a != nil { a.InvalidateRole(...) }
    return nil
}
```

**编译期保证 + 显式依赖 + 易测试 + 易追踪**——4 个 property 全 get。

## 8. 延伸阅读

| 文档 | 内容 |
|---|---|
| [doc/quickstart.md](quickstart.md) | 装 PG、跑 migration、首次 `xin run` |
| [doc/modules.md](modules.md) | 15 个 module 的清单和职责 |
| [doc/database.md](database.md) | 表结构、RLS、迁移机制 |
| [doc/permissions.md](permissions.md) | RBAC + 数据范围 + 平台角色 |
| [doc/developing.md](developing.md) | 新增 module 的标准 8 步 |
| [doc/deployment.md](deployment.md) | 编译、systemd、Docker |
| [doc/api.md](api.md) | 路由的 API 参考 |
