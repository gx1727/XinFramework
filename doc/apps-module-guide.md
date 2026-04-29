# Apps 模块开发指南

## 概述

本文档说明如何在 `apps/` 目录下创建外部插件模块。

## 🎯 核心设计原则

**简化模式（推荐）**：
- ✅ **所有数据访问** → 使用全局 `db.Get()`
  - Framework 的表（users, tenants, roles 等）
  - Apps 自己的表（cms_posts, flags 等）
  - 直接使用 SQL 查询

---

## 架构设计原则

1. **Framework 不依赖 Apps**：framework 是核心，apps 是插件
2. **依赖注入**：framework 向 apps 提供所需的服务
3. **接口隔离**：apps 通过 `pkg/model` 中的接口访问数据，不直接依赖实现

---

## 可用方式

### ✅ 方式 1：使用全局数据库连接（简单场景）

适用于简单的、独立的模块（如 Flag 模块）。

```go
// apps/flag/module.go
package flag

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/db"
    "gx1727.com/xin/framework/pkg/plugin"
)

func (m *module) Register(public *gin.RouterGroup, protected *gin.RouterGroup) {
    // 直接使用全局 db 连接创建自己的 Repository
    pool := db.Get()
    frameRepo = NewFrameRepository(pool)
    
    h := NewHandler()
    Register(public, protected, h)
}
```

**优点**：
- 简单直接
- 不需要 framework 提供额外支持

**缺点**：
- 需要自己实现 Repository
- 无法复用 framework 的 Repository

**适用场景**：
- 模块有自己独特的数据模型
- 不需要访问 framework 的核心实体（User, Tenant 等）

---

### ✅ 方式 2：混合模式（推荐）

结合两种方式的优点：
- Framework 的功能 → 依赖注入
- Apps 自己的数据 → 全局 `db.Get()`

#### 步骤 1：在 apps 模块中定义 Service

```go
// apps/cms/internal/service/service.go
package service

import (
    "context"
    "fmt"
    
    "gx1727.com/xin/framework/pkg/db"
    "gx1727.com/xin/framework/pkg/model"
)

type Service struct {
    // Framework 的 Repository → 通过依赖注入
    userRepo   model.UserRepository
    tenantRepo model.TenantRepository
    
    // CMS 自己的表 → 直接使用 db.Get()
    // 不需要在这里声明 postRepo
}

// NewService 只接收 Framework 的 Repository
func NewService(
    userRepo model.UserRepository,
    tenantRepo model.TenantRepository,
) *Service {
    return &Service{
        userRepo:   userRepo,
        tenantRepo: tenantRepo,
    }
}

// 使用 Framework 的 User Repository（依赖注入）
func (s *Service) GetUserByCode(ctx context.Context, code string) (*model.User, error) {
    return s.userRepo.GetByCode(ctx, code)
}

// 使用 CMS 自己的表（全局 db.Get()）
func (s *Service) GetPost(ctx context.Context, id uint) (*model.CmsPost, error) {
    pool := db.Get()
    if pool == nil {
        return nil, fmt.Errorf("database not initialized")
    }
    
    var post model.CmsPost
    err := pool.QueryRow(ctx,
        "SELECT id, tenant_id, title, content FROM cms_posts WHERE id = $1",
        id,
    ).Scan(&post.ID, &post.TenantID, &post.Title, &post.Content)
    if err != nil {
        return nil, fmt.Errorf("get post: %w", err)
    }
    return &post, nil
}
```

#### 步骤 1：在 apps 模块中定义 Service 接收依赖

```go
// apps/cms/internal/service/service.go
package service

import (
    "context"
    "gx1727.com/xin/framework/pkg/model"
)

type Service struct {
    userRepo   model.UserRepository
    tenantRepo model.TenantRepository
    postRepo   model.CmsPostRepository
}

// 构造函数接收 Repository 依赖
func NewService(
    userRepo model.UserRepository,
    tenantRepo model.TenantRepository,
    postRepo model.CmsPostRepository,
) *Service {
    return &Service{
        userRepo:   userRepo,
        tenantRepo: tenantRepo,
        postRepo:   postRepo,
    }
}

// 通过 user_code 获取用户信息
func (s *Service) GetUserByCode(ctx context.Context, code string) (*model.User, error) {
    return s.userRepo.GetByCode(ctx, code)
}
```

#### 步骤 2：在 module.go 中暴露初始化函数

```go
// apps/cms/module.go
package cms

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/model"
    "gx1727.com/xin/framework/pkg/plugin"
    "gx1727.com/xin/module/cms/internal/handler"
    "gx1727.com/xin/module/cms/internal/service"
)

var (
    cmsService *service.Service
    cmsHandler *handler.Handler
)

type module struct {
    name string
}

func (m *module) Name() string     { return m.name }
func (m *module) Init() error      { return nil }
func (m *module) Shutdown() error  { return nil }

func (m *module) Register(public *gin.RouterGroup, protected *gin.RouterGroup) {
    if cmsHandler != nil {
        Register(cmsHandler, public, protected)
    }
}

// InitService 由 framework 调用，只注入 Framework 的 Repository
// CMS 自己的表（CmsPost）直接使用 db.Get()
func InitService(
    userRepo model.UserRepository,
    tenantRepo model.TenantRepository,
) {
    cmsService = service.NewService(userRepo, tenantRepo)
    cmsHandler = handler.NewHandler(cmsService)
}

func Module() plugin.Module {
    return &module{name: "cms"}
}
```

#### 步骤 3：在 framework.go 中调用初始化

