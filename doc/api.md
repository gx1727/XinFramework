# XinFramework API 接口文档

基础路径: `/api/v1`

## 认证说明

所有受保护的接口需要在请求头中携带 JWT Token:
```
Authorization: Bearer <token>
```

## 响应格式

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 500 | 服务器错误 |

---

## 认证模块 (Auth)

### 公开接口 (无需认证)

#### POST /auth/login - 用户登录

**请求参数:**
```json
{
  "account": "账号",
  "password": "密码",
  "tenant_id": 0
}
```

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "token": "jwt_token",
    "refresh_token": "刷新令牌",
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
  "password": "密码 (6-32字符)",
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
    "token": "jwt_token",
    "refresh_token": "刷新令牌",
    "user": {
      "id": 1,
      "tenant_id": 1,
      "code": "admin",
      "role": "user"
    }
  }
}
```

---

#### POST /auth/refresh - 刷新令牌

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
    "token": "新jwt_token",
    "refresh_token": "新刷新令牌"
  }
}
```

---

### 受保护接口

#### POST /auth/logout - 用户登出

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": null
}
```

---

## 系统模块 (System)

#### GET /health - 健康检查

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "status": "ok"
  }
}
```

---

## 租户模块 (Tenant)

#### GET /tenants - 租户列表

**查询参数:**
| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| size | int | 20 | 每页数量 (最大100) |
| keyword | string | | 搜索关键词 |
| status | int | | 按状态筛选 |

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "list": [
      {
        "id": 1,
        "code": "tenant1",
        "name": "租户名称",
        "status": 1,
        "contact": "联系人",
        "phone": "电话",
        "email": "邮箱",
        "province": "省",
        "city": "市",
        "area": "区",
        "address": "地址",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "total": 100
  }
}
```

---

#### GET /tenants/:id - 获取租户详情

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "code": "tenant1",
    "name": "租户名称",
    "status": 1,
    "contact": "联系人",
    "phone": "电话",
    "email": "邮箱",
    "province": "省",
    "city": "市",
    "area": "区",
    "address": "地址",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

---

#### POST /tenants - 创建租户

**请求参数:**
```json
{
  "code": "编码 (必填, 1-50字符)",
  "name": "名称 (必填, 1-100字符)",
  "contact": "联系人",
  "phone": "电话",
  "email": "邮箱",
  "status": 1
}
```

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "code": "tenant1",
    "name": "租户名称",
    "status": 1,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

#### PUT /tenants/:id - 更新租户

**请求参数:**
```json
{
  "name": "名称 (必填)",
  "contact": "联系人",
  "phone": "电话",
  "email": "邮箱",
  "status": 1,
  "province": "省",
  "city": "市",
  "area": "区",
  "address": "地址"
}
```

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": null
}
```

---

#### DELETE /tenants/:id - 删除租户 (软删除)

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": null
}
```

---

## 用户模块 (User)

#### GET /users - 用户列表

**查询参数:**
| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| keyword | string | | 搜索关键词 |
| page | int | 1 | 页码 |
| size | int | 20 | 每页数量 |

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "list": [
      {
        "id": 1,
        "tenant_id": 1,
        "account_id": 1,
        "code": "user001",
        "status": 1,
        "real_name": "张三",
        "avatar": "头像URL",
        "phone": "手机号",
        "email": "邮箱",
        "role": "user"
      }
    ],
    "total": 100,
    "page": 1,
    "size": 20
  }
}
```

---

#### GET /users/:id - 获取用户详情

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "tenant_id": 1,
    "account_id": 1,
    "code": "user001",
    "status": 1,
    "real_name": "张三",
    "avatar": "头像URL",
    "phone": "手机号",
    "email": "邮箱",
    "role": "user"
  }
}
```

---

#### PUT /users/:id/status - 更新用户状态

**请求参数:**
```json
{
  "id": 1,
  "status": 1
}
```

| 状态值 | 说明 |
|--------|------|
| 1 | 正常 |
| 2 | 停用 |

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": null
}
```

---

## 菜单模块 (Menu)

#### GET /menus - 菜单列表

