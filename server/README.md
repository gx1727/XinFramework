# XinFramework Server

> Go 1.25 + Gin + pgx + PostgreSQL。多租户 SaaS 后端，内建 RBAC、权限中间件、模块化插件框架。
>
> 文档版本：2026-06（config 重构 + platform_menu/platform_tenant 后）

## 启动流程（4 步显式 Build）

入口 [`cmd/xin/main.go`](cmd/xin/main.go)：

```go
cfg, err := config.Load("config/config.yaml")
app, err := framework.Boot(cfg)                      // 1. 装配 *appx.App

modules := []plugin.Module{                          // 2. 显式构造模块列表（16 个）
    // alwaysOn（强制启用，无法关闭）
    auth.Module(app), platformtenant.Module(app), system.Module(app),
    // optOut（默认启用，module 写白名单时过滤）
    menu.Module(app), user.Module(app), role.Module(app),
    organization.Module(app), permission.Module(app), resource.Module(app),
    asset.Module(app), dict.Module(app),
    // optional（默认关，需在 cfg.Module 显式列出）
    refconfig.Module(app), platformmenu.Module(app),
    weixin.Module(app), cms.Module(app), flag.Module(app),
}
framework.Serve(cfg, app, modules)                   // 3. 启动
// 4. SIGINT/SIGTERM → 优雅退出
```

[`framework.Serve`](framework/framework.go) 内部：

```
migrate.Run(app.DB, "migrations")      ← SQL 迁移（幂等）
for _, m := range modules:              ← Init 阶段
    m.Init(ctx, w)                          模块写 own slot
setupRouter(app, modules)                ← 中间件链 + 路由
go app.Server.Start(addr)               ← 后台监听
waitForSignal(app.Server, app)          ← 优雅退出
```

[`framework.Boot`](framework/internal/core/boot/boot.go)（即 `boot.Init`）的 6 步装配：

```go
pool, _      := db.Init(ctx, &cfg.Database)         // ① pgxpool
dict.Init(pool)
cache.Init(&cfg.Redis)                              // ② go-redis
sm            := session.NewRedisSessionManager()   // ③ Session
permCache     := permission.NewRedisPermissionCache()
appCtx        := plugin.NewAppContext(...)          // ④ 依赖容器
ext_impl.InitExtApi(appCtx)
permService   := service.NewPermissionService(...)  // ⑤ RBAC
appCtx.SetAuthz(authz.Wrap(authService))            // ⑥ 跨模块共享
return &appx.App{Config, DB, SessionMgr, Server, PermService, Authz, AppContext}, nil
```

## 文档地图

| 文档 | 用途 |
|---|---|
| [doc/quickstart.md](doc/quickstart.md) | 装 PG、首次跑通 `xin run` |
| [doc/architecture.md](doc/architecture.md) | AppContext / 模块生命周期 / 3 类模块 / 错误码分段 / Phase 历程 |
| [doc/modules.md](doc/modules.md) | 16 个 module 清单、职责、路由、数据表 |
| [doc/api.md](doc/api.md) | HTTP 端点参考（业务 + 平台 + 公开 三空间） |
| [doc/database.md](doc/database.md) | 表结构、RLS 行级安全、JSONB、迁移（含 config_alignment.sql） |
| [doc/permissions.md](doc/permissions.md) | RBAC + 数据范围（5 种）+ 平台角色 + Spec 中间件 |
| [doc/developing.md](doc/developing.md) | 新增模块的 8 步流程（业务模块 + 平台模块两个模板） |
| [doc/deployment.md](doc/deployment.md) | 编译脚本、systemd、Docker、环境变量 |
| [AGENTS.md](AGENTS.md) | 给 AI agent 协作者的高密度参考 |

## 核心特性

