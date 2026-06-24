# XinFramework Server

> Go 1.24 + Gin + pgx + PostgreSQL。多租户 SaaS 后端，内置 RBAC、权限中间件、模块化插件框架。
> Phase 0023 全阶段完成（平台域 / 租户域物理分域）。

## 启动流程：4 步显式 Build

入口 [`cmd/xin/main.go`](cmd/xin/main.go)：

```go
cfg, err := config.Load("config/config.yaml")          // 1. 配置
app, rt, err := framework.Boot(cfg)                    // 2. 装配
modules := []plugin.Module{                            // 3. 显式构造模块列表
    auth.Module(app), tenants.Module(app), system.Module(app),
    sysuser.Module(app), sysrole.Module(app),
    sysmenu.Module(app), syspermission.Module(app),
    menu.Module(app), organization.Module(app), permission.Module(app),
    resource.Module(app), role.Module(app), user.Module(app),
    asset.Module(app), refconfig.Module(app), dict.Module(app), weixin.Module(app),
    cms.Module(app), flag.Module(app),
}
framework.Serve(cfg, app, rt, modules)                 // 4. 启动
```

[`framework.Serve`](framework/framework.go) 内部：

```
migrate.Run(app.DB, "migrations")      → SQL 迁移（幂等）
for _, m := range modules:              → Init 阶段
    m.Init(ctx, w)                          模块写 own slot
setupRouter(app, modules)                → 中间件链 + 路由
go rt.Server.Start(addr)                → 后台监听
waitForSignal(rt, app, modules)          → SIGINT/SIGTERM 优雅退出
```

[`framework.Boot`](framework/internal/core/boot/boot.go)（即 `boot.Init`）的装配：

```go
pool, _      := db.Init(ctx, &cfg.Database)         → pgxpool
dict.Init(pool)
cache.Init(&cfg.Redis)                              → go-redis
sm           := session.NewRedisSessionManager()    → Session
appCtx       := plugin.NewAppContext(pool, cache, cfg, sm)
permService  := service.NewPermissionService(...)   → RBAC
authzService := service.NewAuthorizationService(permService)
appCtx.SetAuthz(authzService)
srv          := server.New(cfg)
app          := &appx.App{Config: cfg, DB: pool}
return app, srv, appCtx, nil
```

## 文档地图

| 文档 | 用途 |
|---|---|
| [doc/quickstart.md](doc/quickstart.md) | 装 PG、首次跑 `xin run` |
| [doc/architecture.md](doc/architecture.md) | AppContext / 模块生命周期 / 3 类模块 / 错误码分段 |
| [doc/modules.md](doc/modules.md) | 19 个 module 清单、路由、数据表 |
| [doc/api.md](doc/api.md) | HTTP 端点参考（业务 + 平台 + 公开 三空间） |
| [doc/database.md](doc/database.md) | 表结构、RLS 行级安全、JSONB、迁移 |
| [doc/permissions.md](doc/permissions.md) | RBAC + 数据范围（5 种）+ 平台角色 + Spec 中间件 |
| [doc/developing.md](doc/developing.md) | 新增模块的 8 步流程（业务模块 + 平台模块两个模板） |
| [doc/deployment.md](doc/deployment.md) | 编译脚本、systemd、Docker、环境变量 |
| [AGENTS.md](AGENTS.md) | 给 AI agent 协作者的高密度参考 |

## 核心特性

