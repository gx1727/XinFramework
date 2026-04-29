# Apps 模块开发指南（简化版）

## 概述

本文档说明如何在 `apps/` 目录下创建外部插件模块。

## 🎯 核心设计原则

**简化模式**：
- ✅ **所有数据访问** → 使用全局 `db.Get()`
  - Framework 的表（users, tenants, roles 等）
  - Apps 自己的表（cms_posts, flags 等）
  - 直接使用 SQL 查询

---

## 快速开始

### 步骤 1：创建模块目录结构

```
apps/mymodule/
├── internal/
│   ├── handler/
│   │   └── handler.go
│   └── service/
│       └── service.go
├── module.go
├── routes.go
├── go.mod
└── go.sum
```

### 步骤 2：实现 Service 层

```go
// apps/mymodule/internal/service/service.go
package service

import (
    "context"
    "fmt"
    
    "gx1727.com/xin/framework/pkg/db"
    "gx1727.com/xin/framework/pkg/model"
)

type Service struct{}

func NewService() *Service {
    return &Service{}
}

// 访问 Framework 的表（如 users）
func (s *Service) GetUser(ctx context.Context, userID uint) (*model.User, error) {
    pool := db.Get()
    if pool == nil {
        return nil, fmt.Errorf("database not initialized")
    }
    
    var user model.User
    err := pool.QueryRow(ctx,
        "SELECT id, tenant_id, code, real_name FROM users WHERE id = $1",
        userID,
    ).Scan(&user.ID, &user.TenantID, &user.Code, &user.RealName)
    if err != nil {
        return nil, fmt.Errorf("get user: %w", err)
    }
    return &user, nil
}

// 访问自己的表
func (s *Service) GetPost(ctx context.Context, id uint) (*model.CmsPost, error) {
    pool := db.Get()
    if pool == nil {
        return nil, fmt.Errorf("database not initialized")
    }
    
    var post model.CmsPost
    err := pool.QueryRow(ctx,
        "SELECT id, title, content FROM my_posts WHERE id = $1",
        id,
    ).Scan(&post.ID, &post.Title, &post.Content)
    if err != nil {
        return nil, fmt.Errorf("get post: %w", err)
    }
    return &post, nil
}
```

### 步骤 3：实现 Handler 层

```go
// apps/mymodule/internal/handler/handler.go
package handler

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/resp"
    "gx1727.com/xin/module/mymodule/internal/service"
)

type Handler struct {
    svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
    return &Handler{svc: svc}
}

func (h *Handler) GetUser(c *gin.Context) {
    userID := parseUserID(c)
    
    user, err := h.svc.GetUser(c.Request.Context(), userID)
    if err != nil {
        resp.Error(c, 500, err.Error())
        return
    }
    
    resp.Success(c, user)
}
```

### 步骤 4：定义路由

```go
// apps/mymodule/routes.go
package mymodule

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/module/mymodule/internal/handler"
)

func Register(h *handler.Handler, public, protected *gin.RouterGroup) {
    api := protected.Group("/mymodule")
    {
        api.GET("/user/:id", h.GetUser)
    }
}
```

### 步骤 5：创建 Module 入口

```go
// apps/mymodule/module.go
package mymodule

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/plugin"
    "gx1727.com/xin/module/mymodule/internal/handler"
    "gx1727.com/xin/module/mymodule/internal/service"
)

var (
    myService *service.Service
    myHandler *handler.Handler
)

func init() {
    // 模块加载时初始化
    myService = service.NewService()
    myHandler = handler.NewHandler(myService)
}

// Module 返回插件模块
func Module() plugin.Module {
    return plugin.NewModule("mymodule", func(public, protected *gin.RouterGroup) {
        Register(myHandler, public, protected)
    })
}
```

### 步骤 6：在 main.go 中注册

```go
// cmd/xin/main.go
import "gx1727.com/xin/apps/mymodule"

var moduleRegistry = map[string]func() plugin.Module{
    "mymodule": mymodule.Module,
}
```

---

## 完整示例：CMS 模块

查看 [`apps/cms/`](file:///D:/work/xin/XinFramework/apps/cms) 作为参考：

- [`module.go`](file:///D:/work/xin/XinFramework/apps/cms/module.go) - 模块入口
- [`internal/service/service.go`](file:///D:/work/xin/XinFramework/apps/cms/internal/service/service.go) - 业务逻辑
- [`internal/handler/handler.go`](file:///D:/work/xin/XinFramework/apps/cms/internal/handler/handler.go) - HTTP 处理
- [`routes.go`](file:///D:/work/xin/XinFramework/apps/cms/routes.go) - 路由定义

---

## 常见问题

### Q1: 如何访问 Framework 的用户信息？

```go
func (s *Service) GetUserByCode(ctx context.Context, code string) (*model.User, error) {
    pool := db.Get()
    var user model.User
    err := pool.QueryRow(ctx,
        "SELECT id, tenant_id, code, real_name FROM users WHERE code = $1",
        code,
    ).Scan(&user.ID, &user.TenantID, &user.Code, &user.RealName)
    return &user, err
}
```

### Q2: 如何实现分页查询？

```go
func (s *Service) ListPosts(ctx context.Context, page, size int) ([]model.CmsPost, int64, error) {
    pool := db.Get()
    offset := (page - 1) * size
    
    // 查询总数
    var total int64
    pool.QueryRow(ctx, "SELECT COUNT(*) FROM posts").Scan(&total)
    
    // 查询列表
    rows, _ := pool.Query(ctx, 
        "SELECT id, title FROM posts ORDER BY created_at DESC LIMIT $1 OFFSET $2",
        size, offset,
    )
    // ... 处理结果
}
```

### Q3: 需要使用事务怎么办？

```go
func (s *Service) CreatePostWithTags(ctx context.Context, post *model.CmsPost, tags []string) error {
    pool := db.Get()
    
    // 开启事务
    tx, err := pool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)
    
    // 插入文章
    _, err = tx.Exec(ctx, "INSERT INTO posts ...")
    if err != nil {
        return err
    }
    
    // 插入标签
    for _, tag := range tags {
        _, err = tx.Exec(ctx, "INSERT INTO post_tags ...")
        if err != nil {
            return err
        }
    }
    
    // 提交事务
    return tx.Commit(ctx)
}
```

---

## 优势

✅ **简单直接**：不需要依赖注入，代码更简洁  
✅ **灵活自由**：可以自由编写 SQL 查询  
✅ **易于理解**：没有复杂的抽象层  
✅ **性能可控**：可以直接优化 SQL  

## 注意事项

⚠️ **SQL 注入防护**：始终使用参数化查询（`$1`, `$2`）  
⚠️ **错误处理**：检查 `db.Get()` 是否返回 nil  
⚠️ **资源释放**：记得关闭 `rows.Close()`  
⚠️ **模型字段**：确保 SQL 查询的字段与 model 结构体匹配
