# 架构详解

> 本文件描述 XinFramework 的代码层架构、各层职责、依赖关系、生命周期。
> 高层概览见 [project-analysis.md](./project-analysis.md)。

---

## 1. 总体分层

```
┌────────────────────────────────────────────────────────────────┐
│  cmd/xin/main.go                  ← 服务入口（4 步 Build）      │
└────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌────────────────────────────────────────────────────────────────┐
│  framework/                      ← 框架内核（必装）             │
│  ├─ framework.go  Serve / Boot   ← 总入口                      │
│  ├─ runtime.go    Runtime{Server, AppCtx}                       │
│  ├─ cmd.go        xin start/stop/reload/status                  │
│  ├─ internal/                                                  │
│  │  ├─ core/boot/boot.go        ← 启动编排（构造 App+Server）  │
│  │  ├─ core/server/server.go    ← HTTP server（包装 gin）      │
│  │  └─ core/middleware/         ← Recovery/RequestID/CORS/...  │
│  └─ pkg/                                                      │
│     ├─ appx/                    ← 进程级资源容器 {Config, DB}  │
│     ├─ config/                  ← YAML 加载 + env 覆盖 + 校验  │
│     ├─ db/                      ← pool + Querier + Tx 工具     │
│     ├─ plugin/                  ← Module 契约 + AppContext     │
│     ├─ authz/                   ← Authorization 接口           │
│     ├─ permission/              ← Spec/Resource/Action/Scope   │
│     ├─ jwt/                     ← token 签发/校验 + Claims     │
│     ├─ session/                 ← Redis/DB SessionManager      │
│     ├─ cache/                   ← Redis 单例                   │
│     ├─ auth/                    ← AccountRepository 公开契约   │
│     ├─ tenant/                  ← TenantRepository 公开契约    │
│     ├─ sysauth/                 ← sys 域 User/Role/... 契约     │
│     ├─ tenant/auth/             ← 租户域 User/Role/Org 契约    │
│     ├─ middleware/              ← Require / RequireSys         │
│     ├─ audit/                   ← db_logs 写入入口             │
│     ├─ migrate/                 ← 启动期 SQL 迁移              │
│     ├─ resp/                    ← 统一响应 + 错误码分段        │
│     ├─ logger/                  ← zap 包装                     │
│     ├─ storage/{local,cos}/     ← 对象存储实现                 │
│     └─ xincontext/              ← 请求上下文 (Xin/User)        │
└────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌────────────────────────────────────────────────────────────────┐
│  apps/                           ← 业务模块（19 个，可装卸）    │
│  ├─ boot/auth/                   ← 登录 / 账号（必装）          │
│  ├─ sys/                         ← sys 管理域                  │
│  │  ├─ tenants/                  ← sys 租户 CRUD               │
│  │  ├─ user/  role/  menu/  permission/                       │
│  ├─ tenant/                      ← 租户域 RBAC                │
│  │  ├─ user/  role/  menu/  organization/  permission/  resource/  message/ │
│  ├─ reference/                   ← 基础设施                    │
│  │  ├─ asset/  config/  dict/  weixin/                        │
│  ├─ system/                      ← health / cache 运维          │
│  ├─ cms/                         ← 示例 CMS（extapi 模式）     │
│  └─ flag/                        ← 头像框 / 空间 / 头像        │
└────────────────────────────────────────────────────────────────┘
```

**单一 Go module**（`gx1727.com/xin`）包含 framework + apps，无 go workspace 拆分。这样业务模块可以"直接修改 framework"，但实际应通过 PR 走 framework 的演进。

---

## 2. 进程级资源：`appx.App`

`framework/pkg/appx/appx.go` 是**进程级资源容器**：

```go
type App struct {
    DB     *pgxpool.Pool
    Config *config.Config
}
```

- 由 `framework.Boot()` 构造
- 通过 `Module(app *appx.App) plugin.Module` 工厂函数传给每个业务模块
- **业务模块可以读** `app.DB / app.Config`，不需要走 `AppContext`
- 但**写**（设置自己的 repository）必须通过 `AppContext.Writer`

设计意图：
- 基础设施（DB/Config）→ 进程级
- 跨模块共享（Repository）→ `AppContext`
- 单一职责分层清晰

---

## 3. 跨模块容器：`plugin.AppContext`

`framework/pkg/plugin/appcontext.go` 是**模块间解耦的核心**。

