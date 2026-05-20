# 认证与权限系统使用指南

## 快速开始

### 1. 基础认证中间件

#### 完整认证（包含权限加载）

```go
import (
    "gx1727.com/xin/framework/internal/core/middleware"
    "gx1727.com/xin/framework/pkg/config"
    "gx1727.com/xin/framework/pkg/session"
)

// 初始化
cfg := config.GetJWTConfig()
sm := session.NewRedisSessionManager() // 或 NewDBSessionManager
permSvc := service.NewPermissionService(permRepo, dsRepo, cache)

// 使用 Auth 中间件 - 会懒加载权限数据
router.Use(middleware.Auth(cfg, sm, permSvc))

// 需要权限检查的路由
protected := router.Group("/api")
protected.GET("/users", 
    middleware.RequirePermission("user", "list"),
    handler.ListUsers,
)
```

#### 轻量级认证（仅身份验证）

```go
// 使用 AuthLite - 只验证 Token，不加载权限
publicRouter.Use(middleware.AuthLite(cfg, sm))

// 个性化但不需要权限检查的接口
publicRouter.GET("/profile", handler.GetMyProfile)
```

#### 可选认证

```go
// OptionalAuth - Token 有效则注入上下文，无效也继续执行
publicRouter.Use(middleware.OptionalAuth(cfg, sm, permSvc))

// 公共接口，登录用户看到个性化内容
publicRouter.GET("/articles", handler.ListArticles)
```

---

### 2. 权限检查

#### 单个权限检查

```go
router.GET("/users", 
    middleware.RequirePermission("user", "list"),
    handler.ListUsers,
)

router.POST("/users", 
    middleware.RequirePermission("user", "create"),
    handler.CreateUser,
)
```

#### 多权限检查（任意一个）

```go
router.DELETE("/posts/:id",
    middleware.RequireAnyPermission("post:delete", "admin:*"),
    handler.DeletePost,
)
```

#### 多权限检查（全部需要）

```go
router.PUT("/roles/:id",
    middleware.RequireAllPermissions("role:update", "permission:assign"),
    handler.UpdateRole,
)
```

---

### 3. 在 Handler 中使用 Context

#### 获取用户身份信息

```go
func GetProfile(c *gin.Context) {
    // 获取轻量级上下文（总是可用）
    xc := xinContext.New(c)
    
    userID := xc.GetUserID()
    tenantID := xc.GetTenantID()
    
    // 查询用户资料...
}
```

#### 获取完整权限上下文

```go
func ListSensitiveData(c *gin.Context) {
    // 获取完整上下文（触发懒加载）
    uc := xinContext.MustNewUserContext(c)
    
    // 检查权限
    if !uc.HasPermission("data", "view") {
        resp.Forbidden(c, "no permission")
        return
    }
    
    // 获取数据范围过滤条件
    whereSQL, args, _ := uc.GetDataScopeFilter()
    
    // 构建查询
    query := "SELECT * FROM sensitive_data"
    if whereSQL != "" {
        query += " WHERE " + whereSQL
    }
    
    rows, _ := db.Query(query, args...)
    // ...
}
```

---

### 4. 数据范围控制

系统支持五种数据访问范围：

| 类型 | 说明 | SQL 示例 |
|------|------|----------|
| DataScopeAll | 全部数据 | 无过滤 |
| DataScopeSelf | 本人数据 | `creator_id = $1` |
| DataScopeDept | 本部门数据 | `org_id = $1` |
| DataScopeDeptAndBelow | 本部门及以下 | CTE 递归查询 |
| DataScopeCustom | 自定义机构列表 | `org_id = ANY($1)` |

#### 使用示例

```go
func ListOrders(c *gin.Context) {
    uc := xinContext.MustNewUserContext(c)
    
    query := "SELECT * FROM orders"
    whereSQL, args, _ := uc.GetDataScopeFilter()
    
    if whereSQL != "" {
        query += " WHERE " + whereSQL
        rows, _ := db.Query(query, args...)
    } else {
        rows, _ := db.Query(query)
    }
    
    // 处理结果...
}
```

---

### 5. 手动权限检查

#### 在 Service 层检查权限

```go
type UserService struct {
    permSvc *service.PermissionService
}

func (s *UserService) CanViewUser(ctx context.Context, userID uint, targetUserID uint) (bool, error) {
    // 检查是否有查看用户的权限
    hasPerm, err := s.permSvc.HasPermission(ctx, userID, "user", "view")
    if err != nil {
        return false, err
    }
    
    if !hasPerm {
        return false, nil
    }
    
    // 检查数据范围
    whereSQL, args, _ := s.permSvc.BuildDataScopeSQL(ctx, userID)
    
    // 验证目标用户是否在允许范围内
    var count int
    query := "SELECT COUNT(*) FROM users WHERE id = $1"
    if whereSQL != "" {
        query += " AND " + whereSQL
        args = append([]any{targetUserID}, args...)
    }
    
    err = db.QueryRow(query, args...).Scan(&count)
    return count > 0, err
}
```

---

### 6. 缓存管理

#### 手动失效缓存

```go
func UpdateUserRole(userID uint, newRole string) error {
    // 更新数据库...
    
    // 失效用户权限缓存
    permSvc.InvalidateUser(context.Background(), userID)
    
    return nil
}
```

