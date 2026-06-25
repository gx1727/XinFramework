# 权限模型

> 本文件描述 XinFramework 的 RBAC + DataScope + 平台角色三层权限模型。
> 中间件 / Spec / Action 详见 [api.md §10](./api.md#10-资源权限码)。

---

## 1. 三层权限

| 层 | 作用 | 存哪 | 检查时机 |
|---|---|---|---|
| **JWT 身份** | 你是谁 | JWT Claims | 每个请求的 Auth 中间件 |
| **RBAC 资源权限** | 你能操作哪些资源 | `tenant_role_resources` / `sys_role_permissions` | `Require(spec)` 中间件 |
| **DataScope 数据范围** | 你能看到哪些行 | `tenant_roles.data_scope` + `tenant_role_data_scopes` | Repository 查询时 `xincontext.ScopeFilterFrom` |
| **平台角色** | 你是不是 super_admin | `sys_user_roles` → `sys_roles.code` | `RequirePlatformRole` 中间件 |

**关系**：
- 平台角色是**正交维度**：与租户内 RBAC 独立，跨租户特权
- RBAC 控制"操作什么"，DataScope 控制"看到什么行"
- 短路：`super_admin` 或 `*:*` 通配 → 跳过 RBAC + DataScope

---

## 2. 资源 × 操作矩阵

`framework/pkg/permission/constants.go`：

### 2.1 资源（Resource）

```go
const (
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
    ResConfig       = "config"
)
```

### 2.2 操作（Action）

```go
const (
    ActList   = "list"
    ActGet    = "get"
    ActCreate = "create"
    ActUpdate = "update"
    ActDelete = "delete"
    ActTree   = "tree"     // 树形结构专用
)
```

### 2.3 Spec 构造（编译期拼写检查）

```go
spec := permission.P(permission.ResUser, permission.ActDelete)  // = "user:delete"
```

**好处**：避免字符串拼写错误（`"user:deltet"` 编译期不报红，运行期 403）。

### 2.4 路由用法

```go
g.GET("/users", middleware.Require(permission.P(permission.ResUser, permission.ActList)), h.List)
g.POST("/users", middleware.Require(permission.P(permission.ResUser, permission.ActCreate)), h.Create)
g.PUT("/users/:id", middleware.Require(permission.P(permission.ResUser, permission.ActUpdate)), h.Update)
g.DELETE("/users/:id", middleware.Require(permission.P(permission.ResUser, permission.ActDelete)), h.Delete)
```

---

## 3. MatchMode

`framework/pkg/permission/spec.go`：

```go
type MatchMode int
const (
    MatchAll MatchMode = iota  // 全部通过
    MatchAny                    // 任一通过
)

func Require(specs ...Spec) gin.HandlerFunc     // 默认 MatchAll
func RequireAny(specs ...Spec) gin.HandlerFunc  // MatchAny
func RequireAll(specs ...Spec) gin.HandlerFunc  // 同 Require
func RequireAuthenticated() gin.HandlerFunc      // 只校验登录
```

---

## 4. DataScope 数据范围

`framework/pkg/permission/types.go`：

| 值 | 名称 | 含义 |
|---|---|---|
| 1 | ALL | 全部 |
| 2 | CUSTOM | 自定义 org 列表 |
| 3 | DEPT | 本部门 |
| 4 | DEPT_AND_BELOW | 本部门及以下 |
| 5 | SELF | 本人 |

**作用**：控制 Repository 查询时 WHERE 子句的过滤条件。

### 4.1 用法

```go
// Repository 内部
type UserQuery struct {
    Keyword string
    OrgID   uint
    Status  int
}

func (r *UserRepository) List(ctx context.Context, q UserQuery) ([]User, error) {
    // 1. 拼业务过滤
    where := []string{"is_deleted = FALSE"}
    args := []any{}
    if q.Keyword != "" { where = append(where, "real_name ILIKE $1"); args = append(args, "%"+q.Keyword+"%") }

    // 2. ctx-aware 注入 DataScope 过滤（一行调用）
    filter, err := xincontext.ScopeFilterFrom(ctx, permission.ScopeColumns{
        SelfColumn: "u.id",
        OrgID:      "u.org_id",
    })
    if err != nil {
        return nil, err
    }
    if !filter.IsEmpty() {
        where = append(where, filter.SQL)
        args  = append(args,  filter.Args...)
    }

    sql := "SELECT * FROM tenant_users WHERE " + strings.Join(where, " AND ")
    // ...
}
```

> `xincontext.ScopeFilterFrom` 会自动从 ctx 取 `UserContext.DataScope`，无需手工取 `UserContext`。
> 旧 API `xincontext.UserContext.DataScopeFilterFor` 仍可用，但已 Deprecated。

### 4.2 SQL 生成

`xincontext.ScopeFilterFrom(ctx, columns) (ScopeFilter, error)` 返回：

| DataScope | 生成的 SQL |
|---|---|
| 1 (ALL) | `""` (无过滤) |
| 2 (CUSTOM) | `org_id IN (?, ?, ...)` — 来自 `tenant_role_data_scopes` |
| 3 (DEPT) | `org_id = ?` — 用户主部门 |
| 4 (DEPT_AND_BELOW) | `org_id IN (递归子部门 ids)` |
| 5 (SELF) | `id = ?` — 用户本人 |

---

## 5. 平台角色

### 5.1 角色定义

`sys_roles` 表存平台角色：

| code | 含义 |
|---|---|
| `super_admin` | 平台级超级管理员，跨租户特权 |

`sys_user_roles` 关联平台用户与平台角色。

### 5.2 JWT 携带

```go
type Claims struct {
    ...
    PlatformRoles []string  // ["super_admin", ...]
}
```

登录时查询 `sys_user_roles` → `sys_roles.code` 填入 `PlatformRoles`。

### 5.3 中间件

```go
g := protected.Group("/tenants",
    pkgmiddleware.RequirePlatformRole(PlatformRoleSuperAdmin),
)
```

`RequirePlatformRole(roles...)` 检查 `XinContext.PlatformRoles` 包含任一 role。

### 5.4 双重防御

```go
g.POST("/tenants", 
    pkgmiddleware.RequirePlatformRole(PlatformRoleSuperAdmin),  // 1. 平台角色
    pkgmiddleware.Require(permission.P(permission.ResTenant, permission.ActCreate)),  // 2. 资源权限
    h.Create)
```

**即使持有 `super_admin`，仍需满足资源权限码**——两个守卫都过才算合法。避免任一 tenant admin 仅凭资源权限码越权。

### 5.5 短路

`requireWithSpecs` 在 `XinContext.HasPlatformRole(super_admin)` 时**短路放行**——但实际上双层守卫更严格。建议**关键操作（删除租户、purge）**显式双层守卫，**普通查询**可以靠 super_admin 短路。

---

## 6. 权限缓存

### 6.1 加载时机

每个请求 `Auth` 中间件：

1. 解析 JWT → Claims
2. 检查 `session.Validate(SessionID)` → session 存活
3. 注入 `XinContext`
4. 注册 `UserContext` 懒加载器（**不立即查 DB**）

### 6.2 懒加载

业务代码首次访问权限时：

```go
uc := xincontext.NewUserContext(c)
perms, err := uc.LoadPermissions(ctx)  // 第一次查 DB
```

后续访问复用缓存。

### 6.3 失效

`authz.Authorization` 接口：

```go
type Authorization interface {
    LoadPermissions(ctx, userID) (map[string]bool, error)
    LoadRoles(ctx, userID) ([]string, error)
    LoadDataScope(ctx, userID) (*permission.DataScope, error)
    LoadUserSecurityContext(ctx, userID) (map[string]bool, []string, *permission.DataScope, int64, error)
    InvalidateUser(ctx, userID) error
    InvalidateRole(ctx, roleID) error
    InvalidateResource(ctx, resourceID) error
}
```

修改角色权限时（如 `PUT /roles/:id/permissions`），调 `InvalidateUser(userID)` 或 `InvalidateRole(roleID)` 清缓存。

**当前实现**：`framework/internal/service/AuthorizationService` 用 in-memory map，可换 Redis。

---

## 7. 通配符

| 模式 | 匹配 |
|---|---|
| `user:list` | 单个 spec |
| `user:*` | 一个资源的所有操作 |
| `*:list` | 所有资源的单个操作 |
| `*:*` | 所有资源所有操作（admin 默认） |

**匹配算法**（`requireWithSpecs`）：

```go
func matchSpec(userPerm string, required string) bool {
    if userPerm == required { return true }
    if userPerm == "*:*" { return true }
    
    userRes, userAct := splitSpec(userPerm)
    reqRes, reqAct := splitSpec(required)
    
    if userRes == "*" || userRes == reqRes {
        if userAct == "*" || userAct == reqAct {
            return true
        }
    }
    return false
}
```

**注意**：通配符只在"用户拥有"的权限码里生效。Spec 不会自己支持 `user:*` 写法——但 `userPerm='*:*'` 会短路放行。

---

## 8. 权限码的存储与查询

### 8.1 租户域

`tenant_role_resources` (role_id, permission_id) → `tenant_permissions` (code)

查询 SQL（`authz.AuthorizationService.LoadPermissions`）：

```sql
SELECT p.code
FROM tenant_role_resources rr
JOIN tenant_permissions p ON p.id = rr.permission_id AND p.is_deleted = FALSE
WHERE rr.role_id IN (
    SELECT role_id FROM tenant_user_roles
    WHERE user_id = $1 AND is_deleted = FALSE
) AND rr.is_deleted = FALSE AND rr.effect = 1
UNION
SELECT '*' WHERE EXISTS (
    SELECT 1 FROM tenant_user_roles ur
    JOIN tenant_role_resources rr ON rr.role_id = ur.role_id
    JOIN tenant_permissions p ON p.id = rr.permission_id
    WHERE ur.user_id = $1 AND p.code = '*' AND ur.is_deleted = FALSE AND rr.is_deleted = FALSE
);
```

### 8.2 平台域

`sys_role_permissions` → `sys_permissions`

类似，作用于 `sys_user_roles`。

---

## 9. 配置中心的特殊规则

`config_items` 的 `is_public` 字段决定是否可被公开访问：

```go
// 中间件：OptionalAuth，未登录也能读 is_public=true 的项
public.GET("/configs", h.GetPublicConfigs)
```

`is_readonly` / `is_system` 字段控制 UI 行为（前端可读但不能改）。

---

## 10. 多身份账号的权限

`auth/login-precheck` 列出所有身份；`select-tenant` / `tenant-login` 后 JWT 的 `Role` 字段是该身份的 tenant role code。

**切换租户**：`refresh_token` + `tenant_id` → 新 JWT。`XinContext.PlatformRoles` 不变（平台角色是账号级，与租户无关）。

---

## 11. 角色继承 / 互斥

**当前不支持**：
- 无角色继承树（每个角色独立）
- 无角色互斥（如 admin 与 guest 不能同存）
- 无父子资源权限

需要这些特性时，建议在 `tenant_role_resources` 上加：
- `parent_role_id`（继承来源）
- `exclude_role_ids []`（互斥）

---

## 12. 审计关联

权限修改操作（创建/删除/绑定）记 `db_logs`：

```go
audit.Log(ctx, pool, audit.Entry{
    Action:    "role:permission_change",
    TableName: "tenant_role_resources",
    RecordID:  roleID,
    OldData:   map[string]any{"permissions": oldPerms},
    NewData:   map[string]any{"permissions": newPerms},
})
```

`actor_id` 从 `XinContext.UserID` 自动取。

---

## 13. 常见模式

### 13.1 业务 Handler 检查

```go
func (h *Handler) Delete(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
    
    // 1. 业务校验
    if id == 1 {
        resp.Error(c, 4001, "不能删除 admin")
        return
    }
    
    // 2. 调 service（service 内调 DataScope 过滤）
    if err := h.svc.Delete(ctx, uc, id); err != nil {
        resp.HandleError(c, err)
        return
    }
    
    // 3. 审计
    audit.Log(ctx, h.pool, audit.Entry{
        Action:    "user:delete",
        TableName: "tenant_users",
        RecordID:  uint(id),
    })
    
    resp.Success(c, gin.H{"ok": true})
}
```

### 13.2 Repository 注入 DataScope

```go
func (r *UserRepository) List(ctx context.Context, uc *xincontext.UserContext, q ListQuery) ([]User, error) {
    var (
        conditions []string
        args       []any
    )
    conditions = append(conditions, "is_deleted = FALSE")
    conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", len(args)+1))
    args = append(args, uc.TenantID)
    
    // DataScope 过滤
    if filter, err := xincontext.ScopeFilterFrom(ctx, permission.ScopeColumns{SelfColumn: "u.id", OrgID: "u.org_id"}); err != nil {
        return err
    } else if !filter.IsEmpty() {
        conditions = append(conditions, filter.SQL)
        args = append(args, filter.Args...)
    }
    
    sql := "SELECT * FROM tenant_users WHERE " + strings.Join(conditions, " AND ")
    // ...
}
```

### 13.3 前端按钮级

```tsx
<Auth permission="user:create">
  <Button onClick={handleAdd}>添加用户</Button>
</Auth>
```

`<Auth>` 仅控制 UI 显示；后端 `Require(spec)` 强制。

---

## 14. 调试技巧

```go
// 中间件顺序（debug 时打印）
r.Use(middleware.Recovery())
r.Use(middleware.RequestID())
r.Use(middleware.CORS(...))
r.Use(middleware.ClientIP())
r.Use(middleware.Logger())  // 看每条请求的 method / path / status / duration
```

```sql
-- 查看某用户实际拥有的权限码
SELECT u.real_name, p.code
FROM tenant_users u
JOIN tenant_user_roles ur ON ur.user_id = u.id AND ur.is_deleted = FALSE
JOIN tenant_role_resources rr ON rr.role_id = ur.role_id AND rr.is_deleted = FALSE
JOIN tenant_permissions p ON p.id = rr.permission_id AND p.is_deleted = FALSE
WHERE u.id = 1;
```

```bash
# 调试 403
curl -i http://localhost:8087/api/v1/users
# → 检查响应 code 是 4001 (无权限) 还是 4002 (角色冲突) 还是 1001 (token 失效)
```