| 特性 | 实现 |
|---|---|
| **多租户** | `tenants` + `accounts` / `users` 双层模型；`FORCE ROW LEVEL SECURITY` + `db.RunInTenantTx` |
| **平台管理** | `/api/v1/platform/*` 域，统一 `RequirePlatformRole("super_admin")` 守卫 |
| **RBAC** | `role → user_role → user`，`permission → role_permission`；支持 `user:*` / `*:*` 通配 |
| **数据范围** | 5 种 `DataScopeType`（All / Custom / Dept / DeptAndBelow / Self）编译期生成 SQL |
| **平台角色** | 跨租户特权 `super_admin`，自动 bypass 所有 spec |
| **JWT + Session** | HS256，token 内含 SessionID；登出即 revoke |
| **可插拔模块** | main.go 显式 `Module(app)`，`AppContext.Reader/Writer` 注入，无全局依赖 |
| **配置中心** | `config` 模块支持 Platform / Override / Visibility / Resolve 四层模型 |
| **文件存储** | local 默认，可切换 COS（腾讯云） |
| **缓存** | Redis 可选；不可用时 graceful degradation 到 DB session |
| **CORS / 审计 / 日志** | 中间件，可热插拔 |
| **JSONB 写入安全** | 所有 `[]byte` / `string` 写入 JSONB 列时 SQL 显式 `::jsonb` cast |

## 关键约定

1. **统一响应**：`{code, msg, data}`，业务码按段管理（[resp/errors.go](framework/pkg/resp/errors.go)）
2. **认证中间件**：`Auth`（必须登录）/ `OptionalAuth`（可选）/ `AuthLite`（只注入身份）
3. **权限中间件**：`Require(spec)` / `RequireAny(specs)` / `RequireAll(specs)` / `RequirePlatformRole(roles)`
4. **平台角色守卫**：平台模块在 `adminGroup := protected.Group("/admin", RequirePlatformRole("super_admin"))` 分组级守卫，叠加资源码细粒度
5. **错误**：业务错误用 `resp.Err(code, msg)`；系统错误用 `fmt.Errorf` 包上下文，最终 `HandleError` 兜底
6. **租户上下文**：从 JWT claims 读 `TenantID`，`db.RunInTenantTx(ctx, pool, tenantID, fn)` 自动设 RLS；平台域用 `db.RunInPlatformTx` 跳过 RLS
7. **JSONB 写入**：用 `[]byte`/`string` 写 JSONB 列时，SQL 必须 `::jsonb` cast（pgx 默认发 text/bytea 会报 42804）

## 路由空间

```
/api/v1/
   ├── /<resource>                  → 业务（protected + Require ResX）
   ├── /platform/<platform_resource> → 平台（protected + RequirePlatformRole(super_admin) + Require ResX）
   └── /public/<resource>           → 公开（OptionalAuth）
```

| 空间 | 中间件栈 | 示例 |
|---|---|---|
| 业务 | `Auth` + `Require(ResX)` | `POST /users`、`GET /configs`、`PUT /dicts/:id` |
| 平台 | `Auth` + `RequirePlatformRole("super_admin")` + `Require(ResX)` | `POST /platform/tenants`、`PUT /platform/sys-menus/:id` |
| 公开 | `OptionalAuth`（可不登录） | `GET /public/configs`、`POST /auth/tenant-login` |

完整中间件顺序（[framework.go `setupRouter`](framework/framework.go)）：

```
Recovery → RequestID → CORS → ClientIP → Logger
  → (v1 group)
public    → OptionalAuth
tenant    → Auth → RequireTenantContext
protected → Auth
  → RequirePlatformRole? Require(spec)?
  → handler
```

## 模块（19 个）

| 类型 | 模块 |
|---|---|
| **alwaysOn**（强制启用） | `auth`, `tenants`（platform_tenant）, `system` |
| **optOut**（默认启，白名单过滤） | `menu`, `user`, `role`, `resource`, `permission`, `organization`, `asset`, `dict` |
| **optional**（默认关） | `config`, `sys_user`, `sys_role`, `sys_menu`, `sys_permission`, `weixin`, `cms`, `flag` |

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
| Go | 是 | 1.24+ |
| PostgreSQL | 是 | 14+（需要 `ltree`、`pg_trgm` 扩展） |
| Redis | 否 | 7+（`enabled: true` 时启用，不可用时自动降级到 DB session） |

详细安装步骤见 [doc/quickstart.md](doc/quickstart.md)。

## 仓库结构