| 特性 | 实现 |
|---|---|
| **多租户** | `tenants` + `accounts` / `users` 双层模型；`FORCE ROW LEVEL SECURITY` + `db.RunInTenantTx` |
| **平台管理** | `/api/v1/admin/*` 域，统一 `RequirePlatformRole("super_admin")` 守卫 |
| **RBAC** | `role → user_role → user`，`resource → role_resource`；支持 `user:*` / `*:*` 通配 |
| **数据范围** | 5 种 `DataScopeType`（All / Custom / Dept / DeptAndBelow / Self）编译期生成 SQL |
| **平台角色** | 跨租户特权 `super_admin`，自动 bypass 所有 spec |
| **JWT + Session** | HS256，token 内含 SessionID；登出即 revoke |
| **可插拔模块** | `main.go` 显式 `Module(app)`，`AppContext.Reader/Writer` 注入，无全局依赖 |
| **配置中心** | `config` 模块支持 Platform / Override / Visibility / Resolve 四层（Phase 0022） |
| **资源/对象存储** | local 默认，可切 COS（腾讯云） |
| **缓存** | Redis 可选；不可用时 graceful degradation 到 DB session |
| **CORS / 审计 / 日志** | 中间件，可热插拔 |
| **JSONB 写入安全** | 所有 `[]byte` / `string` 写入 JSONB 列时 SQL 显式 `::jsonb` cast |

## 关键约定

1. **统一响应**：`{code, msg, data}`，业务码按段管理（[resp/errors.go](framework/pkg/resp/errors.go)）
2. **认证中间件**：`Auth`（必须登录） / `OptionalAuth`（可选） / `AuthLite`（只注入身份）
3. **权限中间件**：`Require(spec)` / `RequireAny(specs)` / `RequireAll(specs)` / `RequirePlatformRole(roles)`
4. **平台角色守卫**：平台模块在 `adminGroup := protected.Group("/admin", RequirePlatformRole("super_admin"))` 分组级守卫，叠加资源码细分
5. **错误**：业务错误用 `resp.Err(code, msg)`；系统错误 `fmt.Errorf` 包上下文，最终 `HandleError` 兜底
6. **租户上下文**：从 JWT claims 取 `TenantID`，`db.RunInTenantTx(ctx, pool, tenantID, fn)` 自动套 RLS；平台域用 `db.RunInPlatformTx` 跳过 RLS
7. **JSONB 写入**：用 `[]byte`/`string` 写 JSONB 列时，SQL 必须 `::jsonb` cast（pgx 默认发 text/bytea 会报 42804）

## 路由空间

```
/api/v1/
   ├── /<resource>                  ← 业务（protected + Require ResX）
   ├── /admin/<platform_resource>   ← 平台（protected + RequirePlatformRole(super_admin) + Require ResX）
   └── /public/<resource>           ← 公开（OptionalAuth）
```

| 空间 | 中间件栈 | 示例 |
|---|---|---|
| 业务 | `Auth` + `Require(ResX)` | `POST /users`、`GET /configs`、`PUT /dicts/:id` |
| 平台 | `Auth` + `RequirePlatformRole("super_admin")` + `Require(ResX)` | `POST /admin/platform-tenants`、`PUT /admin/platform-menus/:id` |
| 公开 | `OptionalAuth`（可不登录） | `GET /public/configs`、`POST /auth/login` |

完整中间件顺序（[framework.go `setupRouter`](framework/framework.go)）：

```
Recovery → RequestID → CORS → ClientIP → Logger
  ↓ (v1 group)
public  → OptionalAuth
protected → Auth
  ↓
RequirePlatformRole? Require(spec)?
  ↓
handler
```

## 模块（16 个）

| 类型 | 模块 |
|---|---|
| **alwaysOn**（强制启） | `auth`, `platform_tenant`, `system` |
| **optOut**（默认启，白名单过滤） | `menu`, `user`, `role`, `resource`, `permission`, `organization`, `asset`, `dict` |
| **optional**（默认关） | `config`, `platform_menu`, `weixin`, `cms`, `flag` |

详见 [doc/modules.md](doc/modules.md)。

## 命令行

```bash
xin start          # 守护进程启动
xin stop           # 停止
xin restart        # 重启
xin reload         # 平滑重载（目前等价 restart）
xin run            # 前台运行（开发用）
xin status         # 查看 PID 文件状态
xin hot-restart    # 不中断服务的热重启
xin help           # 帮助
```

## 构建

```bash
./build.sh                                    # Linux / macOS → bin/xin
.\build.ps1                                   # Windows → bin/xin.exe

# 手编（交叉编译）
go build -ldflags="-s -w" -o bin/xin ./cmd/xin
```

## 平台支持

| OS | 部署 | 入口 |
|---|---|---|
| Linux | systemd | [framework/xin-server.service](framework/xin-server.service) |
| Windows | 直接运行 | [build.ps1](build.ps1) |
| macOS / Linux | tarball | [build.sh](build.sh) |

## 依赖环境