### 3.1 Reader / Writer 接口分离

```go
type Reader interface {
    // 基础设施（启动期填充，必非 nil）
    DB() *pgxpool.Pool
    Cache() *redis.Client         // 可能为 nil（Redis 关闭时）
    Config() *config.Config
    Session() session.SessionManager
    Authz() authz.Authorization

    // 跨模块服务（Init 后填充）
    AccountRepo() auth.AccountRepository
    AccountAuthRepo() auth.AccountAuthRepository
    TenantRepo() tenant.TenantRepository
    UserRepo() pkgauth.UserRepository
    RoleRepo() pkgauth.RoleRepository
    OrgRepo() pkgauth.OrganizationRepository
    PermRepo() pkgauth.RoleResourceRepository
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

`AppContext` 同时实现这两个接口。**Init 阶段**：模块拿 `Writer` 写自己负责的槽；**Register 阶段**：拿 `Reader` 读别人的槽。

### 3.2 槽位所有权

| 槽 | 写方（Init 阶段） | 读方（任意阶段） |
|---|---|---|
| `DB() / Cache() / Config() / Session()` | `framework/boot` | 所有模块 |
| `Authz()` | `framework/boot` | middleware（间接） |
| `AccountRepo() / AccountAuthRepo()` | `apps/boot/auth` | `apps/tenant/user` 等 |
| `TenantRepo()` | `apps/sys/tenants` | `apps/boot/auth` 等 |
| `UserRepo()` | `apps/tenant/user` | `apps/tenant/role` 等 |
| `RoleRepo()` | `apps/tenant/role` | `apps/tenant/user` 等 |
| `OrgRepo()` | `apps/tenant/organization` | `apps/tenant/user` 等 |
| `PermRepo()` | `apps/tenant/permission` | `apps/tenant/role` 等 |

Reader 方法**返回 nil 表示对应模块未启用**——调用方必须 nil-check。这避免"启动时未启用的模块被错误引用"。

### 3.3 为什么是接口而不是结构体？

- **Init 阶段**模块不能误调 `SetX()`（接口分离保证）
- **测试**可以构造一个 fake Reader，不用拉起整个 AppContext
- **编译期保证**：业务模块依赖的是 `Reader` 类型，不是 `*AppContext`，未来扩展槽位不影响调用方

---

## 4. HTTP 路由空间

`framework/framework.go:registerModules` 把 `v1 := r.Group("/api/v1")` 拆成三组 RouterGroup：

```
v1 = /api/v1
├── public   = v1.Group("")            # OptionalAuth
│   ├─ /auth/*                         # 登录/注册/refresh/logout
│   ├─ /health                         # health 检查
│   ├─ /public/configs                 # 公开读配置
│   └─ /flag/frames*, /flag/avatar-categories
│
├── tenant   = v1.Group("")            # Auth + RequireTenantContext
│   ├─ /users        /user/profile     # 业务用户
│   ├─ /organizations                  # 组织
│   ├─ /roles        /roles/:id/menus  # 角色
│   ├─ /menus        /menus/tree
│   ├─ /resources                       # 资源权限码
│   ├─ /messages                        # 站内信
│   ├─ /dicts         /dicts/resolve
│   ├─ /configs       /configs/resolve
│   ├─ /asset/upload  /asset/:id
│   └─ /flag/*                          # 头像框
│
└── protected = v1.Group("/sys")    # Auth
    ├─ /sys/tenants/*                  # super_admin 限定
    ├─ /sys/sys-users/*
    ├─ /sys/sys-roles/*
    ├─ /sys/menus/*
    ├─ /sys/sys-permissions/*
    ├─ /sys/dicts/*
    └─ /sys/configs/*
```

**关键不变量**：
- 业务域**无 `/t` 前缀**（历史 `/api/v1/t/users` 已弃用）
- sys 域统一 `/api/v1/sys/*` 前缀
- 公开域走 `/api/v1/*` 但用 `OptionalAuth` 中间件

---

## 5. 启动编排 `internal/core/boot`

`framework/internal/core/boot/boot.go` 实现 `Init(cfg) → (*appx.App, *server.Server, *plugin.AppContext, error)`：

```
boot.Init(cfg)
├── logger.Init                          # zap
├── db.Init → *pgxpool.Pool              # pgxpool + 连接池配置
├── dict.Init(pool)                      # 字典热加载缓存
├── cache.Init → Redis (可降级 nil)
├── session.Init (Redis 或 DB SessionManager)
├── plugin.NewAppContext(db, cache, cfg, session)
├── service.NewAuthorizationService(...) → appCtx.SetAuthz(...)
└── server.New(...)                      # 包装 gin.Engine
```

`framework.Boot(cfg)` 是 `boot.Init` 的公开包外包装（`main.go` 不能直接 import internal）。

---

## 6. 模块生命周期

每个业务模块都通过 `Module(app *appx.App) plugin.Module` 工厂函数返回 `*plugin.BaseModule`：

```go
func Module(app *appx.App) plugin.Module {
    return &plugin.BaseModule{
        NameStr: "user",
        InitFn: func(_ plugin.Reader, w plugin.Writer) error {
            pool := app.DB
            w.SetUserRepo(NewUserRepository(pool))
            return nil
        },
        RegFn: func(ctx plugin.Reader, _ *gin.RouterGroup,
                     tenant *gin.RouterGroup, protected *gin.RouterGroup) {
            // 构造 Handler/Service/Repository，挂路由
        },
        StopFn: func(_ plugin.Reader) error {
            // 释放资源（连接、文件句柄等）
            return nil
        },
    }
}
```

### 6.1 Init 阶段（框架内顺序）

```
framework.Serve
└── for m in modules:
    └── m.Init(ctx, w)    // 模块写自己的 slot
```

**关键**：Init 阶段可以依赖"前面模块已写入的 slot"。但**顺序由 main.go 显式声明的 modules 切片决定**，框架不保证任何隐式顺序。

### 6.2 Register 阶段

```
framework.Serve
├── setupRouter
│   └── for m in modules:
│       └── m.Register(ctx, public, tenant, protected)
```

**关键**：到 Register 阶段时，所有模块的 Init 已完成，Reader 上能拿到所有 Repository。

### 6.3 Shutdown 阶段（信号触发后）

```
waitForSignal → shutdownModules
└── for m in modules:  m.Shutdown(ctx)
```

`Shutdown(ctx Reader)` 接收的 reader 为 nil——目前没有模块需要它；将来若有"读别人 slot 来清理自己"的场景再扩展。

---

## 7. 中间件链

`framework/framework.go:setupRouter` 安装全局中间件（按执行顺序）：

```
1. Recovery()       // panic 恢复，最先执行
2. RequestID()      // 请求 ID
3. CORS(&cfg.CORS)  // 跨域
4. ClientIP()       // 客户端 IP（供 audit 使用）
5. Logger()         // 日志（依赖 RequestID）
```

分组中间件（在 `registerModules` 内）：

| 组 | 链 |
|---|---|
| `public` | `OptionalAuth(&cfg.JWT, sm, authzSvc, db)` |
| `tenant` | `Auth(&cfg.JWT, sm, authzSvc, db)` + `RequireTenantContext()` |
| `protected` | `Auth(...)`（模块内部再追加 `RequireSysRole("super_admin")`） |

`Auth` 注入 `Context`（轻量身份）和 `UserContext` 懒加载器（RBAC + DataScope）；`OptionalAuth` 是 Auth 的弱化版（无 token 时从 `X-Tenant-ID` header 兜底注入 `Context.TenantID`）。

---

## 8. 数据库访问层

`framework/pkg/db/db.go` 是无 ORM 的最小抽象：

### 8.1 核心 API

```go
func Init(ctx context.Context, cfg *config.DatabaseConfig) (*pgxpool.Pool, error)
type Querier interface { Exec / Query / QueryRow }
func WithTx(ctx context.Context, tx pgx.Tx) context.Context
func GetQuerier(ctx context.Context, pool *pgxpool.Pool) (Querier, error)
func RunInTx(ctx context.Context, pool *pgxpool.Pool, fn func(ctx context.Context) error) error
func RunInTenantTx(ctx context.Context, pool *pgxpool.Pool, tenantID uint, fn func(ctx context.Context) error) error
func RunInSysTx(ctx context.Context, pool *pgxpool.Pool, fn func(ctx context.Context) error) error
```

### 8.2 事务工具

- `RunInTx`：**嵌套安全**——已存在事务则复用
- `RunInTenantTx`：套 `app.tenant_id = $tenantID`，触发 RLS 让 SQL 只能看到本租户数据
- `RunInSysTx`：设 `app.tenant_id='0'` + `app.bypass_rls='on'`，sys 域跨租户访问

### 8.3 Repository 写法

```go
// 入口拿 Querier（自动 join 当前事务或 fallback pool）
q, err := db.GetQuerier(ctx, r.db)
if err != nil { return err }

err = q.QueryRow(ctx, `SELECT ...`, id).Scan(&v)
```

业务模块**不直接 import `pgxpool.Pool` 之外的 DB 类型**；Repository 方法第一个参数是 `ctx`，不要在内部新建 context。

---

## 9. 请求上下文

`framework/pkg/xincontext/context.go`：

```go
type Context struct {
    TenantID       uint
    UserID         uint
    SessionID      string
    Role           string
    SysRoles       []string
}

type UserContext struct {
    *Context
    // RBAC + DataScope 懒加载
}
```

`Context` 注入到 `gin.Context`（`Auth` 中间件）：

```go
type Handler struct{}
func (h *Handler) Get(c *gin.Context) {
    xc := xincontext.New(c)
    tenantID := xc.GetTenantID()
    userID := xc.GetUserID()
    // ...
}
```

`UserContext` 包装 `Context`，提供 `LoadPermissions / LoadRoles / LoadDataScope` 懒加载，按需查 DB（由 `authz.AuthorizationService` 实现）。

---

## 10. 错误处理三层

```
┌────────────────────────────────────────┐
│ Repository 层                          │
│   var ErrUserNotFoundDB = errors.New()│  ← 命名 sentinel
│   var ErrDuplicateEmail = errors.New()│
└────────────────────────────────────────┘
              ↓ errors.Is
┌────────────────────────────────────────┐
│ Service 层                             │
│   func mapRepoError(err) error {       │
│     if errors.Is(err, ErrUserNotFoundDB)│
│       return resp.Err(2001, "用户不存在")│
│     return err                          │
│   }                                     │
└────────────────────────────────────────┘
              ↓
┌────────────────────────────────────────┐
│ Handler 层                             │
│   if err != nil {                       │
│     resp.HandleError(c, err)            │
│     return                             │
│   }                                     │
└────────────────────────────────────────┘
              ↓
┌────────────────────────────────────────┐
│ resp.HandleError                       │
│   if errors.As(err, &bizErr) {          │
│     httpStatus := CodeToHTTPStatus(...)│
│     c.JSON(httpStatus, Response{...})   │
│   } else { /* 500 兜底 */ }            │
└────────────────────────────────────────┘
```

**关键约定**：
- DB 层必须用命名 sentinel
- Service 层用 `errors.Is` 翻译
- Handler 层用 `resp.HandleError` 出口
- 业务模块**不自己判断 HTTP 状态码**

---

## 11. 鉴权链路

```
Request → Auth 中间件
    ├─ 解析 Authorization: Bearer <token>
    ├─ jwt.Validate(token, cfg) → Claims
    ├─ session.Validate(claims.SessionID) → 检查 session 存活
    ├─ 把 Claims 灌进 Context
    └─ 注册 UserContext 懒加载器（不立即查 DB）

