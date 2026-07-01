# 项目综合分析

> XinFramework 全栈项目分析报告。
>
> 本文件是项目整体的全景图：定位、架构、技术选型、模块、约定、设计亮点。
> 各子主题的深入细节见 `server/doc/` 同目录下其它文件：
> [architecture.md](./architecture.md) / [quickstart.md](./quickstart.md) /
> [modules.md](./modules.md) / [api.md](./api.md) / [database.md](./database.md) /
> [permissions.md](./permissions.md) / [developing.md](./developing.md) /
> [deployment.md](./deployment.md)

---

## 1. 一句话定位

**XinFramework** 是一个面向多租户 SaaS 后台的可扩展框架，包含：

- **后端**：Go 1.24+ / Gin / pgx / PostgreSQL 14+，单一 Go Module（`gx1727.com/xin`）
- **前端**：React 19 / TypeScript 5.9 / Vite 7 / shadcn/ui
- **核心场景**：多租户隔离 + RBAC + 数据范围 + 平台/租户双域
- **模块规模**：19 个模块（3 alwaysOn + 13 optOut + 3 optional），全部跑在同一进程
- **当前阶段**：Phase 0023 平台/租户域分域完成（2026-06-23）

---

## 2. 仓库结构

```
XinFramework/
├── server/                        # Go 后端（单一 module gx1727.com/xin）
│   ├── cmd/
│   │   ├── xin/                   # 服务入口（start/stop/run/reload/hot-restart/status）
│   │   └── rotate-admin-password/ # 独立工具：重置 admin 密码
│   ├── config/                    # YAML 配置（config.yaml / dev / prod / cms.yaml）
│   ├── migrations/                # SQL 迁移（init_schema / init_seed / asset/cms/flag/message）
│   ├── framework/                 # 框架内核（必装）
│   │   ├── framework.go           # Serve / Boot 入口
│   │   ├── internal/              # core/boot + core/server + core/middleware
│   │   └── pkg/                   # appx / authz / db / plugin / resp / permission / jwt / session
│   │                             # cache / audit / migrate / storage / tenant / xincontext / ...
│   ├── apps/                      # 业务模块（19 个）
│   │   ├── boot/auth/             # 登录 / 账号（必装）
│   │   ├── sys/                   # sys 管理域：tenants / user / role / menu / permission / org
│   │   ├── tenant/                # 租户域 RBAC：user / role / menu / organization / permission / resource
│   │   │                          # + tenant/message（站内信）
│   │   ├── reference/             # 基础设施：asset / config / dict / weixin
│   │   ├── system/                # health / cache 运维（必装）
│   │   ├── cms/                   # 示例 CMS（extapi 模式）
│   │   └── flag/                  # 头像框 / 空间 / 头像（optional）
│   ├── doc/                       # 项目文档（本目录）
│   ├── bin/                       # 构建产物（xin.exe / rotate-admin-password.exe）
│   ├── scripts/strip_bom.py       # BOM 检测 / 剥离工具（CI gate）
│   ├── AGENTS.md                  # 后端设计参考（高密度速查）
│   ├── go.mod / go.sum / go.work.sum
│   └── build.sh / build.ps1
└── UI/                            # React 前端
    └── src/
        ├── api/                   # 19 个 API 模块（auth/user/role/menu/...）
        ├── components/
        │   ├── ui/                # shadcn/ui（~25 个组件）
        │   ├── schema/            # DynamicForm + DynamicTable + showIfEvaluator
        │   └── permission/        # Auth.tsx（按钮级）+ DynamicRouter.tsx（路由级）
        ├── locales/               # i18n（仅 zh-CN，类型源头）
        ├── pages/                 # 24 个页面
        ├── stores/                # zustand（authStore / menuStore / configStore / permissionStore）
        ├── types/schema.ts        # FormSchema / TableSchema
        └── App.tsx                # 路由（scope-based guard）
```

---

