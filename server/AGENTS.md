# AGENTS.md — XinFramework AI Agent 高密度参考

> 这是给 AI agent 的速查表,内容假设你已经读过 [README.md](README.md) 和 [doc/architecture.md](doc/architecture.md)。

## 1. 仓库结构(记住这张图)

```
server/
├── cmd/xin/main.go              ← 入口
├── config/                      ← YAML 配置
├── migrations/                  ← SQL 迁移
├── framework/                   ← 框架本体
│   ├── framework.go             ← Boot() / Serve() 入口
│   ├── internal/                ← 框架核心实现
│   │   ├── core/{boot,middleware,server,ext_impl}/
│   │   └── service/{authorization,permission}_service.go
│   └── pkg/                     ← 框架公共包（含 appx/appx.go）
└── apps/                        ← 业务模块
    ├── boot/{auth,tenant}/      ← 平台级(alwaysOn)
    ├── rbac/{menu,organization,permission,resource,role,user}/
    ├── reference/{asset,config,dict,weixin}/
    ├── system/                  ← health / cache 运维
    ├── cms/, flag/              ← 示例业务
```

## 2. 模块依赖方向

```
cmd/xin ──→ framework ──→ apps
            (internal)     (internal ← 不可)
```

- `framework` 不能 import `apps`（用 `internal/` 强制）
- `apps` 不能 import `cmd/xin`
- `apps` 间跨模块通信**只能**通过 `plugin.AppContext`（Reader/Writer）
- 单一 go module（Phase 1 已合并 framework + apps + cmd）

修改前确认:你改的文件属于哪个目录?是否跨边界?

## 3. 关键路径速查

| 想找... | 路径 |
|---|---|
| AppContext 设计 | [framework/pkg/plugin/appcontext.go](framework/pkg/plugin/appcontext.go) |
| Module 接口 | [framework/pkg/plugin/plugin.go](framework/pkg/plugin/plugin.go) |
| 启动流程 | [framework/framework.go](framework/framework.go) |
| 启动期装配 | [framework/internal/core/boot/boot.go](framework/internal/core/boot/boot.go) |
| Auth 中间件 | [framework/internal/core/middleware/auth.go](framework/internal/core/middleware/auth.go) |
| RBAC 中间件 | [framework/pkg/middleware/auth.go](framework/pkg/middleware/auth.go) |
| DataScope 编译期 | [framework/pkg/permission/scope.go](framework/pkg/permission/scope.go) |
| 错误码分段 | [framework/pkg/resp/errors.go](framework/pkg/resp/errors.go) |
| Bootstrap | [framework/internal/core/boot/bootstrap.go](framework/internal/core/boot/bootstrap.go) |
| 资源码常量 | [framework/pkg/permission/constants.go](framework/pkg/permission/constants.go) |
| extapi Provider | [framework/internal/core/ext_impl/provider.go](framework/internal/core/ext_impl/provider.go) |
| 用户模块样板 | [apps/rbac/user/module.go](apps/rbac/user/module.go) |
| Dict 模块样板(最小) | [apps/reference/dict/module.go](apps/reference/dict/module.go) |

## 4. 跨模块通信:必须用 AppContext

❌ 错误做法:

```go
import "gx1727.com/xin/apps/rbac/user"
// 编译错误(framework 不能 import apps)
```

❌ 老做法(Phase 5 之前,已删):

```go
import "gx1727.com/xin/framework/pkg/auth"
auth.SetAccountRepository(myRepo)  // 写全局
// 别处:auth.GetAccountRepository()  // 读全局
```

✅ 正确做法(Phase 5 之后,唯一允许):

```go
// 1. 模块 Init 阶段:写 own slot
func (m *Module) InitFn(_ plugin.Reader, w plugin.Writer) error {
    w.SetAccountRepo(&myRepo{...})
    return nil
}

// 2. 其他模块 Register 阶段:读别人 slot
func OtherModule() plugin.Module {
    return &plugin.BaseModule{
        RegFn: func(ctx plugin.Reader, _, protected *gin.RouterGroup) {
            accountRepo := ctx.AccountRepo()
            if accountRepo == nil {
                // producer 模块被关闭,nil 友好
            }
            // ...
        },
    }
}
```

**新增跨模块接口的步骤**:

