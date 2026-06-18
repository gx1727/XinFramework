# 架构总览

> 这是 XinFramework 最关键的文档。建议第一次接触代码的人从这里开始。

## 1. 双 Go Module 结构

仓库根目录下有两个独立的 Go module,通过 `go.mod` 隔离:

```
server/
├── go.mod         (path: gx1727.com/xin/server)      # cmd/ + migrations/
├── framework/
│   └── go.mod     (path: gx1727.com/xin/framework)  # 框架本体
└── apps/
    └── go.mod     (path: gx1727.com/xin/apps)       # 业务模块
```

**为什么双 module?** 业务模块和框架的发布节奏完全不同——框架一旦稳定就不应该动,业务模块可以高频迭代。Go module 的依赖方向必须严格:

```
cmd/xin ──→ framework ──→ apps
            (框架核心)    (业务)
```

不允许反向:`framework` 不能 `import apps`,`apps` 不能 `import cmd/xin`。这两个边界用 `framework/internal/` 强制(`internal` 包的 import 限制在同一 module 树内)。

## 2. 启动时序

`framework.Run(cfg)` 执行的精确顺序,见 [framework/framework.go](framework/framework.go):

```go
func Run(cfg *config.Config) {
    // 1. 命令行分发(start/stop/restart/run/help)
    // 2. runServer(cfg)
    runServer(cfg)
}

func runServer(cfg *config.Config) {
    // 3. boot.Init 装载全部基础设施
    app, err := boot.Init(cfg)  // panic on failure
    // 4. 遍历 plugin.Apps(),按 cfg.Module 白名单 Init/Register
    initModules(app)
    // 5. 数据库迁移(./migrations/*.sql)
    runMigrations()
    // 6. 装配全局中间件 + 各 module Register 路由
    setupRouter(app)
    // 7. 后台启动 HTTP server
    go app.Server.Start(addr)
    // 8. 等待 SIGINT/SIGTERM,优雅退出
    waitForSignal(app.Server, app)
}
```

`boot.Init` 的内部 6 步装配([framework/internal/core/boot/boot.go](framework/internal/core/boot/boot.go)):

```go
func Init(cfg *config.Config) (*App, error) {
    logger.Init(cfg.Log.Dir, cfg.Log.Level)
    db.Init(&cfg.Database)                        // ① pgxpool
    dict.Init(db.Get())
    cache.Init(&cfg.Redis)                        // ② go-redis (enabled)
    sm := session.NewRedisSessionManager()        // ③ Redis 优先
    permCache := permission.NewRedisPermissionCache()
    appCtx := plugin.NewAppContext(...)           // ④ 唯一的依赖容器
    ext_impl.InitExtApi(appCtx)
    permService := service.NewPermissionService(...)  // ⑤ RBAC 服务
    appCtx.SetAuthz(authz.Wrap(authzService))     // ⑥ 跨 module 共享
    return &App{...}, nil
}
```

## 3. 模块生命周期:Init / Register / Shutdown

每个 module 实现 `plugin.Module` 接口([framework/pkg/plugin/plugin.go](framework/pkg/plugin/plugin.go)):

```go
type Module interface {
    Name() string
    Init(ctx Reader, w Writer) error                              // 写 own slots
    Register(ctx Reader, public, protected *gin.RouterGroup)      // 路由
    Shutdown(ctx Reader) error                                   // 释放资源
}
```

推荐用 `BaseModule` struct(避免每个 module 写自己的 method set):

```go
return &plugin.BaseModule{
    NameStr: "tenant",
    InitFn: func(_ plugin.Reader, w plugin.Writer) error {
        w.SetTenantRepo(&tenantPkgAdapter{repo: NewTenantRepository(db.Get())})
        return nil
    },
    RegFn: func(_ plugin.Reader, _, protected *gin.RouterGroup) {
        h := NewHandler(NewService(NewTenantRepository(db.Get())))
        Register(protected, h)
    },
}
```

### 3.1 为什么 Init 阶段 Writer 要传 nil-friendly Reader?