## 3. 核心能力

| 能力 | 实现要点 |
|---|---|
| **多租户隔离** | 业务表带 `tenant_id` + RLS `FORCE ROW LEVEL SECURITY`；所有 SQL 强制 tenant 过滤 |
| **RBAC + 数据范围** | 用户 → 角色 → 资源权限码（`user:list` / `flag:create` 等）；角色携带 data_scope（全部/部门/本人/自定义/本部门及以下） |
| **平台角色** | 跨租户特权（`super_admin`），独立于租户内 RBAC，自动 bypass 所有 spec |
| **统一响应** | `{code, msg, data}`，`code=0` 成功；code 分段 → HTTP 状态码单点映射 |
| **插件化模块** | 内置模块（boot）与外部 app（cms / flag）走同一 `Module(app)` 工厂注册；按 `cfg.Module` 白名单启停 |
| **JSONB 安全** | 所有 JSONB 列在 SQL 里显式 `::jsonb` cast（避免 pgx 把 string/[]byte 当 text/bytea 发） |
| **多身份账号** | 路径 B：`login-precheck` 列出所有身份 → 用户选择 → `select-tenant` 签 token；`refresh` 可切租户 |
| **配置中心** | `config_items.value JSONB` + `is_override + platform_item_id` 支持租户覆盖平台项 |
| **审计** | `audit.Log` 写 `db_logs`，失败仅记日志不返回（业务路径不被审计写库失败打断） |
| **自动迁移** | 启动期 `migrate.Run` 跑 SQL，`_schema_migrations` 记录版本；DDL/Seed 全部幂等 |

---

## 4. 数据域分层（Phase 0023 终态）

| 域 | 标识 | RLS | API 守卫 |
|---|---|---|---|
| **sys 域** `sys_*` | 无 `tenant_id` | ❌ 不启用 | `RequireSysRole("super_admin")` + `db.RunInPlatformTx` |
| **租户域** `tenant_*` | 带 `tenant_id` | ✅ 全部启用 | `RequireTenantContext`（tenant_id > 0） |
| **共享层** | `accounts` / `tenants` / `auth_sessions` | ❌ | 由对应业务模块自己路由 |

RLS policy 模板：
```sql
CREATE POLICY tenant_isolation_policy ON tenant_xxx USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);
```

字典/配置类表有 `tenant_id = 0` 短路：平台级跨租户共享。

---

## 5. 启动流程（4 步显式 Build）

```
cmd/xin/main.go
  ├─ config.Load("config/config.yaml")          # YAML + .env + 环境变量
  ├─ framework.Boot(cfg)                        # → (*appx.App, *Runtime)
  │    └─ boot.Init(cfg)
  │         ├─ logger.Init
  │         ├─ db.Init  → *pgxpool.Pool
  │         ├─ dict.Init(pool)
  │         ├─ cache.Init  → Redis（可降级）
  │         ├─ session.Init（Redis 或 DB 实现）
  │         ├─ plugin.NewAppContext(...)
  │         └─ service.NewAuthorizationService + appCtx.SetAuthz
  ├─ modules := []plugin.Module{ ... }          # main.go 显式列
  └─ framework.Serve(cfg, app, rt, modules)
       ├─ migrate.Run(app.DB, "migrations")
       ├─ for m in modules:  m.Init(ctx, w)    # 各模块写自己的 slot
       ├─ setupRouter：全局中间件 + 三组 RouterGroup
       ├─ for m in modules:  m.Register(ctx, public, tenant, protected)
       └─ 阻塞，收到信号后按顺序调 m.Shutdown
```

`cmd/xin/main.go` 同时是 `xin start/stop/reload/hot-restart/status` 命令的入口（`framework/cmd.go`）。

---

## 6. 三组 RouterGroup 语义

`framework/framework.go:registerModules` 把 `v1 := r.Group("/api/v1")` 拆成：

