# 权限系统

> XinFramework 的权限分三层：**资源码 RBAC**（能不能调 API） + **数据范围 DataScope**（能看哪些行） + **平台角色 PlatformRole**（跨租户特权）。

## 1. 三层权限一览

| 层 | 解决什么 | 实现位置 |
|---|---|---|
| **资源码 RBAC** | 这个用户**能不能调**这个 API | `resources` + `role_resources` + `Require(spec)` 中间件 |
| **数据范围 DataScope** | 这个用户**能看哪些行** | `roles.data_scope` + `BuildDataScopeFilter` + 业务 SQL WHERE |
| **平台角色 PlatformRole** | 这个账号**跨租户**的特权 | `account_roles.role` + `RequirePlatformRole` 中间件 |

**重要**：`super_admin` 平台角色**自动 bypass** 资源码 RBAC（详见 §6）。

---

## 2. 资源码 RBAC

### 2.1 资源码格式

`resource:action`，例如：

- `user:list` — 列出用户
- `user:create` — 创建用户
- `flag:create` — 创建 flag 业务记录
- `config:update` — 修改配置项

支持的 action 见 [`framework/pkg/permission/constants.go`](../framework/pkg/permission/constants.go)：

```go
ActList   = "list"
ActGet    = "get"
ActCreate = "create"
ActUpdate = "update"
ActDelete = "delete"
ActTree   = "tree"
```

支持的 resource（14 个）：

```go
ResSystem       = "system"
ResAsset        = "asset"
ResDict         = "dict"
ResTenant       = "tenant"
ResOrganization = "organization"
ResResource     = "resource"
ResMenu         = "menu"
ResRole         = "role"
ResUser         = "user"
ResPermission   = "permission"
ResWeixin       = "weixin"
ResAuth         = "auth"
ResFlag         = "flag"
ResConfig       = "config"     // apps/reference/config
```

### 2.2 Spec 类型

[`framework/pkg/permission/spec.go`](../framework/pkg/permission/spec.go) 定义：

```go
type Spec struct {
    Resource      string  // 资源码
    Action        string  // 操作
    Authenticated bool    // 是否需要登录（默认 true）
}
```

构造方式：

```go
spec := permission.P("user", "list")       // resource:action, 默认需要登录
spec := permission.AuthOnly()              // 只需要登录，任何资源都行
```

### 2.3 中间件

[`framework/pkg/middleware/auth.go`](../framework/pkg/middleware/auth.go) 提供：

| 函数 | 行为 |
|---|---|
| `Require(spec)` | 必须满足该 spec（单一） |
| `RequireAny(specs...)` | 任一满足 |
| `RequireAll(specs...)` | 全部满足 |
| `RequireAuthenticated()` | 仅登录（等价 `Require(AuthOnly())`） |
| `RequirePlatformRole(roles...)` | 必须持有平台角色（详见 §6） |

### 2.4 使用示例

```go
// 业务模块的 routes.go
import (
    "gx1727.com/xin/framework/pkg/middleware"
    "gx1727.com/xin/framework/pkg/permission"
)

func Register(protected *gin.RouterGroup, h *Handler) {
    protected.GET("/users",
        middleware.Require(permission.P(permission.ResUser, permission.ActList)),
        h.List)
    protected.POST("/users",
        middleware.Require(permission.P(permission.ResUser, permission.ActCreate)),
        h.Create)

    // 多选一
    protected.DELETE("/admin",
        middleware.RequireAny(
            permission.P(permission.ResUser, permission.ActDelete),
            permission.P(permission.ResRole, permission.ActDelete),
        ),
        h.Delete)

    // 全部需要
    protected.PUT("/super",
        middleware.RequireAll(
            permission.P(permission.ResRole, permission.ActUpdate),
            permission.P(permission.ResPermission, permission.ActUpdate),
        ),
        h.Update)
}
```

### 2.5 通配匹配

[`permission.HasPermission`](../framework/pkg/permission/types.go) 支持 3 级匹配：

```go
// 优先级 1：精确匹配
perms["user:list"] = true

// 优先级 2：资源级通配
perms["user:*"] = true      // 任意 user:xxx 都通过

// 优先级 3：全局通配（super_admin）
perms["*:*"] = true         // 任意 任意 都通过
```