**查询参数:**
| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| size | int | 20 | 每页数量 |
| root | bool | false | 仅查询顶级菜单 |

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "list": [
      {
        "id": 1,
        "tenant_id": 1,
        "code": "menu1",
        "name": "菜单名称",
        "subtitle": "副标题",
        "url": "/url",
        "path": "/path",
        "icon": "icon",
        "sort": 1,
        "parent_id": 0,
        "ancestors": "0",
        "visible": true,
        "enabled": true,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "total": 50
  }
}
```

---

#### GET /menus/tree - 获取菜单树形结构

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": [
    {
      "id": 1,
      "tenant_id": 1,
      "code": "menu1",
      "name": "菜单名称",
      "subtitle": "副标题",
      "url": "/url",
      "path": "/path",
      "icon": "icon",
      "sort": 1,
      "parent_id": 0,
      "ancestors": "0",
      "visible": true,
      "enabled": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "children": [
        {
          "id": 2,
          "parent_id": 1,
          "children": []
        }
      ]
    }
  ]
}
```

---

#### GET /menus/:id - 获取菜单详情

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "tenant_id": 1,
    "code": "menu1",
    "name": "菜单名称",
    "subtitle": "副标题",
    "url": "/url",
    "path": "/path",
    "icon": "icon",
    "sort": 1,
    "parent_id": 0,
    "ancestors": "0",
    "visible": true,
    "enabled": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

---

#### POST /menus - 创建菜单

**请求参数:**
```json
{
  "code": "编码 (必填)",
  "name": "名称 (必填)",
  "subtitle": "副标题",
  "url": "URL",
  "path": "路径",
  "icon": "图标",
  "sort": 1,
  "parent_id": 0,
  "ancestors": "0,1",
  "visible": true,
  "enabled": true
}
```

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "code": "menu1",
    "name": "菜单名称"
  }
}
```

---

#### PUT /menus/:id - 更新菜单

**请求参数:**
```json
{
  "code": "编码",
  "name": "名称",
  "subtitle": "副标题",
  "url": "URL",
  "path": "路径",
  "icon": "图标",
  "sort": 1,
  "visible": true,
  "enabled": true
}
```

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": null
}
```

---

#### DELETE /menus/:id - 删除菜单 (软删除)

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": null
}
```

---

## 字典模块 (Dict)

#### GET /dicts - 字典列表

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "list": [
      {
        "id": 1,
        "tenant_id": 1,
        "code": "status",
        "name": "状态",
        "extend": {},
        "items": [
          {
            "id": 1,
            "dict_id": 1,
            "code": "1",
            "name": "启用",
            "sort": 1,
            "extend": {}
          }
        ]
      }
    ],
    "total": 10
  }
}
```

---

#### GET /dicts/:code - 获取字典详情

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "tenant_id": 1,
    "code": "status",
    "name": "状态",
    "extend": {},
    "items": [
      {
        "id": 1,
        "dict_id": 1,
        "code": "1",
        "name": "启用",
        "sort": 1,
        "extend": {}
      }
    ]
  }
}
```

---

#### POST /dicts - 创建字典

**请求参数:**
```json
{
  "code": "编码 (必填)",
  "name": "名称 (必填)",
  "extend": {}
}
```

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "code": "status",
    "name": "状态"
  }
}
```

---

#### PUT /dicts/:id - 更新字典

**请求参数:**
```json
{
  "name": "名称",
  "extend": {}
}
```

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "ok": true
  }
}
```

---

#### DELETE /dicts/:id - 删除字典 (软删除)

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "ok": true
  }
}
```

---

#### POST /dicts/:id/items - 创建字典项

**请求参数:**
```json
{
  "code": "编码 (必填)",
  "name": "名称 (必填)",
  "sort": 1,
  "extend": {}
}
```

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "dict_id": 1,
    "code": "1",
    "name": "启用",
    "sort": 1
  }
}
```

---

#### PUT /dicts/:id/items/:item_id - 更新字典项

**请求参数:**
```json
{
  "name": "名称",
  "sort": 1,
  "extend": {}
}
```

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "ok": true
  }
}
```

---

