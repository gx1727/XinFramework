# 权限

> RBAC + 数据范围 + 平台角色 三层模型。

## 总览

```
┌─────────────────────────────────────────────────────────┐
│  平台角色 (Platform Role)                                │
│  例：super_admin — 跨租户操作                            │
│  中间件：middleware.RequirePlatformRole("super_admin")    │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  资源权限码 (Resource Permission)                         │
│  例：user:list, user:create, role:update                 │
│  中间件：middleware.Require(permission.P("user", "list")) │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  数据范围 (Data Scope)                                    │
│  All / Custom / Dept / DeptAndBelow / Self               │
│  由 role.data_scope_type 决定；repository 自动应用        │
└─────────────────────────────────────────────────────────┘
```

## 平台角色

平台角色是**跨租户**的特权，独立于租户内 RBAC。

- 存储：`platform_roles(account_id, role)` 联合主键
- 常用值：`super_admin`
- 中间件：[middleware.RequirePlatformRole](file:///d:\work\xin\XinFramework\server\framework\pkg\middleware\auth.go)

```go
tenants := protected.Group("/tenants")
tenants.Use(middleware.RequirePlatformRole("super_admin"))
```

登录成功后 `user.platform_roles` 字段会带上平台角色，前端可用于显示"系统管理"菜单。

## 资源权限

资源码格式：`<resource>:<action>`，如 `user:list`、`role:update`。

存储：`resources(code, type, name, method, path)`，由 `permission` 模块维护。

绑定：`role_resources(role_id, resource_id)`。

中间件使用：

```go
import (
    "gx1727.com/xin/framework/pkg/middleware"
    "gx1727.com/xin/framework/pkg/permission"
)

items.GET("", middleware.Require(permission.P(permission.ResUser, permission.ActList)), h.List)
items.POST("", middleware.Require(permission.P(permission.ResUser, permission.ActCreate)), h.Create)

// 多选其一
items.GET("/:id", middleware.RequireAny(
    permission.P(permission.ResUser, permission.ActList),
    permission.P(permission.ResUser, permission.ActUpdate),
), h.Get)

// 必须全部满足
items.POST("/batch", middleware.RequireAll(
    permission.P(permission.ResUser, permission.ActCreate),
    permission.P(permission.ResPermission, permission.ActCreate),
), h.BatchCreate)
```

### 内置资源码常量

[framework/pkg/permission/constants.go](file:///d:\work\xin\XinFramework\server\framework\pkg\permission\constants.go) 定义：

```go
const (
    ResTenant       = "tenant"
    ResUser         = "user"
    ResRole         = "role"
    ResMenu         = "menu"
    ResResource     = "resource"
    ResPermission   = "permission"
    ResOrganization = "org"
    ResDict         = "dict"
    ResAttachment   = "attachment"
    ResSystem       = "system"

    ActList   = "list"
    ActGet    = "get"
    ActCreate = "create"
    ActUpdate = "update"
    ActDelete = "delete"
)
```

新模块加资源码：

1. 在 `constants.go` 加常量
2. 在 `migrations/framework.sql` 的 `resources` 表插对应行
3. 在路由上用 `middleware.Require(permission.P(...))`

## 数据范围

数据范围决定"看多少数据"，与"能不能看"正交。

### 类型

| type | 名称 | 行为 |
| --- | --- | --- |
| 1 | All | 全部数据 |
| 2 | Custom | 仅 `data_scope_org_ids` 指定的组织 |
| 3 | Dept | 仅本组织 |
| 4 | DeptAndBelow | 本组织 + 子组织 |
| 5 | Self | 仅自己 |

### 存储

```sql
ALTER TABLE roles ADD COLUMN
    data_scope_type SMALLINT NOT NULL DEFAULT 5,
    data_scope_org_ids BIGINT[] DEFAULT '{}';
```

### 应用

在 repository 查询里自动根据当前用户的 data_scope 改写 WHERE：

```go
func (r *UserRepository) List(ctx context.Context, filter ListFilter) ([]User, error) {
    xc := xinContext.FromContext(ctx)
    if xc == nil {
        return nil, errors.New("missing xin context")
    }

    where := []string{"tenant_id = $1", "is_deleted = FALSE"}
    args := []any{xc.TenantID}

    // 应用 data_scope
    switch xc.DataScope.Type {
    case 1: // All - 不加额外过滤
    case 2: // Custom
        where = append(where, fmt.Sprintf("org_id = ANY($%d)", len(args)+1))
        args = append(args, xc.DataScope.OrgIDs)
    case 3: // Dept
        where = append(where, fmt.Sprintf("org_id = $%d", len(args)+1))
        args = append(args, xc.OrgID)
    case 4: // DeptAndBelow
        where = append(where, fmt.Sprintf("org_id = ANY($%d)", len(args)+1))
        args = append(args, collectOrgSubtreeIDs(xc.OrgID))
    case 5: // Self
        where = append(where, fmt.Sprintf("id = $%d", len(args)+1))
        args = append(args, xc.UserID)
    }

    query := "SELECT ... FROM users WHERE " + strings.Join(where, " AND ") + " ..."
    ...
}
```

`xc.DataScope` 在 `middleware.Auth` 里注入到 `gin.Context`，handler 通过 `xinContext.FromContext(c)` 拿。

### 前端配合

前端列表组件可加 `org_subtree` 参数手动控制范围（用于"切换查看"）：

```typescript
api.users.list({ org_id: 5, org_subtree: 1 })  // 看 5 号组织及其子
api.users.list({ org_id: 5 })                   // 仅看 5 号
```

## 超级管理员

`super_admin` 是平台角色，不受 data_scope 约束。在 repository 里：

```go
if xc.HasPlatformRole("super_admin") {
    // 跳过 data_scope
    return doListAll()
}
```

## 缓存

权限检查结果在 `framework/pkg/permission/permission_cache` 里缓存（默认 5 分钟 TTL）：

- `(user_id, resource, action) -> bool`
- `(user_id) -> DataScope`
- `(user_id) -> []Roles`

角色变更时通过 `permission_cache.Invalidate(userID)` 主动失效。

## 调试

开启 `debug: true` 配置后：

```yaml
app:
  debug: true
```

中间件会打印权限检查日志到 stderr，便于调试：

```
[perm] user=1 spec=user:list result=true cache=hit
[perm] user=1 spec=user:create result=false cache=miss
```

生产环境关掉 debug。

## 实战：新增一个业务模块的权限

假设加 `apps/order`，要支持：

1. 资源码常量加到 `framework/pkg/permission/constants.go`：

```go
const (
    ResOrder = "order"
)
```

2. 在 `migrations/framework.sql` 插入：

```sql
INSERT INTO resources (code, type, name, method, path) VALUES
('order:list',   'api', '订单列表', 'GET',    '/api/v1/orders'),
('order:create', 'api', '新建订单', 'POST',   '/api/v1/orders'),
('order:update', 'api', '更新订单', 'PUT',    '/api/v1/orders/:id'),
('order:delete', 'api', '删除订单', 'DELETE', '/api/v1/orders/:id');
```

3. 路由注册：

```go
orders := protected.Group("/orders")
orders.GET("", middleware.Require(permission.P(permission.ResOrder, permission.ActList)), h.List)
orders.POST("", middleware.Require(permission.P(permission.ResOrder, permission.ActCreate)), h.Create)
orders.PUT("/:id", middleware.Require(permission.P(permission.ResOrder, permission.ActUpdate)), h.Update)
orders.DELETE("/:id", middleware.Require(permission.P(permission.ResOrder, permission.ActDelete)), h.Delete)
```

4. 数据范围：默认用租户级隔离。如果订单需要"部门级"，在 `users` 表查询时加 `org_id` 过滤，按 data_scope 自动应用。

## 常见错误

| 现象 | 原因 | 排查 |
| --- | --- | --- |
| 403 Forbidden | 当前用户没有 `resource:action` | 检查 `role_resources` 绑定 |
| 列表为空 | data_scope 太严 | 看 `xc.DataScope.Type` |
| super_admin 也看不到全部 | 没在 repository 检查 platform_role | 改 repository 加 `if xc.HasPlatformRole(...)` |
| 改角色后权限不更新 | 缓存没失效 | 调 `permission_cache.Invalidate(userID)` |