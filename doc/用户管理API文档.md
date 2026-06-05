# 用户管理 API 文档

## 概述

用户管理模块提供用户 CRUD、状态管理、个人资料更新等功能。

**基础路径**: `/api/v1`

**认证说明**: 除登录相关接口外，其他接口需要在请求头中携带 `Authorization: Bearer <token>`

---

## 统一响应格式

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

| code | 说明 |
|------|------|
| 0    | 成功 |
| 400  | 请求参数错误 |
| 401  | 未登录 |
| 403  | 无权限 |
| 404  | 资源不存在 |
| 409  | 资源冲突（如用户名已存在） |
| 500  | 服务器内部错误 |

---

## 用户信息响应字段

`GET /user/profile`、`GET /users/:id`、`GET /users`、`PUT /users/:id`、`PATCH /users/:id` 五个接口返回的用户对象字段如下：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 用户ID |
| tenant_id | uint | 租户ID |
| account_id | uint | 关联的全局账号ID（0=未绑定） |
| code | string | 用户编码 |
| nickname | string | 昵称 |
| real_name | string | 真实姓名 |
| avatar | string | 头像URL |
| phone | string | 手机号（来自 accounts JOIN，未绑定时为空） |
| email | string | 邮箱（同上） |
| role | string | 主角色 code |
| status | int8 | 状态：0/2=禁用，1=正常 |

> `POST /users`（创建）的响应是 `createResponse`，只含 `id / tenant_id / code / username / real_name / phone / status`，**没有** `account_id / nickname / avatar / email / role`——按需查询上面那五个接口。

---

## 1. 获取当前用户资料

获取已登录用户的信息。

**接口**: `GET /api/v1/user/profile`

**权限**: 仅需登录（`RequireAuthenticated`）

**请求头**:
```
Authorization: Bearer <token>
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "tenant_id": 1,
    "account_id": 1,
    "code": "U0000001",
    "nickname": "",
    "real_name": "管理员",
    "avatar": "",
    "phone": "13800138000",
    "email": "admin@example.com",
    "role": "admin",
    "status": 1
  }
}
```

---

## 2. 更新个人资料

更新已登录用户的昵称和头像。

**接口**: `PUT /api/v1/user/profile`

**权限**: 仅需登录

**请求头**:
```
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体**:
```json
{
  "nickName": "新昵称",
  "avatarUrl": "https://example.com/avatar.jpg"
}
```

| 字段      | 类型   | 必填 | 说明   |
|-----------|--------|------|--------|
| nickName  | string | 是   | 昵称   |
| avatarUrl | string | 否   | 头像URL |

**响应示例**:
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

## 3. 上传头像

上传用户头像文件。

**接口**: `POST /api/v1/user/avatar`

**权限**: 仅需登录

**请求头**:
```
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

**请求体**: `file` (文件表单字段)

**响应示例**:
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "url": "https://cdn.example.com/avatars/1/avatar.jpg"
  }
}
```

---

## 4. 用户列表

获取当前租户下的用户列表。

**接口**: `GET /api/v1/users`

**权限**: `user:list`

**请求头**:
```
Authorization: Bearer <token>
```

**Query 参数**:
| 参数   | 类型   | 必填 | 说明               | 默认值 |
|--------|--------|------|--------------------|--------|
| keyword | string | 否   | 搜索关键词（账号/昵称/姓名/手机） |       |
| page    | int    | 否   | 页码               | 1      |
| size    | int    | 否   | 每页数量           | 20     |

**响应示例**:
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
        "code": "U0000001",
        "nickname": "",
        "real_name": "管理员",
        "avatar": "",
        "phone": "13800138000",
        "email": "admin@example.com",
        "role": "admin",
        "status": 1
      }
    ],
    "total": 1,
    "page": 1,
    "size": 20
  }
}
```

---

## 5. 获取用户详情

根据用户ID获取用户详情。

**接口**: `GET /api/v1/users/:id`

**权限**: `user:list`

**路径参数**:
| 参数 | 类型    | 必填 | 说明   |
|------|---------|------|--------|
| id   | integer | 是   | 用户ID |

**响应示例**:
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "tenant_id": 1,
    "account_id": 1,
    "code": "U0000001",
    "nickname": "",
    "real_name": "管理员",
    "avatar": "",
    "phone": "13800138000",
    "email": "admin@example.com",
    "role": "admin",
    "status": 1
  }
}
```

---

## 6. 创建用户

在当前租户下创建新用户，同时创建对应的账号。

**接口**: `POST /api/v1/users`

**权限**: `user:create`

**请求头**:
```
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体**:
```json
{
  "username": "testuser",
  "phone": "13800138001",
  "email": "test@example.com",
  "real_name": "测试用户",
  "password": "123456",
  "status": 1
}
```

| 字段     | 类型   | 必填 | 说明                          |
|----------|--------|------|------------------------------|
| username | string | 是   | 用户名（登录账号）            |
| phone    | string | 是   | 手机号                       |
| email    | string | 否   | 邮箱                         |
| real_name| string | 是   | 真实姓名                     |
| password | string | 是   | 密码（至少6位）              |
| status   | int    | 否   | 1=正常，2=禁用，默认 1        |