#### DELETE /dicts/:id/items/:item_id - 删除字典项

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "ok": true
  }
}
```

---

## 角色模块 (Role)

#### GET /roles - 角色列表

**查询参数:**
| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| keyword | string | | 按编码或名称搜索 |
| page | int | 1 | 页码 |
| size | int | 20 | 每页数量 |

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "list": [
      {
        "id": 1,
        "tenant_id": 1,
        "org_id": 1,
        "code": "admin",
        "name": "管理员",
        "description": "管理员角色",
        "data_scope": 1,
        "extend": "{}",
        "is_default": false,
        "sort": 1,
        "status": 1,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "total": 10
  }
}
```

| 数据范围 | 说明 |
|----------|------|
| 1 | 全部数据 |
| 2 | 自定义机构范围 |
| 3 | 仅本部门 |
| 4 | 本部门及下级 |
| 5 | 仅本人 |

---

#### GET /roles/:id - 获取角色详情

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "tenant_id": 1,
    "org_id": 1,
    "code": "admin",
    "name": "管理员",
    "description": "管理员角色",
    "data_scope": 1,
    "extend": "{}",
    "is_default": false,
    "sort": 1,
    "status": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

---

#### POST /roles - 创建角色

**请求参数:**
```json
{
  "code": "编码 (必填)",
  "name": "名称 (必填)",
  "description": "描述",
  "data_scope": 1,
  "is_default": false,
  "sort": 1,
  "status": 1
}
```

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "code": "admin",
    "name": "管理员"
  }
}
```

---

#### PUT /roles/:id - 更新角色

**请求参数:**
```json
{
  "name": "名称 (必填)",
  "description": "描述",
  "data_scope": 1,
  "is_default": false,
  "sort": 1,
  "status": 1
}
```

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": null
}
```

---

#### DELETE /roles/:id - 删除角色 (软删除)

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": null
}
```

---

#### GET /roles/:id/data-scopes - 获取角色数据权限

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "org_ids": [1, 2, 3]
  }
}
```

---

#### PUT /roles/:id/data-scopes - 更新角色数据权限

**请求参数:**
```json
{
  "org_ids": [1, 2, 3]
}
```

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": null
}
```

---

## 资源模块 (Resource)

#### GET /resources - 资源列表

**查询参数:**
| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| menu_id | uint | | 按菜单ID筛选 |
| action | string | | 按操作筛选 |
| page | int | 1 | 页码 |
| size | int | 20 | 每页数量 |

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "list": [
      {
        "id": 1,
        "tenant_id": 1,
        "menu_id": 1,
        "code": "user:create",
        "name": "创建用户",
        "action": "create",
        "description": "创建新用户",
        "sort": 1,
        "status": 1,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "total": 50
  }
}
```

---

#### GET /resources/:id - 获取资源详情

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "tenant_id": 1,
    "menu_id": 1,
    "code": "user:create",
    "name": "创建用户",
    "action": "create",
    "description": "创建新用户",
    "sort": 1,
    "status": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

---

#### POST /resources - 创建资源

**请求参数:**
```json
{
  "menu_id": 1,
  "code": "编码 (必填)",
  "name": "名称 (必填)",
  "action": "操作 (必填)",
  "description": "描述",
  "sort": 1,
  "status": 1
}
```

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "code": "user:create",
    "name": "创建用户"
  }
}
```

---

#### PUT /resources/:id - 更新资源

**请求参数:**
```json
{
  "name": "名称 (必填)",
  "action": "操作",
  "description": "描述",
  "sort": 1,
  "status": 1
}
```

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": null
}
```

---

#### DELETE /resources/:id - 删除资源 (软删除)

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": null
}
```

---

#### GET /resources/by-menu/:menu_id - 按菜单获取资源

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": [
    {
      "id": 1,
      "tenant_id": 1,
      "menu_id": 1,
      "code": "user:create",
      "name": "创建用户",
      "action": "create",
      "description": "创建新用户",
      "sort": 1,
      "status": 1,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

## 组织机构模块 (Organization)

#### GET /organizations - 组织列表

**查询参数:**
| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| keyword | string | | 搜索关键词 |
| parent_id | uint | | 按父级ID筛选 |
| page | int | 1 | 页码 |
| size | int | 20 | 每页数量 |

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "list": [
      {
        "id": 1,
        "tenant_id": 1,
        "code": "org1",
        "name": "总公司",
        "type": "company",
        "description": "集团总部",
        "admin_code": "admin",
        "parent_id": 0,
        "ancestors": "0",
        "sort": 1,
        "status": 1,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "total": 20
  }
}
```