业务 Handler
    └─ xincontext.NewUserContext(c).LoadPermissions(ctx)
            ↓
    └─ authz.AuthorizationService.LoadUserSecurityContext(ctx, userID)
            ↓
    └─ SELECT ... FROM tenant_role_resources + sys_role_permissions
            ↓
    └─ 返回 map[resource:action]bool + []roles + *DataScope

Require(spec) 中间件
    └─ 检查 spec 出现在 user perms 中（或 super_admin 短路）
```

---

## 12. 关键设计原则

| 原则 | 体现 |
|---|---|
| **依赖倒置** | `AppContext` 注入而非全局；`Authz Authorization` 接口而非具体类型 |
| **显式 Build** | `main.go` 列 `[]plugin.Module{...}`，无 `init()` 注册表 |
| **纵深防御** | DB RLS + 应用层 `RunInTenantTx` + 中间件 `RequireTenantContext` |
| **单一来源** | 配置 → `config.Config`；DB pool → `appx.App`；context 身份 → `Context` |
| **失败不抛** | `audit.Log` 失败仅记日志；Redis 不可用自动降级到 DB session |
| **显式错误** | DB sentinel + service mapRepoError + handler HandleError |
| **可关闭副作用** | Redis `required=false` 时不可用继续；JWT secret 校验仅 prod 强制 |