1. 在 `framework/pkg/{rbac,auth,tenant,...}/xxx.go` 定义窄 interface + struct
2. 在 [framework/pkg/plugin/appcontext.go](framework/pkg/plugin/appcontext.go) `Reader` / `Writer` 接口加方法
3. 在 `AppContext` struct 加字段 + getter/setter
4. provider module 在 `InitFn` 调 `w.SetXxx(myRepo)`

**编译会引导你走完整链路** —— 任何缺漏都会编译失败。

## 5. 命名约定

| 类别 | 约定 | 例子 |
|---|---|---|
| Module name | 小写,无下划线 | `"tenant"`, `"flag"` |
| Resource code | 小写,无下划线 | `"user"`, `"flag"` |
| Action | 小写动词 | `"list"`, `"create"`, `"update"`, `"delete"`, `"get"`, `"tree"` |
| Spec | `permission.P(code, action)` | `permission.P(permission.ResUser, permission.ActList)` |
| Error code | 每个 module 一个区段 | `1001-1999` for auth |
| Endpoint | `/<resource>` 复数 | `/users`, `/roles`, `/tenants` |
| ID param | `:id` | `/users/:id` |
| PATCH vs PUT | PATCH 部分,PUT 整体 | `PATCH /users/:id` vs `PUT /users/:id` |
| 软删除 | `is_deleted BOOLEAN DEFAULT FALSE` | 索引都加 `WHERE is_deleted = FALSE` |
| 时间戳 | `TIMESTAMPTZ DEFAULT NOW()` | `created_at` / `updated_at` |

## 6. 常见模式(直接 copy)

### 6.1 模块骨架

Phase 5 之后**统一形态**：`Module(app *appx.App) plugin.Module`，由 main.go 显式 import 并放进 `[]plugin.Module`：

```go
// apps/feedback/module.go
package feedback

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/appx"
    "gx1727.com/xin/framework/pkg/plugin"
)

// Module returns the feedback module as a BaseModule.
func Module(app *appx.App) plugin.Module {
    return &plugin.BaseModule{
        NameStr: "feedback",
        InitFn: func(_ plugin.Reader, _ plugin.Writer) error { return nil },
        RegFn: func(_ plugin.Reader, _, protected *gin.RouterGroup) {
            pool := app.DB
            svc := NewService(pool)
            Register(protected, NewHandler(svc))
        },
    }
}
```

然后在 [cmd/xin/main.go](cmd/xin/main.go) 显式 import（无 `_`）并放进模块列表：

```go
import "gx1727.com/xin/apps/feedback"

modules := []plugin.Module{
    // ...
    feedback.Module(app),
    // ...
}
```

### 6.2 路由 + 中间件

```go
// apps/feedback/routes.go
import (
    "gx1727.com/xin/framework/pkg/middleware"
    "gx1727.com/xin/framework/pkg/permission"
)

func Register(protected *gin.RouterGroup, h *Handler) {
    g := protected.Group("/feedbacks")
    {
        g.GET("",    middleware.Require(permission.P(permission.ResFeedback, permission.ActList)),   h.List)
        g.POST("",   middleware.Require(permission.P(permission.ResFeedback, permission.ActCreate)), h.Create)
        g.GET("/:id",middleware.Require(permission.P(permission.ResFeedback, permission.ActList)),   h.Get)
        g.PUT("/:id",middleware.Require(permission.P(permission.ResFeedback, permission.ActUpdate)), h.Update)
        g.DELETE("/:id",middleware.Require(permission.P(permission.ResFeedback, permission.ActDelete)), h.Delete)
    }
}
```

### 6.3 Handler + resp

```go
// apps/feedback/handler.go
import "gx1727.com/xin/framework/pkg/resp"

func (h *Handler) Create(c *gin.Context) {
    var req struct {
        Title string `json:"title" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        resp.BadRequest(c, err.Error())
        return
    }
    id, err := h.svc.Create(c.Request.Context(), req.Title)
    if err != nil {
        resp.HandleError(c, err)  // 区分 BizError / 其他 error
        return
    }
    resp.Success(c, gin.H{"id": id})
}
```

### 6.4 Service 拿当前用户

```go
import xinContext "gx1727.com/xin/framework/pkg/context"

func (s *Service) List(ctx context.Context, page, size int) ([]Item, int64, error) {
    uc := xinContext.MustNewUserContext(ctx)  // panic if Auth 没挂载
    return s.repo.List(ctx, uc.TenantID, uc.UserID, page, size)
}
```

### 6.5 业务 SQL 套 DataScope

```go
import (
    "gx1727.com/xin/framework/pkg/context"
    "gx1727.com/xin/framework/pkg/db"
    "gx1727.com/xin/framework/pkg/permission"
)

