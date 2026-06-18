# 模块清单

> 当前共 **14 个 module**。`alwaysOn` 的 3 个无法关闭,其余可在 `cfg.Module` 中通过白名单 / 全留两种模式控制。

## 总览

| Name | 类型 | 错误码段 | 数据表 | 默认 |
|---|---|---|---|---|
| [system](#system) | alwaysOn | 11001-11999 | — | ✅ |
| [auth](#auth) | alwaysOn | 1001-1999 | accounts / account_auths / account_roles / user_codes | ✅ |
| [tenant](#tenant) | alwaysOn | 3001-3999 | tenants | ✅ |
| [user](#user) | optOut | 2001-2999 | users / user_roles | ✅ |
| [role](#role) | optOut | 4001-4999 | roles | ✅ |
| [menu](#menu) | optOut | 5001-5999 | menus / role_menus | ✅ |
| [organization](#organization) | optOut | 6001-6999 | organizations | ✅ |
| [permission](#permission) | optOut | 7001-7999 | role_resources | ✅ |
| [resource](#resource) | optOut | 8001-8999 | resources / menus | ✅ |
| [dict](#dict) | optOut | 10001-10999 | dicts / dict_items | ✅ |
| [asset](#asset) | optOut | 9001-9999 | file_assets | ✅ |
| [cms](#cms) | reference | 11000+(示例) | — | optional |
| [flag](#flag) | reference | 13001-13999 | frames / frame_categories / spaces / avatars / avatar_categories | optional |
| [weixin](#weixin) | reference | 12001-12999 | (无,纯配置 + handler) | optional |

> `alwaysOn` = 启动必需,无法关闭([config/config.go](framework/pkg/config/config.go) 中硬编码)。
> `optOut` = 默认启用,但用户写 `module:` 时视为白名单,需要显式列出来。
> `optional` = 默认不启用,需要在 `cfg.Module` 显式列出。

## 配置示例

```yaml
# config/config.yaml
module: []                    # 留空 = 启用全部 optOut
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
  # - flag                     # optional 模块不列就不开
  # - cms
  # - weixin
```

`alwaysOn` 的 system / auth / tenant **永远会加载**,即使从 module 列表里删了也会自动加回去。

---

## system

**职责**:health check + 运维 cache 操作入口。

**路由**:

| Method | Path | Auth | Spec | Handler |
|---|---|---|---|---|
| GET | `/health` | public | — | `health` |
| GET | `/system/server-info` | protected | `system:list` | `ServerInfo` |
| POST | `/system/clear-cache` | protected | `system:update` | `ClearCache` |
| GET | `/system/cache/info` | protected | `system:list` | `CacheInfo` |
| GET | `/system/cache/keys` | protected | `system:list` | `GetCacheKeys` |
| GET | `/system/cache/value/*key` | protected | `system:list` | `GetCacheValue` |
| DELETE | `/system/cache/keys/*key` | protected | `system:update` | `DeleteCacheKey` |

**数据表**:无(纯服务)。

---

## auth

**职责**:账号、登录、注册、JWT 颁发与撤销。

**路由**:

| Method | Path | Auth | Spec | Handler |
|---|---|---|---|---|
| POST | `/auth/login` | public | — | `Login` |
| POST | `/auth/register` | public | — | `Register` |
| POST | `/auth/refresh` | public | — | `Refresh` |
| POST | `/auth/logout` | protected | — | `Logout` |

**数据表**:

| 表 | 说明 |
|---|---|
| `accounts` | 全局账号(username/phone/email + 密码) |
| `account_auths` | 第三方授权(wechat / oauth 等) |
| `account_roles` | 平台级角色(`super_admin` 等) |
| `user_codes` | 验证码(短信 / 邮件) |

**跨模块依赖**:写 `AppContext.AccountRepo` + `AppContext.AccountAuthRepo`。

**关键约束**:accounts 表**不受 RLS 限制**(全局唯一),users 表受 RLS 限制(每租户隔离)。LoginIdentity 查询时需要在租户事务内 join。

---

## tenant

**职责**:租户 CRUD。这是**唯一必须挂平台角色守卫**的模块。

**路由**(全部 protected,且额外要求 `super_admin` 平台角色):

| Method | Path | Spec | Handler |
|---|---|---|---|
| POST | `/tenants` | `tenant:create` | `Create` |
| PUT | `/tenants/:id` | `tenant:update` | `Update` |
| PUT | `/tenants/:id/status` | `tenant:update` | `UpdateStatus` |
| DELETE | `/tenants/:id` | `tenant:delete` | `Delete`(软删) |
| POST | `/tenants/:id/purge` | `tenant:delete` | `Purge`(硬删) |
| GET | `/tenants/:id` | `tenant:list` | `Get` |
| GET | `/tenants` | `tenant:list` | `List` |

**数据表**:`tenants`(不受 RLS)。

**安全设计**:租户管理属于**跨租户特权**,`RequirePlatformRole("super_admin")` 在分组级别强制要求,普通租户内 admin 仅凭资源权限码也无法访问。

**跨模块依赖**:写 `AppContext.TenantRepo`,被 `ext_impl` 的 `TenantFacade` 读取(给 CMS 等外部模块提供数据)。

---

## user

**职责**:租户内用户 CRUD + 当前用户信息。

**路由**:

| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/users` | `user:list` | `List` |
| POST | `/users` | `user:create` | `Create` |
| GET | `/users/:id` | `user:list` | `Get` |
| PUT | `/users/:id` | `user:update` | `Update` |
| PATCH | `/users/:id` | `user:update` | `Patch`(部分更新) |
| PUT | `/users/:id/status` | `user:update` | `UpdateStatus` |
| PUT | `/users/:id/org` | `user:update` | `UpdateOrg` |
| GET | `/user/profile` | — | `Profile` |
| POST | `/user/avatar` | — | `UploadAvatar` |
| PUT | `/user/profile` | — | `UpdateProfile` |

**数据表**:`users` / `user_roles`(受 RLS)。

**跨模块依赖**:写 `AppContext.UserRepo`,被 `auth.Login` 读取(users join accounts),被 `ext_impl.UserFacade` 读取(CMS 跨服务查询)。

**特殊**:`/user/profile` 系列不带 RBAC spec —— 因为用户改自己 profile 不需要 RBAC 校验(已经登录了)。

---

## role

**职责**:角色 CRUD + 数据范围 + 角色-菜单分配。

**路由**:

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

**数据表**:`roles` / `user_roles` / `role_menus` / `role_resources`。

**跨模块依赖**:写 `AppContext.RoleRepo`,从 `AppContext.Authz()` 取 Authorization 用于 **失效缓存**(角色变更 → 关联用户权限缓存全失效)。

---

## menu

**职责**:菜单树 CRUD(纯 UI 导航数据,不参与 RBAC)。

**路由**:

| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/menus/tree` | `menu:list` | `Tree` |
| GET | `/menus` | `menu:list` | `List` |
| GET | `/menus/:id` | `menu:list` | `Get` |
| POST | `/menus` | `menu:create` | `Create` |
| PUT | `/menus/:id` | `menu:update` | `Update` |
| DELETE | `/menus/:id` | `menu:delete` | `Delete` |

**数据表**:`menus` / `role_menus`。

**注意**:`menus` 自身不含 RBAC 权限 —— 菜单是导航,真正决定按钮是否可点的是 `resources`(按钮/API)。`/roles/:id/menus` 在 role 模块里管理。

---

## organization

**职责**:组织架构树(支持递归 CTE)。

**路由**:

| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/organizations/tree` | `organization:list` | `Tree` |
| GET | `/organizations` | `organization:list` | `List` |
| GET | `/organizations/:id` | `organization:list` | `Get` |
| POST | `/organizations` | `organization:create` | `Create` |
| PUT | `/organizations/:id` | `organization:update` | `Update` |
| DELETE | `/organizations/:id` | `organization:delete` | `Delete` |

**数据表**:`organizations`(含 `parent_id` 递归 + `ancestors` 物化路径)。

**DataScope 集成**:`DataScopeDept` / `DataScopeDeptAndBelow` 都用 `organizations` 表做递归 CTE,见 [permission/scope.go](framework/pkg/permission/scope.go)。

---

## permission

**职责**:角色-资源(按钮/API) 分配。**菜单权限已迁移到 role 模块的 `/roles/:id/menus`**。

**路由**:

| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/roles/:id/permissions` | `role:list` | `GetPermissions` |
| POST | `/roles/:id/permissions` | `role:update` | `AssignPermissions` |
| PUT | `/roles/:id/permissions` | `role:update` | `AssignPermissions`(幂等) |
| GET | `/roles/:id/resources` | `role:list` | `GetResources` |

**数据表**:`role_resources`(M:N join 表)。

**跨模块依赖**:写 `AppContext.PermRepo`(实际是 RoleResourceRepository)。从 `AppContext.Authz()` 取 Authorization,角色权限变更 → 失效缓存。

---

## resource

**职责**:资源(按钮/API) CRUD + 当前用户的资源列表。

**路由**:

| Method | Path | Spec | Handler |
|---|---|---|---|
| GET | `/resources` | `resource:list` | `List` |
| GET | `/resources/:id` | `resource:list` | `Get` |
| POST | `/resources` | `resource:create` | `Create` |
| PUT | `/resources/:id` | `resource:update` | `Update` |
| DELETE | `/resources/:id` | `resource:delete` | `Delete` |
| GET | `/resources/by-menu/:menu_id` | `resource:list` | `GetByMenu` |
| GET | `/resources/my` | — | `GetMyResources`(返回当前用户可见的资源) |

**数据表**:`resources`。

**注意**:`/resources/my` 是给前端用的:列出当前用户能点哪些按钮,用来动态渲染 UI。不带 RBAC spec,因为它本身就是返回"我有啥权限"的接口。

---

## dict

**职责**:数据字典(可层级化,带 items 子表)。

**路由**:

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

**数据表**:`dicts` / `dict_items`。

---

## asset

**职责**:文件上传/删除(local 或 COS)。

**路由**:

| Method | Path | Spec | Handler |
|---|---|---|---|
| POST | `/asset/upload` | `asset:create` | `Upload` |
| DELETE | `/asset/:id` | `asset:delete` | `Delete` |

**数据表**:`file_assets`。

**存储后端**:`cfg.storage.provider = "local"` (默认, 写到 `./uploads/`) 或 `"cos"` (腾讯云)。

---

## cms

**职责**:**示例 CMS**,展示 extapi 模式(cms module 自己不连 DB,而是通过 extapi.UserFacade / TenantFacade 跨 module 查询)。

**路由**:

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

**数据表**:无(cms 模块自带 `migrations/cms.sql`,但表是 cms 自有)。

**设计意图**:展示外部 module 如何只通过 `extapi.Provider` 调用平台能力(查用户、查租户),而不直接 import `apps/rbac/user` 等模块。

---

## flag

**职责**:**示例业务** —— 头像、相框、虚拟空间管理。展示完整的多表关联 + 数据范围应用。

**路由**:

| Method | Path | Auth | Spec | Handler |
|---|---|---|---|---|
| GET | `/flag/frames` | public | — | `ListFrames` |
| GET | `/flag/frames/:id` | public | — | `GetFrame` |
| POST | `/flag/frames` | protected | `flag:create` | `CreateFrame` |
| PUT | `/flag/frames/:id` | protected | `flag:update` | `UpdateFrame` |
| DELETE | `/flag/frames/:id` | protected | `flag:delete` | `DeleteFrame` |
| GET | `/flag/frames-categories` | public | — | `ListFrameCategories` |
| POST | `/flag/frames-categories` | protected | `flag:create` | `CreateFrameCategory` |
| PUT | `/flag/frames-categories/:id` | protected | `flag:update` | `UpdateFrameCategory` |
| DELETE | `/flag/frames-categories/:id` | protected | `flag:delete` | `DeleteFrameCategory` |
| GET | `/flag/spaces/:code` | public | — | `GetSpaceByCode` |
| POST | `/flag/spaces` | protected | `flag:create` | `CreateSpace` |
| PUT | `/flag/spaces/:id` | protected | `flag:update` | `UpdateSpace` |
| DELETE | `/flag/spaces/:id` | protected | `flag:delete` | `DeleteSpace` |
| GET | `/flag/spaces` | protected | `flag:list` | `ListSpaces` |
| GET | `/flag/avatar-categories` | public | — | `ListAvatarCategories` |
| POST | `/flag/avatar-categories` | protected | `flag:create` | `CreateAvatarCategory` |
| PUT | `/flag/avatar-categories/:id` | protected | `flag:update` | `UpdateAvatarCategory` |
| DELETE | `/flag/avatar-categories/:id` | protected | `flag:delete` | `DeleteAvatarCategory` |
| GET | `/flag/avatars` | public | — | `ListAvatars` |
| GET | `/flag/avatars/:id` | public | — | `GetAvatar` |
| POST | `/flag/avatars` | protected | `flag:create` | `CreateAvatar` |
| PUT | `/flag/avatars/:id` | protected | `flag:update` | `UpdateAvatar` |
| DELETE | `/flag/avatars/:id` | protected | `flag:delete` | `DeleteAvatar` |
| POST | `/flag/generate` | protected | `flag:create` | `GenerateAvatar` |
| GET | `/flag/my-avatars` | protected | `flag:list` | `ListMyAvatars` |

**数据表**:`frames` / `frame_categories` / `spaces` / `avatars` / `avatar_categories`(见 [migrations/flag.sql](../migrations/flag.sql))。

**DataScope 集成**:`ListMyAvatars` 用 `DataScopeSelf` 自动只返回 `creator_id = 当前 userID` 的记录。

---

## weixin

**职责**:微信小程序登录 + 手机号绑定。**无数据表**,纯配置 + handler。

**路由**:

| Method | Path | Auth | Handler |
|---|---|---|---|
| GET | `/weixin/ping` | public | `ping` |
| POST | `/weixin/login` | public | `Login`(code2Session) |
| POST | `/weixin/phone` | public | `GetPhoneNumber` |
| POST | `/weixin/bind-phone` | protected | `BindPhone` |

**配置**:从 `config/weixin.yaml` 读 `appid` / `secret` / `token` / `encoding_aes_key`。

**跨模块依赖**:从 `AppContext.AccountRepo` / `AccountAuthRepo` / `TenantRepo` / `RoleRepo` / `UserRepo` 读 —— 微信登录本质是把 code → openid → 绑定到已有的 account/user。

---

## 附录:启动日志示例

正常启动会看到 14 条 `module X initialized`:

```
2026/06/18 08:33:06 module cms initialized
2026/06/18 08:33:06 module weixin initialized
2026/06/18 08:33:06 module tenant initialized
2026/06/18 08:33:06 module auth initialized
2026/06/18 08:33:06 module flag initialized
2026/06/18 08:33:06 module menu initialized
2026/06/18 08:33:06 module organization initialized
2026/06/18 08:33:06 module permission initialized
2026/06/18 08:33:06 module resource initialized
2026/06/18 08:33:06 module role initialized
2026/06/18 08:33:06 module asset initialized
2026/06/18 08:33:06 module user initialized
2026/06/18 08:33:06 module dict initialized
2026/06/18 08:33:06 module system initialized
```

如果某 module 没在 `cfg.Module` 列表里,会打:

```
2026/06/18 08:33:06 module cms registered but not enabled (skip)
```