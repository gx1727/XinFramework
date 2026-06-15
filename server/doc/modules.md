# 模块清单

> 当前所有模块的清单、依赖、路由前缀、状态。

## 状态总览

| 状态 | 含义 |
| --- | --- |
| ✅ | 已就位 |
| 🚧 | Phase 3 待迁移 |
| ➕ | 可由开发者新增 |

| 模块 | 位置 | 状态 | 路由前缀 |
| --- | --- | --- | --- |
| `system` | `framework/internal/module/system` | 🚧 | `/system` |
| `auth` | `apps/boot/auth` | ✅ | `/auth` |
| `tenant` | `apps/boot/tenant` | ✅ | `/tenants` |
| `user` | `framework/internal/module/user` | 🚧 | `/users` |
| `account` | (auth 模块内) | ✅ | `/accounts` |
| `role` | `framework/internal/module/role` | 🚧 | `/roles` |
| `menu` | `framework/internal/module/menu` | 🚧 | `/menus` |
| `resource` | `framework/internal/module/resource` | 🚧 | `/resources` |
| `permission` | `framework/internal/module/permission` | 🚧 | `/role-resources` |
| `organization` | `framework/internal/module/organization` | 🚧 | `/orgs` |
| `dict` | `framework/internal/module/dict` | 🚧 | `/dicts` |
| `asset` | `framework/internal/module/asset` | 🚧 | `/attachments` |
| `weixin` | `framework/internal/module/weixin` | 🚧 | `/weixin/wxxcx` |
| `cms` | `apps/cms` | ✅ | `/cms/*` |
| `flag` | `apps/flag` | ✅ | `/flag/*` |

---

## 内置模块详情

### auth — 认证