- Writer 是**写自己负责的 slot**:`SetAccountRepo` / `SetTenantRepo` / `SetUserRepo` ...
- Reader 是**读别人贡献的 slot**:`AccountRepo()` / `TenantRepo()` / `UserRepo()` ...
- 模块必须 nil-check Reader 返回的 Repository(可能 producer module 被 `cfg.Module` 关闭了)
- 模块**永远不会**拿到写别人 slot 的 Writer,这是编译期类型保证

### 3.2 Register 阶段拿到完整 Reader

到 Register 时,所有模块的 Init 都已完成。`framework/framework.go::registerModules` 把 `app.AppContext` 作为 Reader 传给每个 module。

## 4. AppContext:唯一的依赖容器

[framework/pkg/plugin/appcontext.go](framework/pkg/plugin/appcontext.go) 是整个重构的成果物。**两件不变量**:

1. **构造一次,终身不变** —— 在 `boot.Init` 中构造,后续只读
2. **Reader / Writer 接口分离** —— 让"读了别人的 repo" 和 "写了别人的 repo" 在类型系统上不可能

### 4.1 接口定义

```go
type Reader interface {
    // 基础设施 (Init 之前就填好)
    DB()       *pgxpool.Pool
    Cache()    *redis.Client           // 可能 nil
    Config()   *config.Config
    Session()  session.SessionManager

    // 跨模块贡献 (Init 完成后填好)
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

### 4.2 为什么 AppContext 是 concrete struct 而非 interface?

- **构造期 panic**:NewAppContext 校验 db / cfg 非 nil,在启动期暴露配置错误
- **零运行时断言**:Reader/Writer 是接口,struct 同时实现两者,编译期 `var _ Reader = (*AppContext)(nil)`
- **测试友好**:测试可以传一个 `&AppContext{db: fakePool}` 而非 mock 整个 interface

## 5. 中间件链

[framework/framework.go](framework/framework.go) `setupRouter` 注册全局中间件(按顺序):

```go
srv.Engine.Use(
    middleware.Recovery(),       // 1. panic recover,最先
    middleware.RequestID(),      // 2. 注入 X-Request-ID
    middleware.CORS(&cfg.CORS),  // 3. CORS 预检
    middleware.ClientIP(),       // 4. 客户端 IP(供审计)
    middleware.Logger(),         // 5. access log(依赖 RequestID)
)
```

然后在 `/api/v1` 路由组里挂两个分组:

```go
public := v1.Group("")
public.Use(middleware.OptionalAuth(...))   // 可选登录

protected := v1.Group("")
protected.Use(middleware.Auth(...))        // 必须登录
```

### 5.1 Auth 中间件做了什么

[framework/internal/core/middleware/auth.go](framework/internal/core/middleware/auth.go):

1. 从 `Authorization: Bearer <jwt>` 提取 token
2. JWT 验证(HS256 + `cfg.JWT.Secret`)
3. Session 验证(去 Redis 或 DB 查 SessionID)
4. 把 `XinContext` 注入到 `c.Request.Context()`
5. **懒加载** `UserContextLoader`(第一次有人 `MustNewUserContext(c)` 才查 DB)

为什么不立即查权限?

- `List` 路由可能只查元数据,不需要权限校验
- `GetCurrentUser` 路由只需要身份,不需要数据范围
- 懒加载 + `sync.Once` 保证单请求只查一次

### 5.2 RBAC 中间件

[framework/pkg/middleware/auth.go](framework/pkg/middleware/auth.go) 暴露给业务模块用:

| 函数 | 行为 |
|---|---|
| `Require(spec)` | 一个 spec 必须满足 |
| `RequireAny(specs...)` | 任一 spec 满足即可 |
| `RequireAll(specs...)` | 所有 spec 都必须满足 |
| `RequireAuthenticated()` | 登录即可,不查 RBAC |
| `RequirePlatformRole(roles...)` | 必须持有平台角色(跨租户) |

`Spec` 由 [framework/pkg/permission/spec.go](framework/pkg/permission/spec.go) 定义:

```go
spec := permission.P("user", "list")  // resource=user, action=list
spec := permission.AuthOnly()          // 仅登录
```

`super_admin` 平台角色自动 bypass 所有 RBAC(spec 不需要写通配)。

## 6. 响应协议

[framework/pkg/resp/resp.go](framework/pkg/resp/resp.go):

```json
// 成功
{ "code": 0, "msg": "ok", "data": { ... } }

