# 权限系统

> XinFramework 的权限分三层�?*资源�?RBAC**（能不能�?API�?+ **数据范围 DataScope**（能看哪些行�?+ **平台角色 PlatformRole**（跨租户特权）�?>
> 文档版本�?026-06（sys_menu / platform_tenant 模块化后�?breaking change�?
## 1. 三层权限一�?
| �?| 解决什�?| 实现位置 |
|---|---|---|
| **资源�?RBAC** | 这个用户**能不能调**这个 API | `tenant_permissions` + `tenant_role_tenant_permissions` + `Require(spec)` 中间�?|
| **数据范围 DataScope** | 这个用户**能看哪些�?* | `roles.data_scope` + `BuildDataScopeFilter` + 业务 SQL WHERE |
| **平台角色 PlatformRole** | 这个账号**跨租�?*的特�?| `<dropped:account_roles>.role` + `RequirePlatformRole` 中间�?|

**重要**：`super_admin` 平台角色**自动 bypass** 资源�?RBAC（详�?§6）�?
---

## 2. 资源�?RBAC

### 2.1 资源码格�?
`resource:action`，例如：

- `user:list` �?列出用户
- `flag:create` �?创建 flag 业务记录
- `config:update` �?修改配置�?- `tenant:create` �?创建租户（仅 super_admin 域）

支持�?action �?[`framework/pkg/permission/constants.go`](../framework/pkg/permission/constants.go)�?
```go
ActList   = "list"
ActGet    = "get"
ActCreate = "create"
ActUpdate = "update"
ActDelete = "delete"
ActTree   = "tree"
```

支持�?resource�?4 个）�?
```go
ResSystem       = "system"
ResAsset        = "asset"
ResDict         = "dict"
ResTenant       = "tenant"          // 平台租户管理复用
ResOrganization = "organization"
ResResource     = "resource"
ResMenu         = "menu"            // 平台菜单管理复用
ResRole         = "role"
ResUser         = "user"
ResPermission   = "permission"
ResWeixin       = "weixin"
ResAuth         = "auth"
ResFlag         = "flag"
ResConfig       = "config"
```

> **复用原则**：`sys_menu` �?`menu` 共用 `ResMenu`；`platform_tenant` 和未来业务层 `tenant` 共用 `ResTenant`。资源码代表"操作这种资源的权�?，与具体路径无关�?
### 2.2 Spec 类型

[`framework/pkg/permission/spec.go`](../framework/pkg/permission/spec.go) 定义�?
```go
type Spec struct {
    Resource      string
    Action        string
    Authenticated bool    // 默认 true
}
```

构造方式：

```go
spec := permission.P("user", "list")       // resource:action, 默认需要登�?spec := permission.AuthOnly()              // 只需要登录，任何资源都行
```

### 2.3 中间�?
[`framework/pkg/middleware/auth.go`](../framework/pkg/middleware/auth.go) 提供�?
| 函数 | 行为 |
|---|---|
| `Require(spec)` | 必须满足�?spec（单一�?|
| `RequireAny(specs...)` | 任一满足 |
| `RequireAll(specs...)` | 全部满足 |
| `RequireAuthenticated()` | 仅登录（等价 `Require(AuthOnly())`�?|
| `RequirePlatformRole(roles...)` | 必须持有平台角色（详�?§6�?|

### 2.4 使用示例

```go
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

    protected.DELETE("/admin",
        middleware.RequireAny(
            permission.P(permission.ResUser, permission.ActDelete),
            permission.P(permission.ResRole, permission.ActDelete),
        ),
        h.Delete)

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
// 优先�?1：精确匹�?perms["user:list"] = true

// 优先�?2：资源级通配
perms["user:*"] = true      // 任意 user:xxx 都通过

// 优先�?3：全局通配（super_admin�?perms["*:*"] = true         // 任意 任意 都通过
```

| 用户有权�?| `user:list` | `role:list` |
|---|---|---|
| `{"user:list": true}` | �?| �?|
| `{"user:*": true}` | ✅（2 级） | �?|
| `{"*:*": true}` | ✅（3 级） | �?|

**特别说明**：`*:*` 等价�?`super_admin`，在中间件层直接 `c.Next()`，不需要查数据库�?
---

## 3. 数据范围 DataScope

资源码回�?能不能调"，数据范围回�?能看哪些�?�?
### 3.1 五种类型

[`framework/pkg/permission/types.go`](../framework/pkg/permission/types.go)�?
```go
const (
    DataScopeAll          DataScopeType = 1
    DataScopeCustom       DataScopeType = 2
    DataScopeDept         DataScopeType = 3
    DataScopeDeptAndBelow DataScopeType = 4
    DataScopeSelf         DataScopeType = 5
)
```

### 3.2 存储

