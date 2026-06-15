# server/AGENTS.md

> 给 AI agent / Codex 看的后端速查。读这一份，比重新读 100 个 .go 快得多。

---

## 1. 仓库速览

```
server/
├── cmd/xin/main.go             # 入口（side-effect import + framework.Run）
├── go.mod                      # 根 module: gx1727.com/xin
├── go.work                     # 多 module 编排（. + ./framework + ./apps）
├── config/                     # yaml 配置
├── migrations/                 # *.sql 迁移
│
├── framework/                  # 框架核心 module: gx1727.com/xin/framework
│   ├── pkg/                    #   公开 SDK
│   │   ├── auth/               #     pkgauth.Account/AccountAuth/HashPassword
│   │   ├── tenant/             #     pkgtenant.TenantRecord + 注册钩子
│   │   ├── module/             #     Module 接口 / AppContext / Manifest（Phase 3 启用）
│   │   ├── middleware/         #     Require / RequireAny / RequireAll / RequirePlatformRole
│   │   ├── permission/         #     P() / Spec / Constants / Scope
│   │   ├── plugin/             #     Module 接口 + Register / Apps
│   │   ├── resp/               #     OK / Fail / ErrXxx 预置错误
│   │   ├── session/            #     会话管理
│   │   ├── jwt/                #     JWT 签发 / 校验
│   │   ├── db/                 #     pgxpool + 事务
│   │   ├── cache/              #     内存缓存（dict 使用）
│   │   ├── migrate/            #     SQL 迁移执行器
│   │   ├── config/             #     yaml 加载 + Get()
│   │   ├── audit/              #     审计日志
│   │   ├── logger/             #     日志
│   │   ├── storage/            #     存储抽象（local / cos）
│   │   └── extapi/             #     跨模块 Provider 接口
│   ├── internal/
│   │   ├── core/
│   │   │   ├── boot/           #   Init(cfg) → *App{Server, SessionMgr, Authz}
│   │   │   ├── server/         #   Server 包装（Engine + Start/Stop）
│   │   │   ├── middleware/     #   框架内部中间件
│   │   │   └── ext_impl/       #   extapi.Provider 的 default 实现
│   │   ├── service/            #   框架级服务（Authz / Perm）
│   │   └── module/             #   仍在 framework 的业务模块
│   │       ├── asset/  dict/   menu/  organization/  permission/
│   │       ├── resource/  role/  system/  user/  weixin/
│   └── xin-server.service      # systemd unit
│
└── apps/                       # 业务模块 module: gx1727.com/xin/apps
    ├── boot/
    │   ├── auth/               # Phase 2 已迁入
    │   └── tenant/             # Phase 2 已迁入
    ├── rbac/                   # Phase 3 待迁入（user/role/menu/...）
    ├── reference/              # Phase 3 待迁入（dict/asset/weixin）
    ├── cms/
    └── flag/
```

---

## 2. 启动流程（必读）

```
cmd/xin/main.go
└── import (
    "gx1727.com/xin/framework"          # builtin_modules.go 侧效引入 10 个内部模块
    _ "gx1727.com/xin/apps/boot/auth"   # side-effect：init() 调用 plugin.Register
    _ "gx1727.com/xin/apps/boot/tenant"
    _ "gx1727.com/xin/apps/cms"
    _ "gx1727.com/xin/apps/flag"
  )
└── framework.Run(cfg)
    ├── boot.Init(cfg)                # 加载配置 / 初始化 pgxpool / SessionMgr / Authz
    ├── initModules(cfg)              # 遍历 plugin.Apps()，按 cfg.Module 过滤后 Init()
    ├── runMigrations()               # 执行 migrations/ 目录下的 SQL（按文件名排序）
    ├── setupRouter(app)              # 全局中间件 + 各模块 Register
    └── app.Server.Start(addr) + waitForSignal
```

**关键点**：

