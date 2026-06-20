# AGENTS.md — XinFramework AI Agent 高密度参考

> 给 AI agent 协作者的速查表。内容假设你已经读过 [README.md](README.md) 和 [doc/architecture.md](doc/architecture.md)。

## 1. 仓库结构

```
server/
├── cmd/xin/main.go              # 入口（4 步显式 Build）
├── config/                       # YAML 配置
├── migrations/                   # SQL 迁移（framework / asset / config / dict / flag / cms）
├── scripts/
│   └── strip_bom.py             # BOM 检测 / 剥离（--check 用于 CI）
├── framework/                   # 框架本体
│   ├── framework.go             # Boot() / Serve() 入口
│   ├── internal/                # 框架核心实现
│   │   ├── core/{boot,middleware,server,ext_impl}/
│   │   └── service/{authorization,permission}_service.go
│   └── pkg/                     # 公共包（含 appx/appx.go）
└── apps/                        # 业务模块（同 module）
    ├── boot/{auth,tenant}/      # 平台级 alwaysOn
    ├── rbac/{menu,organization,permission,resource,role,user}/
    ├── reference/{asset,config,dict,weixin}/
    ├── system/                  # health / cache 运维 alwaysOn
    ├── cms/, flag/              # 示例业务 optional
```

## 2. 依赖方向

```
cmd/xin ──→ framework ──→ apps
            (internal)     (internal ← 不可)
```

- `framework` 不能 import `apps`（`internal/` 强制）
- `apps` 不能 import `cmd/xin`
- `apps` 间跨模块通信**只能**通过 `plugin.AppContext`（Reader/Writer 接口）
- **单 Go module** `gx1727.com/xin`（Phase 1 已合并 framework + apps + cmd；不要回退到 multi-module / `go.work`）

修改前确认：你改的文件属于哪个目录？是否跨边界？

## 3. 15 个模块名（NameStr）

| Name | 类型 | 路径 | 备注 |
|---|---|---|---|
| `auth` | alwaysOn | apps/boot/auth | 登录 / JWT |
| `tenant` | alwaysOn | apps/boot/tenant | 租户管理（需 super_admin） |
| `system` | alwaysOn | apps/system | /health + cache 运维 |
| `user` | optOut | apps/rbac/user | |
| `role` | optOut | apps/rbac/role | |
| `menu` | optOut | apps/rbac/menu | |
| `organization` | optOut | apps/rbac/organization | |
| `permission` | optOut | apps/rbac/permission | 角色-资源分配 |
| `resource` | optOut | apps/rbac/resource | |
| `asset` | optOut | apps/reference/asset | |
| `dict` | optOut | apps/reference/dict | |
| `config` | optOut | apps/reference/config | 租户配置中心 |
| `weixin` | optional | apps/reference/weixin | 微信小程序登录 |
| `cms` | optional | apps/cms | 示例 CMS（extapi 模式） |
| `flag` | optional | apps/flag | 头像框 / 空间 / 头像 |

## 4. 关键路径速查

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
| 资源码常量 | [framework/pkg/permission/constants.go](framework/pkg/permission/constants.go) |
| extapi Provider | [framework/internal/core/ext_impl/provider.go](framework/internal/core/ext_impl/provider.go) |
| 审计日志 | [framework/pkg/audit/audit.go](framework/pkg/audit/audit.go) |
| 用户模块样板 | [apps/rbac/user/module.go](apps/rbac/user/module.go) |
| 字典模块样板（最小） | [apps/reference/dict/module.go](apps/reference/dict/module.go) |
| 配置模块样板（最新） | [apps/reference/config/module.go](apps/reference/config/module.go) |
| Flag 模块样板（业务） | [apps/flag/module.go](apps/flag/module.go) |

## 5. 跨模块通信：必须用 AppContext

❌ 错误做法：

```go
import "gx1727.com/xin/apps/rbac/user"
// 编译错误（framework 不能 import apps）
```

❌ 老做法（Phase 5 之前，已删）：

```go
import "gx1727.com/xin/framework/pkg/auth"
auth.SetAccountRepository(myRepo)  // 写全局
// 别处：auth.GetAccountRepository()  // 读全局
```

✅ 正确做法（Phase 5 之后，唯一允许）：