**位置**：[apps/boot/auth](file:///d:\work\xin\XinFramework\server\apps\boot\auth)

**职责**：账号登录、登出、刷新 token、第三方登录（微信小程序）。

**关键依赖**：`tenant`（查询租户）、`permission`（平台角色）。

**路由**：

| Method | Path | Auth | 说明 |
| --- | --- | --- | --- |
| POST | `/auth/login` | public | 账号密码登录 |
| POST | `/auth/logout` | auth | 登出（撤销 session） |
| POST | `/auth/refresh` | public | 刷新 token |
| GET | `/auth/me` | auth | 当前会话信息 |
| POST | `/auth/wxxcx/code2session` | public | 小程序 code 换 session |

**对外暴露**（[framework/pkg/auth](file:///d:\work\xin\XinFramework\server\framework\pkg\auth)）：

- `Account`、`AccountAuth` 结构体（type alias 到 apps/boot/auth）
- `AccountRepository`、`AccountAuthRepository` 接口
- `HashPassword`、`VerifyPassword`
- 注册钩子 `pkgauth.Register(func() AccountRepository)`

### tenant — 租户管理

**位置**：[apps/boot/tenant](file:///d:\work\xin\XinFramework\server\apps\boot\tenant)

**职责**：租户 CRUD、状态切换、初始化（首次安装）。

**关键依赖**：`platform_role`（super_admin）。

**路由**：

| Method | Path | Auth | 说明 |
| --- | --- | --- | --- |
| GET | `/tenants` | super_admin + `tenant:list` | 列表 |
| GET | `/tenants/:id` | super_admin + `tenant:list` | 详情 |
| POST | `/tenants` | super_admin + `tenant:create` | 新建 |
| PUT | `/tenants/:id` | super_admin + `tenant:update` | 更新 |
| PUT | `/tenants/:id/status` | super_admin + `tenant:update` | 启停 |
| DELETE | `/tenants/:id` | super_admin + `tenant:delete` | 软删 |
| POST | `/tenants/:id/purge` | super_admin + `tenant:delete` | 硬删（不可恢复） |

**对外暴露**（[framework/pkg/tenant](file:///d:\work\xin\XinFramework\server\framework\pkg\tenant)）：

- `TenantRecord` 结构体
- `TenantRepository` 接口
- 注册钩子 `pkgtenant.Register`

### user — 用户管理

**位置**：[framework/internal/module/user](file:///d:\work\xin\XinFramework\server\framework\internal\module\user)（Phase 3 迁至 `apps/rbac/user`）

**职责**：租户内用户 CRUD、状态、头像、附件上传。

**关键依赖**：`auth`（创建账号）、`role`、`organization`、`asset`。

**路由**：

| Method | Path | Auth | 说明 |
| --- | --- | --- | --- |
| GET | `/users` | `user:list` | 列表（支持按 org 过滤） |
| GET | `/users/:id` | `user:list` | 详情 |
| POST | `/users` | `user:create` | 新建（同时创建 account） |
| PUT | `/users/:id` | `user:update` | 更新 |
| PUT | `/users/:id/status` | `user:update` | 启停 |
| PUT | `/users/:id/password` | `user:update` | 重置密码 |
| POST | `/users/:id/avatar` | `user:update` | 上传头像 |
| DELETE | `/users/:id` | `user:delete` | 删除 |

### role — 角色

**位置**：[framework/internal/module/role](file:///d:\work\xin\XinFramework\server\framework\internal\module\role)（Phase 3 迁至 `apps/rbac/role`）

**职责**：租户内角色 CRUD、菜单绑定、数据范围配置。

**路由**：`/roles`、`/roles/:id/menus`、`/roles/:id/users`、`/roles/:id/data-scope`

### menu — 菜单

**位置**：[framework/internal/module/menu](file:///d:\work\xin\XinFramework\server\framework\internal\module\menu)（Phase 3 迁至 `apps/rbac/menu`）

**职责**：前端菜单树维护（路由 + 图标 + 国际化 key）。

**路由**：`/menus`（CRUD + 拖拽排序）

### resource — 资源（按钮/接口）

**位置**：[framework/internal/module/resource](file:///d:\work\xin\XinFramework\server\framework\internal\module\resource)（Phase 3 迁至 `apps/rbac/resource`）

**职责**：维护后端资源码（resource:action），用于 `middleware.Require`。

**路由**：`/resources`

### permission — 角色-资源绑定

**位置**：[framework/internal/module/permission](file:///d:\work\xin\XinFramework\server\framework\internal\module\permission)（Phase 3 迁至 `apps/rbac/permission`）

**职责**：维护 `role_resources` 关联表。

**路由**：`/role-resources`

### organization — 组织

**位置**：[framework/internal/module/organization](file:///d:\work\xin\XinFramework\server\framework\internal\module\organization)（Phase 3 迁至 `apps/rbac/organization`）

**职责**：组织树 CRUD、与 user 的绑定。

**路由**：`/orgs`

### dict — 字典

**位置**：[framework/internal/module/dict](file:///d:\work\xin\XinFramework\server\framework\internal\module\dict)（Phase 3 迁至 `apps/reference/dict`）

**职责**：字典分类 + 字典项 + 内存缓存。

**路由**：`/dicts`、`/dicts/:code/items`、`/dict-items`

**对外暴露**：[framework/pkg/dict](file:///d:\work\xin\XinFramework\server\framework\pkg\dict) — 内存缓存接口，业务模块可用 `dict.Get("user_status")` 取所有项。

### asset — 附件

**位置**：[framework/internal/module/asset](file:///d:\work\xin\XinFramework\server\framework\internal\module\asset)（Phase 3 迁至 `apps/reference/asset`）

**职责**：附件元信息 + 文件上传/下载（local 或 COS）。

**路由**：`/attachments`、`/files/:id`

### weixin — 微信

**位置**：[framework/internal/module/weixin](file:///d:\work\xin\XinFramework\server\framework\internal\module\weixin)（Phase 3 迁至 `apps/reference/weixin`）

**职责**：小程序 code 换 openid、自动创建 user/account/session。

**路由**：`/weixin/wxxcx/code2session`

### system — 系统

**位置**：[framework/internal/module/system](file:///d:\work\xin\XinFramework\server\framework\internal\module\system)（保留）

**职责**：健康检查、缓存刷新、缓存统计。

**路由**：`/system/health`、`/system/cache/stats`、`/system/cache/refresh`

---

## Apps 详情

### cms — 内容管理

**位置**：[apps/cms](file:///d:\work\xin\XinFramework\server\apps\cms)

**职责**：内容/分类/标签管理（自定义业务，与 framework 无依赖关系）。

**路由**：`/cms/articles`、`/cms/categories`、`/cms/tags`

### flag — 头像框生成器

**位置**：[apps/flag](file:///d:\work\xin\XinFramework\server\apps\flag)

**职责**：头像框模板、活动头像工具。

**API 文档**：[apps/flag/doc/api.md](file:///d:\work\xin\XinFramework\server\apps\flag\doc\api.md)

**路由**：`/flag/frames`、`/flag/avatars`、`/flag/categories`

---

## 模块间依赖图

```
system ─────────────────────────────┐
                                    │
auth ─→ tenant ─→ permission ─→ role
  │
  └→ user ─→ organization ─→ role
  └→ asset

dict  (无依赖)
weixin ─→ auth ─→ user
cms  (无依赖)
flag  (无依赖)
```

箭头表示"依赖"，启动顺序按依赖图拓扑排序（当前为注册顺序，未启用自动拓扑）。

---

## 配置文件中的启用方式

```yaml
# config/config.yaml
module:
  - system
  - auth
  - tenant
  - user
  - role
  - menu
  - resource
  - permission
  - organization
  - dict
  - asset
  - weixin
  - cms
  - flag
```

**注意**：`main.go` 的 side-effect import 是"打包"开关，`module:` 是"运行"开关。两者必须配合：

| 想做的事 | 改哪里 |
| --- | --- |
| 加一个新 app | `main.go` 加 `_ "gx1727.com/xin/apps/<x>"` + `module:` 加 `- <x>` |
| 临时关闭一个模块 | `module:` 删一行 |
| 永久移除一个模块 | `main.go` 删除 import + 删除 `apps/<x>/` |

---

## 如何知道某个模块启用了

启动日志：

```
2026-06-15 10:00:00 [INFO] module auth registered
2026-06-15 10:00:00 [INFO] module tenant registered
2026-06-15 10:00:00 [INFO] module user registered
...
2026-06-15 10:00:00 [INFO] module flag initialized
```

未在 `module:` 列表中的会被跳过并打 `module X registered but not enabled (skip)`。

---

## 新增模块

详见 [developing.md](file:///d:\work\xin\XinFramework\server\doc\developing.md)。