- 加新模块：只要在 `cmd/xin/main.go` 加一行 `_ "gx1727.com/xin/apps/<x>"`，并确保该模块包内有 `init() { plugin.Register(Module()) }`
- 删除模块：从 main.go 删除 import 行即可，模块本身不需要改
- 模块启停：在 `config/config.yaml` 的 `module:` 列表里增删名称

---

## 3. SDK 公开契约（直接 import 即可）

| 包 | 内容 | 谁该用 |
| --- | --- | --- |
| `framework/pkg/auth` | `Account`, `AccountAuth`, `AccountRepository`, `AccountAuthRepository`, `HashPassword`, `VerifyPassword` | apps 模块 + framework 内部 |
| `framework/pkg/tenant` | `TenantRecord`, `TenantRepository`, `Register(func)` | apps + framework/internal |
| `framework/pkg/permission` | `P(resource, action) Spec`, 常量（ResUser/ResRole/... ActList/...）, DataScope 类型 | 所有路由 handler |
| `framework/pkg/middleware` | `Require(spec)` / `RequireAny` / `RequireAll` / `RequirePlatformRole` | 路由注册 |
| `framework/pkg/resp` | `OK(c, data)`, `Fail(c, code, msg)`, 业务错误常量 | handler 返回值 |
| `framework/pkg/plugin` | `Module`, `Register(m)`, `Apps()` | main.go / 模块注册 |
| `framework/pkg/session` | 会话存储（内存 / Redis） | login handler |
| `framework/pkg/jwt` | `Issue(token, claims)`, `Verify(token)` | middleware/auth |
| `framework/pkg/db` | `Get()`, `GetQuerier(ctx)`, `WithTx(ctx, fn)` | repository |
| `framework/pkg/cache` | 进程内 LRU | dict / permission |
| `framework/pkg/migrate` | `Run("migrations")` | framework.Run |
| `framework/pkg/config` | `Get()`, `Load(path)` | 配置读取 |
| `framework/pkg/audit` | `WithContext(c, op, target)` | 业务关键操作 |
| `framework/pkg/storage` | `Storage` 接口 + `local` / `cos` 实现 | 文件上传 |

**注意**：`framework/pkg/module/` 是 Phase 3 才会落地的 DI SDK，目前只是接口声明，不要在生产代码里依赖。

---

## 4. 跨模块依赖规则

### 4.1 apps ↔ framework

| 方向 | 规则 | 例 |
| --- | --- | --- |
| apps → framework | ✅ 直接 import | `apps/boot/auth/handler.go` → `framework/pkg/...` |
| framework → apps | ❌ 直接 import 会被 `internal/` 拒绝 | —— |
| framework → apps（间接） | ✅ 通过 pkg 钩子 | framework 的 user 包通过 `pkgauth.Get()` 拿到 apps/boot/auth 的 AccountRepository |

### 4.2 同模块内

| 方向 | 规则 |
| --- | --- |
| apps/X → apps/Y | ✅ 同 module，可直接 import |
| apps/X → framework/internal | ❌ internal/ 规则 |
| framework/internal/X → framework/internal/Y | ✅ 同 module |
| framework/internal/X → framework/pkg/Y | ✅ |
| framework/pkg/X → 任何 internal | ❌ |

---

## 5. 新增一个 app（apps/Y/）