```go
// 1. 模块 Init 阶段：写 own slot
func (m *Module) InitFn(_ plugin.Reader, w plugin.Writer) error {
    w.SetAccountRepo(&myRepo{...})
    return nil
}

// 2. 其他模块 Register 阶段：读别人 slot
func OtherModule() plugin.Module {
    return &plugin.BaseModule{
        RegFn: func(ctx plugin.Reader, _, protected *gin.RouterGroup) {
            accountRepo := ctx.AccountRepo()
            if accountRepo == nil {
                // producer 模块被关闭，nil 友好
            }
            // ...
        },
    }
}
```

**新增跨模块接口的步骤**：

1. 在 `framework/pkg/{rbac,auth,tenant,...}/xxx.go` 定义窄 interface + struct
2. 在 [framework/pkg/plugin/appcontext.go](framework/pkg/plugin/appcontext.go) `Reader` / `Writer` 接口加方法
3. 在 `AppContext` struct 加字段 + getter/setter
4. provider module 在 `InitFn` 调 `w.SetXxx(myRepo)`

**编译会引导你走完整链路**——任何缺漏都会编译失败。

## 6. 命名约定

| 类别 | 约定 | 例子 |
|---|---|---|
| Module name | 小写，无下划线 | `"tenant"`, `"flag"` |
| Resource code | 小写，无下划线 | `"user"`, `"flag"`, `"config"` |
| Action | 小写动词 | `list` / `get` / `create` / `update` / `delete` / `tree` |
| Spec | `permission.P(code, action)` | `permission.P(permission.ResUser, permission.ActList)` |
| Error code | 每个 module 一个区段 | `1001-1999` for auth |
| Endpoint | `/<resource>` 复数 | `/users`, `/roles`, `/tenants` |
| ID param | `:id` | `/users/:id` |
| PATCH vs PUT | PATCH 部分，PUT 整体 | `PATCH /users/:id` vs `PUT /users/:id` |
| 软删除 | `is_deleted BOOLEAN DEFAULT FALSE` | 索引都加 `WHERE is_deleted = FALSE` |
| 时间戳 | `TIMESTAMPTZ DEFAULT NOW()` | `created_at` / `updated_at` |
| JSONB 写入 | SQL 显式 `::jsonb` cast | `NULLIF($6, '')::jsonb` |

## 7. 常见模式（直接 copy）

### 7.1 模块骨架

Phase 5 之后**统一形态**：`Module(app *appx.App) plugin.Module`，由 main.go 显式 import 并放进 `[]plugin.Module`：

```go
// apps/feedback/module.go
package feedback

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/appx"
    "gx1727.com/xin/framework/pkg/plugin"
)

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

### 7.2 路由 + 中间件

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

### 7.3 Handler + resp

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

### 7.4 Service 拿当前用户

```go
import xinContext "gx1727.com/xin/framework/pkg/context"

func (s *Service) List(ctx context.Context, page, size int) ([]Item, int64, error) {
    uc := xinContext.MustNewUserContext(ctx)  // panic if Auth 没挂载
    return s.repo.List(ctx, uc.TenantID, uc.UserID, page, size)
}
```

### 7.5 业务 SQL 套 DataScope

```go
import (
    "gx1727.com/xin/framework/pkg/context"
    "gx1727.com/xin/framework/pkg/db"
    "gx1727.com/xin/framework/pkg/permission"
)

func (s *Service) List(ctx context.Context, page, size int) ([]Item, int64, error) {
    uc := context.MustNewUserContext(ctx)
    filter, err := uc.GetDataScopeFilter()  // 默认列：creator_id / org_id
    if err != nil { return nil, 0, err }

    return db.RunInTenantTx(ctx, s.pool, uc.TenantID, func(txCtx context.Context) ([]Item, int64, error) {
        q, _ := db.GetQuerier(txCtx, s.pool)
        // SELECT * FROM items WHERE <filter.SQL> ORDER BY id LIMIT $2 OFFSET $3
        // ...
    })
}
```

### 7.6 缓存失效

```go
func (s *Service) Update(ctx context.Context, roleID uint, req UpdateReq) error {
    if err := s.repo.Update(ctx, roleID, req); err != nil { return err }
    if s.authz != nil {
        _ = s.authz.InvalidateRole(context.Background(), roleID)  // 后台失效
    }
    return nil
}
```

### 7.7 审计日志（在事务内）

```go
import "gx1727.com/xin/framework/pkg/audit"

