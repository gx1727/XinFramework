# API 参考

> HTTP API 总览。具体某个模块的端点见各模块 doc 或 UI/AGENTS.md。

## 通用约定

### Base URL

所有 API 前缀：`/api/v1`

### 认证

```
Authorization: Bearer <jwt-token>
```

`/auth/login` 成功后返回 `token` 与 `refresh_token`。前端用 zustand authStore 持久化。

### 响应格式

```json
{
  "code": 0,
  "msg": "ok",
  "data": { /* 业务数据 */ }
}
```

### 错误码

| 段 | 范围 | 含义 |
| --- | --- | --- |
| 0 | 0 | 成功 |
| 4xxx | 4000-4999 | 客户端错误（参数、未授权、资源不存在） |
| 5xxx | 5000-5999 | 服务端错误（系统异常、数据库错误） |
| 9xxx | 9000-9999 | 业务错误（用户不存在、状态冲突、并发冲突） |

详见 [framework/pkg/resp/errors.go](file:///d:\work\xin\XinFramework\server\framework\pkg\resp\errors.go)。

### 分页

请求：

```
GET /api/v1/users?page=1&size=20&keyword=admin
```

响应 `data`：

```json
{
  "list": [...],
  "total": 100,
  "page": 1,
  "size": 20
}
```

---

## 认证 (`/auth`)

### POST `/auth/login`

**权限**：public

**Body**：

```json
{
  "tenant_code": "default",
  "account": "admin",
  "password": "secret"
}
```

**成功响应**：

```json
{
  "code": 0,
  "data": {
    "token": "eyJ...",
    "refresh_token": "eyJ...",
    "expires_in": 3600,
    "user": {
      "id": 1,
      "tenant_id": 1,
      "account": "admin",
      "name": "管理员",
      "email": "admin@example.com",
      "phone": "13800138000",
      "platform_roles": ["super_admin"]
    }
  }
}
```

**错误**：
- `4001` 未授权：账号或密码错误
- `4004` 资源不存在：租户编码不存在

### POST `/auth/logout`

**权限**：auth

撤销当前 session。客户端需要丢弃 token。

### POST `/auth/refresh`

**权限**：public

**Body**：

```json
{ "refresh_token": "eyJ..." }
```

返回新的 token / refresh_token。

### GET `/auth/me`

**权限**：auth

返回当前登录账号的完整信息（user + roles + permissions + data_scope + platform_roles）。

### POST `/auth/wxxcx/code2session`

**权限**：public

**Body**：

```json
{ "code": "081abc..." }
```

小程序登录专用。详见 [apps/boot/auth/handler.go](file:///d:\work\xin\XinFramework\server\apps\boot\auth\handler.go)。

---

## 租户 (`/tenants`)

**权限**：全部接口要求 `super_admin` 平台角色 + 对应资源权限。

| Method | Path | 资源权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/tenants` | `tenant:list` | 列表 |
| GET | `/tenants/:id` | `tenant:list` | 详情 |
| POST | `/tenants` | `tenant:create` | 新建 |
| PUT | `/tenants/:id` | `tenant:update` | 更新 |
| PUT | `/tenants/:id/status` | `tenant:update` | 启停 |
| DELETE | `/tenants/:id` | `tenant:delete` | 软删 |
| POST | `/tenants/:id/purge` | `tenant:delete` | 硬删 |

---

## 用户 (`/users`)

**权限**：租户内 RBAC，要求对应资源权限。

### GET `/users`

查询参数：

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `page` | int | 页码，默认 1 |
| `size` | int | 每页，默认 20 |
| `keyword` | string | 模糊匹配 code / name |
| `org_id` | uint | 按组织过滤 |
| `org_subtree` | 0/1 | 是否包含子组织 |
| `status` | int | 状态过滤 |

**自动应用 data_scope**：
- `All`（1）：返回全部
- `Custom`（2）：返回 org 列表内
- `Dept`（3）：返回本组织
- `DeptAndBelow`（4）：返回本组织 + 子组织
- `Self`（5）：仅返回自己

### POST `/users`

```json
{
  "code": "u001",
  "name": "张三",
  "email": "zhangsan@example.com",
  "phone": "13800138000",
  "account": "zhangsan",
  "password": "initial-pwd",
  "org_id": 5,
  "role_ids": [2]
}
```

同时创建 `account` 表记录（密码由后端 argon2 哈希）。

---

## 角色 (`/roles`)

| Method | Path | 说明 |
| --- | --- | --- |
| GET | `/roles` | 列表 |
| POST | `/roles` | 新建（带数据范围） |
| PUT | `/roles/:id` | 更新 |
| DELETE | `/roles/:id` | 删除 |
| GET | `/roles/:id/menus` | 已绑定的菜单 |
| PUT | `/roles/:id/menus` | 替换菜单绑定 |
| GET | `/roles/:id/users` | 已分配的用户 |
| PUT | `/roles/:id/users` | 替换用户绑定 |
| GET | `/roles/:id/data-scope` | 数据范围详情 |
| PUT | `/roles/:id/data-scope` | 更新数据范围 |

**数据范围**：

```json
{
  "type": 2,                // 1=All, 2=Custom, 3=Dept, 4=DeptAndBelow, 5=Self
  "org_ids": [3, 4]         // 仅 type=Custom 时使用
}
```

---

## 菜单 (`/menus`)

树形结构：

```json
{
  "id": 1,
  "parent_id": 0,
  "code": "dashboard",
  "title_i18n_key": "menu.dashboard",
  "path": "/dashboard",
  "icon": "LayoutDashboard",
  "sort": 1,
  "children": [...]
}
```

支持 `PUT /menus/sort` 批量更新排序。

---

## 资源 (`/resources`)

维护后端接口与按钮的权限码：

```json
{
  "code": "user:create",
  "resource_type": "api",     // api | button
  "name": "新建用户",
  "method": "POST",
  "path": "/api/v1/users"
}
```

---

## 角色-资源 (`/role-resources`)

```json
{
  "role_id": 2,
  "resource_codes": ["user:list", "user:create", "user:update"]
}
```

替换式绑定（全量覆盖）。

---

## 组织 (`/orgs`)

树形 CRUD：

| Method | Path | 说明 |
| --- | --- | --- |
| GET | `/orgs/tree` | 完整树 |
| POST | `/orgs` | 新建（指定 parent_id） |
| PUT | `/orgs/:id` | 更新 |
| PUT | `/orgs/:id/move` | 移动到新父节点 |
| DELETE | `/orgs/:id` | 删除（含子节点校验） |

---

## 字典 (`/dicts`)

| Method | Path | 说明 |
| --- | --- | --- |
| GET | `/dicts` | 字典分类列表 |
| POST | `/dicts` | 新建分类 |
| GET | `/dicts/:code` | 单个分类 + 项 |
| PUT | `/dicts/:id` | 更新分类 |
| DELETE | `/dicts/:id` | 删除分类（需先清空项） |
| GET | `/dicts/:code/items` | 字典项列表 |
| POST | `/dicts/:code/items` | 新建项 |
| PUT | `/dicts/:code/items/:id` | 更新项 |
| DELETE | `/dicts/:code/items/:id` | 删除项 |

`GET /api/v1/dicts/:code` 是 **public** 路由（前端可在未登录时取字典）。

---

## 附件 (`/attachments`)

| Method | Path | 说明 |
| --- | --- | --- |
| GET | `/attachments/:id` | 元信息 |
| GET | `/files/:id` | 下载（按 storage provider） |
| POST | `/attachments/upload` | 上传（multipart/form-data） |
| DELETE | `/attachments/:id` | 删除 |

---

## 微信 (`/weixin`)

| Method | Path | 说明 |
| --- | --- | --- |
| POST | `/weixin/wxxcx/code2session` | 小程序 code 换 session |

详见 [apps/boot/auth](file:///d:\work\xin\XinFramework\server\apps\boot\auth) 中的 weixin 集成部分。

---

## CMS (`/cms/*`)

详见 [apps/cms/handler.go](file:///d:\work\xin\XinFramework\server\apps\cms\handler.go)（路由注册），模块自带 API 文档。

---

## Flag (`/flag/*`)

详见 [apps/flag/doc/api.md](file:///d:\work\xin\XinFramework\server\apps\flag\doc\api.md)。

---

## 系统 (`/system`)

| Method | Path | Auth | 说明 |
| --- | --- | --- | --- |
| GET | `/system/health` | public | 健康检查（ping DB / cache） |
| GET | `/system/cache/stats` | auth | 缓存命中率 |
| POST | `/system/cache/refresh` | `system:cache` | 刷新所有缓存 |

---

## 错误响应示例

```json
{
  "code": 4001,
  "msg": "账号或密码错误",
  "data": null
}
```

```json
{
  "code": 9001,
  "msg": "用户不存在",
  "data": { "user_id": 999 }
}
```

```json
{
  "code": 9400,
  "msg": "状态冲突：用户已禁用",
  "data": { "current_status": 0, "target_status": 1 }
}
```

前端处理：

```typescript
// UI/src/api/client.ts
if (resp.code !== 0) {
  throw new ApiError(resp.code, resp.msg, resp.data)
}
```

具体错误码语义见 [resp/errors.go](file:///d:\work\xin\XinFramework\server\framework\pkg\resp\errors.go)。