| 组件 | 必需 | 版本 |
|---|---|---|
| Go | ✅ | 1.25+ |
| PostgreSQL | ✅ | 14+（需要 `ltree`、`pg_trgm` 扩展） |
| Redis | ❌ | 7+（`enabled: true` 时启用，不可用时自动降级到 DB session） |

详细安装步骤见 [doc/quickstart.md](doc/quickstart.md)。

## 仓库结构

```
server/
├── cmd/xin/main.go              # 入口（4 步显式 Build）
├── config/                       # YAML 配置（config.yaml + 子模块 yaml）
├── migrations/                   # SQL 迁移
│   ├── framework.sql              # 核心表（tenants / accounts / rbac / ...）
│   ├── asset.sql                  # 业务表
│   ├── config.sql                 # 配置中心
│   ├── config_alignment.sql       # Phase 0022：scope/visibility/override
│   ├── dict.sql
│   ├── flag.sql
│   ├── weixin.sql
│   └── cms.sql
├── scripts/
│   └── strip_bom.py              # BOM 检测 / 剥离（含 --check CI gate）
├── framework/                    # 框架本体
│   ├── framework.go              # Boot() / Serve() 入口
│   ├── internal/
│   │   ├── core/                 # boot / middleware / server / ext_impl
│   │   └── service/              # authorization / permission
│   └── pkg/                      # appx / audit / auth / authz / cache /
│                                  # config / context / db / dict / extapi /
│                                  # jwt / logger / middleware / migrate /
│                                  # model / permission / plugin / rbac /
│                                  # resp / session / storage / tenant
└── apps/                         # 业务模块（同 module）
    ├── boot/                     # auth（alwaysOn）
    ├── admin/                    # 平台管理域
    │   ├── platform_menu/          # optional
    │   └── platform_tenant/        # alwaysOn（替代旧 boot/tenant）
    ├── rbac/                     # menu, organization, permission,
    │                             # resource, role, user
    ├── reference/                # asset, config, dict, weixin
    ├── system/                   # health / cache 运维
    ├── cms/                      # 示例 CMS（extapi 模式）
    └── flag/                     # 头像框 / 空间 / 头像
```

## 当前状态（2026-06）

| 维度 | 状态 |
|---|---|
| Go modules | **单 module** `gx1727.com/xin`（Phase 1 合并 framework + apps + cmd） |
| 模块数 | **16**（3 alwaysOn + 8 optOut + 5 optional） |
| 路由空间 | 3 个（业务 `/` + 平台 `/admin/*` + 公开 `/public/*`） |
| 跨模块全局 | 1 个（`authz.Authorization` interface，无状态） |
| `db.Get` / `config.Get` / `bootx` | 已删（Phase 4-5） |
| main.go | 4 步显式 Build |
| 模块入口 | 全部 `Module(app *appx.App) plugin.Module`，main.go 显式注册 |
| 中间件 | 无 wrapper 重复；Require 全在 `pkg/middleware` |
| extapi | Provider 模式；facade 从 ctx 拿 repo |
| JSONB 列 | 10 列（含 Phase 0022 新加），全部 `::jsonb` cast |
| 错误码段 | 14 段已用（auth/user/tenant/role/menu/org/permission/resource/asset/dict/system/weixin/flag/platform_menu/config） |
| P0 单测 | 36 个，覆盖 permission / middleware / plugin 三包 |

## Phase 历程（精简）

| Phase | 内容 |
|---|---|
| 0 | 摸底：找到 16 个跨模块全局，409 处引用 |
| 1-2 | 拆 module / AppContext 骨架 |
| 3-4c | 删全局变量（authz/registry/ext_impl/middleware wrapper） |
| 5 | 单 module + main.go 4 步显式 Build |
| 001x | cms/flag/cms 等示例业务补全 |
| 0020 | platform_tenant 从 `apps/boot/tenant` 迁到 `apps/admin/platform_tenant` |
| 0021 | 新增 platform_menu 模块 |
| **0022** | **config 完全重构**（路由 `/config/*` → `/t/configs/*`，Scope/Visibility/Override/Resolve，错误码段迁移到 18xxx） |

## 贡献

提交前必跑：

```bash
go build ./...                                # 必须 EXIT=0
go vet ./...                                  # 必须 EXIT=0
go test ./...                                 # 必须全 PASS
python scripts/strip_bom.py --check .         # 必须无 BOM
```