// 在 db.RunInTenantTx 闭包内
audit.Log(ctx, s.pool, audit.Entry{
    Action:    "config_item:update",
    TableName: "config_items",
    RecordID:  int64(id),
    Old:       oldObj,
    New:       newObj,
})
// 注意：JSONB 列已经在 SQL 里 `::jsonb` cast，这里 Old/New 会被 json.Marshal 后传入
```

### 7.8 RLS 事务

```go
err := db.RunInTenantTx(ctx, s.pool, tenantID, func(txCtx context.Context) error {
    q, _ := db.GetQuerier(txCtx, s.pool)  // 自动拿 tx
    // txCtx 上自动 SET LOCAL app.tenant_id = tenantID
    rows, err := q.Query(txCtx, "SELECT * FROM users WHERE ...", ...)
    // ...
})
```

### 7.9 super_admin 守卫（跨租户操作）

```go
import (
    pkgmiddleware "gx1727.com/xin/framework/pkg/middleware"
)

tenants := protected.Group("/tenants")
tenants.Use(pkgmiddleware.RequirePlatformRole("super_admin"))
tenants.POST("", h.Create)
// ...
```

### 7.10 JSONB 写入（重要！）

```go
// ❌ 错：pgx 把 []byte 当 bytea、把 string 当 text 发，PG 报 42804
valueJSON, _ := json.Marshal(req.Value)
_, err := q.Exec(ctx, `UPDATE t SET value = $1 WHERE id = $2`, valueJSON, id)

// ✅ 对：SQL 显式 ::jsonb cast
valueJSON, _ := json.Marshal(req.Value)
_, err := q.Exec(ctx, `UPDATE t SET value = $1::jsonb WHERE id = $2`, valueJSON, id)

// ✅ 也对：COALESCE 场景
valueJSON := toJSON(req.Value)
_, err := q.Exec(ctx, `UPDATE t SET value = COALESCE($1::jsonb, value) WHERE id = $2`, valueJSON, id)
```

## 8. 不要做的事

❌ **不要**新增跨模块全局变量（Phase 0-8 重构就是为了消灭它们）
❌ **不要**在 `framework/` 里 import `apps/`（编译就过不去）
❌ **不要**绕过 `Require(spec)` 中间件直接调 handler
❌ **不要**用 `db.Get()` 在业务代码里——改用 `app.DB`（module 注入）或 `s.pool`（service 持有）
❌ **不要**写 `gorm` 或 `database/sql`——项目统一用 `pgx`
❌ **不要**在 handler 里写 SQL——走 service → repository
❌ **不要**新增 `var globalXxx` 包级变量——用 `AppContext`
❌ **不要**写 `gin.WrapH(legacyMux)` 之类的兼容层——旧 mux 早就删了
❌ **不要**写"先 work 再说"的全局 map——早晚会并发出问题
❌ **不要**用 `[]byte` / `string` 直接写 JSONB 列——SQL 必须 `::jsonb` cast
❌ **不要**让源文件带 UTF-8 BOM——用 `python scripts/strip_bom.py --check .` 验证

## 9. 验证清单（提交前必跑）

```bash
cd server
go build ./...                                # 必须 EXIT=0
go vet ./...                                  # 必须 EXIT=0
go test -count=1 ./framework/pkg/...          # 36 个 P0 单测必须全过
python scripts/strip_bom.py --check .         # 必须无 BOM
# 烟测
go run ./cmd/xin run &
sleep 3
curl -s http://localhost:8087/api/v1/health   # 必须 200
pkill xin
```

## 10. 数据流总结

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
  ├─ repo.SQL(JSONB 列 ::jsonb)  ← pgx 类型正确
  ├─ audit.Log(...)              ← 审计（事务内，失败不回滚业务）
  └─ authz.InvalidateRole(...)   ← 失效缓存
  ↓
resp.OK(c, data) / resp.HandleError(c, err)
```

## 11. 重构后的关键不变量

