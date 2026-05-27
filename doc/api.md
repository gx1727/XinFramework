# XinFramework API 接口文档

基础路径: `/api/v1`

---

## 1. 认证说明

### 1.1 认证方式

受保护接口需要在请求头中携带 JWT Token:
```
Authorization: Bearer <token>
```

### 1.2 路由组说明

| 路由组 | 前缀 | 认证要求 |
|--------|------|----------|
| `public` | `/api/v1` | 可选认证（Token 有效则注入上下文） |
| `protected` | `/api/v1` | 必须认证（有效 JWT Token） |

---

## 2. 响应格式

### 2.1 统一响应结构

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

### 2.2 错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 500 | 服务器错误 |
| 1001 | 用户名或密码错误 |
| 2002 | Token 已过期 |

---

## 3. 认证模块 (Auth)

### 公开接口

#### POST /auth/login - 用户登录

**请求参数:**
```json
{
  "account": "账号（手机号/邮箱）",
  "password": "密码",
  "tenant_id": 1
}
```

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "tenant_id": 1,
      "code": "admin",
      "role": "super_admin"
    }
  }
}
```

---

#### POST /auth/register - 用户注册

**请求参数:**
```json
{
  "account": "账号",
  "password": "密码（6-32字符）",
  "tenant_id": 0,
  "real_name": "真实姓名"
}
```

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "token": "...",
    "refresh_token": "...",
    "user": {
      "id": 1,
      "tenant_id": 1,
      "code": "user001",
      "role": "user"
    }
  }
}
```

---

#### POST /auth/refresh - 刷新Token

**请求参数:**
```json
{
  "refresh_token": "刷新令牌"
}
```

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "token": "新JWT令牌",
    "refresh_token": "新刷新令牌"
  }
}
```

---

### 受保护接口

#### POST /auth/logout - 用户登出

**请求头:** `Authorization: Bearer <token>`

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": { "ok": true }
}
```

---

## 4. 用户模块 (User)

### 受保护接口

#### GET /users - 用户列表

**请求参数:**
| 参数 | 类型 | 说明 |
|------|------|------|
| page | int | 页码（默认1） |
| size | int | 每页数量（默认20） |
| keyword | string | 搜索关键字（可选） |

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "list": [
      { "id": 1, "tenant_id": 1, "code": "admin", "real_name": "管理员", "status": 1 }
    ],
    "total": 100,
    "page": 1,
    "size": 20
  }
}
```

---

#### GET /users/:id - 获取用户

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "tenant_id": 1,
    "code": "admin",
    "real_name": "管理员",
    "phone": "13800138000",
    "email": "admin@example.com",
    "org_id": 1,
    "status": 1,
    "created_at": "2026-01-01T00:00:00Z"
  }
}
```

---

#### PUT /users/:id/status - 更新用户状态

**请求参数:**
```json
{
  "status": 1
}
```

---

#### GET /user/profile - 获取当前用户信息

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "tenant_id": 1,
    "code": "admin",
    "real_name": "管理员",
    "role": "super_admin"
  }
}
```

---

## 5. 租户模块 (Tenant)

### 受保护接口

#### GET /tenants - 租户列表

**请求参数:**
| 参数 | 类型 | 说明 |
|------|------|------|
| page | int | 页码 |
| size | int | 每页数量 |

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "list": [
      { "id": 1, "code": "default", "name": "默认租户", "status": 1 }
    ],
    "total": 10,
    "page": 1,
    "size": 20
  }
}
```

---

#### GET /tenants/:id - 获取租户

---

#### POST /tenants - 创建租户

**请求参数:**
```json
{
  "code": "new_tenant",
  "name": "新租户",
  "contact": "联系人",
  "phone": "13800138000"
}
```

---

#### PUT /tenants/:id - 更新租户

---

#### DELETE /tenants/:id - 删除租户

---

## 6. 菜单模块 (Menu)

### 受保护接口

#### GET /menus - 菜单列表

**请求参数:**
| 参数 | 类型 | 说明 |
|------|------|------|
| page | int | 页码 |
| size | int | 每页数量 |

---

#### GET /menus/tree - 菜单树

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": [
    {
      "id": 1,
      "name": "系统管理",
      "path": "/system",
      "icon": "setting",
      "children": [
        { "id": 2, "name": "用户管理", "path": "/system/users" }
      ]
    }
  ]
}
```

---

#### GET /menus/:id - 获取菜单

---

#### POST /menus - 创建菜单

**请求参数:**
```json
{
  "name": "菜单名称",
  "path": "/path",
  "icon": "icon-name",
  "parent_id": 0,
  "sort": 1
}
```

---

#### PUT /menus/:id - 更新菜单

---

#### DELETE /menus/:id - 删除菜单

---

## 7. 角色模块 (Role)

### 受保护接口

#### GET /roles - 角色列表

---

#### GET /roles/:id - 获取角色

---

#### POST /roles - 创建角色

**请求参数:**
```json
{
  "code": "admin",
  "name": "管理员",
  "data_scope": 1,
  "remark": "备注"
}
```

| data_scope | 说明 |
|------------|------|
| 1 | 全部数据 |
| 2 | 自定义数据范围 |
| 3 | 本部门数据 |
| 4 | 本部门及下级 |
| 5 | 仅本人数据 |

---

#### PUT /roles/:id - 更新角色

---

#### DELETE /roles/:id - 删除角色

---

#### GET /roles/:id/data-scopes - 获取角色数据范围

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "data_scope": 2,
    "org_ids": [1, 2, 3]
  }
}
```

---

#### PUT /roles/:id/data-scopes - 更新角色数据范围