```
server/
├── cmd/xin/main.go              # 入口（4 步显式 Build）
├── config/                       # YAML 配置（config.yaml + .env）
├── .env                          # 首次启动 bootstrap token
├── migrations/                   # SQL 迁移（init_schema / init_seed / asset / cms / flag）
├── scripts/
│   └── strip_bom.py             # BOM 检测/剥离（含 --check CI gate）
├── framework/                    # 框架本体
│   ├── framework.go             # Boot() / Serve() 入口
│   ├── internal/
│   │   ├── core/{boot,middleware,server,ext_impl}/
│   │   └── service/{authorization,permission}_service.go
│   └── pkg/                     # appx / audit / auth / authz / cache /
│                                # config / context / db / dict / extapi /
│                                # jwt / logger / middleware / migrate /
│                                # model / permission / plugin / rbac /
│                                # resp / session / storage / tenant
└── apps/                         # 业务模块（同 module）
    ├── boot/auth/               # 登录（alwaysOn）
    ├── platform/{tenants,sys_user,sys_role,sys_menu,sys_permission}/
    ├── tenant/{menu,organization,permission,resource,role,user}/
    ├── reference/{asset,config,dict,weixin}/
    ├── system/                  # health / cache 运维 alwaysOn
    ├── cms/                     # 示例 CMS（extapi 模式）
    └── flag/                    # 头像框 / 空间 / 头像
```

## 当前状态（2026）

| 维度 | 状态 |
|---|---|
| Go modules | **单 module** `gx1727.com/xin` |
| 模块数 | **19**（3 alwaysOn + 8 optOut + 8 optional） |
| 路由空间 | 3 个（业务 `/` + 平台 `/platform/*` + 公开 `/public/*`） |
| 跨模块全局 | 1 个（`authz.Authorization` interface，无状态） |
| 模块入口 | 全部 `Module(app *appx.App) plugin.Module`，main.go 显式注册 |
| 中间件 | 无 wrapper 重复；Require 全在 `pkg/middleware` |
| extapi | Provider 模式；facade 从 ctx 拿 repo |
| JSONB 列 | 11 列，全部 `::jsonb` cast |
| 错误码段 | 15 段已用 |
| P0 单测 | 36+，覆盖 permission / middleware / plugin 三包 |

## Phase 历程（精简）

| Phase | 内容 |
|---|---|
| 0 | 摸底：找到 16 个跨模块全局、109 处引用 |
| 1-2 | 建 module / AppContext 骨架 |
| 3-4c | 删全局变量（authz/registry/ext_impl/middleware wrapper） |
| 5 | 新 module + main.go 4 步显式 Build |
| 001x | cms/flag 等示例业务补齐 |
| 0020 | platform_tenant 从 `apps/boot/tenant` 迁到 `apps/platform/tenants` |
| 0021 | 新增 sys_menu 模块 |
| **0022** | config 完全重构 + 三域路由（业务/平台/公开） |
| **0023** | 平台/租户域物理拆分：9 张表 rename、`account_roles` drop、Go 包 rename（`apps/rbac→apps/tenant`）、新增 `sys_user/sys_role/sys_permission` |

## 首次启动

通过 `.env` 注入 bootstrap 凭据：

```env
XIN_BOOTSTRAP_TOKEN=your-32+char-secret
XIN_BOOTSTRAP_ACCOUNT=root
XIN_BOOTSTRAP_PASSWORD=change-me-please
XIN_BOOTSTRAP_REAL_NAME=System Root
XIN_BOOTSTRAP_ROLE=super_admin
XIN_BOOTSTRAP_TENANT_CODE=default
```

首次启动会自动跑迁移+bootstrap（见 [doc/quickstart.md](doc/quickstart.md)）。

## 贡献

提交前必跑：

```bash
go build ./...                                # 必须 EXIT=0
go vet ./...                                  # 必须 EXIT=0
go test ./...                                 # 必须全 PASS
python scripts/strip_bom.py --check .         # 必须无 BOM
```