| 组 | URL 前缀 | 中间件 | 用途 |
|---|---|---|---|
| `public` | `/api/v1/*` | `OptionalAuth` | 登录、健康、公开读、`<module>/public/*` 隔离子资源 |
| `tenant` | `/api/v1/*`（**无 `/t` 前缀**） | `Auth` + `RequireTenantContext` | 业务域；模块直接挂资源路径 |
| `protected` | `/api/v1/sys/*` | `Auth` | sys 域；模块内部追加 `RequireSysRole("super_admin")` |

---

## 7. 模块契约 `plugin.Module`

```go
type Module interface {
    Name() string
    Init(ctx Reader, w Writer) error                // 写 AppContext 槽位
    Register(ctx Reader, public, tenant, protected *gin.RouterGroup)
    Shutdown(ctx Reader) error
}
```

字段填充用 `plugin.BaseModule`（`NameStr` / `InitFn` / `RegFn` / `StopFn`），不需的方法留 nil。

**`plugin.AppContext` 是核心 DI 容器**，通过 `Reader / Writer` 接口分离读写角色——模块只能往自己拥有的槽位写，避免误覆盖别人。

| 槽 | Reader | Writer（写方） |
|---|---|---|
| `DB() / Cache() / Config() / Session()` | 读 | `framework/boot` 启动期填充 |
| `Authz()` | 读 | `framework/boot`（AuthorizationService） |
| `AccountRepo() / AccountAuthRepo()` | 读 | `apps/boot/auth` |
| `TenantRepo()` | 读 | `apps/sys/tenants` |
| `UserRepo() / RoleRepo() / OrgRepo() / PermRepo()` | 读 | `apps/tenant/{user,role,organization,permission}` |

Reader 返回 nil 表示对应模块未启用，**调用方必须 nil-check**。

---

## 8. 19 个模块清单

| 档 | 数量 | 含义 | 模块 |
|---|---|---|---|
| `alwaysOn` | 3 | 必装 | `system` / `auth` / `sys_tenant` |
| `optOut` | 13 | 框架默认启用，不写即为开 | RBAC：`menu / user / role / resource / organization / permission`<br>基础设施：`dict / asset / config`<br>sys 管理：`sys_user / sys_role / sys_menu / sys_permission` |
| `optional` | 3 | 默认关，必须显式列出 | `weixin` / `cms` / `flag` |

`cfg.module` 现在的语义是**累加 optional**，不是白名单。生产环境必须显式声明 `module:` 列表（`config.prod.yaml` 强制）。

---

## 9. 鉴权与会话

**JWT Claims**（`framework/pkg/jwt/jwt.go`）：
```go
type Claims struct {
    UserID                 uint     // tenant login: tenant_users.id；sys login: account_id
    TenantID               uint     // 0 = sys 域
    Role                   string   // tenant role code / "_sys"（sys 域占位）
    SessionID              string   // 服务端 session 标识
    TokenType              string   // "access" | "refresh"
    SysRoles               []string // super_admin 等
    ImpersonatedBy         uint     // 模拟登录：原 sys 账号 ID
    ImpersonationSessionID string   // 模拟登录：原 sys session id
    jwt.RegisteredClaims
}
```

**登录入口**（按 scope 拆分）：

| 端点 | scope | tenant_id | 适用 |
|---|---|---|---|
| `POST /auth/tenant-login` | `tenant` | 必填 | 业务用户登录 |
| `POST /auth/sys-login` | `sys` | 0 | super_admin 登录 |
| `POST /auth/login-precheck` | — | — | 多身份账号列出可用身份 |
| `POST /auth/select-tenant` | `tenant` | 必填 | precheck 后选身份签 token |
| `POST /auth/refresh` | 同原 token | 可选 | 切租户时传 `tenant_id`；sys token 不允许切租户 |
| `POST /auth/logout` | — | — | 撤销 session |
| `POST /auth/register` | `tenant` | 必填 | 注册新用户 |
| `POST /auth/login` | `tenant` | 必填 | 旧入口，**等价转发到 tenant-login** |
| `POST /auth/platform-login` | `sys` | 0 | 兼容期路径，**转发到 /auth/sys-login** |