`roles.data_scope` 字段�?JSONB�?
```sql
ALTER TABLE roles ADD COLUMN data_scope JSONB DEFAULT '{"type": 5}';
-- 默认每个人只能看自己创建的数�?```

```json
{"type": 1}
{"type": 2, "org_ids": [3,5,7]}
{"type": 3}
{"type": 4}
{"type": 5}
```

### 3.3 编译期类�?
```go
type DataScope struct {
    Type   DataScopeType `json:"type"`
    OrgIDs []int64       `json:"org_ids,omitempty"`
}
```

### 3.4 编译期过滤器

[`framework/pkg/permission/scope.go`](../framework/pkg/permission/scope.go)�?
```go
func BuildDataScopeFilter(
    ds DataScope,
    userID uint,
    orgID int64,
    columns ScopeColumns,
) (ScopeFilter, error)
```

### 3.5 五种类型生成�?SQL

默认 `ScopeColumns{CreatorID: "creator_id", OrgID: "org_id"}`�?
| Type | SQL | Args |
|---|---|---|
| `DataScopeAll` (1) | `""` | `nil` |
| `DataScopeCustom` (2) �?`org_ids=[]` | `"creator_id = $1"` | `[userID]` |
| `DataScopeCustom` (2) �?`org_ids=[3,5]` | `"org_id = ANY($1)"` | `[[3,5]]` |
| `DataScopeDept` (3) �?`org_id=0` | `"creator_id = $1"` | `[userID]` |
| `DataScopeDept` (3) �?`org_id=7` | `"org_id = $1"` | `[7]` |
| `DataScopeDeptAndBelow` (4) �?`org_id=7` | 递归 CTE | `[7]` |
| `DataScopeSelf` (5) | `"creator_id = $1"` | `[userID]` |

**DeptAndBelow 的递归 SQL**�?
```sql
org_id = $1
OR org_id IN (
    WITH RECURSIVE org_tree AS (
        SELECT id FROM tenant_organizations WHERE id = $1
        UNION ALL
        SELECT o.id FROM tenant_organizations o
        JOIN org_tree ot ON o.parent_id = ot.id
    )
    SELECT id FROM org_tree
)
```

### 3.6 在业务里�?
```go
filter, _ := uc.GetDataScopeFilter()

sql := `SELECT ... FROM avatars WHERE ` + filter.SQL
args := append(filter.Args, size, (page-1)*size)
sql += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`
```

或者自定义列名�?
```go
filter, _ := uc.GetDataScopeFilterFor(permission.ScopeColumns{
    SelfColumn: "user_id",
    OrgID:      "department_id",
})
```

---

## 4. 权限加载

[`framework/pkg/permission/permission_impl.go`](../framework/pkg/permission/permission_impl.go) 实现�?
### 4.1 GetUserPermissions

```sql
SELECT DISTINCT res.code, res.action
FROM tenant_users u
JOIN tenant_user_roles ur ON ur.user_id = u.id
JOIN roles rol ON rol.id = ur.role_id
JOIN tenant_role_resources_OLD rr ON rr.role_id = rol.id
JOIN resources_OLD res ON res.id = rr.resource_id
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

[`framework/pkg/permission/permission_cache.go`](../framework/pkg/permission/permission_cache.go) �?Redis 缓存权限 + DataScope�?
```go
uc := &UserContext{
    Permissions: perms,
    DataScope:   ds,
    Roles:       roles,
}
```

每次"懒加�?只查一�?DB（`/auth` 注入 loader 后第一�?`MustNewUserContext` 触发），后续�?Redis cache�?
**失效**：role / permission 模块变更时调 `authz.InvalidateRole(roleID)` 清缓存，�?§5�?
---

## 5. 缓存失效

[`framework/internal/service/authorization_service.go`](../framework/internal/service/authorization_service.go) 提供 `Authorization` interface，apps 通过 `AppContext.Authz()` 拿到�?
### 5.1 接口

```go
type Authorization interface {
    InvalidateRole(ctx context.Context, roleID uint) error
    InvalidateUser(ctx context.Context, userID uint) error
}
```

### 5.2 何时�?
| 操作 | 调用 |
|---|---|
| 创建/更新/删除 role | `authz.InvalidateRole(roleID)` |
| �?role 加减 resource | `authz.InvalidateRole(roleID)` |
| �?user 加减 role | `authz.InvalidateUser(userID)` |
| �?role �?data_scope | `authz.InvalidateRole(roleID)` |

调用示例�?[`apps/tenant/role/service.go`](../apps/tenant/role/service.go)�?
```go
func (s *Service) Update(ctx context.Context, roleID uint, req UpdateRoleReq) error {
    if err := s.repo.Update(ctx, roleID, req); err != nil { return err }
    if s.authz != nil {
        _ = s.authz.InvalidateRole(context.Background(), roleID)
    }
    return nil
}
```