func (s *Service) List(ctx context.Context, page, size int) ([]Item, int64, error) {
    uc := context.MustNewUserContext(ctx)
    filter, err := uc.GetDataScopeFilter()  // 用默认列:creator_id / org_id
    if err != nil { return nil, 0, err }

    // Phase 4+：s.pool 是 service 在 NewService(pool, ...) 时显式持有的，不再调 db.Get()
    return db.RunInTenantTx(ctx, s.pool, uc.TenantID, func(txCtx context.Context) ([]Item, int64, error) {
        q, _ := db.GetQuerier(txCtx, s.pool)
        // SELECT * FROM items WHERE <filter.SQL> ORDER BY id LIMIT $2 OFFSET $3
        // filter.Args 已经包含 userID / orgIDs 等
        // ...
    })
}
```

### 6.6 缓存失效

```go
func (s *Service) Update(ctx context.Context, roleID uint, req UpdateReq) error {
    if err := s.repo.Update(ctx, roleID, req); err != nil { return err }
    if s.authz != nil {
        _ = s.authz.InvalidateRole(context.Background(), roleID)  // 后台失效
    }
    return nil
}
```

### 6.7 RLS 事务

```go
// service 持有 pool；不再用 db.Get() 拿全局 pool
err := db.RunInTenantTx(ctx, s.pool, tenantID, func(txCtx context.Context) error {
    q, _ := db.GetQuerier(txCtx, s.pool)  // 自动拿 tx
    // txCtx 上自动 SET LOCAL app.tenant_id = tenantID
    rows, err := q.Query(txCtx, "SELECT * FROM users WHERE ...", ...)
    // ...
})
```

### 6.8 super_admin 守卫(跨租户操作)

```go
import (
    pkgmiddleware "gx1727.com/xin/framework/pkg/middleware"
)

tenants := protected.Group("/tenants")
tenants.Use(pkgmiddleware.RequirePlatformRole("super_admin"))
tenants.POST("", h.Create)
// ...
```

## 7. 不要做的事

❌ **不要**新增跨模块全局变量(Phase 0-8 重构就是为了消灭它们)
❌ **不要**在 `framework/` 里 import `apps/`(编译就过不去)
❌ **不要**绕过 `Require(spec)` 中间件直接调 handler
❌ **不要**用 `db.Get()` 在业务代码里 — 改用 `app.DB`（module 注入）或 `s.pool`（service 持有）
❌ **不要**写 `gorm` 或 `database/sql` — 项目统一用 `pgx`
❌ **不要**在 handler 里写 SQL — 走 service → repository
❌ **不要**新增 `var globalXxx` 包级变量 — 用 `AppContext`
❌ **不要**写 `gin.WrapH(legacyMux)` 之类的兼容层 — 旧 mux 早就删了
❌ **不要**写"先 work 再说"的全局 map — 早晚会并发出问题

## 8. 验证清单(提交前必跑)

```bash
cd server
go build ./...                              # 必须 EXIT=0
go vet ./...                                # 必须 EXIT=0
go test -count=1 ./framework/pkg/...        # 36 个 P0 单测必须全过
./scripts/xin_main_check.exe &              # 烟测
sleep 3
curl -s http://localhost:8087/api/v1/health # 必须 200
curl -s -X POST http://localhost:8087/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"account":"x","password":"x","tenant_code":"x"}'   # 必须返回 BizError(不能 panic)
pkill xin_main_check
```

## 9. 数据流总结

```
请求进来
  ↓
[Recovery / RequestID / CORS / ClientIP / Logger]      ← 全局中间件
  ↓
[Auth 或 OptionalAuth]                                  ← framework/internal/core/middleware
  ├─ 验 JWT 签名
  ├─ 验 Session (Redis/DB)
  ├─ 注入 XinContext{ UserID, TenantID, PlatformRoles }
  └─ 注册 UserContextLoader (懒)
  ↓
[RequirePlatformRole? Require?]                         ← framework/pkg/middleware
  ├─ IsSuperAdmin → 放行
  └─ HasPermission → 否则 403
  ↓
