# 组织管理 API 文档

## 基础信息

- **Base Path**: `/api/v1`
- **认证**: 需要在 Header 中携带有效的 Token
- **权限说明**: 所有接口都需要 `organization` 资源对应的权限（路由由 `Require` 中间件统一鉴权）
- **数据隔离**: 默认走租户隔离 + 当前用户数据范围（DataScope）过滤，仅返回当前用户可见的组织

### 权限码

| 操作 | 权限码 |
|------|--------|
| 查看组织树 / 列表 / 详情 | `organization:list` |
| 创建组织 | `organization:create` |
| 更新组织 | `organization:update` |
| 删除组织 | `organization:delete` |

> 超级管理员角色使用通配 `*:*`，可绕过上述校验。

---

## 1. 获取组织树

返回当前租户 + 当前用户数据范围下完整的组织树结构（按 `ancestors` 排序后组装成父子层级）。

### 请求

```
GET /api/v1/organizations/tree
```

### 响应

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "tree": [
      {
        "id": 1,
        "tenant_id": 1,
        "code": "HQ",
        "name": "总部",
        "type": "company",
        "description": "集团总部",
        "admin_code": "ceo",
        "parent_id": 0,
        "ancestors": "0",
        "sort": 0,
        "status": 1,
        "created_at": "2026-04-26 10:00:00",
        "updated_at": "2026-04-26 10:00:00",
        "children": [
          {
            "id": 2,
            "tenant_id": 1,
            "code": "TECH",
            "name": "技术中心",
            "type": "department",
            "description": "研发中心",
            "admin_code": "cto",
            "parent_id": 1,
            "ancestors": "0.1",
            "sort": 1,
            "status": 1,
            "created_at": "2026-04-26 10:05:00",
            "updated_at": "2026-04-26 10:05:00",
            "children": []
          }
        ]
      }
    ]
  }
}
```

---

## 2. 组织列表（扁平）

返回当前租户下（受数据范围过滤）的扁平组织列表；支持按 `parent_id` 取直接下级，按 `keyword` 在内存中模糊匹配 `name` / `code`。

> 当前实现为「全量返回 + 内存过滤」，`page` / `size` 仅作为入参占位，未真正分页；前端可自行对结果分页或改用 `tree` 接口。

### 请求

```
GET /api/v1/organizations?keyword=&parent_id=&page=1&size=20
```

### 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keyword | string | 否 | 搜索关键词，匹配 `name` 或 `code`（contains） |
| parent_id | uint | 否 | 指定父组织 ID；传值后只返回该组织的直接下级；不传或传 0 返回当前租户下所有可见组织 |
| page | int | 否 | 页码（占位，默认 1） |
| size | int | 否 | 每页数量（占位，默认 20） |

### 响应

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "list": [
      {
        "id": 2,
        "tenant_id": 1,
        "code": "TECH",
        "name": "技术中心",
        "type": "department",
        "description": "研发中心",
        "admin_code": "cto",
        "parent_id": 1,
        "ancestors": "0.1",
        "sort": 1,
        "status": 1,
        "created_at": "2026-04-26 10:05:00",
        "updated_at": "2026-04-26 10:05:00"
      }
    ],
    "total": 1
  }
}
```

---

## 3. 组织详情

### 请求

```
GET /api/v1/organizations/:id
```

### 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 组织 ID |

### 响应

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 2,
    "tenant_id": 1,
    "code": "TECH",
    "name": "技术中心",
    "type": "department",
    "description": "研发中心",
    "admin_code": "cto",
    "parent_id": 1,
    "ancestors": "0.1",
    "sort": 1,
    "status": 1,
    "created_at": "2026-04-26 10:05:00",
    "updated_at": "2026-04-26 10:05:00"
  }
}
```

### 错误

- `400`: 组织 ID 非法
- `404` / 业务码 `6001`: 组织不存在
- `403` / 业务码 `403`: 当前用户不在该组织的数据范围内

---

## 4. 创建组织

新建组织时，框架会按 `parent_id` 自动拼接 `ancestors`：根组织 `ancestors = "0"`，子组织 `ancestors = "<父级 ancestors>.<父级 id>"`。

### 请求

```
POST /api/v1/organizations
Content-Type: application/json

{
  "code": "TECH",
  "name": "技术中心",
  "type": "department",
  "description": "研发中心",
  "admin_code": "cto",
  "parent_id": 1,
  "sort": 1,
  "status": 1
}
```

### 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| code | string | 是 | 组织编码（租户内唯一） |
| name | string | 是 | 组织名称 |
| type | string | 是 | 组织类型，如 `company` / `department` / `team` |
| description | string | 否 | 组织描述 |
| admin_code | string | 否 | 管理员账号编码（业务自定义） |
| parent_id | uint | 否 | 父组织 ID；不传或传 0 表示根组织 |
| sort | int | 否 | 排序号，ASC 排序，默认 0 |
| status | int8 | 否 | 状态（1=启用，0=禁用），默认 1 |

### 响应

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 2,
    "tenant_id": 1,
    "code": "TECH",
    "name": "技术中心",
    "type": "department",
    "description": "研发中心",
    "admin_code": "cto",
    "parent_id": 1,
    "ancestors": "0.1",
    "sort": 1,
    "status": 1,
    "created_at": "2026-04-26 10:05:00",
    "updated_at": "2026-04-26 10:05:00"
  }
}
```

