# Apps 模块开发指南

## 概述

本文档说明如何在 `apps/` 目录下创建外部插件模块。

**当前架构**：扁平化结构（Handler → Repository），无 Service 层。

---

## 1. 标准项目结构

```
apps/{name}/
├── module.go          # 模块定义和注册
├── routes.go          # 路由注册
├── handler.go         # HTTP 处理器（业务逻辑）
├── repository.go      # 数据访问层
├── types.go           # 类型定义
├── errors.go          # 错误定义
├── helpers.go         # 辅助函数
└── go.mod
```

---

## 2. 开发流程

以 `flag` 模块为例，说明标准的开发流程。

### 步骤 1：定义 Types（types.go）

```go
package flag

import "time"

// Frame 相框
type Frame struct {
    ID           uint      `json:"id"`
    TenantID     uint      `json:"tenant_id"`
    CategoryID   uint      `json:"category_id"`
    Name         string    `json:"name"`
    Description  string    `json:"description,omitempty"`
    PreviewURL   string    `json:"preview_url,omitempty"`
    TemplateURL  string    `json:"template_url,omitempty"`
    TemplateConfig string  `json:"template_config,omitempty"`
    Type         int       `json:"type"`
    Sort         int       `json:"sort"`
    Status       int       `json:"status"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

### 步骤 2：定义 Repository（*_repository.go）

Repository 只负责数据访问，使用 `db.GetQuerier(ctx)` 自动适配事务上下文。

```go
package flag

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5"
    "gx1727.com/xin/framework/pkg/db"
)

// FrameRepository 相框数据访问层
type FrameRepository struct {
    db *pgxpool.Pool
}

func NewFrameRepository(pool *pgxpool.Pool) *FrameRepository {
    return &FrameRepository{db: pool}
}

// GetByID 查询单个
func (r *FrameRepository) GetByID(ctx context.Context, id uint) (*Frame, error) {
    q, err := db.GetQuerier(ctx)
    if err != nil {
        return nil, err
    }

    var f Frame
    var description, previewURL *string
    err = q.QueryRow(ctx, `
        SELECT id, tenant_id, category_id, name, description, preview_url, type, sort, status
        FROM flag_frames
        WHERE is_deleted = FALSE AND id = $1`, id).Scan(
        &f.ID, &f.TenantID, &f.CategoryID, &f.Name, &description, &previewURL,
        &f.Type, &f.Sort, &f.Status,
    )
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrFrameNotFound
        }
        return nil, err
    }
    // 处理 NULL 指针
    if description != nil {
        f.Description = *description
    }
    if previewURL != nil {
        f.PreviewURL = *previewURL
    }
    return &f, nil
}

// Create 创建
func (r *FrameRepository) Create(ctx context.Context, frame *Frame) (*Frame, error) {
    q, err := db.GetQuerier(ctx)
    if err != nil {
        return nil, err
    }

    var f Frame
    err = q.QueryRow(ctx, `
        INSERT INTO flag_frames (tenant_id, category_id, name, description, preview_url, type, sort, status)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, tenant_id, category_id, name, description, preview_url, type, sort, status, created_at, updated_at`,
        frame.TenantID, frame.CategoryID, frame.Name, nullStr(frame.Description),
        nullStr(frame.PreviewURL), frame.Type, frame.Sort, frame.Status,
    ).Scan(&f.ID, &f.TenantID, &f.CategoryID, &f.Name, &description, &previewURL,
        &f.Type, &f.Sort, &f.Status, &f.CreatedAt, &f.UpdatedAt)
    if err != nil {
        return nil, fmt.Errorf("create frame: %w", err)
    }
    return &f, nil
}
```

### 步骤 3：定义 Handler（handler.go）

Handler 负责业务逻辑和事务控制。