**中间件**：

| 中间件 | 行为 |
|---|---|
| `Auth` | 必登录，注入 `Context`（轻量身份） + 注册 `UserContext` 懒加载器（RBAC + DataScope） |
| `AuthLite` | 必登录，只注入 `Context`；不挂权限加载器 |
| `OptionalAuth` | 有 token 走 `Auth`；无 token 时从 `X-Tenant-ID` header 兜底注入 `Context.TenantID` |
| `Require(spec)` | 全部 spec 校验（`MatchAll`），缺一个 403 |
| `RequireAny(specs...)` / `RequireAll(specs...)` | 任一通过 / 全部通过 |
| `RequireAuthenticated()` | 只校验登录 |
| `RequireSysRole(roles...)` | 校验 `Context.SysRoles` 包含任一 role |
| `RequireAnySysRole()` | 校验 `Context.SysRoles` 非空（任何 sys 角色） |
| `RequireSysScope()` | `Context.TenantID == 0`；挡住 tenant token |
| `RequireTenantContext()` | `Context.TenantID > 0`；挡住 sys token |

**短路**：拥有 `*:*` 资源权限或 `super_admin` 平台角色时，`Require*` 全部放行。

**密码**：`apps/boot/auth/password.go` 用 argon2id（m=64MiB, t=3, p=4），格式 `$argon2id$v=19$m=...,t=...,p=...$<salt>$<hash>`。

**Session**：`framework/pkg/session` 提供 `RedisSessionManager` / `DBSessionManager` 两套实现；Redis 可用时优先；DB 模式自动 `CREATE TABLE IF NOT EXISTS auth_sessions`。

---

## 10. 错误与响应

**统一响应** `{ code, msg, data }`，`code=0` 成功。所有 API **禁止手写 JSON**。

**Code 分段 → HTTP 状态码**（`resp.CodeToHTTPStatus` 单点决定）：

| 段 | HTTP | 用途 |
|---|---|---|
| `1xxx` | 200 | 鉴权 / 账号 / 通用业务（body Code 表达真实结果） |
| `2xxx` | 400 | 参数校验 / 业务规则 |
| `3xxx` | 404 | 资源不存在 / 租户级 |
| `4xxx` | 403 | 权限不足 / 角色冲突 |
| `5xxx+` | 500 | 服务端故障 / 系统异常 |

**模块码段分配**（`framework/pkg/resp/errors.go`）：

| 段 | 模块 | 段 | 模块 |
|---|---|---|---|
| 1000 | auth | 8000 | resource |
| 2000 | user | 9000 | asset |
| 3000 | tenant | 10000 | dict |
| 4000 | role | 11000 | system |
| 5000 | menu | 12000 | weixin |
| 6000 | organization | 13000 | flag |
| 7000 | permission | 14000 | cms |

**Handler 模板**：
```go
func (h *Handler) Get(c *gin.Context) {
    id, err := parseIDParam(c, "id")
    if err != nil { resp.BadRequest(c, "无效的ID参数"); return }
    out, err := h.svc.GetByID(c.Request.Context(), id)
    if err != nil { resp.HandleError(c, err); return }
    resp.Success(c, out)
}
```

---

## 11. 前端技术栈与约定

