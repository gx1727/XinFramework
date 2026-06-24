# 模块清单

> 当前�?**19 �?module**。按 `cfg.Module` 行为�?3 类：3 �?alwaysOn�? �?optOut�? �?optional�?>
> 文档版本�?026-06（config 重构 + sys_menu/platform_tenant 模块化后�?
## 总览

| Name | 类型 | 错误码段 | 主要�?/ 资源 | 默认 |
|---|---|---|---|---|
| [system](#system) | alwaysOn | 11001-11999 | �?| �?|
| [auth](#auth) | alwaysOn | 1001-1999 | accounts / auth_sessions | �?|
| [platform_tenant](#platform_tenant) | alwaysOn | 3001-3999 | tenants | �?|
| [user](#user) | optOut | 2001-2999 | tenant_users / tenant_user_roles | �?|
| [role](#role) | optOut | 4001-4999 | tenant_roles / tenant_role_data_scopes / tenant_user_roles / tenant_role_menus / tenant_tenant_role_resources | �?|
| [menu](#menu) | optOut | 5001-5999 | tenant_menus / tenant_role_menus | �?|
| [organization](#organization) | optOut | 6001-6999 | tenant_organizations | �?|
| [permission](#permission) | optOut | 7001-7999 | tenant_role_resources | �?|
| [resource](#resource) | optOut | 8001-8999 | tenant_permissions | �?|
| [dict](#dict) | optOut | 10001-10999 | dicts / dict_items | �?|
| [asset](#asset) | optOut | 9001-9999 | file_assets | �?|
| [config](#config) | optional | 18001-18999 | config_categories / config_items / config_visibility | 显式 |
| [sys_menu](#sys_menu) | optional | 15001-15999 | sys_menus / sys_role_menus | 显式 |
| [weixin](#weixin) | optional | 12001-12999 | �?| 显式 |
| [cms](#cms) | optional | �?(示例) | posts | 显式 |
| [flag](#flag) | optional | 13001-13999 | frames / spaces / avatars | 显式 |

> `alwaysOn` = 启动必需，无法关闭（�?[`framework/pkg/config/config.go`](../framework/pkg/config/config.go) 中硬编码）�?> `optOut` = 默认启用，写 `module:` 时视为白名单，需要显式列出来�?> `optional` = 默认不启用，需要在 `cfg.Module` 显式列出�?
## 配置示例

```yaml
# config/config.yaml
module: []                    # 留空 = 启用 alwaysOn + 全部 optOut�?1 个），不启用 optional
# �?module:
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
  # - sys_menu
  # - weixin
  # - cms
  # - flag
```

`alwaysOn` �?`system` / `auth` / `platform_tenant` **永远会加�?*，即使从 module 列表里删了也会自动加回去�?
---

## system

**职责**：health check + 运维 cache 操作入口�?
**路由**（前缀 `/api/v1`）：

| Method | Path | Auth | Spec | Handler |
|---|---|---|---|---|
| GET | `/health` | public | �?| `Health` |
| GET | `/system/server-info` | protected | `system:list` | `ServerInfo` |
| POST | `/system/clear-cache` | protected | `system:update` | `ClearCache` |
| GET | `/system/cache/info` | protected | `system:list` | `CacheInfo` |
| GET | `/system/cache/keys` | protected | `system:list` | `GetCacheKeys` |
| GET | `/system/cache/value/*key` | protected | `system:list` | `GetCacheValue` |
| DELETE | `/system/cache/keys/*key` | protected | `system:update` | `DeleteCacheKey` |

**数据�?*：无（纯服务）�?
---

## auth

**职责**：账号、登录、注册、JWT 颁发与撤销�?
**路由**�?
| Method | Path | Auth | Handler |
|---|---|---|---|
| POST | `/auth/tenant-login` | public | `Login` |
| POST | `/auth/register` | public | `Register` |
| POST | `/auth/refresh` | public | `Refresh` |
| POST | `/auth/logout` | protected | `Logout` |

**数据�?*�?
| �?| 说明 |
|---|---|
| `accounts` | 全局账号（username / phone / email + 密码 hash�?|
| `account_auths` | 第三方授权（wechat / oauth 等） |
| `account_roles` | 平台级角色（`super_admin` 等） |
| `user_codes` | 验证码（短信 / 邮件�?|

**跨模块依�?*：写 `AppContext.AccountRepo` + `AppContext.AccountAuthRepo`�?
**关键约束**：`accounts` �?*不受 RLS 限制**（全局唯一），`users` 表受 RLS 限制（每租户隔离）。LoginIdentity 查询时需要在租户事务�?join�?
---

## platform_tenant  �?(alwaysOn, Phase 0020)

**职责**：租�?CRUD�?*唯一 alwaysOn 平台管理模块**，强�?`super_admin` 平台角色�?
**路由**（全部位�?`/api/v1/platform/tenants`，`RequirePlatformRole("super_admin")` + `Require(ResTenant.*)` 双层守卫）：

| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/platform/tenants` | `tenant:list` | `List` |
| GET | `/platform/tenants/:id` | `tenant:list` | `Get` |
| POST | `/platform/tenants` | `tenant:create` | `Create` |
| PUT | `/platform/tenants/:id` | `tenant:update` | `Update` |
| PUT | `/platform/tenants/:id/status` | `tenant:update` | `UpdateStatus` |
| DELETE | `/platform/tenants/:id` | `tenant:delete` | `Delete`（软删） |
| POST | `/platform/tenants/:id/purge` | `tenant:delete` | `Purge`（硬删） |

**数据�?*：`tenants`（不�?RLS）�?
**安全设计**：group �?`RequirePlatformRole("super_admin")` 短路所有非 super_admin 请求；super_admin 仍需满足资源权限码（`tenant:create` / `update` / `delete` / `list`）做细粒度校验�?
**演进**：Phase 0020 之前位于 `apps/boot/tenant`，路由为 `/api/v1/tenants`；Phase 0022 迁到 `/api/v1/platform/tenants`；路由重构后迁到 `/api/v1/platform/tenants`。错误码段沿�?3001-3999（与未来业务�?tenant 模块共用）�?
---

## user

**职责**：租户内用户 CRUD + 当前用户信息�?
**路由**�?
| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/users` | `user:list` | `List` |
| POST | `/users` | `user:create` | `Create` |
| GET | `/users/:id` | `user:list` | `Get` |
| PUT | `/users/:id` | `user:update` | `Update` |
| PATCH | `/users/:id` | `user:update` | `Patch` |
| PUT | `/users/:id/status` | `user:update` | `UpdateStatus` |
| PUT | `/users/:id/org` | `user:update` | `UpdateOrg` |
| GET | `/user/profile` | �?| `Profile` |
| POST | `/user/avatar` | �?| `UploadAvatar` |
| PUT | `/user/profile` | �?| `UpdateProfile` |

**数据�?*：`users` / `user_roles`（受 RLS）�?
---

## role

**职责**：角�?CRUD + 数据范围 + 角色-菜单分配�?
**路由**�?
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
| GET | `/roles/:id/resources` | `role:list` | `GetResources` |

**数据�?*：`tenant_roles` / `tenant_user_roles` / `tenant_role_menus` / `tenant_role_resources` / `tenant_role_data_scopes`�?
---

## menu

**职责**：租户内菜单�?CRUD（平台菜单由 `sys_menu` 模块管理）�?
**路由**�?
| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/menus/tree` | `menu:list` | `Tree` |
| GET | `/menus` | `menu:list` | `List` |
| GET | `/menus/:id` | `menu:list` | `Get` |
| POST | `/menus` | `menu:create` | `Create` |
| PUT | `/menus/:id` | `menu:update` | `Update` |
| DELETE | `/menus/:id` | `menu:delete` | `Delete` |

**数据�?*：`menus` / `role_menus`（受 RLS）�?
---

## sys_menu  �?(optional, Phase 0021)

**职责**：平台级菜单管理。`menus` 表中 `tenant_id = 0` 的子集，�?super_admin 跨租户维护�?
**路由**（全部位�?`/api/v1/sys_menus`，group �?`RequirePlatformRole("super_admin")`）：

| Method | Path | Handler |
|---|---|---|
| GET | `/sys_menus` | `List` |
| GET | `/sys_menus/tree` | `Tree` |
| GET | `/sys_menus/:id` | `Get` |
| POST | `/sys_menus` | `Create` |
| PUT | `/sys_menus/:id` | `Update` |
| DELETE | `/sys_menus/:id` | `Delete` |

**数据�?*：`menus`（用 `WHERE tenant_id = 0` 过滤；强�?`platformTenantID = 0`）�?
**关键约定**�?
- 平台菜单与租户菜单共�?`menus` 表，�?`tenant_id` 区分（`0` = 平台�?- 所有写操作�?`db.RunInPlatformTx` 跳过 RLS
- 模板参�?[`apps/sys_menu/`](apps/sys_menu/)（Phase 0021 后新加的命名约定样板�?
---

## organization

**职责**：组织架构树（支持递归 CTE + 物化路径）�?
**路由**�?
| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/organizations/tree` | `organization:list` | `Tree` |
| GET | `/organizations` | `organization:list` | `List` |
| GET | `/organizations/:id` | `organization:list` | `Get` |
| POST | `/organizations` | `organization:create` | `Create` |
| PUT | `/organizations/:id` | `organization:update` | `Update` |
| DELETE | `/organizations/:id` | `organization:delete` | `Delete` |

**数据�?*：`organizations`（含 `parent_id` 递归 + `ancestors` 物化路径）�?
**DataScope 集成**：`DataScopeDept` / `DataScopeDeptAndBelow` 都用 `organizations` 表做递归 CTE / 物化路径，详�?[permission/scope.go](../framework/pkg/permission/scope.go)�?
---

## permission

**职责**：角�?资源（按�?API）分配�?
**路由**�?
| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/roles/:id/permissions` | `role:list` | `GetPermissions` |
| POST | `/roles/:id/permissions` | `role:update` | `AssignPermissions` |
| PUT | `/roles/:id/permissions` | `role:update` | `AssignPermissions`（幂等） |
| GET | `/roles/:id/resources` | `role:list` | `GetResources` |

**数据�?*：`role_resources`（M:N join 表）�?
---

## resource

**职责**：资源（按钮/API）CRUD + 当前用户的资源列表�?
**路由**�?
| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/resources` | `resource:list` | `List` |
| GET | `/resources/:id` | `resource:list` | `Get` |
| POST | `/resources` | `resource:create` | `Create` |
| PUT | `/resources/:id` | `resource:update` | `Update` |
| DELETE | `/resources/:id` | `resource:delete` | `Delete` |
| GET | `/resources/by-menu/:menu_id` | `resource:list` | `GetByMenu` |
| GET | `/resources/my` | �?| `GetMyResources` |

**数据�?*：`tenant_permissions`����`resources`��0023.3 rename���?
---

## dict

**职责**：数据字典（�?items 子表，支持层级）�?
**路由**�?
| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/dicts` | `dict:list` | `List` |
| GET | `/dicts/:id` | `dict:get` | `Get` |
| POST | `/dicts` | `dict:create` | `Create` |
| PUT | `/dicts/:id` | `dict:update` | `Update` |
| DELETE | `/dicts/:id` | `dict:delete` | `Delete` |
| GET | `/dicts/:id/items` | `dict:list` | `ListItems` |
| POST | `/dicts/:id/items` | `dict:update` | `CreateItem` |
| PUT | `/dicts/:id/items/:item_id` | `dict:update` | `UpdateItem` |
| DELETE | `/dicts/:id/items/:item_id` | `dict:update` | `DeleteItem` |
| GET | `/dicts/resolve` | `dict:get` | `Resolve` |
| POST | `/dicts/resolve/batch` | `dict:get` | `ResolveBatch` |
| PUT | `/dicts/:id/items/:item_id/override` | `dict:update` | `UpsertOverride` |
| DELETE | `/dicts/:id/items/:item_id/override` | `dict:update` | `DeleteOverride` |
| GET | `/platform/dicts/:id/items` | `dict:list`（plat�?| `ListPlatformItems` |
| POST | `/platform/dicts/:id/items` | `dict:create`（plat�?| `CreatePlatformItem` |
| PUT | `/platform/dicts/:id/items/:item_id` | `dict:update`（plat�?| `UpdatePlatformItem` |
| DELETE | `/platform/dicts/:id/items/:item_id` | `dict:delete`（plat�?| `DeletePlatformItem` |
| GET | `/platform/dicts/:id/visibility` | `dict:list`（plat�?| `ListVisibility` |
| POST | `/platform/dicts/:id/visibility` | `dict:update`（plat�?| `UpsertVisibility` |
| DELETE | `/platform/dicts/:id/visibility/:tenant_id` | `dict:update`（plat�?| `DeleteVisibility` |

**数据�?*：`dicts` / `dict_items`�?
**JSONB**：`dicts.extend` / `dict_items.extend` 都是 `JSONB`（SQL 显式 `::jsonb` cast）�?
---

## asset

**职责**：文件上�?删除（local �?COS）�?
**路由**�?
| Method | Path | Spec | Handler |
|---|---|---|---|
| POST | `/asset/upload` | `asset:create` | `Upload` |
| DELETE | `/asset/:id` | `asset:delete` | `Delete` |

**数据�?*：`file_assets`�?
**存储后端**：`cfg.storage.provider = "local"`（默认，写到 `./uploads/`）或 `"cos"`（腾讯云）�?
---

## config  �?(Phase 0022 重构)

**职责**：租户配置中心（分组 + 键值项），支持 **Platform / Override / Visibility / Resolve** 四层模型�?
**路由**�?
#### 业务消费（`/api/v1/configs`�?
| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/configs` | `config:list` | `ListGroups` |
| GET | `/configs/:id` | `config:get` | `GetGroup` |
| GET | `/configs/:id/items` | `config:list` | `ListItemsByGroup` |
| POST | `/configs/:id/items/:item_id/override` | `config:update` | `UpsertOverride` |
| DELETE | `/configs/:id/items/:item_id/override` | `config:update` | `DeleteOverride` |
| GET | `/configs/resolve` | `config:list` | `Resolve`（`?code=`�?|
| POST | `/configs/resolve/batch` | `config:list` | `ResolveBatch` |

#### 平台管理（`/api/v1/platform/configs`，强�?`super_admin`�?
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

#### 公开读（`/api/v1/public/configs`，无需 auth�?
| Method | Path | Handler |
|---|---|---|
| GET | `/public/configs` | `GetPublic` |

**数据�?*：`config_groups` / `config_items` / `config_visibility`（受 RLS）�?
**Scope 模型**�?
- `scope = 'platform'` �?全平台共享，存于 `config_items` �?`platform_item_id IS NULL` �?`tenant_id = 0`
- `scope = 'tenant'` �?租户私有，存�?`config_items` �?`tenant_id = <租户>`
- `is_override = TRUE` �?租户覆盖平台 item
- `platform_item_id` �?指向被覆盖的平台 item

**Visibility 模型**（`config_visibility` 表）�?
- `access = 'invisible'` �?租户不可�?- `access = 'readonly'` �?租户可读不可�?- `access = 'editable'` �?租户可读可改（默认）
- 可针对单租户 `tenant_id`，或�?`*`（通配�?
**Resolve 算法**（`GET /configs/resolve?code=site`）：

1. �?platform group（按 code 找到 `tenant_id = 0` �?group�?2. 检查当前租户在 `config_visibility` 表的 access
3. 合并 platform items + 租户 overrides + visibility 规则
4. 返回扁平 `map[string]interface{}` 形式

**JSONB**：`config_items.value` / `default_value` / `options` / `validation` 均为 `JSONB`（SQL 显式 `::jsonb` cast）�?
**审计**：`POST/PUT/DELETE` �?`db.RunInPlatformTx` / `db.RunInTenantTx` 内调 `audit.Log` �?`db_logs`（事务内）。审计失败不回滚业务�?
**迁移**：`migrations/config_alignment.sql`（Phase 0022 新加）做 `ALTER TABLE ... ADD COLUMN IF NOT EXISTS scope / visibility` + 新建 `config_visibility` 表，幂等�?
---

## cms

**职责**：*示例 CMS**，展示 plugin.Reader 模式（cms module 自己不连 RBAC 表，而是通过 AppContext 拿 user/tenant repo）。
**路由**�?
| Method | Path | Auth | Handler |
|---|---|---|---|
| GET | `/cms/ping` | public | `Ping` |
| GET | `/cms/me` | protected | `GetCurrentUser` |
| GET | `/cms/users` | protected | `ListUsers` |
| GET | `/cms/tenant` | protected | `GetTenant` |
| GET | `/cms/posts` | protected | `ListPosts` |
| GET | `/cms/posts/:id` | protected | `GetPost` |
| POST | `/cms/posts` | protected | `CreatePost` |
| PUT | `/cms/posts/:id` | protected | `UpdatePost` |
| DELETE | `/cms/posts/:id` | protected | `DeletePost` |

**数据�?*：`posts`（cms 自有，`migrations/cms.sql`）�?
**设计意图**：展示外部 module 如何通过 AppContext（plugin.Reader）拿其他模块的 repo（user/tenant），而不直接 import `apps/tenant/user` 等模块。
---

## flag

**职责**�?*示例业务**——头像、相框、虚拟空间管理。展示完整的多表关联 + 数据范围应用�?
**路由**（节选，完整�?[`apps/flag/doc/api.md`](../apps/flag/doc/api.md)）：

| Method | Path | Auth | Spec | Handler |
|---|---|---|---|---|
| GET | `/flag/frames` | public | �?| `ListFrames` |
| GET | `/flag/frames/:id` | public | �?| `GetFrame` |
| POST | `/flag/frames` | protected | `flag:create` | `CreateFrame` |
| PUT | `/flag/frames/:id` | protected | `flag:update` | `UpdateFrame` |
| DELETE | `/flag/frames/:id` | protected | `flag:delete` | `DeleteFrame` |
| GET | `/flag/spaces` | protected | `flag:list` | `ListSpaces` |
| POST | `/flag/spaces` | protected | `flag:create` | `CreateSpace` |
| GET | `/flag/avatars` | public | �?| `ListAvatars` |
| POST | `/flag/generate` | protected | `flag:create` | `GenerateAvatar` |
| GET | `/flag/my-avatars` | protected | `flag:list` | `ListMyAvatars`（自�?DataScopeSelf�?|

**数据�?*：`frames` / `frame_categories` / `spaces` / `avatars` / `avatar_categories`（见 [migrations/flag.sql](../migrations/flag.sql)）�?
**DataScope 集成**：`ListMyAvatars` �?`DataScopeSelf` 自动只返�?`creator_id = 当前 userID` 的记录�?
**JSONB**：`flag_frames.template_config` �?`JSONB`（SQL 显式 `::jsonb` cast）�?
---

## weixin

**职责**：微信小程序登录 + 手机号绑定�?*无数据表**，纯配置 + handler�?
**路由**�?
| Method | Path | Auth | Handler |
|---|---|---|---|
| GET | `/weixin/ping` | public | `Ping` |
| POST | `/weixin/login` | public | `Login`（code2Session�?|
| POST | `/weixin/phone` | public | `GetPhoneNumber` |
| POST | `/weixin/bind-phone` | protected | `BindPhone` |

**配置**：从 `config/weixin.yaml` �?`appid` / `secret` / `token` / `encoding_aes_key`�?
**跨模块依�?*：从 `AppContext.AccountRepo` / `AccountAuthRepo` / `TenantRepo` / `RoleRepo` / `UserRepo` 读——微信登录本质是�?`code �?openid �?绑定到已有的 account/user`�?
---

## 附录：启动日志示�?
正常启动会看�?16 �?`module X initialized`�?
```
2026/06/23 module auth initialized
2026/06/21 module platform_tenant initialized
2026/06/21 module sys_menu initialized
2026/06/21 module menu initialized
2026/06/21 module organization initialized
2026/06/21 module permission initialized
2026/06/21 module resource initialized
2026/06/21 module role initialized
2026/06/21 module user initialized
2026/06/21 module asset initialized
2026/06/21 module dict initialized
2026/06/21 module config initialized
2026/06/21 module weixin initialized
2026/06/21 module system initialized
2026/06/21 module cms initialized
2026/06/21 module flag initialized
```

如果�?optional module 没在 `cfg.Module` 列表里，会打�?
```
2026/06/21 module config registered but not enabled (skip)
```

`alwaysOn` 模块不会�?skip，无�?`cfg.Module` 怎么配都会加载�

## 0023 Status Update (2026-06-23)

> modules.md has historical drift from Phase 0023. Below is the source-of-truth mapping for cross-reference.

### Table Renames (0023.3)

| Old | New | Domain |
|---|---|---|
| `users` | `tenant_users` | tenant |
| `roles` | `tenant_roles` | tenant |
| `organizations` | `tenant_organizations` | tenant |
| `user_roles` | `tenant_user_roles` | tenant |
| `role_menus` | `tenant_role_menus` | tenant |
| `role_resources` | `tenant_role_resources` | tenant |
| `role_data_scopes` | `tenant_role_data_scopes` | tenant |
| `resources` | `tenant_permissions` | tenant |
| `menus WHERE scope=tenant` | `tenant_menus` (no `scope` field) | tenant |
| `menus WHERE scope=platform` | `sys_menus` (independent table) | platform |
| `resources WHERE scope=platform` | `sys_permissions` | platform |
| `account_auths` | **dropped** (was never used) | - |
| `account_roles` | **dropped** (replaced by `sys_user_roles + sys_roles`) | - |
| `user_codes` | **dropped** (was never used) | - |
| `config_groups` | `config_categories` | mixed |

### Module Additions (0023.0+)

| Module | Path | Purpose |
|---|---|---|
| `sys_user` | apps/platform/sys_user | Platform user identity (sys_users) |
| `sys_role` | apps/platform/sys_role | Platform roles (incl super_admin) |
| `sys_menu` | apps/platform/sys_menu | Platform menus (sys_menus) -- replaces apps/platform/menu (deleted 0023.4) |
| `sys_permission` | apps/platform/sys_permission | Platform permission codes (sys_permissions) |

### Total Module Count

Before 0023: 16 modules
After 0023.4: **19 modules** (3 alwaysOn + 8 optOut + 8 optional)

### Code Path

- Platform SQL **must** go through `db.RunInPlatformTx` (no RLS by design)
- Tenant SQL **must** go through `db.RunInTenantTx` (RLS enforced)
- Login: `accounts.id` -> `sys_users.account_id` -> `sys_user_roles.role_id` -> `sys_roles.code = 'super_admin'`
- See `doc/database.md` for full schema details
- See `doc/refactor/0023-split-platform-tenant.md` for the complete plan