```go
package flag

import (
    "context"
    "github.com/gin-gonic/gin"
    xincontext "gx1727.com/xin/framework/pkg/context"
    "gx1727.com/xin/framework/pkg/db"
    "gx1727.com/xin/framework/pkg/resp"
)

type Handler struct{}

func NewHandler() *Handler {
    return &Handler{}
}

func (h *Handler) GetFrame(c *gin.Context) {
    uc := xincontext.NewUserContext(c)
    var req getFrameRequest
    if err := c.ShouldBindUri(&req); err != nil {
        resp.BadRequest(c, FormatValidationError(err))
        return
    }

    var frame *Frame
    // 使用 RunInTenantTx 提供租户上下文（触发 RLS）
    err := db.RunInTenantTx(c.Request.Context(), db.Get(), uc.TenantID, func(ctx context.Context) error {
        var err error
        frame, err = frameRepo.GetByID(ctx, req.ID)
        return err
    })

    if err != nil {
        resp.HandleError(c, err)
        return
    }

    resp.Success(c, frame)
}

func (h *Handler) CreateFrame(c *gin.Context) {
    uc := xincontext.NewUserContext(c)
    if uc.TenantID == 0 {
        resp.Unauthorized(c, "未登录")
        return
    }

    var req createFrameRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        resp.BadRequest(c, FormatValidationError(err))
        return
    }

    frame := &Frame{
        TenantID:   uc.TenantID,
        CategoryID: req.CategoryID,
        Name:       req.Name,
        // ...
    }

    var result *Frame
    err := db.RunInTenantTx(c.Request.Context(), db.Get(), uc.TenantID, func(ctx context.Context) error {
        var err error
        result, err = frameRepo.Create(ctx, frame)
        return err
    })

    if err != nil {
        resp.HandleError(c, err)
        return
    }

    resp.Success(c, result)
}
```

### 步骤 4：定义 Routes（routes.go）

```go
package flag

import "github.com/gin-gonic/gin"

func Register(public *gin.RouterGroup, protected *gin.RouterGroup, h *Handler) {
    // 公开路由（可选认证）
    public.GET("/flag/frames", h.ListFrames)
    public.GET("/flag/frames/:id", h.GetFrame)

    // 受保护路由（需要认证）
    protected.POST("/flag/frames", h.CreateFrame)
    protected.PUT("/flag/frames/:id", h.UpdateFrame)
    protected.DELETE("/flag/frames/:id", h.DeleteFrame)
}
```

### 步骤 5：定义 Module（module.go）

```go
package flag

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/db"
    "gx1727.com/xin/framework/pkg/plugin"
)

type module struct {
    name string
}

func (m *module) Name() string { return m.name }
func (m *module) Init() error   { return nil }
func (m *module) Shutdown() error { return nil }

func (m *module) Register(public *gin.RouterGroup, protected *gin.RouterGroup) {
    // 初始化 Repository（持有 db pool 引用）
    InitRepositories(db.Get())

    // 创建 Handler（直接调用 Repository，无 Service 层）
    h := NewHandler()
    Register(public, protected, h)
}

func Module() plugin.Module {
    return &module{name: "flag"}
}
```

### 步骤 6：注册模块（init.go）

```go
package flag

func init() {
    plugin.Register(Module())
}
```

---

## 3. 核心规则

### 3.1 事务边界

| 层级 | 职责 | 规则 |
|------|------|------|
| **Handler** | 业务逻辑、事务控制 | 必须使用 `db.RunInTenantTx` 定义事务边界 |
| **Repository** | 数据访问 | 使用 `db.GetQuerier(ctx)` 获取查询器 |

### 3.2 租户上下文传递

```
HTTP 请求
    ↓
Auth 中间件（注入 XinContext）
    ↓
Handler（通过 XinContext 获取 TenantID）
    ↓
db.RunInTenantTx（设置 app.tenant_id 触发 RLS）
    ↓
Repository（使用 db.GetQuerier 获取已设置租户上下文的事务）
    ↓
PostgreSQL RLS 策略自动过滤
```

### 3.3 禁止事项

| 禁止 | 原因 |
|------|------|
| Repository 层直接使用 `db.Get()` | 无法适配事务上下文 |
| Handler 层不使用 `RunInTenantTx` | RLS 不生效，租户隔离失效 |
| 在 RLS 策略中加入 `is_deleted = FALSE` | 软删除时会失败 |

---

## 4. 辅助函数（helpers.go）

```go
package flag

func nullStr(s string) *string {
    if s == "" {
        return nil
    }
    return &s
}

func nilIfZero(n uint) *uint {
    if n == 0 {
        return nil
    }
    return &n
}
```

---

## 5. 错误处理（errors.go）

```go
package flag

import "errors"

var (
    ErrFrameNotFound      = errors.New("frame not found")
    ErrCategoryNotFound   = errors.New("category not found")
    ErrSpaceNotFound      = errors.New("space not found")
    // ...
)
```

---

## 6. 验证清单

创建新模块时，按以下清单检查：

- [ ] `module.go` - 实现 Module 接口
- [ ] `routes.go` - 定义公开和受保护路由
- [ ] `handler.go` - 使用 `db.RunInTenantTx` 处理事务
- [ ] `*_repository.go` - 使用 `db.GetQuerier(ctx)`
- [ ] `init()` - 调用 `plugin.Register(Module())`
- [ ] `config.yaml` - 在 `apps:` 中启用

---

## 7. 参考

- Flag 模块：`apps/flag/`
- CMS 模块：`apps/cms/`