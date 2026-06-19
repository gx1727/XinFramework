# XinFramework Server

> Go 1.25 + Gin + pgx + PostgreSQL。多租户 SaaS 后端，内建 RBAC、权限中间件、模块化插件框架。

## 启动流程（4 步显式 Build）

入口 [`cmd/xin/main.go`](cmd/xin/main.go)：

```go
cfg, _ := config.Load("config/config.yaml")
app, _ := framework.Boot(cfg)                          // 1. 装配 *appx.App
modules := []plugin.Module{                            // 2. 显式构造模块列表
    auth.Module(app), tenant.Module(app),
    user.Module(app), role.Module(app),
    // ... 共 15 个
}
framework.Serve(cfg, app, modules)                     // 3. 启动
// 4. SIGINT/SIGTERM → 优雅退出
```

[`framework.Serve`](framework/framework.go) 内部：

```
migrate.Run(app.DB, "migrations")      ← SQL 迁移（幂等）
for _, m := range modules:              ← Init 阶段
    m.Init(ctx, w)
setupRouter(app)                        ← 中间件链 + 路由
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
| [doc/architecture.md](doc/architecture.md) | AppContext / 模块生命周期 / Phase 重构背景 |
| [doc/modules.md](doc/modules.md) | 15 个 module 清单、职责、路由、数据表 |
| [doc/api.md](doc/api.md) | HTTP 端点参考（认证、用户、租户、RBAC、配置、字典、flag） |
| [doc/database.md](doc/database.md) | 表结构、RLS 行级安全、JSONB、迁移说明 |
| [doc/permissions.md](doc/permissions.md) | RBAC + 数据范围（5 种）+ 平台角色 + Spec 中间件 |
| [doc/developing.md](doc/developing.md) | 新增模块的标准 8 步流程 |
| [doc/deployment.md](doc/deployment.md) | 编译脚本、systemd、Docker、环境变量 |
| [AGENTS.md](AGENTS.md) | 给 AI agent 协作者的高密度参考 |

## 核心特性

| 特性 | 实现 |
|---|---|
| **多租户** | `tenants` + `accounts` / `users` 双层模型；`FORCE ROW LEVEL SECURITY` |
| **RBAC** | `role → user_role → user`，`resource → role_resource`；支持 `user:*` / `*:*` 通配 |
| **数据范围** | 5 种 `DataScopeType`（All / Custom / Dept / DeptAndBelow / Self）编译期生成 SQL |
| **平台角色** | 跨租户特权 `super_admin`，自动 bypass 所有 spec |
| **JWT + Session** | HS256，token 内含 SessionID；登出即 revoke |
| **可插拔模块** | `main.go` 显式 `Module(app)`，`AppContext` 注入，无全局依赖 |
| **资源/对象存储** | local 默认，可切 COS（腾讯云） |
| **缓存** | Redis 可选；不可用时 graceful degradation 到 DB session |
| **CORS / 审计 / 日志** | 中间件，可热插拔 |
| **JSONB 写入安全** | 所有 `[]byte` / `string` 写入 JSONB 列时 SQL 显式 `::jsonb` cast |

## 关键约定

1. **统一响应**：`{code, msg, data}`，业务码按段管理（[resp/errors.go](framework/pkg/resp/errors.go)）
2. **认证中间件**：`Auth`（必须登录） / `OptionalAuth`（可选） / `AuthLite`（只注入身份）
3. **权限中间件**：`Require(spec)` / `RequireAny(specs)` / `RequireAll(specs)` / `RequirePlatformRole(roles)`
4. **平台角色守卫**：`RequirePlatformRole("super_admin")` 必须挂在具体路由上，不全局放行
5. **错误**：业务错误用 `resp.Err(code, msg)`；系统错误 `fmt.Errorf` 包上下文，最终 `HandleError` 兜底
6. **租户上下文**：从 JWT claims 取 `TenantID`，`db.RunInTenantTx(ctx, pool, tenantID, fn)` 自动套 RLS
7. **JSONB 写入**：用 `[]byte`/`string` 写 JSONB 列时，SQL 必须 `::jsonb` cast（pgx 默认发 text/bytea 会报 42804）

## 路由命名

```
/api/v1/{public|protected}/<module>/<resource>
         ↑                ↑
         |                |
         OptionalAuth     Auth
         ↓                ↓
         游客可访问       必须登录
```

| 分组 | 中间件栈 | 示例 |
|---|---|---|
| `public` | `OptionalAuth` + `Require(spec)` | `POST /auth/login`、`GET /flag/frames` |
| `protected` | `Auth` + `Require(spec)` | `POST /users`、`DELETE /roles/:id` |
| `protected` + `/tenants/*` | `Auth` + `RequirePlatformRole("super_admin")` + `Require(spec)` | `POST /tenants` |

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

## 模块（15 个）

按启用方式分组：

| 类型 | 模块 |
|---|---|
| **alwaysOn**（强制启） | `auth`, `tenant`, `system` |
| **optOut**（默认启，白名单过滤） | `user`, `role`, `menu`, `resource`, `permission`, `organization`, `asset`, `dict`, `config` |
| **optional**（默认关） | `weixin`, `cms`, `flag` |

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
# Linux / macOS
./build.sh                                    # → bin/xin

# Windows
.\build.ps1                                   # → bin/xin.exe

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
├── migrations/                   # SQL 迁移（framework / cms / flag / dict / asset / config）
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
    ├── boot/                     # auth, tenant（alwaysOn）
    ├── rbac/                     # menu, organization, permission,
    │                             # resource, role, user
    ├── reference/                # asset, config, dict, weixin
    ├── system/                   # health / cache 运维
    ├── cms/                      # 示例 CMS（extapi 模式）
    └── flag/                     # 头像框 / 空间 / 头像
```

## 当前状态（2026）

| 维度 | 状态 |
|---|---|
| Go modules | **单 module** `gx1727.com/xin`（Phase 1 已合并 framework + apps + cmd） |
| 跨模块全局 | 1 个（`authz.Authorization` interface，无状态） |
| `db.Get` / `config.Get` / `bootx` | 已删（Phase 4-5） |
| main.go | 4 步显式 Build：`config.Load → framework.Boot → 构造 []plugin.Module → framework.Serve` |
| 模块入口 | 全部 `Module(app *appx.App) plugin.Module`，main.go 显式注册 |
| 中间件 | 无 wrapper 重复；Require 全在 `pkg/middleware` |
| P0 单测 | 36 个，覆盖 permission / middleware / plugin 三包 |

## 贡献

提交前必跑：

```bash
go build ./...                                # 必须 EXIT=0
go vet ./...                                  # 必须 EXIT=0
go test ./...                                 # 必须全 PASS
python scripts/strip_bom.py --check .         # 必须无 BOM
```
