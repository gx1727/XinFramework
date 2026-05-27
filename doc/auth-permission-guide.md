# 认证与权限系统使用指南

## 1. 概述

XinFramework 的认证与权限系统基于以下组件：
- **JWT Token** - 用户身份凭证
- **Session Manager** - Session 存储（Redis 或 DB）
- **PermissionService** - 权限数据加载（权限、角色、DataScope）
- **RLS（行级安全）** - 数据库层面的租户隔离

---

## 2. 认证中间件

### 2.1 三种认证中间件

| 中间件 | Token验证 | XinContext | UserContext | 适用场景 |
|--------|-----------|------------|------------|----------|
| `Auth()` | 必需 | 注入 | 懒加载 | 受保护路由（需要权限检查） |
| `AuthLite()` | 必需 | 注入 | 不加载 | 轻量认证（仅需身份识别） |
| `OptionalAuth()` | 可选 | 有则注入 | 有则懒加载 | 公共接口（登录用户看个性化内容） |

### 2.2 使用示例

```go
func setupRouter(app *boot.App) {
    srv := app.Server
    cfg := app.Config

    srv.Engine.Use(middleware.Recovery())
    srv.Engine.Use(middleware.RequestID())
    srv.Engine.Use(middleware.CORS(&cfg.CORS))
    srv.Engine.Use(middleware.Logger())

    v1 := r.Group("/api/v1")

    // 公开路由组（可选认证）
    public := v1.Group("")
    public.Use(middleware.OptionalAuth(&cfg.JWT, app.SessionMgr, app.PermService))

    // 受保护路由组（必须认证）
    protected := v1.Group("")
    protected.Use(middleware.Auth(&cfg.JWT, app.SessionMgr, app.PermService))

    // 注册模块路由
    // ...
}
```

---

## 3. 权限检查

### 3.1 权限格式

```
"resource_code:action"
```

示例：
| 权限 | 说明 |
|------|------|
| `user:list` | 查看用户列表 |
| `user:create` | 创建用户 |
| `user:update` | 修改用户 |
| `user:delete` | 删除用户 |
| `*:*` | 超级管理员（所有权限） |

### 3.2 中间件方式

**单个权限检查：**

```go
protected.GET("/users",
    middleware.RequirePermission(permission.ResUser, permission.ActList),
    h.List,
)

protected.POST("/users",
    middleware.RequirePermission(permission.ResUser, permission.ActCreate),
    h.Create,
)
```

**任意权限检查（满足一个即可）：**

```go
protected.DELETE("/posts/:id",
    middleware.RequireAnyPermission("post:delete", "admin:*"),
    h.Delete,
)
```

**全部权限检查（必须拥有所有）：**

```go
protected.PUT("/roles/:id",
    middleware.RequireAllPermissions("role:update", "permission:assign"),
    h.Update,
)
```

### 3.3 代码方式

```go
func MyHandler(c *gin.Context) {
    xc := xincontext.New(c)
    uc := xc.MustNewUserContext(c.Request.Context())

    // 检查单个权限
    if !uc.HasPermission("post", "create") {
        resp.Forbidden(c, "权限不足")
        return
    }

    // 检查多个权限（AND 逻辑）
    if !uc.HasPermissions("post:update", "post:delete") {
        resp.Forbidden(c, "权限不足")
        return
    }

    // 检查角色
    if uc.IsSuperAdmin() {
        // 超级管理员逻辑
    }

    if uc.HasRole("admin") {
        // 管理员逻辑
    }
}
```

---

## 4. 上下文系统

### 4.1 XinContext（轻量级身份）

始终可用，包含基本身份信息：

```go
type XinContext struct {
    TenantID  uint
    UserID    uint
    SessionID string
    Role      string
}

xc := xincontext.New(c)
userID := xc.GetUserID()
tenantID := xc.GetTenantID()
role := xc.GetRole()
```

### 4.2 UserContext（完整上下文）

懒加载，包含权限和角色信息：

```go
type UserContext struct {
    *XinContext
    OrgID       int64
    Roles       []string
    Permissions map[string]bool
    DataScope   DataScope
}

// 懒加载：首次调用时从数据库加载
uc := xc.MustNewUserContext(ctx)
```

### 4.3 获取当前用户信息

```go
func GetProfile(c *gin.Context) {
    xc := xincontext.New(c)

    // 轻量级信息（始终可用）
    userID := xc.GetUserID()
    tenantID := xc.GetTenantID()

    // 完整上下文（如需要权限）
    uc := xc.MustNewUserContext(c.Request.Context())
    perms := uc.Permissions
    roles := uc.Roles

    resp.Success(c, gin.H{
        "user_id": userID,
        "tenant_id": tenantID,
        "roles": roles,
    })
}
```

---

## 5. 数据权限（DataScope）

### 5.1 DataScope 类型

| 值 | 名称 | 说明 |
|----|------|------|
| 1 | DataScopeAll | 租户内所有数据 |
| 2 | DataScopeCustom | 自定义机构范围 |
| 3 | DataScopeDept | 本部门数据 |
| 4 | DataScopeDeptAndBelow | 本部门及下级数据 |
| 5 | DataScopeSelf | 仅本人数据 |

### 5.2 DataScope 使用

