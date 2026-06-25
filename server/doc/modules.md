# 模块清单

> 本文件列出 XinFramework 全部 19 个业务模块、所属数据域、依赖关系、关键路由。
> 三档分类详见 [project-analysis.md §8](./project-analysis.md#8-19-个模块清单)。

---

## 1. 模块总览

| Name | Type | 位置 | 数据表 | 说明 |
|---|---|---|---|---|
| `auth` | alwaysOn | `apps/boot/auth/` | `accounts` / `auth_sessions` | 登录 / 注册 / JWT / 多身份 |
| `tenants` | alwaysOn | `apps/platform/tenants/` | `tenants` | 平台租户 CRUD |
| `system` | alwaysOn | `apps/system/` | — | `/health` + 运维 cache |
| `user` | optOut | `apps/tenant/user/` | `tenant_users` / `tenant_user_roles` | 租户内用户 CRUD |
| `role` | optOut | `apps/tenant/role/` | `tenant_roles` / `tenant_role_data_scopes` / `tenant_user_roles` / `tenant_role_menus` / `tenant_role_resources` | 角色 + 数据范围 |
| `menu` | optOut | `apps/tenant/menu/` | `tenant_menus` / `tenant_role_menus` | 租户菜单树 |
| `organization` | optOut | `apps/tenant/organization/` | `tenant_organizations` | 租户组织（递归 CTE + 物化路径） |
| `permission` | optOut | `apps/tenant/permission/` | `tenant_role_resources` | 租户角色-权限码分配 |
| `resource` | optOut | `apps/tenant/resource/` | `tenant_permissions` | 租户权限码 CRUD |
| `message` | optOut | `apps/tenant/message/` | `messages` | 站内信 |
| `asset` | optOut | `apps/reference/asset/` | `assets` | 文件上传（local / COS） |
| `dict` | optOut | `apps/reference/dict/` | `dicts` / `dict_items` / `dict_visibility` | 数据字典（平台 + 租户二级） |
| `config` | optOut | `apps/reference/config/` | `config_categories` / `config_items` / `config_visibility` | 租户配置中心 |
| `sys_user` | optOut | `apps/platform/sys_user/` | `sys_users` / `sys_orgs` / `sys_user_roles` | 平台域用户身份 |
| `sys_role` | optOut | `apps/platform/sys_role/` | `sys_roles` / `sys_user_roles` | 平台域角色（含 super_admin） |
| `sys_menu` | optOut | `apps/platform/sys_menu/` | `sys_menus` / `sys_role_menus` | 平台域菜单 |
| `sys_permission` | optOut | `apps/platform/sys_permission/` | `sys_permissions` / `sys_role_permissions` | 平台域权限码 |
| `weixin` | optional | `apps/reference/weixin/` | — | 微信小程序登录 |
| `cms` | optional | `apps/cms/` | `posts` | 示例 CMS（extapi 模式） |
| `flag` | optional | `apps/flag/` | `flag_*` | 头像框 / 空间 / 头像 |

---

## 2. 依赖关系

### 2.1 alwaysOn（必装）

```
auth ──┬── depends on: db, jwt, session, cache, audit
       └── writes: AccountRepo, AccountAuthRepo
                ↓
tenants ─┬── depends on: db
         └── writes: TenantRepo
                ↑
        (auth 读 TenantRepo 验证登录)
                ↓
system ─┬── depends on: db
        └── 无外部依赖
```

### 2.2 optOut（默认启用）

**租户域 RBAC 链路**：

```
organization ─┬── writes: OrgRepo
              └── reads: db
                  
user ─┬── writes: UserRepo
      ├── reads: OrgRepo, RoleRepo, AccountRepo, asset service
      └── (本模块最复杂：依赖所有 RBAC)

role ─┬── writes: RoleRepo
      ├── reads: db
      └── (与 menu/permission/resource 双向引用)

menu ─┬── writes: (无；纯读)
      └── reads: db

permission ─┬── writes: PermRepo
            └── reads: db, RoleRepo

resource ─┬── writes: (无；纯读)
          └── reads: db

message ─┬── writes: (无；纯读)
         └── reads: db
```

**平台域链路**（Phase 0023+）：

```
sys_user ─┬── writes: (无；纯读)
          └── reads: db, sys_role/sys_user_roles

sys_role ─┬── writes: (无；纯读)
          └── reads: db, sys_permission/sys_role_permissions

sys_menu ─┬── writes: (无；纯读)
          └── reads: db, sys_role/sys_role_menus

sys_permission ─┬── writes: (无；纯读)
                └── reads: db
```

**基础设施**：

```
dict ─┬── writes: (无；纯读)
      └── reads: db

config ─┬── writes: (无；纯读)
        └── reads: db, 依赖 dictpkg 缓存

asset ─┬── writes: (无；纯读)
      └── reads: db, storage
```

### 2.3 optional（默认关闭）

```
weixin ─┬── writes: (无；纯读)
        └── reads: db, weixin API SDK

cms ──── (extapi 模式，无跨模块依赖)

flag ─── (avatar/frame 模块，依赖 asset 上传)
```

---

## 3. 路由清单

### 3.1 public 域（`/api/v1/*`，OptionalAuth）

| 路径 | 模块 | 说明 |
|---|---|---|
| `POST /auth/tenant-login` | auth | 租户域登录 |
| `POST /auth/platform-login` | auth | 平台域登录 |
| `POST /auth/login-precheck` | auth | 多身份账号列身份 |
| `POST /auth/select-tenant` | auth | precheck 后选身份签 token |
| `POST /auth/register` | auth | 注册新用户 |
| `POST /auth/refresh` | auth | 刷新 access token |
| `POST /auth/login` | auth | 旧入口，等价转发到 tenant-login |
| `GET /health` | system | 健康检查 |
| `GET /public/configs` | config | 公开读配置（站点名等） |
| `GET /flag/frames*` | flag | 公开相框资源 |
| `GET /flag/spaces/:code` | flag | 公开活动空间 |
| `GET /flag/avatar-categories` | flag | 公开头像分类 |
| `GET /flag/avatars` | flag | 公开头像列表 |
| `POST /weixin/login` | weixin | 微信小程序登录 |
| `POST /weixin/phone` | weixin | 微信手机号授权 |
| `GET /weixin/ping` | weixin | 微信模块 ping |
| `GET /cms/ping` | cms | CMS 模块 ping |

### 3.2 tenant 域（`/api/v1/*`，Auth + RequireTenantContext）

| 路径 | 模块 |
|---|---|
| `GET /users` / `POST /users` / `GET/PUT/PATCH/DELETE /users/:id` | user |
| `PUT /users/:id/status` | user |
| `PUT /users/:id/org` | user |
| `GET /user/profile` / `PUT /user/profile` / `POST /user/avatar` | user |
| `GET /organizations` / `POST /organizations` / `GET/PUT/DELETE /organizations/:id` | organization |
| `GET /organizations/tree` | organization |
| `GET /roles` / `POST /roles` / `GET/PUT/DELETE /roles/:id` | role |
| `GET /roles/:id/menus` / `PUT /roles/:id/menus` | role |
| `GET /roles/:id/data-scopes` / `PUT /roles/:id/data-scopes` | role |
| `GET /menus` / `POST /menus` / `GET/PUT/DELETE /menus/:id` | menu |
| `GET /menus/tree` | menu |
| `GET /resources` / `POST /resources` / `GET/PUT/DELETE /resources/:id` | resource |
| `GET /resources/my` | resource |
| `GET /resources/by-menu/:menu_id` | resource |
| `GET /roles/:id/permissions` / `PUT /roles/:id/permissions` | permission |
| `GET /roles/:id/resources` / `PUT /roles/:id/resources` | permission |
| `GET /messages` / `POST /messages` | message |
| `GET /dicts` / `POST /dicts` / `GET/PUT/DELETE /dicts/:id` | dict |
| `GET /dicts/resolve` | dict |
| `GET /dicts/:id/items` | dict |
| `GET /configs` / `POST /configs` / `GET/PUT/DELETE /configs/:id` | config |
| `GET /configs/resolve` | config |
| `GET /configs/:id/items` | config |
| `POST /asset/upload` / `GET /asset/:id` | asset |
| `GET /flag/frames` / `POST /flag/frames` | flag |
| `GET /flag/frames-categories` | flag |
| `GET /flag/spaces` / `POST /flag/spaces` | flag |
| `GET /flag/avatars` / `POST /flag/avatars` | flag |
| `GET /flag/avatar-categories` | flag |
| `POST /flag/generate` | flag |
| `GET /flag/my-avatars` | flag |
| `POST /weixin/bind-phone` | weixin |
| `GET /system/server-info` / `POST /system/clear-cache` | system |
| `GET /system/cache/*` | system |
| `GET /cms/me` / `GET /cms/users` / `GET /cms/tenant` | cms |
| `GET /cms/posts*` / `POST /cms/posts*` | cms |

### 3.3 platform 域（`/api/v1/platform/*`，Auth + RequirePlatformRole）

| 路径 | 模块 |
|---|---|
| `GET /platform/tenants` / `POST /platform/tenants` | tenants |
| `GET/PUT/DELETE /platform/tenants/:id` | tenants |
| `PUT /platform/tenants/:id/status` | tenants |
| `POST /platform/tenants/:id/purge` | tenants |
| `GET /platform/sys-users` / `POST /platform/sys-users` | sys_user |
| `GET/PUT/DELETE /platform/sys-users/:id` | sys_user |
| `PUT /platform/sys-users/:id/status` | sys_user |
| `PUT /platform/sys-users/:id/roles` | sys_user |
| `GET /platform/sys-roles` / `POST /platform/sys-roles` | sys_role |
| `GET/PUT/DELETE /platform/sys-roles/:id` | sys_role |
| `GET/PUT /platform/sys-roles/:id/menus` | sys_role |
| `GET/PUT /platform/sys-roles/:id/permissions` | sys_role |
| `GET /platform/sys-menus` / `POST /platform/sys-menus` | sys_menu |
| `GET/PUT/DELETE /platform/sys-menus/:id` | sys_menu |
| `GET /platform/sys-menus/tree` | sys_menu |
| `GET /platform/sys-permissions` / `POST /platform/sys-permissions` | sys_permission |
| `GET/PUT/DELETE /platform/sys-permissions/:id` | sys_permission |
| `GET /platform/dicts` / `POST /platform/dicts` | dict |
| `GET/PUT/DELETE /platform/dicts/:id` | dict |
| `GET /platform/dicts/:id/items` | dict |
| `GET /platform/dicts/:id/visibility` | dict |
| `GET /platform/configs` / `POST /platform/configs` | config |
| `GET/PUT/DELETE /platform/configs/:id` | config |
| `GET /platform/configs/:id/items` | config |
| `GET /platform/configs/:id/visibility` | config |

---

## 4. 鉴权层级

每个路由都至少经过：

1. **JWT 解析**（Auth / OptionalAuth）— 注入 `XinContext`
2. **资源权限**（`middleware.Require(permission.P(ResXxx, ActYyy))`）— 细粒度 RBAC

`platform` 域路由额外叠加 `RequirePlatformRole("super_admin")`。

**短路**：拥有 `*:*` 通配权限或 `super_admin` 平台角色时，`Require*` 全部放行。

---

## 5. 模块启用配置

`config/config.yaml`：

```yaml
module:
  - weixin
  - cms
  - flag
```

**语义**：累加 optional（默认关闭），不动 alwaysOn + optOut。

| 档 | 行为 |
|---|---|
| `alwaysOn` | 不可关，必启 |
| `optOut` | 默认全开；写 `module: [...]` 时视为白名单，不列就关 |
| `optional` | 默认关；必须显式列出 |

生产环境 `config.prod.yaml` 强制显式声明 `module:` 列表。

---

## 6. 新增模块的入口

详见 [developing.md](./developing.md)。简要步骤：

1. 在 `apps/<domain>/<module>/` 创建包
2. 写 8 个文件（module.go / routes.go / handler.go / service.go / repository.go / model.go / types.go / errors.go）
3. 在 `cmd/xin/main.go` 的 `modules := []plugin.Module{...}` 列表中追加
4. 申请 `resp.CodeXxx` 错误码段
5. 在 `migrations/<module>.sql` 加 DDL（带 `IF NOT EXISTS`）
6. 写 menu / permission seed（可选）

---

## 7. 关键文件索引

| 关注点 | 路径 |
|---|---|
| 模块契约 | `framework/pkg/plugin/plugin.go` |
| 启动入口 | `cmd/xin/main.go` |
| 启动编排 | `framework/internal/core/boot/boot.go` |
| 路由装配 + 模块过滤 | `framework/framework.go` |
| 进程级资源 | `framework/pkg/appx/appx.go` |
| 跨模块容器 | `framework/pkg/plugin/appcontext.go` |
| 鉴权中间件（internal） | `framework/internal/core/middleware/auth.go` |
| 鉴权守卫（公开） | `framework/pkg/middleware/auth.go` |
| 登录 / 账号 | `apps/boot/auth/` |
| 平台租户 | `apps/platform/tenants/` |
| 平台域 sys_* | `apps/platform/sys_{user,role,menu,permission}/` |
| 租户域 RBAC | `apps/tenant/{user,role,menu,resource,organization,permission}/` |
| 字典 / 配置 | `apps/reference/{dict,config}/` |
| 响应 / 错误码 | `framework/pkg/resp/` |
| 权限 Spec | `framework/pkg/permission/spec.go` |
| 权限常量 | `framework/pkg/permission/constants.go` |
| JWT | `framework/pkg/jwt/jwt.go` |
| Session | `framework/pkg/session/session.go` |
| 事务工具 | `framework/pkg/db/db.go` |
| 审计 | `framework/pkg/audit/audit.go` |
| 迁移 | `framework/pkg/migrate/migrate.go` |
| Schema | `migrations/init_schema.sql` |
| Seed | `migrations/init_seed.sql` |
| 配置 | `config/config.yaml` |
| 请求上下文 | `framework/pkg/xincontext/context.go` |