**例子**：

| 用户有权限 | 是否能调 `user:list` | 是否能调 `role:list` |
|---|---|---|
| `{"user:list": true}` | ✅ | ❌ |
| `{"user:*": true}` | ✅（2 级匹配） | ❌ |
| `{"*:*": true}` | ✅（3 级） | ✅ |

**特别说明**：`*:*` 等价于 `super_admin`，在中间件层直接 `c.Next()`，不需要查数据库。

---

## 3. 数据范围 DataScope

资源码回答"能不能调"，数据范围回答"能看哪些行"。

### 3.1 五种类型

[`framework/pkg/permission/types.go`](../framework/pkg/permission/types.go)：

```go
const (
    DataScopeAll          DataScopeType = 1  // 全部数据
    DataScopeCustom       DataScopeType = 2  // 自定义数据（org_ids）
    DataScopeDept         DataScopeType = 3  // 本部门数据
    DataScopeDeptAndBelow DataScopeType = 4  // 本部门及以下数据（递归）
    DataScopeSelf         DataScopeType = 5  // 仅本人数据
)
```

### 3.2 存储

`roles.data_scope` 字段是 JSONB：

```sql
ALTER TABLE roles ADD COLUMN data_scope JSONB DEFAULT '{"type": 5}';
-- 默认每个人只能看自己创建的数据
```

例子：

```json
{"type": 1}                       // 看全部
{"type": 2, "org_ids": [3,5,7]}   // 看指定组织
{"type": 3}                       // 看本部门
{"type": 4}                       // 看本部门 + 子部门
{"type": 5}                       // 仅本人
```

### 3.3 编译期类型

```go
type DataScope struct {
    Type   DataScopeType `json:"type"`
    OrgIDs []int64       `json:"org_ids,omitempty"`
}
```

### 3.4 编译期过滤器

[`framework/pkg/permission/scope.go`](../framework/pkg/permission/scope.go)：

```go
func BuildDataScopeFilter(
    ds DataScope,
    userID uint,         // 当前用户 ID
    orgID int64,         // 当前用户组织 ID
    columns ScopeColumns, // 表列映射
) (ScopeFilter, error)
```

返回：

```go
type ScopeFilter struct {
    SQL  string
    Args []any
}
```

### 3.5 五种类型生成的 SQL

默认 `ScopeColumns{CreatorID: "creator_id", OrgID: "org_id"}`：

| Type | SQL | Args |
|---|---|---|
| `DataScopeAll` (1) | `""`（空） | `nil` |
| `DataScopeCustom` (2) 且 `org_ids=[]` | `"creator_id = $1"` | `[userID]` |
| `DataScopeCustom` (2) 且 `org_ids=[3,5]` | `"org_id = ANY($1)"` | `[[3,5]]` |
| `DataScopeDept` (3) 且 `org_id=0` | `"creator_id = $1"` | `[userID]` |
| `DataScopeDept` (3) 且 `org_id=7` | `"org_id = $1"` | `[7]` |
| `DataScopeDeptAndBelow` (4) 且 `org_id=7` | 递归 CTE | `[7]` |
| `DataScopeSelf` (5) | `"creator_id = $1"` | `[userID]` |
| 未知 type | `"creator_id = $1"`（防御 fallback） | `[userID]` |

**DeptAndBelow 的递归 SQL**：

```sql
org_id = $1
OR org_id IN (
    WITH RECURSIVE org_tree AS (
        SELECT id FROM organizations WHERE id = $1
        UNION ALL
        SELECT o.id FROM organizations o
        JOIN org_tree ot ON o.parent_id = ot.id
    )
    SELECT id FROM org_tree
)
```

### 3.6 在业务里用

