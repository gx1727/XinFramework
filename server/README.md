# XinFramework Server

> Go 1.25 + Gin + pgx + PostgreSQL. 多租户 SaaS 后端,内建 RBAC、权限中间件、模块化插件框架。

## 一句话概览

```
main.go ──→ framework.Run(cfg)
              ├─ logger.Init
              ├─ db.Init            (pgxpool)
              ├─ cache.Init         (go-redis, 可选)
              ├─ session.New        (Redis or DB)
              ├─ permService + authzService
              ├─ AppContext = { DB, Cache, Cfg, Session, Authz, Repo×8 }
              ├─ for m in plugin.Apps():
              │     if m.Name() ∈ cfg.Module:
              │         m.Init(ctx, w)        ← 模块推 own Repos
              │     m.Register(ctx, public, protected)
              ├─ migrate.Run(./migrations)
              ├─ recovery → request_id → CORS → client_ip → logger
              ├─ listen :cfg.App.Port
              └─ wait SIGINT/SIGTERM → 优雅退出
```

## 文档地图

| 文档 | 用途 |
|---|---|
| [doc/architecture.md](doc/architecture.md) | **必读** Go module 切分、模块生命周期、AppContext 设计、Phase 0-8 重构背景 |
| [doc/quickstart.md](doc/quickstart.md) | 装依赖、配 DB、首次跑通 `xin run` |
| [doc/modules.md](doc/modules.md) | 内置 + 业务模块清单、每个模块的职责/路由/数据表 |
| [doc/api.md](doc/api.md) | HTTP 端点全表(100+ routes)、请求/响应字段 |
| [doc/database.md](doc/database.md) | 表结构、RLS 行级安全、迁移说明 |
| [doc/permissions.md](doc/permissions.md) | RBAC、数据范围(5 种)、平台角色、Spec 中间件用法 |
| [doc/developing.md](doc/developing.md) | 新增模块的标准 8 步流程 |
| [doc/deployment.md](doc/deployment.md) | 编译脚本、systemd、Docker、环境变量 |

## 核心特性

| 特性 | 实现 |
|---|---|
| **多租户** | tenants + account/user 两层模型,RLS 行级隔离 |
| **RBAC** | role → user_role → user,resource → role_resource,支持 `user:*`、`*:*` 通配 |
| **数据范围** | 5 种 Scope:All/Custom/Dept/DeptAndBelow/Self |
| **平台角色** | 跨租户特权,`super_admin` 走 `RequirePlatformRole` |
| **JWT + Session** | HS256,JWT 内含 SessionID,登出即 revoke |
| **可插拔模块** | 启动时 side-effect import,AppContext 注入,无全局依赖 |
| **资源/对象存储** | local 默认,可切 COS(腾讯云) |
| **缓存** | Redis(可选),不可用时 graceful degradation 到 DB session |
| **CORS / 审计 / 日志** | 中间件,可热插拔 |

## 关键约定

1. **统一响应**:`{code, msg, data}`,业务码分段管理([resp/errors.go](framework/pkg/resp/errors.go))
2. **认证中间件**:`Auth`(必须登录) / `OptionalAuth`(可选) / `AuthLite`(只注入身份)
3. **权限中间件**:`Require(spec)` / `RequireAny(specs)` / `RequireAll(specs)` / `RequirePlatformRole(roles)`
4. **平台角色守卫**:`RequirePlatformRole("super_admin")` 必须挂在具体路由上,不会全局放行
5. **错误**:业务错误用 `resp.Err(code, msg)` 返回;系统错误用 `fmt.Errorf` 包上下文,最终 `HandleError` 兜底
6. **租户上下文**:从 JWT claims 取 `TenantID`,通过 `xinContext.WithTenantID(ctx, id)` 注入,`db.RunInTenantTx(ctx, pool, tenantID, fn)` 自动套 RLS

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
| `protected`+`/tenants/*` | `Auth` + `RequirePlatformRole("super_admin")` + `Require(spec)` | `POST /tenants` |

## 命令行

```bash
xin start          # 守护进程启动
xin stop           # 停止
xin restart        # 重启
xin reload         # 平滑重载(目前等价 restart)
xin run            # 前台运行(开发用)
xin status         # 查看 PID 文件状态
xin hot-restart    # 不中断服务的热重启
xin help           # 帮助
```

## 构建

```bash
# Linux/macOS
./build.sh                                    # → bin/xin

# Windows
.\build.ps1                                   # → bin/xin.exe

# 手编(交叉编译)
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
| PostgreSQL | ✅ | 14+(`ltree`、`pg_trgm` 扩展) |
| Redis | ❌ | 7+(`enabled: true` 时启用,不可用时自动降级) |

详细安装步骤见 [doc/quickstart.md](doc/quickstart.md)。

## 仓库结构

```
server/
├── cmd/xin/                   # 主入口
├── config/                    # YAML 配置(config.{yaml,dev,prod} + 子模块 yaml)
├── migrations/                # SQL 迁移(framework / cms / flag / dict / asset)
├── framework/                 # 内置 Go module(framework 框架本体)
│   ├── framework.go           # Run() 入口
│   ├── internal/
│   │   ├── core/              # boot / middleware / server / ext_impl
│   │   └── service/           # authorization / permission service
│   └── pkg/                   # audit / auth / authz / cache / config /
│                             # context / db / dict / extapi / jwt /
│                             # logger / middleware / migrate / model /
│                             # permission / plugin / rbac / resp /
│                             # session / storage / tenant
└── apps/                      # 业务模块 Go module(独立 go.mod)
    ├── boot/                  # auth / tenant(平台级)
    ├── rbac/                  # menu / organization / permission /
    │                          # resource / role / user
    ├── reference/             # asset / dict / weixin
    ├── system/                # health / cache 运维
    ├── cms/                   # 示例 CMS
    └── flag/                  # 头像/相框/空间(示例业务)
```

## Phase 0-8 重构(已完成)

通过 [doc/architecture.md#重构历程](doc/architecture.md) 了解为什么代码现在长这样。

| Phase | 内容 | 状态 |
|---|---|---|
| 0 | 摸底:16 个跨模块全局,409 处引用 | ✅ |
| 1 | go.mod 修复(server + apps 双模块独立) | ✅ |
| 2 | AppContext Reader/Writer 接口骨架 | ✅ |
| 3 | auth + tenant 模块迁移,删 2 个 registry.go | ✅ |
| 4 | rbac 4 件套迁移(user / role / org / perm) | ✅ |
| 5 | authz 完全收官,8 处 apps cache 失效切换 | ✅ |
| 6 | 删 ext_impl/registry.go 死代码(189 行) | ✅ |
| 7 | 删 internal/middleware wrapper 重复(53 行) | ✅ |
| 8 | P0 单测:permission / middleware / plugin(36 测试) | ✅ |

**净收益**:跨模块全局从 12 → 1(仅 authz.Authorization interface,无副作用);删 dead code 525 行;3 包覆盖率 48.4%。

## 贡献

提交前必跑:

```bash
go build ./...         # 必须 EXIT=0
go vet ./...           # 必须 EXIT=0
go test ./...          # 必须全 PASS
./scripts/xin_main_check.exe   # 烟测,确认服务能启动
```