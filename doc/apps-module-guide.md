# Apps 模块开发指南

## 概述

本文档说明如何在 `apps/` 目录下创建外部插件模块。

## 🎯 核心设计原则

在最新的 XinFramework 架构中，Apps 模块的开发必须遵循**统一事务与上下文传递**的原则。

**核心规则**：
- ✅ **业务编排层 (Handler / Service)**：必须使用 `db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx) error)` 定义事务边界，以触发 RLS（行级安全）和租户隔离。
- ✅ **数据访问层 (Repository)**：必须使用 `db.GetQuerier(ctx)` 获取执行器，无论是操作 Framework 的表（users, tenants）还是 Apps 自己的表（cms_posts, flags）。绝对禁止在业务层直接使用全局 `db.Get()` 执行跨表操作。

---

## 架构设计原则

1. **Framework 不依赖 Apps**：framework 是核心，apps 是插件
2. **纯粹的 Repository**：数据访问层必须是纯粹的，只做增删改查，不负责事务的 Begin/Commit，也不负责租户环境变量的修改。
3. **闭包事务**：通过闭包将事务和上下文向下穿透。

---

## 开发规范与示例

以 `flag` 模块为例，说明标准的开发流程。

### 步骤 1：定义 Repository 层

Repository 只负责数据访问，使用 `db.GetQuerier` 自动提取上下文事务。

```go
// apps/flag/avatar_repository.go
package flag

import (
    "context"
    "gx1727.com/xin/framework/pkg/db"
)

type AvatarRepository struct{}

func NewAvatarRepository() *AvatarRepository {
    return &AvatarRepository{}
}

func (r *AvatarRepository) GetByID(ctx context.Context, id uint) (*Avatar, error) {
    // 自动获取外层事务或连接池
    q, err := db.GetQuerier(ctx)
    if err != nil {
        return nil, err
    }

    var a Avatar
    err = q.QueryRow(ctx, "SELECT id, name FROM flag_avatars WHERE is_deleted = FALSE AND id = $1", id).Scan(&a.ID, &a.Name)
    return &a, err
}

func (r *AvatarRepository) Create(ctx context.Context, a *Avatar) error {
    q, err := db.GetQuerier(ctx)
    if err != nil {
        return err
    }

    // 插入时不需要显式指定 tenant_id（如果是在 RunInTenantTx 闭包中，通常已经由框架自动处理或需要业务层赋值）
    _, err = q.Exec(ctx, "INSERT INTO flag_avatars (name, tenant_id) VALUES ($1, $2)", a.Name, a.TenantID)
    return err
}
```

### 步骤 2：实现 Handler/Service 层（事务控制）

所有的业务逻辑，包括多步操作，都在 `db.RunInTenantTx` 的闭包中进行。

```go
// apps/flag/handler.go
package flag

import (
    "context"
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/db"
    "gx1727.com/xin/framework/pkg/resp"
    xincontext "gx1727.com/xin/framework/pkg/context"
)

type Handler struct {
    avatarRepo *AvatarRepository
}

func NewHandler() *Handler {
    return &Handler{
        avatarRepo: NewAvatarRepository(),
    }
}

func (h *Handler) CreateAvatar(c *gin.Context) {
    uc := xincontext.NewUserContext(c)
    
    var req CreateAvatarReq
    if err := c.ShouldBindJSON(&req); err != nil {
        resp.BadRequest(c, err.Error())
        return
    }

    // 核心：使用 RunInTenantTx 开启租户事务
    err := db.RunInTenantTx(c.Request.Context(), db.Get(), uc.TenantID, func(ctx context.Context) error {
        avatar := &Avatar{Name: req.Name, TenantID: uc.TenantID}
        // 这里的 Create 会自动复用上述事务，通过 RLS 安全检查
        return h.avatarRepo.Create(ctx, avatar)
    })

    if err != nil {
        resp.HandleError(c, err)
        return
    }

    resp.Success(c, nil)
}
```

### 步骤 3：定义路由与入口

```go
// apps/flag/module.go
package flag

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/plugin"
)

func Register(public *gin.RouterGroup, protected *gin.RouterGroup) {
    h := NewHandler()
    
    api := protected.Group("/flag")
    {
        api.POST("/avatars", h.CreateAvatar)
    }
}

func Module() plugin.Module {
    return plugin.NewModule("flag", Register)
}
```

---

## 常见问题

### Q1: 如果我要在 Apps 模块中查询 Framework 的核心表（如 users）怎么办？

**A**: 在您的 Apps Repository 中，同样使用 `db.GetQuerier(ctx)` 去执行 SQL 即可。只要 Handler 层包裹了 `db.RunInTenantTx`，对 `users` 表的查询同样受限于 RLS 策略，只能查出当前租户的用户，完全安全。

```go
func (r *MyRepo) GetUserCode(ctx context.Context, userID uint) (string, error) {
    q, err := db.GetQuerier(ctx)
    if err != nil {
        return "", err
    }
    var code string
    // 由于在 RunInTenantTx 中，如果 userID 不属于当前租户，将查不到数据（ErrNoRows）
    err = q.QueryRow(ctx, "SELECT code FROM users WHERE is_deleted = FALSE AND id = $1", userID).Scan(&code)
    return code, err
}
```

### Q2: 为什么以前可以直接 `pool := db.Get(); pool.QueryRow(...)`，现在不行了？

**A**: 以前的写法会导致 **RLS（行级安全）策略拦截**。
当系统启用 RLS 后，如果不注入 `app.tenant_id`，数据库会拒绝所有对租户表的读写操作（表现为查询返回空或插入报错）。
通过 `db.RunInTenantTx` + `db.GetQuerier(ctx)` 的组合，框架会自动在底层执行 `SET app.tenant_id = ?`，确保数据库的 RLS 校验能正确通过，并彻底防止越权访问。

### Q3: 软删除记录（is_deleted = TRUE）如何处理？

**A**: 在最新的架构中，软删除属于**纯业务逻辑**。
- **查询时**：请在 Repo 的 SQL 中显式加上 `WHERE is_deleted = FALSE`。
- **软删除时**：只需执行 `UPDATE xxx SET is_deleted = TRUE`，因为我们已经从 RLS 策略中移除了对 `is_deleted` 的限制，所以它会顺畅执行，不会触发 RLS 新行违规报错。

---

## 优势

✅ **绝对的数据安全**：强制使用闭包事务触发 RLS，杜绝了串租户、越权访问的可能。  
✅ **更清晰的职责分离**：Repository 不再包含事务控制代码，纯粹处理 SQL，逻辑极为精简。  
✅ **原子性保障**：Service/Handler 层可以轻松地将多个 Repo 方法组合在一个事务中。  