---

#### GET /organizations/tree - 获取组织树形结构

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "tree": [
      {
        "id": 1,
        "tenant_id": 1,
        "code": "org1",
        "name": "总公司",
        "type": "company",
        "description": "集团总部",
        "admin_code": "admin",
        "parent_id": 0,
        "ancestors": "0",
        "sort": 1,
        "status": 1,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z",
        "children": [
          {
            "id": 2,
            "parent_id": 1,
            "children": []
          }
        ]
      }
    ]
  }
}
```

---

#### GET /organizations/:id - 获取组织详情

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "tenant_id": 1,
    "code": "org1",
    "name": "总公司",
    "type": "company",
    "description": "集团总部",
    "admin_code": "admin",
    "parent_id": 0,
    "ancestors": "0",
    "sort": 1,
    "status": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

---

#### POST /organizations - 创建组织

**请求参数:**
```json
{
  "code": "编码 (必填)",
  "name": "名称 (必填)",
  "type": "类型 (必填)",
  "description": "描述",
  "admin_code": "管理员账号",
  "parent_id": 0,
  "sort": 1,
  "status": 1
}
```

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "code": "org1",
    "name": "总公司"
  }
}
```

---

#### PUT /organizations/:id - 更新组织

**请求参数:**
```json
{
  "name": "名称 (必填)",
  "type": "类型",
  "description": "描述",
  "admin_code": "管理员账号",
  "sort": 1,
  "status": 1
}
```

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": null
}
```

---

#### DELETE /organizations/:id - 删除组织 (软删除)

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": null
}
```

---

## 权限模块 (Permission)

#### GET /roles/:id/permissions - 获取角色所有权限

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "menus": [
      {
        "id": 1,
        "code": "menu1",
        "name": "菜单1",
        "effect": 1
      }
    ],
    "resources": [
      {
        "id": 1,
        "code": "user:create",
        "name": "创建用户",
        "action": "create",
        "effect": 1
      }
    ]
  }
}
```

| effect | 说明 |
|--------|------|
| 1 | 允许 |
| 0 | 拒绝 |

---

#### POST /roles/:id/permissions - 分配权限给角色

**请求参数:**
```json
{
  "permissions": [
    {
      "resource_type": "menu",
      "resource_id": 1,
      "resource_code": "menu1",
      "effect": 1
    },
    {
      "resource_type": "resource",
      "resource_id": 1,
      "resource_code": "user:create",
      "effect": 1
    }
  ]
}
```

| resource_type | 说明 |
|---------------|------|
| menu | 菜单权限 |
| resource | 资源/按钮权限 |
| route | API路由权限 |

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": null
}
```

---

#### PUT /roles/:id/permissions - 更新角色权限 (替换所有)

**请求参数:**
同 POST

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": null
}
```

---

#### GET /roles/:id/menus - 获取角色菜单权限

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": [
    {
      "id": 1,
      "code": "menu1",
      "name": "菜单1",
      "effect": 1
    }
  ]
}
```

---

#### GET /roles/:id/resources - 获取角色资源权限

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": [
    {
      "id": 1,
      "code": "user:create",
      "name": "创建用户",
      "action": "create",
      "effect": 1
    }
  ]
}
```

---

## 微信模块 (WeChat)

#### GET /weixin/ping - 微信模块健康检查

**响应示例:**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "domain": "weixin",
    "status": "enabled"
  }
}
```

---

## 接口统计

| 模块 | 接口数量 | 说明 |
|------|----------|------|
| auth | 4 | 登录/注册/刷新/登出 |
| system | 1 | 健康检查 |
| tenant | 5 | 租户CRUD |
| user | 3 | 用户查询 |
| menu | 6 | 菜单CRUD |
| dict | 8 | 字典及字典项CRUD |
| role | 7 | 角色CRUD + 数据权限 |
| resource | 6 | 资源CRUD |
| organization | 6 | 组织CRUD |
| permission | 5 | 角色权限分配 |
| weixin | 1 | 微信模块 |
| **合计** | **52** | |