最小骨架（参考 [apps/cms](file:///d:\work\xin\XinFramework\server\apps\cms)）：

```
apps/Y/
├── go.mod         # 不需要独立 go.mod，apps 整体一个 module
├── module.go      # module struct + Module() + init() { plugin.Register(Module()) }
├── handler.go     # HTTP handlers
├── service.go     # 业务逻辑
├── repository.go  # DB 访问（pgx）
├── types.go       # DTO
├── errors.go      # 业务错误（实现 Error() + Is()）
├── routes.go      # Register(public, protected, h)
├── config.go      # 可选：本模块私有配置
└── doc/api.md     # 模块 API 文档
```

注册到 main：

```go
// cmd/xin/main.go
_ "gx1727.com/xin/apps/Y"
```

配置启用：

```yaml
# config/config.yaml
module:
  - Y
```

---

## 6. 写一个路由的标准模板

```go
// apps/Y/handler.go
func (h *Handler) Create(c *gin.Context) {
    var req createRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        resp.Fail(c, resp.ErrBadRequest, err.Error())
        return
    }

    item, err := h.svc.Create(c.Request.Context(), req)
    if err != nil {
        resp.Fail(c, codeFromErr(err), err.Error())
        return
    }

    // 审计（可选）
    audit.WithContext(c, "Y.create", item.ID)

    resp.OK(c, item)
}

// apps/Y/routes.go
func Register(protected *gin.RouterGroup, h *Handler) {
    items := protected.Group("/Y")
    items.POST("", middleware.Require(permission.P(permission.ResY, permission.ActCreate)), h.Create)
    items.GET("", middleware.Require(permission.P(permission.ResY, permission.ActList)), h.List)
    items.GET("/:id", middleware.Require(permission.P(permission.ResY, permission.ActList)), h.Get)
    items.PUT("/:id", middleware.Require(permission.P(permission.ResY, permission.ActUpdate)), h.Update)
    items.DELETE("/:id", middleware.Require(permission.P(permission.ResY, permission.ActDelete)), h.Delete)
}
```

---

## 7. 数据库约定

- 表名：复数 snake_case（`users`, `role_resources`, `account_auths`）
- 主键：`id BIGSERIAL`
- 软删：`is_deleted BOOLEAN DEFAULT FALSE` + `deleted_at TIMESTAMPTZ`
- 审计字段：`created_at`, `updated_at`, `created_by`, `updated_by`
- 租户字段：**多租户表必带** `tenant_id BIGINT NOT NULL`，所有查询强制过滤
- 索引：外键必加；`tenant_id + is_deleted + created_at DESC` 是常规列表索引
- 字符集：UTF-8；金额：`NUMERIC(20,4)`；枚举：`SMALLINT`（不要用 PostgreSQL 枚举类型）

详见 [doc/database.md](file:///d:\work\xin\XinFramework\server\doc\database.md)。

---

## 8. 权限模型

```
super_admin (平台角色)
  └─ 可访问所有租户
tenant_admin (租户内角色)
  └─ 通过 (resource, action) 资源权限码控制
  └─ 通过 data_scope (All / Custom / Dept / DeptAndBelow / Self) 控制数据范围
```

```go
// 路由级
items.GET("", middleware.Require(permission.P("user", "list")), h.List)

// 跨租户
items.POST("", middleware.RequirePlatformRole("super_admin"), h.Create)
```

详见 [doc/permissions.md](file:///d:\work\xin\XinFramework\server\doc\permissions.md)。

---

## 9. 响应格式

```json
{
  "code": 0,        // 0 = 成功；非 0 = 业务错误码
  "msg": "ok",      // 文案（前端展示）
  "data": { ... }   // 业务数据；列表分页包 list/total/page/size
}
```

错误码：
- 4xxx：客户端错误（参数、资源不存在、未授权）
- 5xxx：服务端错误（系统异常、数据库错误）
- 9xxx：业务错误（用户/角色不存在、状态冲突）

详见 [doc/api.md](file:///d:\work\xin\XinFramework\server\doc\api.md)。

---

## 10. 踩坑与决策

### 10.1 编码（最重要）

- **PowerShell 默认 GBK**，`Get-Content` / `Set-Content` 会把中文 mangle 成 `?`
- 可靠方案（参考 [server/AGENTS.md 历史版本](file:///d:\work\xin\XinFramework\server\.claude\settings.local.json)）：
  - 写中文文件用 `[System.IO.File]::WriteAllText($path, $content, [System.Text.UTF8Encoding]::new($false))`
  - 读二进制看编码：`[System.IO.File]::ReadAllBytes($p)`
- Vite/esbuild 期望 **UTF-8 无 BOM**

### 10.2 internal/ 规则

- `framework/internal/X` 只允许 `gx1727.com/xin/framework/...` 内引用
- apps 不能直接 import `framework/internal/...`
- 解决方案：把需要跨边界的类型提到 `framework/pkg/...`

### 10.3 多 Go module 切分

- 三个 module：root / framework / apps
- 靠 `go.work` + 各自的 `go.mod` 编排
- 加新 app 不要建独立 go.mod，直接在 apps/ 下加目录

### 10.4 注册钩子模式

跨 module 共享接口时（apps/boot/auth 的 AccountRepository 被 framework/internal/module/user 消费）：

1. 在 `framework/pkg/X/` 定义接口和注册函数 `Register(func() X)`
2. apps/X 的 `init()` 调用 `pkgX.Register(...)`
3. framework 内部用 `pkgX.Get()` 拿到工厂

示例：`framework/pkg/auth/registry.go` ↔ `apps/boot/auth/module.go`。

### 10.5 模块重复注册保护

`plugin.Register()` 对同名模块只注册一次。误重复 import 不会 panic，只是 silently skip。

### 10.6 数据库连接

- 默认走 `framework/pkg/db.Get()` 返回 `*pgxpool.Pool`
- 事务：`db.WithTx(ctx, func(tx) error { ... })`
- 跨请求共享 query：`db.GetQuerier(ctx)`（事务内返回 tx，否则返回 pool 的 conn）

### 10.7 错误传递

- 数据库错误：直接 `return err`（handler 里映射成 500）
- 业务错误：定义 `var ErrXxx = errors.New(...)`，handler 里用 `errors.Is(err, Xxx.ErrXxx)` 判别

### 10.8 验证

```bash
# 编译
go build ./...

# 静态检查
go vet ./...

# 跑（前台）
go run ./cmd/xin run

# 迁移（一般自动跑，手动可）
psql -U xin -d xin -f migrations/framework.sql
```

### 10.9 不要做的事

- ❌ 不要在 `framework/internal/module/` 加新业务模块（应该放进 `apps/`）
- ❌ 不要让 apps 直接 import `framework/internal/...`
- ❌ 不要在 handler 里直接调 `db.Get()`（应该走注入的 repository）
- ❌ 不要返回裸 `{code, msg, data}` 字面量（用 `resp.OK/Fail`）
- ❌ 不要把 `apps/cms` / `apps/flag` 写成独立 go.mod
- ❌ 不要修改 `cmd/xin/main.go` 的 modules map（已废弃）

---

## 11. 关键文件索引

| 关注点 | 路径 |
| --- | --- |
| 入口 | `server/cmd/xin/main.go` |
| 框架入口 | `server/framework/framework.go` |
| 内置模块汇总（side-effect） | `server/framework/builtin_modules.go` |
| 启动流程 | `server/framework/internal/core/boot/boot.go` |
| 全局中间件 | `server/framework/framework.go` 的 `setupRouter` |
| 业务模块注册 API | `server/framework/pkg/plugin/plugin.go` |
| 跨 module 钩子（auth） | `server/framework/pkg/auth/registry.go` ↔ `server/apps/boot/auth/module.go` |
| 跨 module 钩子（tenant） | `server/framework/pkg/tenant/registry.go` ↔ `server/apps/boot/tenant/module.go` |
| 平台角色中间件 | `server/framework/pkg/middleware/auth.go` |
| 数据库迁移 | `server/migrations/{framework,cms,flag}.sql` |
| 配置示例 | `server/config/config.{yaml,dev.yaml,prod.yaml}` |
| systemd 单元 | `server/framework/xin-server.service` |
| 构建脚本 | `server/build.sh`（Linux/macOS）/ `server/build.ps1`（Windows） |