| 维度 | 选型 |
|---|---|
| 构建 | Vite 7.3 + @vitejs/plugin-react |
| 框架 | React 19.2 + TypeScript 5.9 |
| 样式 | Tailwind CSS v4（`@tailwindcss/vite` 插件） + tw-animate-css |
| UI 组件 | shadcn/ui（基于 Radix UI），~25 个组件 |
| 图标 | lucide-react + 自研 `IconPicker` |
| 路由 | react-router-dom v7（`App.tsx` 集中 lazy + `RequireScope` guard） |
| 状态 | zustand v5（`authStore` localStorage 持久化） |
| 文案 | 自研：仅简体中文，`src/locales/zh-CN.ts` 是类型源头 |
| HTTP | 原生 `fetch` + `ApiError`（带 JWT 自动 refresh + 指数退避重试） |
| 表格 | @tanstack/react-table v8 |
| 图表 | Recharts v3 |
| 通知 | Sonner v2 |
| 表单验证 | Zod v4 |

**关键约定**：

- **i18n**：`zh-CN.ts` 是类型源头；用 `t.xxx.yyy` 对象访问，无 hook
- **Schema 驱动**：表单 `FormSchema { items: FormItemSchema[] }`，表格 `TableSchema { columns, search?, actions? }`；字段类型 `text / number / select / radio / checkbox / switch / date / icon / divider / slot`；`showIf` 条件显示
- **API 客户端**：自动加 `Authorization: Bearer <token>`、自动 401 refresh、指数退避重试
- **路由**：前端 `/app/*`（tenant）、`/sys/*`（sys），与后端三组 RouterGroup 对齐
- **Mock 兜底**：不再静默，catch 内必须 `setError`；mock 仅在 `useMockFallback=true` 时使用
- **编码**：**所有源文件 UTF-8 无 BOM**（PowerShell 默认 GBK 会破坏中文）

---

## 12. 前端 ↔ 后端 API 契约

### 12.1 统一响应

```json
{ "code": 0, "msg": "ok", "data": { /* 业务数据 */ } }
```

分页接口 `data` 结构：
```json
{ "total": 100, "list": [ /* ... */ ], "page": 1, "size": 20 }
```

### 12.2 路由空间对照

| 域 | 后端 | 前端 |
|---|---|---|
| public | `POST /auth/*`、`GET /health`、`GET /public/configs`、`GET /flag/frames*` | `/login`、`/sys/login` |
| tenant | `/users` `/roles` `/menus` `/resources` `/organizations` `/dicts` `/configs` `/asset` `/flag/*` `/messages` 等 | `/app/*` |
| sys | `/sys/tenants`、`/sys/sys-users`、`/sys/sys-roles`、`/sys/menus`、`/sys/sys-permissions`、`/sys/dicts`、`/sys/configs`、`/sys/system/*` | `/sys/*`（仅 super_admin） |

### 12.3 关键模块映射

| 后端模块 | 前端页面 | 路由 | 前端 API | 后端路径 |
|---|---|---|---|---|
| `auth` | Login.tsx / TenantLogin.tsx / SysLogin.tsx | `/login`, `/sys/login` | `authApi` | `/auth/*` |
| `user` | Users.tsx | `/app/users` | `userApi` | `/users/*` |
| `role` | Roles.tsx | `/app/roles` | `roleApi` | `/roles/*` |
| `menu`（租户域） | Menus.tsx | `/app/menus` | `menuApi` | `/menus/*` |
| `menu`（sys 域） | Menus.tsx（Tab） | `/sys/menus` | `sysMenuApi` | `/sys/menus/*` |
| `organization` | Organizations.tsx | `/app/organizations` | `organizationApi` | `/organizations/*` |
| `resource` | Resources.tsx | `/app/resources` | `resourceApi` | `/resources/*` |
| `asset` | （无独立页） | — | `assetApi` | `/asset/*` |
| `dict` | Dicts.tsx | `/app/dicts` | `dictApi` | `/dicts/*` |
| `config` | Configs.tsx / SysConfigs.tsx | `/app/configs`, `/sys/configs` | `configApi` | `/configs/*`, `/sys/configs/*`, `/public/configs` |
| `flag` | Frames.tsx / FrameCategories.tsx / Avatars.tsx / AvatarCategories.tsx | `/app/frames` 等 | `frameApi` 等 | `/flag/*` |
| `cms` | — | — | — | `/cms/*` |
| `tenants` | Tenants.tsx | `/sys/tenants` | `tenantApi` | `/sys/tenants/*` |
| `system` | Cache.tsx | `/sys/cache` | `systemApi` | `/sys/system/cache/*`, `/sys/system/clear-cache` |
| `weixin` | （无独立页） | — | — | `/weixin/*` |
| `sys_user` | — | `/sys/users` | — | `/sys/sys-users/*` |
| `sys_role` | — | `/sys/roles` | — | `/sys/sys-roles/*` |
| `sys_menu` | Menus.tsx（Tab） | `/sys/menus` | `sysMenuApi` | `/sys/menus/*` |

