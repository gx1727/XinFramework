# Apps 模块开发指南（简化版）

## 概述

本文档说明如何在 `apps/` 目录下快速创建一个轻量级的外部插件模块（如 Flag 模块）。

## 🎯 核心架构原则

在最新架构中，即便是最简单的模块，也**必须**遵守事务和 RLS 上下文的传递规范。

**核心规则**：
- **业务入口**：在 Handler 中使用 `db.RunInTenantTx` 开启事务闭包。
- **数据访问**：在 Repository 中使用 `db.GetQuerier(ctx)` 执行 SQL。

---

## 快速开始

### 步骤 1：创建模块目录结构

```
apps/mymodule/
├── repository.go    # 纯粹的数据访问层
├── handler.go       # 业务逻辑与 HTTP 层
├── module.go        # 插件入口
└── routes.go        # 路由注册
```

### 步骤 2：实现 Repository 层

```go
// apps/mymodule/repository.go
package mymodule

import (
    "context"
    "gx1727.com/xin/framework/pkg/db"
)

type Repository struct{}

func NewRepository() *Repository {
    return &Repository{}
}

// 访问 Framework 的表（如 users）
func (r *Repository) GetUser(ctx context.Context, userID uint) (string, error) {
    q, err := db.GetQuerier(ctx)
    if err != nil {
        return "", err
    }
    
    var name string
    err = q.QueryRow(ctx, "SELECT real_name FROM users WHERE is_deleted = FALSE AND id = $1", userID).Scan(&name)
    return name, err
}

// 访问自己的表
func (r *Repository) GetPost(ctx context.Context, id uint) (string, error) {
    q, err := db.GetQuerier(ctx)
    if err != nil {
        return "", err
    }
    
    var title string
    err = q.QueryRow(ctx, "SELECT title FROM my_posts WHERE is_deleted = FALSE AND id = $1", id).Scan(&title)
    return title, err
}
```

### 步骤 3：实现 Handler 层

```go
// apps/mymodule/handler.go
package mymodule

import (
    "context"
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/db"
    "gx1727.com/xin/framework/pkg/resp"
    xincontext "gx1727.com/xin/framework/pkg/context"
)

type Handler struct {
    repo *Repository
}

func NewHandler() *Handler {
    return &Handler{
        repo: NewRepository(),
    }
}

func (h *Handler) GetUserInfo(c *gin.Context) {
    uc := xincontext.NewUserContext(c)
    userID := parseUserID(c) // 假设存在该函数
    
    var name string
    // 使用 RunInTenantTx 注入上下文，以穿透 users 表的 RLS
    err := db.RunInTenantTx(c.Request.Context(), db.Get(), uc.TenantID, func(ctx context.Context) error {
        var err error
        name, err = h.repo.GetUser(ctx, userID)
        return err
    })

    if err != nil {
        resp.Error(c, 500, err.Error())
        return
    }
    
    resp.Success(c, name)
}
```

### 步骤 4：定义路由与入口

```go
// apps/mymodule/module.go
package mymodule

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/plugin"
)

func Register(public *gin.RouterGroup, protected *gin.RouterGroup) {
    h := NewHandler()
    
    api := protected.Group("/mymodule")
    {
        api.GET("/user/:id", h.GetUserInfo)
    }
}

func Module() plugin.Module {
    return plugin.NewModule("mymodule", Register)
}
```

---

## 注意事项

⚠️ **为什么不直接使用 `pool := db.Get()`？**
如果你直接使用全局 `db.Get()` 执行查询，会导致查询操作没有租户上下文，进而被底层数据库的 **RLS（行级安全策略）**拦截，返回空结果或报错。

⚠️ **软删除处理**
软删除记录只需在 Repo 层 SQL 加 `WHERE is_deleted = FALSE`。更新删除状态只需 `UPDATE ... SET is_deleted = TRUE` 即可，RLS 策略已彻底与 `is_deleted` 解耦。