handler → service → repository
  ├─ MustNewUserContext(c)        → 触发 loader → DB 查权限 → 缓存
  ├─ RunInTenantTx(ctx, pool, tenantID, fn)   ← RLS 自动隔离
  ├─ BuildDataScopeFilter        ← SQL WHERE
  └─ authz.InvalidateRole(...)   ← 失效缓存
  ↓
resp.OK(c, data) / resp.HandleError(c, err)
```

## 10. 重构后的关键不变量

| 不变量 | 怎么保证 |
|---|---|
| 模块不能 import 另一模块的内部包 | `framework/internal/` 边界 |
| 跨模块依赖必须显式 | `AppContext.Reader/Writer` 接口 |
| 业务不能写全局变量 | 重构已删,新增会 grep 出来 |
| 软删除不可绕过 | 所有索引都是 partial `WHERE is_deleted = FALSE` |
| RLS 不能关 | DB 用 `FORCE ROW LEVEL SECURITY` |
| JWT secret 不能弱 | `validateJWTSecret` prod 校验 ≥32 字节 |

## 11. 调试技巧

| 现象 | 排查路径 |
|---|---|
| 启动 panic | 看 `cmd/xin/main.go` + `framework.Serve` 链路 |
| 路由 404 | `git grep "/xxx" apps/` 找路由定义 |
| 权限 403 | `middleware.Require` 的 spec 是否对应 `resources` 表里有 seed |
| 跨租户泄漏 | 业务 SQL 是否包在 `RunInTenantTx` 里 |
| 缓存不失效 | 是否调了 `authz.InvalidateRole/User` |
| Module 没启动 | `cfg.Module` 是否包含它,看启动日志 |
| 启动慢 | `db.Init` 连不上? `migrate.Run` 跑很多 SQL? |
| API 500 | gin log + `resp.HandleError` 日志 + 看 `reqID` trace 整个链路 |

## 12. 当前状态(2026)

| 维度 | 状态 |
|---|---|
| Go | 1.25+ |
| Go modules | 单 module `gx1727.com/xin`（Phase 1 合并） |
| 跨模块全局 | 1 个（`authz.Authorization` interface,无状态） |
| db.Get / config.Get / bootx | 已删（Phase 4-5） |
| main.go | 4 步显式 Build：`config.Load` → `framework.Boot` → 构造 `[]plugin.Module` → `framework.Serve` |
| 模块入口 | 全部 `Module(app *appx.App) plugin.Module`，main.go 显式注册 |
| 中间件 | 无 wrapper 重复,Require 全在 pkg/middleware |
| extapi | Provider 模式,facade 从 ctx 拿 repo |
| ext_impl/registry.go | 已删（189 行） |
| P0 单测 | 36 个,3 包覆盖率 48.4% |

## 13. 不要改的文件(除非明确要重构)

- [framework/internal/core/middleware/auth.go](framework/internal/core/middleware/auth.go) — 框架内部核心,改完必须跑完整 smoke test
- [framework/pkg/plugin/appcontext.go](framework/pkg/plugin/plugin.go) — 接口设计,新增方法必须考虑下游
- [framework/pkg/permission/scope.go](framework/pkg/permission/scope.go) — DataScope 是合规相关,改 SQL 生成逻辑要做权限审计
- [framework/pkg/middleware/auth.go](framework/pkg/middleware/auth.go) — 所有 RBAC 走这里,改完要看 36 个单测
- [migrations/framework.sql](migrations/framework.sql) — 已部署,不能改 CREATE,只能 ALTER TABLE 加

## 14. 相关文档

- [README.md](README.md) — 项目入口
- [doc/architecture.md](doc/architecture.md) — 架构
- [doc/quickstart.md](doc/quickstart.md) — 快速开始
- [doc/modules.md](doc/modules.md) — 模块清单
- [doc/api.md](doc/api.md) — HTTP API
- [doc/database.md](doc/database.md) — 数据库
- [doc/permissions.md](doc/permissions.md) — 权限
- [doc/developing.md](doc/developing.md) — 新增模块 8 步
- [doc/deployment.md](doc/deployment.md) — 部署
- [doc/refactor/plan.md](doc/refactor/plan.md) — 重构方案
- [doc/refactor/phase0/globals.md](doc/refactor/phase0/globals.md) — Phase 0 摸底数据
- [doc/refactor/phase5_static_analysis.md](doc/refactor/phase5_static_analysis.md) — Phase 5 静态分析