注意 `context.Background()`：缓存失效是后台任务，不应绑请求生命周期�?
---

## 6. 平台角色

### 6.1 概念

平台角色�?*跨租�?*的特权，绑定�?`accounts` 而非 `tenant_users`。典型用途：

- `super_admin`：平台超级管理员，可以管理任意租�?/ 平台菜单 / 平台配置
- 未来可能�?`auditor`（审计员）、`support`（客服）�?
### 6.2 存储

```sql
CREATE TABLE <dropped:account_roles> (
    account_id BIGINT NOT NULL,
    role      VARCHAR(64) NOT NULL,
    ...
);
```

### 6.3 颁发

```sql
INSERT INTO <dropped:account_roles> (account_id, role) VALUES ($1, 'super_admin')
ON CONFLICT (account_id, role) DO NOTHING;
```

### 6.4 �?token 中传�?
`accounts.id` 登录后，框架�?`<dropped:account_roles>` 找出所有平台角色，塞进 JWT claims�?
```go
type Claims struct {
    UserID        uint
    TenantID      uint
    SessionID     string
    Role          string
    PlatformRoles []string
}
```

### 6.5 中间�?
`RequirePlatformRole("super_admin")` 校验 `claims.PlatformRoles` 是否包含任一指定角色�?
### 6.6 自动 bypass RBAC

`super_admin` **自动** bypass 所�?`Require(spec)` 检查：

```go
// 平台超级管理员：无视所有权限规格直接放�?if uc.IsSuperAdmin() {
    c.Next()
    return
}
```

### 6.7 当前用法

平台管理模块（`/api/v1/platform/*`）都强制 `RequirePlatformRole("super_admin")`�?
| 模块 | 路由前缀 | 双层守卫 |
|---|---|---|
| `platform_tenant` | `/platform/platform-tenants` | super_admin + `ResTenant.*` |
| `sys_menu` | `/platform/platform-tenant_menus` | super_admin（单层） |
| `config` 平台�?| `/configs/platform` | super_admin + `ResConfig.*` |
| `dict` 平台�?| `/dicts/platform` | super_admin + `ResDict.*` |

挂载方式�?
```go
adminGroup := protected.Group("/admin",
    pkgmiddleware.RequirePlatformRole("super_admin"))
{
    g := adminGroup.Group("/platform-menus")
    g.GET("", h.List)
}
```

---

## 7. 完整流程�?
```
请求进来
   �?[Recovery �?RequestID �?CORS �?ClientIP �?Logger]   �?全局中间�?   �?[Auth/OptionalAuth]
   ├─ 解析 JWT �?claims
   ├─ 验签 �?失败 401
   ├─ �?session �?失效 401
   ├─ 注入 XinContext{ UserID, TenantID, PlatformRoles }
   └─ 注册 UserContextLoader（懒�?   �?[RequirePlatformRole] (可�? 平台管理路由�?
   └─ 检�?PlatformRoles �?403
   �?[Require(spec) / RequireAny / RequireAll]
   ├─ IsSuperAdmin() �?放行
   ├─ IsAuthOnly()  �?放行
   └─ HasPermission �?403
   �?Handler 业务逻辑
   ├─ MustNewUserContext(c)  �?触发 loader �?�?cache �?�?DB
   ├─ RunInTenantTx / RunInPlatformTx
   ├─ 业务 SQL 拼接 data_scope filter
   ├─ JSONB �?SQL `::jsonb` cast
   └─ resp.OK / resp.Error
```

---

## 9. 安全原则

1. **白名单优于黑名单**：默�?spec `Authenticated=true`，否则中间件拒绝
2. **�?spec 必须 RequireAll**：不要假设一�?spec 通过就够�?3. **跨租户操作必�?RequirePlatformRole**：不要依�?RBAC 资源�?4. **DataScope 不能替代 RLS**：DataScope 是业�?SQL WHERE，绕过方法多；RLS 是数据库最后防�?5. **缓存失效要彻�?*：改 role �?必须 `InvalidateRole(roleID)` �?所有关�?user 立即重新加载
6. **拒绝默认放行**：中间件不存�?= 默认拒绝

---

## 9. 单元测试参�?
- [`framework/pkg/middleware/auth_test.go`](../framework/pkg/middleware/auth_test.go) �?14 测试
- [`framework/pkg/permission/types_test.go`](../framework/pkg/permission/types_test.go) �?6 测试
- [`framework/pkg/permission/scope_test.go`](../framework/pkg/permission/scope_test.go) �?9 测试
- [`framework/pkg/permission/spec_test.go`](../framework/pkg/permission/spec_test.go) �?4 测试

```bash
go test -v ./framework/pkg/permission/... ./framework/pkg/middleware/...
```