```go
// apps/flag/avatar_repository.go
import "gx1727.com/xin/framework/pkg/permission"

func (r *AvatarRepo) ListMy(ctx context.Context, userID uint, page, size int) ([]Avatar, int64, error) {
    uc := xinContext.MustNewUserContext(c)
    filter, _ := uc.GetDataScopeFilter()  // 用默认列映射

    sql := `SELECT ... FROM avatars WHERE ` + filter.SQL
    args := append(filter.Args, size, (page-1)*size)

    // 业务 SQL 末尾追加 LIMIT/OFFSET
    sql += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`

    // ...
}
```

或者用自定义列名：

```go
filter, _ := uc.GetDataScopeFilterFor(permission.ScopeColumns{
    SelfColumn: "user_id",        // 此表的"创建者"列名
    OrgID:      "department_id",  // 此表的"组织"列名
})
```

### 3.7 性能注意

`DataScopeDeptAndBelow` 每次查询都跑递归 CTE。**超过 10 万行 organizations** 时应该用物化路径（`ancestors LIKE '/x/%'`）代替 CTE。

---

## 4. 权限加载

[`framework/pkg/permission/permission_impl.go`](../framework/pkg/permission/permission_impl.go) 实现：

### 4.1 GetUserPermissions

返回 `map[string]bool`，key 格式 `"resource:action"`。

```sql
SELECT DISTINCT res.code, res.action
FROM users u
JOIN user_roles ur ON ur.user_id = u.id
JOIN roles rol ON rol.id = ur.role_id
JOIN role_resources rr ON rr.role_id = rol.id
JOIN resources res ON res.id = rr.resource_id
WHERE u.id = $1
  AND u.is_deleted = FALSE
  AND ur.is_deleted = FALSE
  AND rol.is_deleted = FALSE
  AND rol.status = 1
  AND rr.is_deleted = FALSE
  AND rr.effect = 1
  AND res.is_deleted = FALSE
  AND res.status = 1
```

### 4.2 缓存

[`framework/pkg/permission/permission_cache.go`](../framework/pkg/permission/permission_cache.go) 用 Redis 缓存权限 + DataScope。

```go
// auth middleware 注入 lazy loader
uc := &UserContext{
    Permissions: perms,    // 已加载
    DataScope:   ds,       // 已加载
    Roles:       roles,
}
```

每次"懒加载"只查一次 DB（`/auth` 注入 loader 后第一次 `MustNewUserContext` 触发），后续走 Redis cache。

**失效**：role / permission 模块变更时调 `authz.InvalidateRole(roleID)` 清缓存，见 §5。

---

## 5. 缓存失效

[`framework/internal/service/authorization_service.go`](../framework/internal/service/authorization_service.go) 提供 `Authorization` interface，apps 通过 `AppContext.Authz()` 拿到。

### 5.1 接口

```go
type Authorization interface {
    InvalidateRole(ctx context.Context, roleID uint) error  // 清某角色所有关联 user 的权限缓存
    InvalidateUser(ctx context.Context, userID uint) error  // 清某 user 的权限缓存
    // ...
}
```

### 5.2 何时调

| 操作 | 调用 |
|---|---|
| 创建/更新/删除 role | `authz.InvalidateRole(roleID)` |
| 给 role 加减 resource | `authz.InvalidateRole(roleID)` |
| 给 user 加减 role | `authz.InvalidateUser(userID)` |
| 改 role 的 data_scope | `authz.InvalidateRole(roleID)` |

调用示例见 [`apps/rbac/role/service.go`](../apps/rbac/role/service.go)：

```go
func (s *Service) Update(ctx context.Context, roleID uint, req UpdateRoleReq) error {
    if err := s.repo.Update(ctx, roleID, req); err != nil { return err }
    if s.authz != nil {
        _ = s.authz.InvalidateRole(context.Background(), roleID)
    }
    return nil
}
```

注意 `context.Background()`：缓存失效是后台任务，不应绑请求生命周期。

---

## 6. 平台角色

### 6.1 概念

平台角色是**跨租户**的特权，绑定到 `accounts` 而非 `users`。典型用途：

- `super_admin`：平台超级管理员，可以管理任意租户
- 未来可能有 `auditor`（审计员）、`support`（客服）等

### 6.2 存储

```sql
CREATE TABLE account_roles (
    account_id BIGINT NOT NULL,
    role      VARCHAR(64) NOT NULL,
    ...
);
```

### 6.3 颁发

```sql
INSERT INTO account_roles (account_id, role) VALUES ($1, 'super_admin')
ON CONFLICT (account_id, role) DO NOTHING;
```

### 6.4 在 token 中传递

`accounts.id` 登录后，框架查 `account_roles` 找出所有平台角色，塞进 JWT claims：

```go
type Claims struct {
    UserID        uint
    TenantID      uint
    SessionID     string
    Role          string
    PlatformRoles []string    // ← 这里
    ...
}
```

### 6.5 中间件

`RequirePlatformRole("super_admin")` 校验 `claims.PlatformRoles` 是否包含任一指定角色。

### 6.6 自动 bypass RBAC

`super_admin` **自动** bypass 所有 `Require(spec)` 检查——不需要写 `RequireAny(spec, RequirePlatformRole(...))`，中间件里就已经短路了。

代码见 [`middleware/auth.go::requireWithSpecs`](../framework/pkg/middleware/auth.go)：

```go
// 平台超级管理员：无视所有权限规格直接放行
if uc.IsSuperAdmin() {
    c.Next()
    return
}
```

`uc.IsSuperAdmin()` 检查 `XinContext.PlatformRoles` 是否包含 `super_admin`。

### 6.7 当前用法

只有 `tenant` 模块用了 `RequirePlatformRole`——**租户管理必须 super_admin**。

其他场景如果需要，挂载方式：

```go
protected.Group("/billing").
    Use(middleware.RequirePlatformRole("billing_admin")).
    POST("", h.CreateInvoice)