```go
func ListPosts(c *gin.Context) {
    xc := xincontext.New(c)
    uc := xc.MustNewUserContext(c.Request.Context())

    // 根据 DataScope 构建查询条件
    query := "SELECT * FROM posts WHERE tenant_id = $1"
    args := []interface{}{uc.TenantID}

    switch uc.DataScope {
    case permission.DataScopeAll:
        // 无额外条件
    case permission.DataScopeDept:
        // 仅本部门
        query += " AND org_id = $2"
        args = append(args, uc.OrgID)
    case permission.DataScopeSelf:
        // 仅本人
        query += " AND created_by = $2"
        args = append(args, uc.UserID)
    }

    // 执行查询
    // ...
}
```

---

## 6. 租户隔离

### 6.1 RLS（行级安全）

数据库层面的租户隔离，所有含 `tenant_id` 的表都启用了 RLS。

```sql
CREATE POLICY tenant_isolation_policy ON users
    USING (
        tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    );
```

**逻辑：**
- 设置 `app.tenant_id` → 可访问该租户数据
- 未设置 → NULL 比较 → 拒绝所有访问

### 6.2 事务中的租户设置

框架自动在 `db.RunInTenantTx` 中设置租户上下文：

```go
func (h *Handler) GetUser(c *gin.Context) {
    xc := xincontext.New(c)

    var user *User
    err := db.RunInTenantTx(c.Request.Context(), db.Get(), xc.GetTenantID(), func(ctx context.Context) error {
        // 在这里执行的 SQL 会自动带上 RLS 过滤
        return h.repo.GetByID(ctx, req.ID, &user)
    })
}
```

---

## 7. 完整示例

### 7.1 创建一个需要权限检查的 API

```go
// handler.go
package mymodule

import (
    "context"
    "github.com/gin-gonic/gin"
    xincontext "gx1727.com/xin/framework/pkg/context"
    "gx1727.com/xin/framework/pkg/db"
    "gx1727.com/xin/framework/pkg/resp"
)

type Handler struct {
    repo *Repository
}

func NewHandler(repo *Repository) *Handler {
    return &Handler{repo: repo}
}

func (h *Handler) ListItems(c *gin.Context) {
    xc := xincontext.New(c)
    uc := xc.MustNewUserContext(c.Request.Context())

    // 检查权限
    if !uc.HasPermission("item", "list") {
        resp.Forbidden(c, "权限不足")
        return
    }

    var items []Item
    var total int64
    err := db.RunInTenantTx(c.Request.Context(), db.Get(), uc.TenantID, func(ctx context.Context) error {
        var err error
        items, total, err = h.repo.List(ctx, uc, req.Page, req.Size)
        return err
    })

    if err != nil {
        resp.ServerError(c, err.Error())
        return
    }

    resp.Success(c, gin.H{
        "list":  items,
        "total": total,
        "page":  req.Page,
        "size":  req.Size,
    })
}

func (h *Handler) CreateItem(c *gin.Context) {
    xc := xincontext.New(c)
    uc := xc.MustNewUserContext(c.Request.Context())

    // 检查权限
    if !uc.HasPermission("item", "create") {
        resp.Forbidden(c, "权限不足")
        return
    }

    var req createItemRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        resp.BadRequest(c, FormatValidationError(err))
        return
    }

    item := &Item{
        TenantID: uc.TenantID,
        Name:     req.Name,
        // ...
    }

    var result *Item
    err := db.RunInTenantTx(c.Request.Context(), db.Get(), uc.TenantID, func(ctx context.Context) error {
        var err error
        result, err = h.repo.Create(ctx, item)
        return err
    })

    if err != nil {
        resp.HandleError(c, err)
        return
    }

    resp.Success(c, result)
}
```

### 7.2 注册路由

```go
// routes.go
package mymodule

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/internal/core/middleware"
    "gx1727.com/xin/framework/pkg/permission"
)

func Register(protected *gin.RouterGroup, h *Handler) {
    protected.GET("/items",
        middleware.RequirePermission(permission.ResUser, permission.ActList),
        h.ListItems,
    )

    protected.POST("/items",
        middleware.RequirePermission(permission.ResUser, permission.ActCreate),
        h.CreateItem,
    )
}
```

---

## 8. 常见问题

### Q: 权限检查失败的可能原因？

1. **Token 无效或过期** - 检查 `Authorization` header
2. **Session 被撤销** - 检查 SessionManager 状态
3. **用户没有该权限** - 检查 `permissions` 表数据
4. **权限格式错误** - 检查是否为 `resource:action` 格式

### Q: 如何让某个用户成为超级管理员？

在 `users` 表中设置 `role = 'super_admin'`，或者给用户分配 `*:*` 权限。

### Q: RLS 不生效怎么办？

1. 确保在 `db.RunInTenantTx` 内执行操作
2. 检查 `set_config('app.tenant_id', ...)` 是否成功
3. 检查表是否正确启用了 RLS：`ALTER TABLE xxx ENABLE ROW LEVEL SECURITY;`

### Q: 懒加载失败后会重试吗？

不会。`UserContext` 使用 `sync.Once`，如果加载失败，后续调用会返回空的 UserContext 而不是重试。

### Q: 如何调试权限问题？

1. 打印 `uc.Permissions` 查看用户实际拥有的权限
2. 打印 `uc.Roles` 查看用户的角色
3. 打印 `uc.DataScope` 查看数据范围
4. 检查日志中是否有 RLS 相关的 SQL 错误