### 错误

- `400`: 请求体校验失败
- 业务码 `6002`: 组织编码已存在
- 业务码 `6004`: 后端未初始化 / DB 不可用

---

## 5. 更新组织

仅修改 `name / type / description / admin_code / sort / status` 等可变字段，`code / parent_id / ancestors` 不可通过本接口变更（避免破坏树结构）。

### 请求

```
PUT /api/v1/organizations/:id
Content-Type: application/json

{
  "name": "技术中心（已更名）",
  "type": "department",
  "description": "研发中心 - 含前端、后端、AI 三个组",
  "admin_code": "cto",
  "sort": 2,
  "status": 1
}
```

### 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 组织 ID |
| name | string | 是 | 组织名称 |
| type | string | 否 | 组织类型 |
| description | string | 否 | 组织描述 |
| admin_code | string | 否 | 管理员账号编码 |
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
    "code": "TECH",
    "name": "技术中心（已更名）",
    "type": "department",
    "description": "研发中心 - 含前端、后端、AI 三个组",
    "admin_code": "cto",
    "parent_id": 1,
    "ancestors": "0.1",
    "sort": 2,
    "status": 1,
    "created_at": "2026-04-26 10:05:00",
    "updated_at": "2026-04-26 11:00:00"
  }
}
```

### 错误

- `400`: 组织 ID 非法 / 请求体校验失败
- 业务码 `6001`: 组织不存在
- 业务码 `6004`: 后端未初始化 / DB 不可用

---

## 6. 删除组织

软删除（`is_deleted = TRUE`）。**根组织（`parent_id = 0`）不可删除**。

> 注意：当前实现未对子组织 / 关联用户做级联校验，如需阻断「有子节点或被引用的组织」删除，需要前端先调用 `?parent_id=<id>` 校验是否仍有下级，再调用删除接口。

### 请求

```
DELETE /api/v1/organizations/:id
```

### 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 组织 ID |

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

- `400`: 组织 ID 非法
- 业务码 `6001`: 组织不存在
- 业务码 `6003`: 不能删除根组织

---

## 业务错误码

| Code | HTTP | 说明 |
|------|------|------|
| 6001 | 200 | 组织不存在 |
| 6002 | 200 | 组织编码已存在 |
| 6003 | 200 | 不能删除根组织 |
| 6004 | 500 | 服务后端未初始化或不可用 |
| 400 | 400 | 参数校验失败（ID 非法、Body 解析失败等） |
| 401 | 401 | 未登录 / Token 失效 |
| 403 | 403 | 无访问权限或被数据范围过滤掉 |

---

## 组织对象字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 组织 ID |
| tenant_id | uint | 租户 ID |
| code | string | 组织编码（租户内唯一） |
| name | string | 组织名称 |
| type | string | 组织类型（`company` / `department` / `team` 等，业务自定义） |
| description | string | 组织描述 |
| admin_code | string | 管理员账号编码 |
| parent_id | uint | 父组织 ID，根组织为 0 |
| ancestors | string | 祖先链路（点分 ID 串），例如 `0.1.3`；根组织为 `0` |
| sort | int | 排序号，ASC |
| status | int8 | 状态（1=启用，0=禁用） |
| created_at | string | 创建时间（`YYYY-MM-DD HH:mm:ss`） |
| updated_at | string | 更新时间（`YYYY-MM-DD HH:mm:ss`） |
| children | OrgResp[] | 仅 `tree` 接口出现，子组织数组 |

---

## 数据范围（DataScope）说明

`organization` 模块的所有查询接口都按当前用户的「数据范围」过滤可见组织：

| 值 | 名称 | 说明 |
|----|------|------|
| 1 | DataScopeAll | 租户内全部组织 |
| 2 | DataScopeCustom | 仅 `role_data_scopes` 中配置的自定义组织 ID |
| 3 | DataScopeDept | 仅当前用户所属组织 |
| 4 | DataScopeDeptAndBelow | 当前用户所属组织 + 其所有下级组织 |
| 5 | DataScopeSelf | 仅当前用户所属组织（与 3 行为相同，组织维度无「本人」概念） |

> 提示：给前端做表单下拉时，请直接调 `GET /api/v1/organizations/tree`，避免手工组装父级链路。
