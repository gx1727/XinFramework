# XinFramework 使用手册

## 1. 错误处理规范

### 1.1 错误处理分层策略

本框架采用 **HTTP 状态码 + 业务码** 双层错误处理机制：

| 错误类型 | HTTP 状态码 | 业务码 | 处理方 |
|:--------|:----------|:------|:------|
| 成功 | `200` | `0` | - |
| 未认证 | `401` | `401` | 网关/前端拦截器 |
| 无权限 | `403` | `403` | 网关/前端拦截器 |
| 参数校验失败 | `400` | `400` | 前端表单 |
| 资源不存在 | `404` | `404` | 前端路由 |
| 业务逻辑错误 | `200` | 自定义业务码 | 前端业务代码 |
| 系统内部错误 | `500` | `500` | 监控告警 |

**设计原则**：
- **认证/权限/参数/系统错误** → HTTP 状态码非 200，网关和拦截器能直接处理
- **业务逻辑错误**（如"用户不存在"、"密码错误"）→ HTTP 200，由前端业务代码根据 `code` 处理

### 1.2 业务码范围定义

| 范围 | 类别 | 示例 |
|:-----|:----|:-----|
| `0` | 成功 | - |
| `400` | 参数错误 | 参数校验失败 |
| `401` | 未认证 | Token 为空或无效 |
| `403` | 无权限 | 权限不足 |
| `404` | 资源不存在 | 数据不存在 |
| `1001-1999` | 认证相关 | `1001` 用户名密码错误 |
| `2001-2999` | Token 相关 | `2002` Token 过期 |
| `3001-3999` | 租户相关 | `3001` 租户不存在 |
| `4001-4999` | 资源相关 | `4001` 订单不存在 |
| `5001-5999` | 业务相关 | 自定义业务错误 |
| `500` | 系统错误 | 服务器内部错误 |

### 1.3 响应函数

文件位置：`pkg/resp/resp.go`

```go
// 成功响应
Success(c, data)

// 通用业务错误（HTTP 200）
Error(c, code, msg)

// 未认证（HTTP 401）
Unauthorized(c, msg)

// 无权限（HTTP 403）
Forbidden(c, msg)

// 参数错误（HTTP 400）
BadRequest(c, msg)

// 资源不存在（HTTP 404）
NotFound(c, msg)

// 系统错误（HTTP 500）
ServerError(c, msg)

// 分页列表
Paginate(c, total, data)
```

### 1.4 响应格式

**成功响应**：
```json
{
  "code": 0,
  "msg": "ok",
  "data": { ... }
}
```

**分页响应**：
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "total": 100,
    "list": [...]
  }
}
```

**错误响应**：
```json
{
  "code": 1001,
  "msg": "用户名或密码错误",
  "data": null
}
```

---

## 2. API 接口规范

### 2.1 路径规范

```
/api/v1/{resource}
/api/v1/{resource}/:id
```

示例：
```
GET    /api/v1/users        # 获取用户列表
POST   /api/v1/users        # 创建用户
GET    /api/v1/users/:id    # 获取单个用户
PUT    /api/v1/users/:id    # 更新用户
DELETE /api/v1/users/:id   # 删除用户
```

### 2.2 公共接口 vs 认证接口

公共接口直接在 `v1.RegisterRoutes()` 中注册。

认证接口需使用 `Auth` 中间件分组：

```go
auth := srv.Engine.Group("/api/v1")
auth.Use(middleware.Auth(&cfg.JWT))
{
    // 认证接口
}
```

---

## 3. 多租户支持

### 3.1 租户模式

`config.yaml` 中 `saas.mode` 配置：

| 模式 | 说明 |
|:----|:-----|
| `""`（空） | 单租户模式，不做租户隔离 |
| `shared` | 共享数据库，通过 `tenant_id` 字段隔离 |
| `schema` | PostgreSQL Schema 隔离 |
| `database` | 独立数据库隔离 |

### 3.2 租户上下文传播

请求头：`X-Tenant-ID`

中间件 `Tenant()` 从请求头读取并设置到：
1. `XinContext.TenantID`
2. PostgreSQL 会话变量 `app.tenant_id`

```go
// 设置租户上下文
db.SetTenantID(tenantID)

// 查询时会自动带上租户过滤
// SELECT * FROM users WHERE tenant_id = ? AND is_deleted = FALSE

// 请求结束后清理
defer db.ClearTenantID()
```

### 3.3 租户相关表

以下表为全局表，不带 `tenant_id`：
- `accounts` - 全局账号表
- `plans` - 全局套餐表
- `db_logs` - 审计日志

---

## 4. 配置说明

### 4.1 配置文件

- `config/config.yaml` - 基础配置（提交到版本库）
- `.env` - 本地配置（不提交到版本库）
- `.env.example` - 环境变量模板

### 4.2 环境变量覆盖

环境变量会覆盖 `config.yaml` 中的对应值。变量名规则：大写下划线连接。

| config.yaml | 环境变量 |
|:------------|:---------|
| `app.name` | `APP_NAME` |
| `app.port` | `APP_PORT` |
| `database.host` | `DB_HOST` |
| `jwt.secret` | `JWT_SECRET` |
| `saas.mode` | `SAAS_MODE` |

---

## 5. 中间件说明

### 5.1 中间件执行顺序

```go
srv.Engine.Use(middleware.Logger())   // 1. 请求日志
srv.Engine.Use(middleware.Recovery()) // 2. 异常恢复
srv.Engine.Use(middleware.Tenant(cfg.Saas.Mode)) // 3. 租户隔离
// ... 路由处理 ...
srv.Engine.Group("/api/v1").Use(middleware.Auth(&cfg.JWT)) // 4. 认证
```

### 5.2 自定义中间件

```go
func RateLimit() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 实现限流逻辑
        c.Next()
    }
}
```
