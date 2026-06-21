# 模块清单

> 当前共 **16 个 module**。按 `cfg.Module` 行为分 3 类：3 个 alwaysOn、8 个 optOut、5 个 optional。
>
> 文档版本：2026-06（config 重构 + platform_menu/platform_tenant 模块化后）

## 总览

| Name | 类型 | 错误码段 | 主要表 / 资源 | 默认 |
|---|---|---|---|---|
| [system](#system) | alwaysOn | 11001-11999 | — | ✅ |
| [auth](#auth) | alwaysOn | 1001-1999 | accounts / account_auths / account_roles / user_codes | ✅ |
| [platform_tenant](#platform_tenant) | alwaysOn | 3001-3999 | tenants | ✅ |
| [user](#user) | optOut | 2001-2999 | users / user_roles | ✅ |
| [role](#role) | optOut | 4001-4999 | roles | ✅ |
| [menu](#menu) | optOut | 5001-5999 | menus / role_menus | ✅ |
| [organization](#organization) | optOut | 6001-6999 | organizations | ✅ |
| [permission](#permission) | optOut | 7001-7999 | role_resources | ✅ |
| [resource](#resource) | optOut | 8001-8999 | resources | ✅ |
| [dict](#dict) | optOut | 10001-10999 | dicts / dict_items | ✅ |
| [asset](#asset) | optOut | 9001-9999 | file_assets | ✅ |
| [config](#config) | optional | 18001-18999 | config_groups / config_items / config_visibility | 显式 |
| [platform_menu](#platform_menu) | optional | 15001-15999 | menus（tenant_id=0 子集） | 显式 |
| [weixin](#weixin) | optional | 12001-12999 | — | 显式 |
| [cms](#cms) | optional | — (示例) | posts | 显式 |
| [flag](#flag) | optional | 13001-13999 | frames / spaces / avatars | 显式 |

> `alwaysOn` = 启动必需，无法关闭（在 [`framework/pkg/config/config.go`](../framework/pkg/config/config.go) 中硬编码）。
> `optOut` = 默认启用，写 `module:` 时视为白名单，需要显式列出来。
> `optional` = 默认不启用，需要在 `cfg.Module` 显式列出。

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
  # - platform_menu
  # - weixin
  # - cms
  # - flag
```

`alwaysOn` 的 `system` / `auth` / `platform_tenant` **永远会加载**，即使从 module 列表里删了也会自动加回去。

---

## system

**职责**：health check + 运维 cache 操作入口。

**路由**（前缀 `/api/v1`）：

| Method | Path | Auth | Spec | Handler |
|---|---|---|---|---|
| GET | `/health` | public | — | `Health` |
| GET | `/system/server-info` | protected | `system:list` | `ServerInfo` |
| POST | `/system/clear-cache` | protected | `system:update` | `ClearCache` |
| GET | `/system/cache/info` | protected | `system:list` | `CacheInfo` |
| GET | `/system/cache/keys` | protected | `system:list` | `GetCacheKeys` |
| GET | `/system/cache/value/*key` | protected | `system:list` | `GetCacheValue` |
| DELETE | `/system/cache/keys/*key` | protected | `system:update` | `DeleteCacheKey` |

**数据表**：无（纯服务）。

---

## auth

**职责**：账号、登录、注册、JWT 颁发与撤销。

**路由**：

| Method | Path | Auth | Handler |
|---|---|---|---|
| POST | `/auth/login` | public | `Login` |
| POST | `/auth/register` | public | `Register` |
| POST | `/auth/refresh` | public | `Refresh` |
| POST | `/auth/logout` | protected | `Logout` |

**数据表**：

| 表 | 说明 |
|---|---|
| `accounts` | 全局账号（username / phone / email + 密码 hash） |
| `account_auths` | 第三方授权（wechat / oauth 等） |
| `account_roles` | 平台级角色（`super_admin` 等） |
| `user_codes` | 验证码（短信 / 邮件） |

**跨模块依赖**：写 `AppContext.AccountRepo` + `AppContext.AccountAuthRepo`。

**关键约束**：`accounts` 表**不受 RLS 限制**（全局唯一），`users` 表受 RLS 限制（每租户隔离）。LoginIdentity 查询时需要在租户事务内 join。

---

## platform_tenant  ⭐ (alwaysOn, Phase 0020)

**职责**：租户 CRUD。**唯一 alwaysOn 平台管理模块**，强制 `super_admin` 平台角色。

**路由**（全部位于 `/api/v1/admin/platform-tenants`，`RequirePlatformRole("super_admin")` + `Require(ResTenant.*)` 双层守卫）：

| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/admin/platform-tenants` | `tenant:list` | `List` |
| GET | `/admin/platform-tenants/:id` | `tenant:list` | `Get` |
| POST | `/admin/platform-tenants` | `tenant:create` | `Create` |
| PUT | `/admin/platform-tenants/:id` | `tenant:update` | `Update` |
| PUT | `/admin/platform-tenants/:id/status` | `tenant:update` | `UpdateStatus` |
| DELETE | `/admin/platform-tenants/:id` | `tenant:delete` | `Delete`（软删） |
| POST | `/admin/platform-tenants/:id/purge` | `tenant:delete` | `Purge`（硬删） |

**数据表**：`tenants`（不受 RLS）。

**安全设计**：路由在 `adminGroup` 分组下用 `RequirePlatformRole("super_admin")` 短路所有非 super_admin 请求；super_admin 仍需满足资源权限码（`tenant:create` / `update` / `delete` / `list`）做细粒度校验。

**演进**：Phase 0020 之前位于 `apps/boot/tenant`，路由为 `/api/v1/tenants`；现统一到 `/api/v1/admin/platform-tenants`，错误码段沿用 3001-3999（与未来业务层 tenant 模块共用）。

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
| PUT | `/users/:id/status` | `user:update` | `UpdateStatus` |
| PUT | `/users/:id/org` | `user:update` | `UpdateOrg` |
| GET | `/user/profile` | — | `Profile` |
| POST | `/user/avatar` | — | `UploadAvatar` |
| PUT | `/user/profile` | — | `UpdateProfile` |

**数据表**：`users` / `user_roles`（受 RLS）。

---

## role

**职责**：角色 CRUD + 数据范围 + 角色-菜单分配。

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
| GET | `/roles/:id/resources` | `role:list` | `GetResources` |

**数据表**：`roles` / `user_roles` / `role_menus` / `role_resources`。

---

## menu

**职责**：租户内菜单树 CRUD（平台菜单由 `platform_menu` 模块管理）。

**路由**：

| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/menus/tree` | `menu:list` | `Tree` |
| GET | `/menus` | `menu:list` | `List` |
| GET | `/menus/:id` | `menu:list` | `Get` |
| POST | `/menus` | `menu:create` | `Create` |
| PUT | `/menus/:id` | `menu:update` | `Update` |
| DELETE | `/menus/:id` | `menu:delete` | `Delete` |

**数据表**：`menus` / `role_menus`（受 RLS）。

---

## platform_menu  ⭐ (optional, Phase 0021)

**职责**：平台级菜单管理。`menus` 表中 `tenant_id = 0` 的子集，由 super_admin 跨租户维护。

**路由**（全部位于 `/api/v1/admin/platform-menus`，group 级 `RequirePlatformRole("super_admin")`）：

| Method | Path | Handler |
|---|---|---|
| GET | `/admin/platform-menus` | `List` |
| GET | `/admin/platform-menus/tree` | `Tree` |
| GET | `/admin/platform-menus/:id` | `Get` |
| POST | `/admin/platform-menus` | `Create` |
| PUT | `/admin/platform-menus/:id` | `Update` |
| DELETE | `/admin/platform-menus/:id` | `Delete` |

**数据表**：`menus`（用 `WHERE tenant_id = 0` 过滤；强制 `platformTenantID = 0`）。

**关键约定**：

- 平台菜单与租户菜单共享 `menus` 表，靠 `tenant_id` 区分（`0` = 平台）
- 所有写操作走 `db.RunInPlatformTx` 跳过 RLS
- 模板参考 [`apps/admin/platform_menu/`](apps/admin/platform_menu/)（Phase 0021 后新加的命名约定样板）

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

**数据表**：`organizations`（含 `parent_id` 递归 + `ancestors` 物化路径）。

**DataScope 集成**：`DataScopeDept` / `DataScopeDeptAndBelow` 都用 `organizations` 表做递归 CTE / 物化路径，详见 [permission/scope.go](../framework/pkg/permission/scope.go)。

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

**数据表**：`role_resources`（M:N join 表）。

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

**数据表**：`resources`。

---

## dict

**职责**：数据字典（带 items 子表，支持层级）。

**路由**：

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
| GET | `/dicts/platform/:id/items` | `dict:list`（plat） | `ListPlatformItems` |
| POST | `/dicts/platform/:id/items` | `dict:create`（plat） | `CreatePlatformItem` |
| PUT | `/dicts/platform/:id/items/:item_id` | `dict:update`（plat） | `UpdatePlatformItem` |
| DELETE | `/dicts/platform/:id/items/:item_id` | `dict:delete`（plat） | `DeletePlatformItem` |
| GET | `/dicts/platform/:id/visibility` | `dict:list`（plat） | `ListVisibility` |
| POST | `/dicts/platform/:id/visibility` | `dict:update`（plat） | `UpsertVisibility` |
| DELETE | `/dicts/platform/:id/visibility/:tenant_id` | `dict:update`（plat） | `DeleteVisibility` |

**数据表**：`dicts` / `dict_items`。

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

**存储后端**：`cfg.storage.provider = "local"`（默认，写到 `./uploads/`）或 `"cos"`（腾讯云）。

---

## config  ⭐ (Phase 0022 重构)

**职责**：租户配置中心（分组 + 键值项），支持 **Platform / Override / Visibility / Resolve** 四层模型。

**路由**：

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

#### 平台管理（`/api/v1/configs/platform`，强制 `super_admin`）

| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/configs/platform` | `config:list` | `ListGroups` |
| GET | `/configs/platform/:id` | `config:get` | `GetGroup` |
| POST | `/configs/platform` | `config:create` | `CreateGroup` |
| PUT | `/configs/platform/:id` | `config:update` | `UpdateGroup` |
| DELETE | `/configs/platform/:id` | `config:delete` | `DeleteGroup` |
| GET | `/configs/platform/:id/items` | `config:list` | `ListItems` |
| POST | `/configs/platform/:id/items` | `config:create` | `CreateItem` |
| PUT | `/configs/platform/:id/items/:item_id` | `config:update` | `UpdateItem` |
| DELETE | `/configs/platform/:id/items/:item_id` | `config:delete` | `DeleteItem` |
| GET | `/configs/platform/:id/visibility` | `config:list` | `ListVisibility` |
| POST | `/configs/platform/:id/visibility` | `config:update` | `UpsertVisibility` |
| DELETE | `/configs/platform/:id/visibility/:tenant_id` | `config:update` | `DeleteVisibility` |

#### 公开读（`/api/v1/public/configs`，无需 auth）

| Method | Path | Handler |
|---|---|---|
| GET | `/public/configs` | `GetPublic` |

**数据表**：`config_groups` / `config_items` / `config_visibility`（受 RLS）。

**Scope 模型**：

- `scope = 'platform'` → 全平台共享，存于 `config_items` 但 `platform_item_id IS NULL` 且 `tenant_id = 0`
- `scope = 'tenant'` → 租户私有，存于 `config_items` 且 `tenant_id = <租户>`
- `is_override = TRUE` → 租户覆盖平台 item
- `platform_item_id` → 指向被覆盖的平台 item

**Visibility 模型**（`config_visibility` 表）：

- `access = 'invisible'` → 租户不可见
- `access = 'readonly'` → 租户可读不可改
- `access = 'editable'` → 租户可读可改（默认）
- 可针对单租户 `tenant_id`，或用 `*`（通配）

**Resolve 算法**（`GET /configs/resolve?code=site`）：

1. 取 platform group（按 code 找到 `tenant_id = 0` 的 group）
2. 检查当前租户在 `config_visibility` 表的 access
3. 合并 platform items + 租户 overrides + visibility 规则
4. 返回扁平 `map[string]interface{}` 形式

**JSONB**：`config_items.value` / `default_value` / `options` / `validation` 均为 `JSONB`（SQL 显式 `::jsonb` cast）。

**审计**：`POST/PUT/DELETE` 在 `db.RunInPlatformTx` / `db.RunInTenantTx` 内调 `audit.Log` 写 `db_logs`（事务内）。审计失败不回滚业务。

**迁移**：`migrations/config_alignment.sql`（Phase 0022 新加）做 `ALTER TABLE ... ADD COLUMN IF NOT EXISTS scope / visibility` + 新建 `config_visibility` 表，幂等。

---

## cms

**职责**：**示例 CMS**，展示 extapi 模式（cms module 自己不连 RBAC 表，而是通过 `extapi.UserFacade` / `TenantFacade` 跨 module 查询）。

**路由**：

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

**数据表**：`posts`（cms 自有，`migrations/cms.sql`）。

**设计意图**：展示外部 module 如何只通过 `extapi.Provider` 调用平台能力（查用户、查租户），而不直接 import `apps/rbac/user` 等模块。

---

## flag

**职责**：**示例业务**——头像、相框、虚拟空间管理。展示完整的多表关联 + 数据范围应用。

**路由**（节选，完整见 [`apps/flag/doc/api.md`](../apps/flag/doc/api.md)）：

| Method | Path | Auth | Spec | Handler |
|---|---|---|---|---|
| GET | `/flag/frames` | public | — | `ListFrames` |
| GET | `/flag/frames/:id` | public | — | `GetFrame` |
| POST | `/flag/frames` | protected | `flag:create` | `CreateFrame` |
| PUT | `/flag/frames/:id` | protected | `flag:update` | `UpdateFrame` |
| DELETE | `/flag/frames/:id` | protected | `flag:delete` | `DeleteFrame` |
| GET | `/flag/spaces` | protected | `flag:list` | `ListSpaces` |
| POST | `/flag/spaces` | protected | `flag:create` | `CreateSpace` |
| GET | `/flag/avatars` | public | — | `ListAvatars` |
| POST | `/flag/generate` | protected | `flag:create` | `GenerateAvatar` |
| GET | `/flag/my-avatars` | protected | `flag:list` | `ListMyAvatars`（自动 DataScopeSelf） |

**数据表**：`frames` / `frame_categories` / `spaces` / `avatars` / `avatar_categories`（见 [migrations/flag.sql](../migrations/flag.sql)）。

**DataScope 集成**：`ListMyAvatars` 用 `DataScopeSelf` 自动只返回 `creator_id = 当前 userID` 的记录。

**JSONB**：`flag_frames.template_config` 是 `JSONB`（SQL 显式 `::jsonb` cast）。

---

## weixin

**职责**：微信小程序登录 + 手机号绑定。**无数据表**，纯配置 + handler。

**路由**：

| Method | Path | Auth | Handler |
|---|---|---|---|
| GET | `/weixin/ping` | public | `Ping` |
| POST | `/weixin/login` | public | `Login`（code2Session） |
| POST | `/weixin/phone` | public | `GetPhoneNumber` |
| POST | `/weixin/bind-phone` | protected | `BindPhone` |

**配置**：从 `config/weixin.yaml` 读 `appid` / `secret` / `token` / `encoding_aes_key`。

**跨模块依赖**：从 `AppContext.AccountRepo` / `AccountAuthRepo` / `TenantRepo` / `RoleRepo` / `UserRepo` 读——微信登录本质是把 `code → openid → 绑定到已有的 account/user`。

---

## 附录：启动日志示例

正常启动会看到 16 条 `module X initialized`：

```
2026/06/21 module auth initialized
2026/06/21 module platform_tenant initialized
2026/06/21 module platform_menu initialized
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

如果某 optional module 没在 `cfg.Module` 列表里，会打：

```
2026/06/21 module config registered but not enabled (skip)
```

`alwaysOn` 模块不会被 skip，无论 `cfg.Module` 怎么配都会加载。