**关键约定**：
- 前端路由带 scope 前缀：`/app/*`（tenant）、`/sys/*`（sys）
- 同一前端页面可能调多个后端 API（如 `Menus.tsx` 同时调 `menuApi` 和 `sysMenuApi`）
- `super_admin` 判断：前端用 `useAuthStore().user?.sys_role_codes?.includes("super_admin")`；后端用 `RequireSysRole("super_admin")`

---

## 13. 设计模式与亮点

| 维度 | 模式 |
|---|---|
| 模块化 | 显式 Build，无 `init()` 注册表；`plugin.BaseModule` + `Reader/Writer` 接口分离 |
| 依赖注入 | 跨模块共享走 `plugin.AppContext` 槽位，nil-check 模式 |
| 多租户隔离 | DB 层 RLS + 应用层 `RunInTenantTx` + 中间件 `RequireTenantContext`，三层防御 |
| 平台 vs 租户 | 平台域表无 `tenant_id`、不启用 RLS；`RunInPlatformTx` 设 `bypass_rls=on` |
| 鉴权 | JWT + Session（Redis 优先，DB 降级）+ RBAC 资源权限 + DataScope 数据范围 + 平台角色 |
| 多身份账号 | 路径 B：`login-precheck` 列出身份 → `select-tenant` 签 token；`refresh` 可切租户 |
| 配置中心 | `config_items.value JSONB` + `is_override + platform_item_id` 支持租户覆盖平台项 |
| 错误处理 | 统一响应 + 分段码 → HTTP 单点映射；DB 层命名 sentinel；service 层 `mapRepoError` 翻译 |
| 审计 | `audit.Log` 失败不抛，业务路径不被审计写库失败打断 |
| 迁移 | 启动期自动跑 + `_schema_migrations` 记录；DDL/Seed 全部幂等 |
| 前端 Schema | 表单/表格/搜索全 Schema 驱动，showIf 条件显示，FormDialog 用 `formKey` 重置 |
| 前端 mock | 不再静默兜底；用户主动勾选；显示数据源徽章 |

---

## 14. 项目状态

| 维度 | 状态 |
|---|---|
| 总模块数 | 19（3 alwaysOn + 13 optOut + 3 optional） |
| Go module | 单一 `gx1727.com/xin`（框架 + 业务同 module） |
| 数据库表 | 34 张（init_schema）+ 业务模块独立 DDL |
| Phase | 0023 平台/租户域分域完成（2026-06-23） |
| 前端页面 | 24 个（4 登录 + 14 业务 + 5 平台 + 1 缓存） |
| 多身份登录 | 路径 B 完成（precheck → select-tenant） |
| 文档 | 完整（README + AGENTS.md × 2 + 本目录 8 篇 server/doc/*.md） |
| CI 工具 | BOM 检测 `strip_bom.py`（CI gate） |

**已知的"已知 gap"**（`framework/pkg/resp/resp.go` 包级文档里点出）：menu（5xxx）/ organization（6xxx）等模块的"资源不存在"类错误目前共用 5xxx+ 段，按现有规则会被映射为 HTTP 500。修复路径是给这些模块重新分配段位（如把 menu 调到 35xx 段）。
