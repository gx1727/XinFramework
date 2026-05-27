# 角色管理 API 文档

## 基础信息

- **Base Path**: `/api/v1`
- **认证**: 需要在 Header 中携带有效的 Token
- **权限说明**: 所有接口都需要 `ResRole` 权限

---

## 1. 角色列表

### 请求

```
GET /api/v1/roles?keyword=&page=1&size=20
```

### 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keyword | string | 否 | 搜索关键词（匹配 code 或 name） |
| page | int | 否 | 页码，默认 1 |
| size | int | 否 | 每页数量，默认 20 |

### 响应

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "list": [
      {
        "id": 1,
        "tenant_id": 1,
        "org_id": null,
        "code": "admin",
        "name": "管理员",
        "description": "系统管理员",
        "data_scope": 5,
        "extend": "",
        "is_default": false,
        "sort": 1,
        "status": 1,
        "created_at": "2026-04-26 10:00:00",
        "updated_at": "2026-04-26 10:00:00"
      }
    ],
    "total": 1
  }
}
```

---

## 2. 获取角色详情

### 请求

```
GET /api/v1/roles/:id
```

### 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 角色ID |

### 响应

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "tenant_id": 1,
    "org_id": null,
    "code": "admin",
    "name": "管理员",
    "description": "系统管理员",
    "data_scope": 5,
    "extend": "",
    "is_default": false,
    "sort": 1,
    "status": 1,
    "created_at": "2026-04-26 10:00:00",
    "updated_at": "2026-04-26 10:00:00"
  }
}
```

---

## 3. 创建角色

### 请求

```
POST /api/v1/roles
Content-Type: application/json

{
  "code": "editor",
  "name": "编辑",
  "description": "内容编辑角色",
  "data_scope": 2,
  "is_default": false,
  "sort": 2,
  "status": 1
}
```

### 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| code | string | 是 | 角色编码 |
| name | string | 是 | 角色名称 |
| description | string | 否 | 角色描述 |
| data_scope | int8 | 否 | 数据范围（1-5），默认 1 |
| is_default | bool | 否 | 是否默认角色 |
| sort | int | 否 | 排序号 |
| status | int8 | 否 | 状态（1=启用，0=禁用） |

### 响应

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 2,
    "tenant_id": 1,
    "org_id": null,
    "code": "editor",
    "name": "编辑",
    "description": "内容编辑角色",
    "data_scope": 2,
    "extend": "",
    "is_default": false,
    "sort": 2,
    "status": 1,
    "created_at": "2026-05-27 11:00:00",
    "updated_at": "2026-05-27 11:00:00"
  }
}
```

---

## 4. 更新角色

### 请求

```
PUT /api/v1/roles/:id
Content-Type: application/json

{
  "name": "超级编辑",
  "description": "内容编辑角色（已升级）",
  "data_scope": 3,
  "is_default": false,
  "sort": 3,
  "status": 1
}
```

### 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 角色ID |
| name | string | 是 | 角色名称 |
| description | string | 否 | 角色描述 |
| data_scope | int8 | 否 | 数据范围 |
| is_default | bool | 否 | 是否默认角色 |
| sort | int | 否 | 排序号 |
| status | int8 | 否 | 状态 |

### 响应

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 2,
    "tenant_id": 1,
    "org_id": null,
    "code": "editor",
    "name": "超级编辑",
    "description": "内容编辑角色（已升级）",
    "data_scope": 3,
    "extend": "",
    "is_default": false,
    "sort": 3,
    "status": 1,
    "created_at": "2026-05-27 11:00:00",
    "updated_at": "2026-05-27 12:00:00"
  }
}
```

---

## 5. 删除角色

### 请求

```
DELETE /api/v1/roles/:id
```

### 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 角色ID |

### 响应

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "ok": true
  }
}
```

### 错误

- `400`: 不能删除 admin 角色

---

## 6. 获取角色数据范围

### 请求

```
GET /api/v1/roles/:id/data-scopes
```

### 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 角色ID |

### 响应

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

## 7. 更新角色数据范围

### 请求

```
PUT /api/v1/roles/:id/data-scopes
Content-Type: application/json

{
  "org_ids": [1, 2, 3]
}
```

### 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 角色ID |
| org_ids | []uint | 是 | 组织ID列表（全量覆盖） |

### 响应

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

## 8. 获取角色菜单权限

### 请求

```
GET /api/v1/roles/:id/menus
```

### 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 角色ID |

### 响应

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "menu_ids": [1, 2, 3, 5, 10]
  }
}
```

---

## 9. 分配角色菜单权限

### 请求

```
PUT /api/v1/roles/:id/menus
Content-Type: application/json

{
  "menu_ids": [1, 2, 3, 5, 10]
}
```

### 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 角色ID |
| menu_ids | []uint | 是 | 菜单ID列表（全量覆盖） |

### 响应

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

## 数据范围说明

| 值 | 名称 | 说明 |
|----|------|------|
| 1 | DataScopeAll | 所有数据 |
| 2 | DataScopeCustom | 仅限自定义组织 |
| 3 | DataScopeDept | 仅限本部门 |
| 4 | DataScopeDeptAndBelow | 本部门及下级 |
| 5 | DataScopeSelf | 仅本人 |

---

## 角色响应字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 角色ID |
| tenant_id | uint | 租户ID |
| org_id | *uint | 组织ID（可空） |
| code | string | 角色编码 |
| name | string | 角色名称 |
| description | string | 角色描述 |
| data_scope | int8 | 数据范围 |
| extend | string | 扩展数据（JSON） |
| is_default | bool | 是否默认角色 |
| sort | int | 排序号 |
| status | int8 | 状态（1=启用，0=禁用） |
| created_at | string | 创建时间 |
| updated_at | string | 更新时间 |