```

---

## 7. 完整流程图

```
请求进来
   ↓
[Recovery → RequestID → CORS → ClientIP → Logger]   ← 全局中间件
   ↓
[Auth/OptionalAuth]
   ├─ 解析 JWT → claims
   ├─ 验签 → 失败 401
   ├─ 查 session → 失效 401
   ├─ 注入 XinContext{ UserID, TenantID, PlatformRoles }
   └─ 注册 UserContextLoader（懒）
   ↓
[RequirePlatformRole] (可选, 某些路由组)
   └─ 检查 PlatformRoles → 403
   ↓
[Require(spec) / RequireAny / RequireAll]
   ├─ IsSuperAdmin() → 放行
   ├─ IsAuthOnly()  → 放行
   └─ HasPermission → 403
   ↓
Handler 业务逻辑
   ├─ MustNewUserContext(c)  → 触发 loader → 查 cache → 查 DB
   ├─ RunInTenantTx → 自动 RLS
   ├─ 业务 SQL 拼接 data_scope filter
   ├─ JSONB 列 SQL `::jsonb` cast
   └─ resp.OK / resp.Error
```

---

## 8. 安全原则

1. **白名单优于黑名单**：默认 spec `Authenticated=true`，否则中间件拒绝
2. **多 spec 必须 RequireAll**：不要假设一个 spec 通过就够了
3. **跨租户操作必须 RequirePlatformRole**：不要依赖 RBAC 资源码（`super_admin` 也是一种 RBAC 资源，但单独 guard 更安全）
4. **DataScope 不能替代 RLS**：DataScope 是业务 SQL WHERE，绕过方法多；RLS 是数据库最后防线
5. **缓存失效要彻底**：改 role → 必须 `InvalidateRole(roleID)` → 所有关联 user 立即重新加载
6. **拒绝默认放行**：中间件不存在 = 默认拒绝（gin 的 route 必须显式注册中间件）

---

## 9. 单元测试参考

permission 中间件 / RBAC 决策 / DataScope SQL 生成 都有单测：

- [`framework/pkg/middleware/auth_test.go`](../framework/pkg/middleware/auth_test.go) — 14 测试
- [`framework/pkg/permission/types_test.go`](../framework/pkg/permission/types_test.go) — 6 测试
- [`framework/pkg/permission/scope_test.go`](../framework/pkg/permission/scope_test.go) — 9 测试
- [`framework/pkg/permission/spec_test.go`](../framework/pkg/permission/spec_test.go) — 4 测试

跑：

```bash
go test -v ./framework/pkg/permission/... ./framework/pkg/middleware/...
```