// 业务错误
{ "code": 2001, "msg": "用户不存在", "data": null }

// 分页
{ "code": 0, "msg": "ok", "data": { "total": 100, "list": [ ... ] } }
```

**错误码分段管理**(每个 module 一个区段,[resp/errors.go](framework/pkg/resp/errors.go)):

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

## 7. 重构历程(Phase 0-8)

2026 年完成的一次大重构,把 12 个跨模块全局变量全部迁到 AppContext。

### 7.1 重构前 vs 重构后

| 维度 | 重构前 | 重构后 |
|---|---|---|
| 跨模块全局变量 | 12 个 | 1 个(authz.Authorization interface) |
| 数据流传递方式 | 隐式(全局) | 显式(AppContext) |
| 编译期可追踪 | ✗ | ✓(Reader/Writer 接口) |
| Test mocking | 难(全局副作用) | 易(可注入 fake Reader) |
| 删 dead code | — | 525 行 |
| 新增 P0 单测 | — | 36 个,3 包覆盖率 48.4% |

### 7.2 Phase 时间线

| Phase | 内容 | 关键改动 |
|---|---|---|
| **0** | 摸底 | 找到 16 个跨模块全局,409 处引用 |
| **1** | go.mod 修复 | 拆出 framework / apps 两个独立 module |
| **2** | AppContext 骨架 | 定义 Reader / Writer 接口,BaseModule 引入 |
| **3** | auth + tenant | 删 `framework/pkg/auth/registry.go` 和 `tenant/registry.go` |
| **4** | rbac 4 件套 | user / role / organization / permission 全部走 AppContext |
| **5** | authz | 删 `authz.global`、`service.globalAuthorizationService`、`boot.globalApp + AppInstance()`,8 处 apps cache 失效切到 `ctx.Authz()` |
| **6** | ext_impl | 删 `ext_impl/registry.go`(189 行死代码) |
| **7** | middleware | 删 `internal/middleware/auth.go` 5 个 wrapper 死函数(53 行) |
| **8** | P0 单测 | 36 个测试,permission / middleware / plugin 包 |

### 7.3 为什么这次重构值得做

代码现在的状态:

```go
// apps/rbac/role/module.go
func (m *Module) RegFn(ctx plugin.Reader, _, protected *gin.RouterGroup) {
    // 显式声明依赖:authz 必须存在
    if authz := ctx.Authz(); authz != nil {
        // ... 用 authz.InvalidateRole(roleID) 失效缓存
    }
    // 显式取 own repo
    roleRepo := ctx.RoleRepo()
    svc := NewService(roleRepo, ctx.Authz())
    Register(protected, NewHandler(svc))
}
```

对比之前:

```go
// apps/rbac/role/service.go(Phase 5 之前)
func (s *Service) Update(...) {
    if err := s.repo.Update(...); err != nil { return err }
    // 隐式查全局
    if a := authz.Get(); a != nil { a.InvalidateRole(...) }
    return nil
}
```

**编译期保证 + 显式依赖 + 易测试 + 易追踪** —— 4 个 property 全 get。

详细 Phase 5 静态分析见 [refactor/phase5_static_analysis.md](refactor/phase5_static_analysis.md),Phase 0 摸底数据见 [refactor/phase0/globals.md](refactor/phase0/globals.md),重构方案见 [方案.md](方案.md)。

## 8. 延伸阅读

| 文档 | 内容 |
|---|---|
| [doc/quickstart.md](quickstart.md) | 装 PG、跑 migration、首次 `xin run` |
| [doc/modules.md](modules.md) | 14 个 module 的清单和职责 |
| [doc/database.md](database.md) | 表结构、RLS、迁移机制 |
| [doc/permissions.md](permissions.md) | RBAC + 数据范围 + 平台角色 |
| [doc/developing.md](developing.md) | 新增 module 的标准 8 步 |
| [doc/deployment.md](deployment.md) | 编译、systemd、Docker |
| [doc/api.md](api.md) | 100+ 路由的 API 参考 |