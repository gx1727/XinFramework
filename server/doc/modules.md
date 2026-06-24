# 模块清单

> 当前共 **19 个 module**。按 `cfg.Module` 行为分 3 类：3 个 alwaysOn、8 个 optOut、8 个 optional。
> 文档版本：2026-06（Phase 0023 全阶段完成后）

## 总览

| Name | 类型 | 错误码段 | 主要表 / 资源 | 路径 |
|---|---|---|---|---|---|
| [system](#system) | alwaysOn | 11001-11999 | — | apps/system |
| [auth](#auth) | alwaysOn | 1001-1999 | accounts / auth_sessions | apps/boot/auth |
| [tenants](#tenants) | alwaysOn | 3001-3999 | tenants | apps/platform/tenants |
| [sys_user](#sys_user) | optional | — | sys_users / sys_user_roles | apps/platform/sys_user |
| [sys_role](#sys_role) | optional | — | sys_roles / sys_role_menus / sys_role_permissions | apps/platform/sys_role |
| [sys_menu](#sys_menu) | optional | 15001-15999 | sys_menus / sys_role_menus | apps/platform/sys_menu |
| [sys_permission](#sys_permission) | optional | — | sys_permissions | apps/platform/sys_permission |
| [user](#user) | optOut | 2001-2999 | tenant_users / tenant_user_roles | apps/tenant/user |
| [role](#role) | optOut | 4001-4999 | tenant_roles / tenant_role_data_scopes / tenant_user_roles / tenant_role_menus | apps/tenant/role |
| [menu](#menu) | optOut | 5001-5999 | tenant_menus / tenant_role_menus | apps/tenant/menu |
| [organization](#organization) | optOut | 6001-6999 | tenant_organizations | apps/tenant/organization |
| [permission](#permission) | optOut | 7001-7999 | tenant_role_resources | apps/tenant/permission |
| [resource](#resource) | optOut | 8001-8999 | tenant_permissions | apps/tenant/resource |
| [dict](#dict) | optOut | 10001-10999 | dicts / dict_items | apps/reference/dict |
| [asset](#asset) | optOut | 9001-9999 | file_assets | apps/reference/asset |
| [config](#config) | optional | 18001-18999 | config_categories / config_items / config_visibility | apps/reference/config |
| [weixin](#weixin) | optional | 12001-12999 | — | apps/reference/weixin |
| [cms](#cms) | optional | 14001-14999 | posts | apps/cms |
| [flag](#flag) | optional | 13001-13999 | frames / spaces / avatars | apps/flag |

> `alwaysOn` = 启动必需，无法关闭（在 `framework/pkg/config/config.go` 中硬编码）
> `optOut` = 默认启用，写 `module:` 时视为白名单，需要显式列出来
> `optional` = 默认不启用，需要在 `cfg.Module` 显式列出

## 配置示例

```yaml
# config/config.yaml
module: []                    # 留空 = 启用 alwaysOn + 全部 optOut（11 个），不启用 optional
# 或
module:
  - user
  - role
  - menu
  - organization
  - permission
  - resource
  - asset
  - dict
  # optional 不列就不开
  # - config
  # - sys_user
  # - sys_role
  # - sys_menu
  # - sys_permission
  # - weixin
  # - cms
  # - flag
```

`alwaysOn` 的 `system` / `auth` / `tenants` **永远会加入**，即使从 module 列表里删了也会自动加回去。

---

## system

**职责**：health check + 运维 cache 操作入口。
**路由**（前缀 `/api/v1`）：

| Method | Path | Auth | Spec | Handler |
|---|---|---|---|---|---|
| GET | `/health` | public | — | `Health` |
| GET | `/system/server-info` | tenant | `system:list` | `ServerInfo` |
| POST | `/system/clear-cache` | tenant | `system:update` | `ClearCache` |
| GET | `/system/cache/info` | tenant | `system:list` | `CacheInfo` |
| GET | `/system/cache/keys` | tenant | `system:list` | `GetCacheKeys` |
| GET | `/system/cache/value/*key` | tenant | `system:list` | `GetCacheValue` |
| DELETE | `/system/cache/keys/*key` | tenant | `system:update` | `DeleteCacheKey` |

**数据表**：无（纯服务）。

---

## auth

**职责**：账号、登录、注册、JWT 颁发与撤销。支持多身份登录（tenant-login / platform-login / select-tenant）。
**路由**：

| Method | Path | Auth | Handler |
|---|---|---|---|
| POST | `/auth/tenant-login` | public | `TenantLogin` |
| POST | `/auth/platform-login` | public | `PlatformLogin` |
| POST | `/auth/login-precheck` | public | `LoginPrecheck`（多身份预检） |
| POST | `/auth/select-tenant` | public | `SelectTenant`（多身份第二步） |
| POST | `/auth/register` | public | `Register` |
| POST | `/auth/refresh` | public | `Refresh` |
| POST | `/auth/logout` | protected | `Logout` |

**数据表**：

| 表 | 说明 |
|---|---|
| `accounts` | 全局账号（phone / email unique，password argon2id hash） |
| `auth_sessions` | 会话（account_id, token unique, expires_at） |

> `account_auths` / `account_roles` / `user_codes` 已在 Phase 0023 中 **drop**。平台角色现在通过 `sys_users + sys_user_roles + sys_roles` 管理。

**跨模块依赖**：写 `AppContext.AccountRepo` + `AppContext.AccountAuthRepo`。
**关键约束**：`accounts` 表**不受 RLS 限制**（全局唯一）。

---

## tenants (alwaysOn)

**职责**：租户 CRUD。**唯一 alwaysOn 平台管理模块**，强制 `super_admin` 平台角色。
**路由**（全部位于 `/api/v1/platform/tenants`，`RequirePlatformRole("super_admin")` + `Require(ResTenant.*)` 双层守卫）：

| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/platform/tenants` | `tenant:list` | `List` |
| GET | `/platform/tenants/:id` | `tenant:list` | `Get` |
| POST | `/platform/tenants` | `tenant:create` | `Create` |
| PUT | `/platform/tenants/:id` | `tenant:update` | `Update` |
| PUT | `/platform/tenants/:id/status` | `tenant:update` | `UpdateStatus` |
| DELETE | `/platform/tenants/:id` | `tenant:delete` | `Delete`（软删） |
| POST | `/platform/tenants/:id/purge` | `tenant:delete` | `Purge`（硬删） |

**数据表**：`tenants`（不受 RLS）。
**跨模块依赖**：写 `AppContext.TenantRepo`。
**首次安装**：通过 `.env` 环境变量注入 bootstrap 凭据自动创建初始租户和管理员。

---

## sys_user (optional, Phase 0023+)

**职责**：平台域用户身份管理。
**路由**（`/api/v1/platform/sys-users`，`RequirePlatformRole("super_admin")`）：

| Method | Path | Handler |
|---|---|---|
| GET | `/platform/sys-users` | `List` |
| POST | `/platform/sys-users` | `Create` |
| GET | `/platform/sys-users/:id` | `Get` |
| PUT | `/platform/sys-users/:id` | `Update` |
| PUT | `/platform/sys-users/:id/status` | `UpdateStatus` |
| DELETE | `/platform/sys-users/:id` | `Delete` |
| PUT | `/platform/sys-users/:id/roles` | `AssignRoles` |

**数据表**：`sys_users` / `sys_user_roles`（不受 RLS）。

---

## sys_role (optional, Phase 0023+)

**职责**：平台域角色管理（含 `super_admin`）。
**路由**（`/api/v1/platform/sys-roles`，`RequirePlatformRole("super_admin")`）：

| Method | Path | Handler |
|---|---|---|
| GET | `/platform/sys-roles` | `List` |
| POST | `/platform/sys-roles` | `Create` |
| GET | `/platform/sys-roles/:id` | `Get` |
| PUT | `/platform/sys-roles/:id` | `Update` |
| DELETE | `/platform/sys-roles/:id` | `Delete` |
| PUT | `/platform/sys-roles/:id/menus` | `AssignMenus` |
| PUT | `/platform/sys-roles/:id/permissions` | `AssignPermissions` |

**数据表**：`sys_roles` / `sys_role_menus` / `sys_role_permissions`（不受 RLS）。

---

## sys_menu (optional, Phase 0021+)

**职责**：平台级菜单管理。super_admin 跨租户维护。
**路由**（`/api/v1/platform/sys-menus`，`RequirePlatformRole("super_admin")`）：

| Method | Path | Handler |
|---|---|---|
| GET | `/platform/sys-menus` | `List` |
| GET | `/platform/sys-menus/tree` | `Tree` |
| GET | `/platform/sys-menus/:id` | `Get` |
| POST | `/platform/sys-menus` | `Create` |
| PUT | `/platform/sys-menus/:id` | `Update` |
| DELETE | `/platform/sys-menus/:id` | `Delete` |

**数据表**：`sys_menus`（不受 RLS）。
**关键约定**：所有写操作用 `db.RunInPlatformTx` 跳过 RLS。

---

## sys_permission (optional, Phase 0023+)

**职责**：平台域权限码管理。
**路由**（`/api/v1/platform/sys-permissions`，`RequirePlatformRole("super_admin")`）：

| Method | Path | Handler |
|---|---|---|
| GET | `/platform/sys-permissions` | `List` |
| POST | `/platform/sys-permissions` | `Create` |
| GET | `/platform/sys-permissions/:id` | `Get` |
| PUT | `/platform/sys-permissions/:id` | `Update` |
| DELETE | `/platform/sys-permissions/:id` | `Delete` |

**数据表**：`sys_permissions`（不受 RLS）。

---

## user

**职责**：租户内用户 CRUD + 当前用户信息。
**路由**：

| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/users` | `user:list` | `List` |
| POST | `/users` | `user:create` | `Create` |
| GET | `/users/:id` | `user:list` | `Get` |
| PUT | `/users/:id` | `user:update` | `Update` |
| PATCH | `/users/:id` | `user:update` | `Patch` |
| DELETE | `/users/:id` | `user:delete` | `Delete`（软删） |
| PUT | `/users/:id/status` | `user:update` | `UpdateStatus` |
| PUT | `/users/:id/org` | `user:update` | `UpdateOrg` |
| GET | `/user/profile` | — | `Profile`（当前用户） |
| POST | `/user/avatar` | — | `UploadAvatar` |
| PUT | `/user/profile` | — | `UpdateProfile` |

**数据表**：`tenant_users` / `tenant_user_roles`（受 RLS）。
**跨模块依赖**：写 `AppContext.UserRepo`；读 `AccountRepo()`。

---

## role

**职责**：角色 CRUD + 数据范围 + 角色-菜单/权限分配。
**路由**：

| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/roles` | `role:list` | `List` |
| GET | `/roles/:id` | `role:list` | `Get` |
| POST | `/roles` | `role:create` | `Create` |
| PUT | `/roles/:id` | `role:update` | `Update` |
| PATCH | `/roles/:id` | `role:update` | `Patch` |
| DELETE | `/roles/:id` | `role:delete` | `Delete` |
| GET | `/roles/:id/data-scopes` | `role:list` | `GetDataScopes` |
| PUT | `/roles/:id/data-scopes` | `role:update` | `UpdateDataScopes` |
| GET | `/roles/:id/menus` | `role:list` | `GetMenus` |
| PUT | `/roles/:id/menus` | `role:update` | `AssignMenus` |
| GET | `/roles/:id/permissions` | `role:list` | `GetPermissions` |
| POST | `/roles/:id/permissions` | `role:update` | `AssignPermissions` |

**数据表**：`tenant_roles` / `tenant_user_roles` / `tenant_role_menus` / `tenant_role_data_scopes` / `tenant_role_resources`。
**跨模块依赖**：写 `AppContext.RoleRepo`。

---

## menu

**职责**：租户内菜单树 CRUD（平台菜单由 `sys_menu` 模块管理）。
**路由**：

| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/menus/tree` | `menu:list` | `Tree` |
| GET | `/menus` | `menu:list` | `List` |
| GET | `/menus/:id` | `menu:list` | `Get` |
| POST | `/menus` | `menu:create` | `Create` |
| PUT | `/menus/:id` | `menu:update` | `Update` |
| DELETE | `/menus/:id` | `menu:delete` | `Delete` |

**数据表**：`tenant_menus` / `tenant_role_menus`（受 RLS）。

---

## organization

**职责**：组织架构树（支持递归 CTE + 物化路径）。
**路由**：

| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/organizations/tree` | `organization:list` | `Tree` |
| GET | `/organizations` | `organization:list` | `List` |
| GET | `/organizations/:id` | `organization:list` | `Get` |
| POST | `/organizations` | `organization:create` | `Create` |
| PUT | `/organizations/:id` | `organization:update` | `Update` |
| DELETE | `/organizations/:id` | `organization:delete` | `Delete` |

**数据表**：`tenant_organizations`（含 `parent_id` 递归 + `ancestors` 物化路径）。
**跨模块依赖**：写 `AppContext.OrgRepo`。

---

## permission

**职责**：角色-资源（按钮/API）分配。
**路由**：

| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/roles/:id/permissions` | `role:list` | `GetPermissions` |
| POST | `/roles/:id/permissions` | `role:update` | `AssignPermissions` |
| PUT | `/roles/:id/permissions` | `role:update` | `AssignPermissions`（幂等） |
| GET | `/roles/:id/resources` | `role:list` | `GetResources` |

**数据表**：`tenant_role_resources`（M:N join 表）。
**跨模块依赖**：写 `AppContext.PermRepo`。

---

## resource

**职责**：资源（按钮/API）CRUD + 当前用户的资源列表。
**路由**：

| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/resources` | `resource:list` | `List` |
| GET | `/resources/:id` | `resource:list` | `Get` |
| POST | `/resources` | `resource:create` | `Create` |
| PUT | `/resources/:id` | `resource:update` | `Update` |
| DELETE | `/resources/:id` | `resource:delete` | `Delete` |
| GET | `/resources/by-menu/:menu_id` | `resource:list` | `GetByMenu` |
| GET | `/resources/my` | — | `GetMyResources` |

**数据表**：`tenant_permissions`（0023.3 由 `resources` rename）。

---

## dict

**职责**：数据字典（主表 + items 子表，支持层级 + 租户覆盖）。
**路由**：

#### 业务路由（`/api/v1/dicts`）

| Method | Path | Spec |
|---|---|---|
| GET | `/dicts` | `dict:list` |
| GET | `/dicts/:id` | `dict:get` |
| POST | `/dicts` | `dict:create` |
| PUT | `/dicts/:id` | `dict:update` |
| DELETE | `/dicts/:id` | `dict:delete` |
| GET | `/dicts/:id/items` | `dict:list` |
| POST | `/dicts/:id/items` | `dict:update` |
| PUT | `/dicts/:id/items/:item_id` | `dict:update` |
| DELETE | `/dicts/:id/items/:item_id` | `dict:update` |
| GET | `/dicts/resolve` | `dict:get` |
| POST | `/dicts/resolve/batch` | `dict:get` |
| PUT | `/dicts/:id/items/:item_id/override` | `dict:update` |
| DELETE | `/dicts/:id/items/:item_id/override` | `dict:update` |

#### 平台路由（`/api/v1/platform/dicts`，强制 super_admin）

| Method | Path |
|---|---|
| GET | `/platform/dicts` |
| POST | `/platform/dicts` |
| GET | `/platform/dicts/:id` |
| PUT | `/platform/dicts/:id` |
| DELETE | `/platform/dicts/:id` |
| GET | `/platform/dicts/:id/items` |
| POST | `/platform/dicts/:id/items` |
| PUT | `/platform/dicts/:id/items/:item_id` |
| DELETE | `/platform/dicts/:id/items/:item_id` |
| GET | `/platform/dicts/:id/visibility` |
| POST | `/platform/dicts/:id/visibility` |
| DELETE | `/platform/dicts/:id/visibility/:tenant_id` |

**数据表**：`dicts` / `dict_items` / `dict_visibility`。
**JSONB**：`dicts.extend` / `dict_items.extend` 都是 `JSONB`（SQL 显式 `::jsonb` cast）。

---

## asset

**职责**：文件上传/删除（local 或 COS）。
**路由**：

| Method | Path | Spec | Handler |
|---|---|---|---|
| POST | `/asset/upload` | `asset:create` | `Upload` |
| DELETE | `/asset/:id` | `asset:delete` | `Delete` |

**数据表**：`file_assets`。

---

## config (optional, Phase 0022 重构)

**职责**：租户配置中心（分组 + 键值项），支持 Platform / Override / Visibility / Resolve 四层模型。

#### 业务消费（`/api/v1/configs`）

| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/configs` | `config:list` | `ListGroups` |
| GET | `/configs/:id` | `config:get` | `GetGroup` |
| GET | `/configs/:id/items` | `config:list` | `ListItemsByGroup` |
| POST | `/configs/:id/items/:item_id/override` | `config:update` | `UpsertOverride` |
| DELETE | `/configs/:id/items/:item_id/override` | `config:update` | `DeleteOverride` |
| GET | `/configs/resolve` | `config:list` | `Resolve`（`?code=`） |
| POST | `/configs/resolve/batch` | `config:list` | `ResolveBatch` |

#### 平台管理（`/api/v1/platform/configs`，强制 super_admin）

| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/platform/configs` | `config:list` | `ListGroups` |
| GET | `/platform/configs/:id` | `config:get` | `GetGroup` |
| POST | `/platform/configs` | `config:create` | `CreateGroup` |
| PUT | `/platform/configs/:id` | `config:update` | `UpdateGroup` |
| DELETE | `/platform/configs/:id` | `config:delete` | `DeleteGroup` |
| GET | `/platform/configs/:id/items` | `config:list` | `ListItems` |
| POST | `/platform/configs/:id/items` | `config:create` | `CreateItem` |
| PUT | `/platform/configs/:id/items/:item_id` | `config:update` | `UpdateItem` |
| DELETE | `/platform/configs/:id/items/:item_id` | `config:delete` | `DeleteItem` |
| GET | `/platform/configs/:id/visibility` | `config:list` | `ListVisibility` |
| POST | `/platform/configs/:id/visibility` | `config:update` | `UpsertVisibility` |
| DELETE | `/platform/configs/:id/visibility/:tenant_id` | `config:update` | `DeleteVisibility` |

#### 公开读（`/api/v1/public/configs`，无需 auth）

| Method | Path | Handler |
|---|---|---|
| GET | `/public/configs` | `GetPublic` |

**数据表**：`config_categories` / `config_items` / `config_visibility`（受 RLS）。
**JSONB**：`config_items.value` / `default_value` / `options` / `validation` 均为 `JSONB`。

---

## weixin

**职责**：微信小程序登录 + 手机号绑定。**无数据表**，纯配置 + handler。
**路由**：

| Method | Path | Auth | Handler |
|---|---|---|---|
| POST | `/weixin/login` | public | `Login`（code2Session） |
| POST | `/weixin/phone` | public | `GetPhoneNumber` |
| POST | `/weixin/bind-phone` | protected | `BindPhone` |

**配置**：从 `config/weixin.yaml` 或 `config.yaml` 的 `weixin` 段读取 `appid` / `secret`。
**跨模块依赖**：读 `AccountRepo` / `AccountAuthRepo` / `TenantRepo` / `UserRepo`。

---

## cms

**职责**：**示例 CMS**，展示 plugin.Reader 模式（extapi 调用）。
**路由**：

| Method | Path | Auth | Handler |
|---|---|---|---|
| GET | `/cms/me` | protected | `GetCurrentUser` |
| GET | `/cms/users` | protected | `ListUsers` |
| GET | `/cms/tenant` | protected | `GetTenant` |
| GET/POST/PUT/DELETE | `/cms/posts` | protected | CRUD |

**数据表**：`posts`（cms 自有，`migrations/cms.sql`）。

---

## flag

**职责**：**示例业务**——头像、相框、虚拟空间管理。展示完整的多表关联 + 数据范围应用。
**路由**（节选，完整版见 `apps/flag/doc/api.md`）：

| Method | Path | Auth | Spec | Handler |
|---|---|---|---|---|
| GET | `/flag/frames` | public | — | `ListFrames` |
| POST | `/flag/frames` | protected | `flag:create` | `CreateFrame` |
| GET | `/flag/my-avatars` | protected | `flag:list` | `ListMyAvatars`（DataScopeSelf） |
| POST | `/flag/generate` | protected | `flag:create` | `GenerateAvatar` |

**数据表**：`frames` / `frame_categories` / `spaces` / `avatars` / `avatar_categories`。
**JSONB**：`flag_frames.template_config` 是 `JSONB`。

---

## 附录：启动日志示例

正常启动会看到 19 个 `module X initialized`：

```
module auth initialized
module tenants initialized
module sys_user initialized
module sys_role initialized
module sys_menu initialized
module sys_permission initialized
module menu initialized
module organization initialized
module permission initialized
module resource initialized
module role initialized
module user initialized
module asset initialized
module dict initialized
module config initialized
module weixin initialized
module system initialized
module cms initialized
module flag initialized
```

如果某个 optional module 没在 `cfg.Module` 列表里，会打印：

```
module config not enabled (skip init)
```