| 不变量 | 怎么保证 |
|---|---|
| 模块不能 import 另一模块的内部包 | `framework/internal/` 边界 |
| 跨模块依赖必须显式 | `AppContext.Reader/Writer` 接口 |
| 业务不能写全局变量 | 重构已删，新增会 grep 出来 |
| 软删除不可绕过 | 所有索引都是 partial `WHERE is_deleted = FALSE` |
| RLS 不能关 | DB 用 `FORCE ROW LEVEL SECURITY` |
| JWT secret 不能弱 | `validateJWTSecret` prod 校验 ≥32 字节 |
| JSONB 写入类型正确 | SQL 显式 `::jsonb` cast（pgx 默认发 text/bytea） |
| 源文件不带 BOM | `scripts/strip_bom.py --check` 在 CI / 提交前跑 |

## 12. 调试技巧

| 现象 | 排查路径 |
|---|---|
| 启动 panic | 看 `cmd/xin/main.go` + `framework.Serve` 链路 |
| 路由 404 | `git grep "/xxx" apps/` 找路由定义；确认 `cfg.Module` 包含该模块 |
| 权限 403 | `middleware.Require` 的 spec 是否对应 `resources` 表里有 seed |
| 跨租户泄漏 | 业务 SQL 是否包在 `RunInTenantTx` 里 |
| 缓存不失效 | 是否调了 `authz.InvalidateRole/User` |
| Module 没启动 | `cfg.Module` 是否包含它，看启动日志 `module xxx initialized` |
| 启动慢 | `db.Init` 连不上？`migrate.Run` 跑很多 SQL？ |
| API 500 | gin log + `resp.HandleError` 日志 + 看 `reqID` trace 整个链路 |
| `column X is of type jsonb but expression is of type text/bytea` | SQL 缺 `::jsonb` cast（pgx 把 `string`/`[]byte` 当 text/bytea 发） |
| `invalid BOM in the middle of the file (1:4)` | 源文件双 BOM；跑 `python scripts/strip_bom.py` |
| `undefined: modules` | 忘了把 `modules` 参数往下传（参考 `framework.go:setupRouter`） |

## 13. 当前状态（2026）

| 维度 | 状态 |
|---|---|
| Go modules | 单 module `gx1727.com/xin`（Phase 1 合并） |
| 跨模块全局 | 1 个（`authz.Authorization` interface，无状态） |
| `db.Get` / `config.Get` / `bootx` | 已删（Phase 4-5） |
| main.go | 4 步显式 Build |
| 模块入口 | 全部 `Module(app *appx.App) plugin.Module` |
| 模块数 | 15（3 alwaysOn + 9 optOut + 3 optional） |
| 中间件 | 无 wrapper 重复；Require 全在 `pkg/middleware` |
| extapi | Provider 模式；facade 从 ctx 拿 repo |
| ext_impl/registry.go | 已删（189 行） |
| JSONB 列 | 9 列（audit 2 + config 4 + dict 2 + flag 1），全部 `::jsonb` cast |
| P0 单测 | 36 个，3 包覆盖率 48.4% |

## 14. 不要改的文件（除非明确要重构）

- [framework/internal/core/middleware/auth.go](framework/internal/core/middleware/auth.go) — 框架内部核心，改完必须跑完整 smoke test
- [framework/pkg/plugin/appcontext.go](framework/pkg/plugin/appcontext.go) — 接口设计，新增方法必须考虑下游
- [framework/pkg/permission/scope.go](framework/pkg/permission/scope.go) — DataScope 是合规相关，改 SQL 生成逻辑要做权限审计
- [framework/pkg/middleware/auth.go](framework/pkg/middleware/auth.go) — 所有 RBAC 走这里，改完要看 36 个单测
- [migrations/framework.sql](migrations/framework.sql) — 已部署，不能改 CREATE，只能 ALTER TABLE 加

## 15. 相关文档

- [README.md](README.md) — 项目入口
- [doc/quickstart.md](doc/quickstart.md) — 快速开始
- [doc/architecture.md](doc/architecture.md) — 架构
- [doc/modules.md](doc/modules.md) — 15 个 module 清单
- [doc/api.md](doc/api.md) — HTTP API
- [doc/database.md](doc/database.md) — 数据库
- [doc/permissions.md](doc/permissions.md) — 权限
- [doc/developing.md](doc/developing.md) — 新增模块 8 步
- [doc/deployment.md](doc/deployment.md) — 部署