```go
// framework/framework.go

import (
    "gx1727.com/xin/apps/cms"  // 导入 CMS 模块
)

func initExternalModuleDeps(app *boot.App) {
    // 为 CMS 模块注入 Framework 的 Repository
    if app.Config.AppEnabled("cms") {
        cms.InitService(
            app.Repository.User(),      // Framework 的 User → 依赖注入
            app.Repository.Tenant(),    // Framework 的 Tenant → 依赖注入
            // CmsPost 不需要注入，CMS 内部直接使用 db.Get()
        )
    }
}
```

#### 步骤 4：在 Handler 中使用

```go
// apps/cms/internal/handler/handler.go
package handler

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/context"
    "gx1727.com/xin/framework/pkg/resp"
)

type Handler struct {
    svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
    return &Handler{svc: svc}
}

// 通过 user_code 获取用户信息的 API
func (h *Handler) GetUserByCode(c *gin.Context) {
    code := c.Query("code")
    if code == "" {
        resp.BadRequest(c, "user code is required")
        return
    }
    
    user, err := h.svc.GetUserByCode(c.Request.Context(), code)
    if err != nil {
        resp.HandleError(c, err)
        return
    }
    
    resp.Success(c, user)
}
```

**优点**：
- 完全解耦，符合依赖倒置原则
- 可以复用 framework 的所有 Repository
- 便于测试（可以 mock Repository）

**缺点**：
- 需要在 framework.go 中手动配置依赖注入

**适用场景**：
- 需要访问 framework 核心实体（User, Tenant, Role 等）
- 复杂的业务模块

---

## 实际示例：通过 user_code 获取用户信息

### 在 CMS 模块中实现

```go
// apps/cms/internal/service/service.go

func (s *Service) GetCurrentUserProfile(ctx context.Context, userID uint) (*UserProfile, error) {
    // 1. 通过 ID 获取用户
    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("get user by id: %w", err)
    }
    
    // 2. 也可以通过 code 获取用户
    // user, err := s.userRepo.GetByCode(ctx, userCode)
    
    // 3. 获取租户信息
    tenant, err := s.tenantRepo.GetByID(ctx, user.TenantID)
    if err != nil {
        return nil, fmt.Errorf("get tenant: %w", err)
    }
    
    return &UserProfile{
        User:   user,
        Tenant: tenant,
    }, nil
}
```

### 在 Flag 模块中实现（使用自己的 Repository）

如果 Flag 模块需要获取用户信息，有两种选择：

**选择 1**：直接使用 `db.Get()` 查询

```go
// apps/flag/handler.go
func (h *Handler) GetUserFlag(c *gin.Context) {
    userID := xincontext.New(c).UserID
    
    // 直接使用 SQL 查询
    pool := db.Get()
    var user model.User
    err := pool.QueryRow(c.Request.Context(), 
        "SELECT id, code, nickname FROM users WHERE id = $1", userID).Scan(
        &user.ID, &user.Code, &user.Nickname,
    )
    if err != nil {
        resp.HandleError(c, err)
        return
    }
    
    // 使用 user.Code ...
}
```

**选择 2**：改为依赖注入方式（参考 CMS 模块）

---

## 总结

### 🎯 推荐架构：混合模式

| 数据类型 | 访问方式 | 示例 |
|---------|---------|------|
| **Framework 核心实体** | 依赖注入 | User, Tenant, Role, Menu, Resource, Organization |
| **Apps 自定义表** | 全局 `db.Get()` | CmsPost, Flag, Avatar, Frame |

### 实现要点

1. **Service 构造函数只接收 Framework 的 Repository**
   ```go
   func NewService(
       userRepo model.UserRepository,    // Framework → 注入
       tenantRepo model.TenantRepository, // Framework → 注入
   ) *Service
   ```

2. **在 Service 方法中直接使用 `db.Get()` 访问自己的表**
   ```go
   func (s *Service) GetPost(ctx context.Context, id uint) (*model.CmsPost, error) {
       pool := db.Get()  // Apps 自己的表 → 全局访问
       // ... SQL 查询
   }
   ```

3. **在 framework.go 中配置依赖注入**
   ```go
   func initExternalModuleDeps(app *boot.App) {
       if app.Config.AppEnabled("cms") {
           cms.InitService(
               app.Repository.User(),    // 只注入 Framework 的
               app.Repository.Tenant(),  // 只注入 Framework 的
           )
       }
   }
   ```

### 优势

✅ **清晰的职责分离**：Framework 和 Apps 的数据访问方式明确  
✅ **避免过度注入**：不需要为每个自定义表创建 Repository 接口  
✅ **灵活性高**：Apps 可以自由管理自己的数据模型  
✅ **易于维护**：Framework 的核心功能通过接口解耦，便于测试

---

## 常见问题

### Q1: 为什么不能直接在 apps 中 import internal/repository？

A: Go 的 `internal` 目录只能被同父目录下的包导入。apps 和 framework 是独立的模块，不能互相访问 internal。

### Q2: pkg/model 中的接口是谁实现的？

A: 由 `framework/internal/repository` 中的具体实现类实现，例如 `PostgresUserRepository` 实现了 `model.UserRepository`。

### Q3: 如何添加新的 Repository 到依赖注入？

A: 
1. 在 `pkg/model/interfaces.go` 中定义接口
2. 在 `internal/repository/` 中实现
3. 在 `internal/repository/provider.go` 中添加 getter
4. 在 `initExternalModuleDeps()` 中注入到需要的模块