#### 批量失效（建议在角色权限变更时）

```go
func UpdateRolePermissions(roleID uint, perms []Permission) error {
    // 更新数据库...
    
    // 获取该角色的所有用户
    userIDs, _ := permRepo.GetUserIDsByRole(roleID)
    
    // 批量失效
    for _, userID := range userIDs {
        permSvc.InvalidateUser(context.Background(), userID)
    }
    
    return nil
}
```

---

## 最佳实践

### 1. 选择合适的认证中间件

| 场景 | 推荐中间件 | 原因 |
|------|-----------|------|
| 管理后台 API | `Auth` | 需要完整权限控制 |
| 公开接口但需用户身份 | `AuthLite` | 性能更好，无需权限加载 |
| 混合接口（登录/游客） | `OptionalAuth` | 灵活支持两种模式 |

### 2. 权限粒度设计

```go
// ✅ 好的设计 - 细粒度权限
"user:list"
"user:create"
"user:update"
"user:delete"

// ❌ 避免 - 过于粗粒度
"user:all"
```

### 3. 使用通配符简化配置

```go
// 超级管理员
"*:*"

// 模块管理员
"user:*"     // 用户模块所有操作
"order:*"    // 订单模块所有操作
```

### 4. 数据范围与权限结合

```go
// 先检查功能权限
if !uc.HasPermission("order", "view") {
    return Forbidden
}

// 再应用数据范围过滤
whereSQL, args, _ := uc.GetDataScopeFilter()
query := "SELECT * FROM orders"
if whereSQL != "" {
    query += " WHERE " + whereSQL
}
```

### 5. 错误处理

```go
// ✅ 推荐 - 使用 MustNewUserContext 捕获配置错误
func Handler(c *gin.Context) {
    defer func() {
        if r := recover(); r != nil {
            // 记录日志
            log.Printf("UserContext error: %v", r)
            resp.Error(c, "internal error")
        }
    }()
    
    uc := xinContext.MustNewUserContext(c)
    // 使用 uc...
}

// ❌ 避免 - 忽略错误
uc := xinContext.NewUserContext(c)  // 可能返回空上下文
```

---

## 性能优化建议

### 1. 使用 AuthLite 减少不必要的数据库查询

```go
// 不需要权限检查的接口
publicGroup.Use(middleware.AuthLite(cfg, sm))
publicGroup.GET("/settings", handler.GetSettings)  // 快 30-50ms
```

### 2. 合理设置缓存 TTL

```go
// 根据业务特点调整
- 频繁变更的权限：TTL = 1 分钟
- 稳定的权限：TTL = 10 分钟
- 几乎不变的权限：TTL = 1 小时
```

### 3. 监控缓存命中率

```go
// 定期收集指标
metrics := permSvc.CollectMetrics()
log.Printf("Cache hit rate: %.2f%%", metrics["cache_hit_rate"])

// 命中率 < 80% 时考虑：
// 1. 增加 TTL
// 2. 预加载热点用户权限
// 3. 检查缓存失效策略
```

---

## 常见问题

### Q1: 什么时候会触发 UserContext 加载？

**A**: 当调用以下方法时：
- `xinContext.MustNewUserContext(c)`
- `xinContext.UserContextFrom(c.Request.Context())`
- 任何需要访问 `uc.Permissions`、`uc.Roles`、`uc.DataScope` 的场景

### Q2: 如何避免重复加载权限？

**A**: 
1. 使用缓存（已内置）
2. 在同一请求中复用 UserContext 实例
3. 对于不需要权限的接口使用 `AuthLite`

### Q3: 权限变更后多久生效？

**A**: 
- 默认取决于缓存 TTL
- 可手动调用 `InvalidateUser` 立即生效
- 建议实现自动失效机制（见优化文档）

### Q4: 如何实现行级权限控制？

**A**: 使用数据范围：
```go
whereSQL, args, _ := uc.GetDataScopeFilter()
query := "SELECT * FROM posts WHERE status = 'published'"
if whereSQL != "" {
    query += " AND " + whereSQL
}
```

---

## 迁移指南

### 从旧版本迁移

如果你之前直接使用 JWT Claims，现在可以：

```go
// 旧代码
token := c.GetHeader("Authorization")
claims, _ := jwt.Validate(token, cfg)
userID := claims.UserID

// 新代码
xc := xinContext.New(c)
userID := xc.GetUserID()
```

### 从手动权限检查迁移

```go
// 旧代码
perms, _ := permRepo.GetUserPermissions(ctx, userID)
if !perms["user:list"] {
    return Forbidden
}

// 新代码
uc := xinContext.MustNewUserContext(c)
if !uc.HasPermission("user", "list") {
    return Forbidden
}
```

---

## 示例项目结构

```
cmd/server/
  main.go              # 初始化中间件
internal/
  handler/
    user_handler.go    # 使用 xinContext
  service/
    user_service.go    # 使用 PermissionService
  middleware/
    auth.go            # 认证中间件（框架提供）
pkg/
  context/
    context.go         # Context 管理（框架提供）
  permission/
    types.go           # 权限类型定义（框架提供）
```

---

## 参考资料

- [认证与权限系统优化总结](./auth-permission-optimization.md)
- [Framework 开发手册](./handbook.md)
- [数据库约定](./database-conventions.md)