**响应示例**:
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 2,
    "tenant_id": 1,
    "code": "U0000002",
    "username": "testuser",
    "real_name": "测试用户",
    "phone": "13800138001",
    "status": 1
  }
}
```

**错误示例（用户名已存在）**:
```json
{
  "code": 409,
  "msg": "用户名已存在",
  "data": null
}
```

---

## 7. 更新用户状态

启用或禁用用户账号（仅修改 status 字段的快捷接口）。

**接口**: `PUT /api/v1/users/:id/status`

**权限**: `user:update`

**请求头**:
```
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体**:
```json
{
  "id": 2,
  "status": 2
}
```

| 字段  | 类型 | 必填 | 说明                    |
|-------|------|------|------------------------|
| id    | int  | 是   | 用户ID（与 URL `:id` 一致） |
| status| int  | 是   | 1=正常，2=禁用         |

> 历史接口：`status` 只接受 1/2，无法直接置 0。如需 `status=0` 请用第 9 节 `PATCH /users/:id`。

**响应示例**:
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

## 8. 更新用户（全量替换）

修改指定用户的昵称/真实姓名/头像/状态。一次提交即覆盖所有字段，未提供的字段会被置零/空字符串（status 除外）。

**接口**: `PUT /api/v1/users/:id`

**权限**: `user:update`

**路径参数**:
| 参数 | 类型    | 必填 | 说明   |
|------|---------|------|--------|
| id   | integer | 是   | 用户ID |

**请求体**:
```json
{
  "nickname": "新昵称",
  "real_name": "张三",
  "avatar": "https://example.com/avatar.jpg",
  "status": 1
}
```

| 字段      | 类型   | 必填 | 说明                          |
|-----------|--------|------|------------------------------|
| nickname  | string | 否   | 昵称                          |
| real_name | string | 否   | 真实姓名                      |
| avatar    | string | 否   | 头像URL                       |
| status    | int    | 否   | 0/1/2；0=禁用，1=正常，2=禁用 |

> **重要**：本接口**不修改** `phone` / `email`——这两个字段在 `accounts` 表上，且涉及唯一性约束与换绑验证流程，应走专门的"换绑手机/邮箱"接口（待提供）。

**响应示例**: 用户信息对象，参考"用户信息响应字段"一节。

---

## 9. 更新用户（局部更新）

只修改 body 中显式给出的字段，未提供的字段保持原值。空 body 等价于 GET，不写脏。

**接口**: `PATCH /api/v1/users/:id`

**权限**: `user:update`

**路径参数**: 同上节

**请求体**（任选字段子集）:
```json
{
  "status": 0
}
```

或一次改多个：
```json
{
  "nickname": "新昵称",
  "avatar": "https://example.com/avatar.jpg"
}
```

| 字段      | 类型   | 必填 | 说明                          |
|-----------|--------|------|------------------------------|
| nickname  | string | 否   | 同上                          |
| real_name | string | 否   | 同上                          |
| avatar    | string | 否   | 同上                          |
| status    | int    | 否   | 同上；本接口支持 0=禁用       |

> 指针字段语义：`{"status": 0}` 真的会把用户禁用；body 不带 `status` 字段则完全不动。

**响应示例**: 用户信息对象，参考"用户信息响应字段"一节。

---

## 权限说明

| 接口                     | 需要的权限    | 说明                          |
|--------------------------|---------------|------------------------------|
| GET /user/profile        | 仅需登录       | 获取当前用户信息              |
| PUT /user/profile        | 仅需登录       | 更新当前用户资料              |
| POST /user/avatar        | 仅需登录       | 上传当前用户头像              |
| GET /users               | user:list     | 用户列表                      |
| GET /users/:id           | user:list     | 用户详情                      |
| POST /users              | user:create   | 创建用户                      |
| PUT /users/:id           | user:update   | 全量更新用户（昵称/姓名/头像/状态） |
| PATCH /users/:id         | user:update   | 局部更新用户（任意字段子集）  |
| PUT /users/:id/status    | user:update   | 仅修改 status 的快捷接口      |

---

## 错误码

| HTTP状态码 | code | msg                    | 说明                    |
|------------|------|------------------------|------------------------|
| 200        | 0    | ok                     | 成功                    |
| 200        | 400  | 请求参数格式错误         | 参数校验失败            |
| 200        | 404  | 资源不存在              | 用户不存在              |
| 200        | 409  | 用户名已存在            | 账号冲突                |
| 200        | 500  | 系统错误：默认角色未配置 | 系统配置问题            |
| 401        | 401  | 未登录                  | Token无效或过期         |
| 403        | 403  | permission denied       | 无权限                  |
| 500        | 500  | 服务器内部错误          | 服务器异常              |

---

## 数据范围说明

用户列表和用户详情的查询结果会受到当前用户的数据权限影响：

| 数据范围类型 | 说明                              |
|-------------|-----------------------------------|
| 全部数据     | 可以看到当前租户下所有用户         |
| 本部门       | 只能看到同部门的用户               |
| 本部门及以下 | 只能看到本部门及下级部门的用户     |
| 本人数据     | 只能看到自己的账号                 |

管理员（拥有全部数据权限）不受此限制。