**请求参数:**
```json
{
  "data_scope": 2,
  "org_ids": [1, 2, 3]
}
```

---

## 8. 权限模块 (Permission)

### 受保护接口

#### GET /roles/:id/permissions - 获取角色权限列表

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": [
    { "id": 1, "resource_type": "menu", "resource_code": "user", "action": "list" },
    { "id": 2, "resource_type": "menu", "resource_code": "user", "action": "create" }
  ]
}
```

---

#### POST /roles/:id/permissions - 分配权限

**请求参数:**
```json
{
  "permissions": [
    { "resource_type": "menu", "resource_code": "user", "action": "list" },
    { "resource_type": "menu", "resource_code": "user", "action": "create" }
  ]
}
```

---

#### PUT /roles/:id/permissions - 更新权限（同分配）

---

#### GET /roles/:id/menus - 获取角色可访问菜单

---

#### GET /roles/:id/resources - 获取角色可访问资源

---

## 9. 组织机构模块 (Organization)

### 受保护接口

#### GET /organizations/tree - 组织树

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": [
    {
      "id": 1,
      "name": "总公司",
      "code": "HQ",
      "children": [
        { "id": 2, "name": "研发部", "code": "RD" }
      ]
    }
  ]
}
```

---

#### GET /organizations - 组织列表

---

#### GET /organizations/:id - 获取组织

---

#### POST /organizations - 创建组织

**请求参数:**
```json
{
  "name": "部门名称",
  "code": "DEPT001",
  "parent_id": 0
}
```

---

#### PUT /organizations/:id - 更新组织

---

#### DELETE /organizations/:id - 删除组织

---

## 10. 资源模块 (Resource)

### 受保护接口

#### GET /resources - 资源列表

---

#### GET /resources/by-menu/:menu_id - 获取菜单下的资源

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": [
    { "id": 1, "menu_id": 1, "name": "查看", "code": "list", "action": "list" },
    { "id": 2, "menu_id": 1, "name": "新建", "code": "create", "action": "create" }
  ]
}
```

---

#### POST /resources - 创建资源

**请求参数:**
```json
{
  "menu_id": 1,
  "name": "按钮名称",
  "code": "button_code",
  "action": "create"
}
```

---

#### PUT /resources/:id - 更新资源

---

#### DELETE /resources/:id - 删除资源

---

## 11. 数据字典模块 (Dict)

### 受保护接口

#### GET /dicts - 字典列表

---

#### GET /dicts/:id - 获取字典

---

#### POST /dicts - 创建字典

**请求参数:**
```json
{
  "code": "status",
  "name": "状态",
  "remark": "备注"
}
```

---

#### PUT /dicts/:id - 更新字典

---

#### DELETE /dicts/:id - 删除字典

---

#### GET /dicts/:code/items - 获取字典项

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": [
    { "id": 1, "dict_code": "status", "label": "启用", "value": "1", "sort": 1 },
    { "id": 2, "dict_code": "status", "label": "禁用", "value": "0", "sort": 2 }
  ]
}
```

---

## 12. 系统模块 (System)

### 公开接口

#### GET /health - 健康检查

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": { "status": "ok" }
}
```

---

## 13. Flag 模块

### 公开接口

#### GET /flag/frames - 相框列表

**请求参数:**
| 参数 | 类型 | 说明 |
|------|------|------|
| category_id | int | 分类ID（可选） |
| page | int | 页码 |
| size | int | 每页数量 |

---

#### GET /flag/frames/:id - 获取相框

---

#### GET /flag/frames-categories - 相框分类列表

---

#### GET /flag/spaces/:code - 通过编码获取Space

---

#### GET /flag/avatar-categories - 头像分类列表

---

#### GET /flag/avatars - 头像列表

---

#### GET /flag/avatars/:id - 获取头像

---

### 受保护接口

#### POST /flag/frames - 创建相框

**请求参数:**
```json
{
  "category_id": 1,
  "name": "相框名称",
  "description": "描述",
  "preview_url": "预览图URL",
  "template_url": "模板URL",
  "type": 1,
  "sort": 1
}
```

---

#### PUT /flag/frames/:id - 更新相框

---

#### DELETE /flag/frames/:id - 删除相框

---

#### POST /flag/frames-categories - 创建分类

---

#### PUT /flag/frames-categories/:id - 更新分类

---

#### DELETE /flag/frames-categories/:id - 删除分类

---

#### POST /flag/spaces - 创建Space

---

#### PUT /flag/spaces/:id - 更新Space

---

#### DELETE /flag/spaces/:id - 删除Space

---

#### GET /flag/spaces - Space列表

---

#### POST /flag/avatar-categories - 创建头像分类

---

#### PUT /flag/avatar-categories/:id - 更新头像分类

---

#### DELETE /flag/avatar-categories/:id - 删除头像分类

---

#### POST /flag/avatars - 创建头像

---

#### PUT /flag/avatars/:id - 更新头像

---

#### DELETE /flag/avatars/:id - 删除头像

---

#### POST /flag/generate - 生成头像

---

#### GET /flag/my-avatars - 我的头像列表

---

## 14. 错误响应示例

### 未授权 (401)

```json
{
  "code": 401,
  "msg": "token 已过期",
  "data": null
}
```

### 禁止访问 (403)

```json
{
  "code": 403,
  "msg": "权限不足",
  "data": null
}
```

### 参数错误 (400)

```json
{
  "code": 400,
  "msg": "名称不能为空",
  "data": null
}
```

### 资源不存在 (404)

```json
{
  "code": 404,
  "msg": "用户不存在",
  "data": null
}
```

### 服务器错误 (500)

```json
{
  "code": 500,
  "msg": "服务器内部错误",
  "data": null
}
```