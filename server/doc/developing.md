# 开发指南:新增一个业务模块

> 本文档用一个具体例子带你走完新增模块的全流程。示例:加一个 `feedback`(用户反馈) 模块。

## 0. 前提:理解现有架构

读这些文档后再继续:

1. [architecture.md](architecture.md) — 了解 AppContext / Module 接口 / Init / Register 流程
2. [modules.md](modules.md) — 看现有模块的结构
3. [database.md](database.md) — 了解 RLS / 软删除 / 索引约定

## 1. 标准 8 步流程

```
1. SQL 迁移           migrations/feedback.sql
2. 公共接口定义       framework/pkg/rbac/{feedback}.go(可选)
3. 业务模块           apps/feedback/{handler,service,repository,model,module,routes}.go
4. 错误码             apps/feedback/errors.go
5. 在 main.go 注册    cmd/xin/main.go
6. 在 cfg.Module 启用 config/config.yaml
7. 资源码 seed(可选)  migrations/feedback.sql 末尾 INSERT INTO resources
8. 单元测试(可选)     apps/feedback/*_test.go
```

下面逐步展开。

---

## 2. SQL 迁移 (Step 1)

新建 `migrations/feedback.sql`:

```sql
-- ============================================
-- Feedback 模块表
-- ============================================

CREATE TABLE IF NOT EXISTS feedbacks
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,
    creator_id BIGINT      NOT NULL,
    title      VARCHAR(128) NOT NULL,
    content    TEXT         NOT NULL,
    status     SMALLINT    DEFAULT 1,           -- 1=待处理 2=处理中 3=已处理
    reply      TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);

-- RLS
ALTER TABLE feedbacks ENABLE ROW LEVEL SECURITY;
ALTER TABLE feedbacks FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON feedbacks
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- 索引
CREATE INDEX IF NOT EXISTS idx_feedbacks_tenant_status ON feedbacks (tenant_id, status)
    WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_feedbacks_creator ON feedbacks (creator_id)
    WHERE is_deleted = FALSE;

-- Seed:资源码
INSERT INTO resources (code, action, name, menu_id, status, is_deleted)
VALUES
    ('feedback', 'list',   '查看反馈', NULL, 1, FALSE),
    ('feedback', 'create', '提交反馈', NULL, 1, FALSE),
    ('feedback', 'update', '处理反馈', NULL, 1, FALSE),
    ('feedback', 'delete', '删除反馈', NULL, 1, FALSE)
ON CONFLICT (code, action) DO NOTHING;
```

> 软删除 + RLS + 物化索引 + 资源码 seed 是规范。

---

## 3. 公共接口定义 (Step 2,可选)

如果你的模块要给其他模块提供 Repository(跨模块消费),在 `framework/pkg/rbac/` 里加个窄接口文件:

```go
// framework/pkg/rbac/feedback.go
package rbac

type FeedbackRepository interface {
    List(ctx context.Context, tenantID, creatorID uint, page, size int) ([]Feedback, int64, error)
    GetByID(ctx context.Context, id uint) (*Feedback, error)
    Create(ctx context.Context, tenantID, creatorID uint, title, content string) (uint, error)
    UpdateReply(ctx context.Context, id uint, reply string) error
}

type Feedback struct {
    ID        uint
    TenantID  uint
    CreatorID uint
    Title     string
    Content   string
    Status    int16
    Reply     string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

然后 `AppContext.Reader` 接口加一行:

```go
// framework/pkg/plugin/appcontext.go
type Reader interface {
    // ...
    FeedbackRepo() rbac.FeedbackRepository
}
```

`Writer` 加一行:

```go
type Writer interface {
    // ...
    SetFeedbackRepo(r rbac.FeedbackRepository)
}
```

`AppContext` struct 加一个字段 + 2 个方法。**编译会引导你完成所有必要的接线**(所有用 `ctx.FeedbackRepo()` 的地方都会报错,直到你注册了 module)。

> 如果你的模块**不**给其他模块用,跳过这步,直接在 apps 里实现完整 repo。

---

## 4. 业务模块文件 (Step 3)

新建 `apps/feedback/` 目录,8 个文件:

```
apps/feedback/
├── errors.go            # 错误码 + ErrXxx
├── model.go             # 业务 struct
├── repository.go        # DB 访问
├── service.go           # 业务逻辑
├── handler.go           # gin handler
├── routes.go            # 路由注册
├── module.go            # plugin.Module 实现
└── config.go(可选)      # 模块私有配置
```

### 4.1 errors.go

```go
package feedback

import "gx1727.com/xin/framework/pkg/resp"

var (
    ErrNotFound        = resp.Err(13001, "反馈不存在")
    ErrTitleEmpty      = resp.Err(13002, "标题不能为空")
    ErrContentEmpty    = resp.Err(13003, "内容不能为空")
    ErrStatusInvalid   = resp.Err(13004, "状态值无效")
)
```

错误码走 `CodeFlag` 段(`13001-13999`),如果你的 module 已经在用就要避让。

### 4.2 model.go

```go
package feedback

import "time"

type Feedback struct {
    ID        uint      `json:"id"`
    TenantID  uint      `json:"tenant_id"`
    CreatorID uint      `json:"creator_id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    Status    int16     `json:"status"`
    Reply     string    `json:"reply,omitempty"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### 4.3 repository.go

```go
package feedback

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
    "gx1727.com/xin/framework/pkg/db"
)

type Repository struct {
    db *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
    return &Repository{db: pool}
}

func (r *Repository) List(ctx context.Context, tenantID, creatorID uint, page, size int) ([]Feedback, int64, error) {
    q, err := db.GetQuerier(ctx)
    if err != nil { return nil, 0, err }

    var total int64
    if err := q.QueryRow(ctx,
        `SELECT COUNT(*) FROM feedbacks
         WHERE tenant_id = $1 AND creator_id = $2 AND is_deleted = FALSE`,
        tenantID, creatorID,
    ).Scan(&total); err != nil {
        return nil, 0, err
    }

    rows, err := q.Query(ctx,
        `SELECT id, tenant_id, creator_id, title, content, status, COALESCE(reply, ''), created_at, updated_at
         FROM feedbacks
         WHERE tenant_id = $1 AND creator_id = $2 AND is_deleted = FALSE
         ORDER BY created_at DESC
         LIMIT $3 OFFSET $4`,
        tenantID, creatorID, size, (page-1)*size,
    )
    if err != nil { return nil, 0, err }
    defer rows.Close()

    var out []Feedback
    for rows.Next() {
        var f Feedback
        if err := rows.Scan(&f.ID, &f.TenantID, &f.CreatorID, &f.Title, &f.Content, &f.Status, &f.Reply, &f.CreatedAt, &f.UpdatedAt); err != nil {
            return nil, 0, err
        }
        out = append(out, f)
    }
    return out, total, nil
}

// GetByID, Create, UpdateReply ... 略
```

### 4.4 service.go

```go
package feedback

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
    xinContext "gx1727.com/xin/framework/pkg/context"
)

type Service struct {
    repo *Repository
}

func NewService(pool *pgxpool.Pool) *Service {
    return &Service{repo: NewRepository(pool)}
}

func (s *Service) List(ctx context.Context, page, size int) ([]Feedback, int64, error) {
    uc := xinContext.MustNewUserContext(ctx)
    return s.repo.List(ctx, uc.TenantID, uc.UserID, page, size)
}

// Create, UpdateReply ... 略
```

### 4.5 handler.go

```go
package feedback

import (
    "strconv"
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/resp"
)

type Handler struct {
    svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) List(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

    list, total, err := h.svc.List(c.Request.Context(), page, size)
    if err != nil {
        resp.HandleError(c, err)
        return
    }
    resp.Paginate(c, total, list)
}

func (h *Handler) Create(c *gin.Context) {
    var req struct {
        Title   string `json:"title" binding:"required"`
        Content string `json:"content" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        resp.BadRequest(c, err.Error())
        return
    }
    id, err := h.svc.Create(c.Request.Context(), req.Title, req.Content)
    if err != nil {
        resp.HandleError(c, err)
        return
    }
    resp.Success(c, gin.H{"id": id})
}

// ... UpdateReply, Delete 略
```

### 4.6 routes.go

```go
package feedback

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/middleware"
    "gx1727.com/xin/framework/pkg/permission"
)

func Register(protected *gin.RouterGroup, h *Handler) {
    g := protected.Group("/feedbacks")
    {
        g.GET("", middleware.Require(permission.P(permission.ResFlag, permission.ActList)), h.List)  // 这里沿用 ResFlag 占位
        // ...
    }
}
```

> **重要**:Resource 码需要先在 `resources` 表里 seed,然后才能用 `permission.P(...)` 引用。如果你想新增 `ResFeedback`,先改 [framework/pkg/permission/constants.go](framework/pkg/permission/constants.go) 加常量,再在 migration seed。

### 4.7 module.go

Phase 5 之后的统一形态：`Module(app *appx.App)` 显式接收 `*appx.App`，不再用 `init()` 副作用注册，也不再用 `db.Get()` / `bootx.Pool()` 等全局访问器。

```go
package feedback

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/appx"
    "gx1727.com/xin/framework/pkg/plugin"
)

// Module returns the feedback module as a BaseModule.
func Module(app *appx.App) plugin.Module {
    return &plugin.BaseModule{
        NameStr: "feedback",
        InitFn: func(_ plugin.Reader, w plugin.Writer) error {
            // 初始化逻辑(如资源 seed、配置校验)
            return nil
        },
        RegFn: func(_ plugin.Reader, _, protected *gin.RouterGroup) {
            pool := app.DB
            cfg := app.Config
            _ = cfg
            svc := NewService(pool)
            h := NewHandler(svc)
            Register(protected, h)
        },
    }
}
```

> `app.DB` / `app.Config` 是 `framework/pkg/appx` 公开的进程级资源，由 `framework.Boot(cfg)` 构造。这是当前唯一推荐的访问方式。

---

## 5. 在 main.go 注册 (Step 5)

Phase 5 之后**不再用** `_ "..."` side-effect import。改成在 [cmd/xin/main.go](cmd/xin/main.go) 显式调用：

```go
import (
    "gx1727.com/xin/apps/feedback"   // 显式 import，无下划线
    // ...
)

modules := []plugin.Module{
    // ...
    feedback.Module(app),
    // ...
}
```

## 6. 在 cfg.Module 启用 (Step 6)

[config/config.yaml](config/config.yaml) 加 `- feedback`:

```yaml
module:
  - user
  - role
  - feedback     # ← 加这一行
  # ...
```

或者如果你希望它默认启用,把它加到 [config/config.go](framework/pkg/config/config.go) 的 `optOutModules` 列表里:

```go
var optOutModules = []string{
    "menu", "user", "role", "resource", "organization", "dict", "asset",
    "permission",
    "feedback",   // ← 默认启用
}
```

---

## 7. 资源码 seed (Step 7)

在 `migrations/feedback.sql` 末尾加(见 Step 1 的 SQL 示例):

```sql
INSERT INTO resources (code, action, name, menu_id, status, is_deleted)
VALUES ('feedback', 'list', '查看反馈', NULL, 1, FALSE), ...
ON CONFLICT (code, action) DO NOTHING;
```

然后在 `framework/pkg/permission/constants.go` 加常量:

```go
const (
    ResFeedback = "feedback"
    // ...
)
```

> 如果不加常量,可以在 routes.go 里直接写字符串:`permission.P("feedback", "list")`。

---

## 8. 测试 (Step 8)

`apps/feedback/service_test.go`:

```go
package feedback

import (
    "context"
    "testing"
)

func TestService_List_NoRows(t *testing.T) {
    s := &Service{repo: NewRepository(nil)}
    // 没 db pool,期望走 nil-db fallback 分支
    _, _, err := s.List(context.Background(), 1, 20)
    if err == nil {
        t.Error("expected error from nil pool, got nil")
    }
}
```

跑:

```bash
go test -v ./apps/feedback/...
```

---

## 9. 完整代码模板

如果你是脚手架爱好者,可以直接 copy [apps/reference/dict/](apps/reference/dict/) 当模板 —— 它是最小的"教科书级"模块:

```
apps/reference/dict/
├── errors.go            ← 错误码
├── handler.go           ← gin handler
├── model.go             ← struct
├── module.go            ← BaseModule 完整实现
├── repository.go        ← pgx CRUD
├── routes.go            ← 路由注册(标准模式)
├── service.go           ← 业务
└── types.go             ← Request/Response 类型
```

完整模板在 [apps/reference/dict/module.go](apps/reference/dict/module.go):

```go
package dict

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/appx"
    "gx1727.com/xin/framework/pkg/plugin"
)

// Phase 5：显式接收 *appx.App，依赖显式注入。
func Module(app *appx.App) plugin.Module {
    return &plugin.BaseModule{
        NameStr: "dict",
        RegFn: func(_ plugin.Reader, _, protected *gin.RouterGroup) {
            pool := app.DB
            h := NewHandler(NewService(pool, NewPostgresDictRepository(pool)))
            Register(protected, h)
        },
    }
}
```

---

## 10. 验收清单

新增模块后,提交前跑:

```bash
# 1. 编译
go build ./...

# 2. vet
go vet ./...

# 3. 已有测试不挂
go test ./...

# 4. 启动 + smoke test
./bin/xin run &
sleep 2
curl http://localhost:8087/api/v1/feedbacks   # 应该 200 或 403,不能 panic
curl http://localhost:8087/api/v1/health      # 必须 200
```

启动日志应该看到:

```
2026/06/18 ... module feedback initialized
```

---

## 11. 常见陷阱

| 陷阱 | 解决 |
|---|---|
| 忘记在 main.go 显式 import 模块（不再用 `_` 副作用） | `feedback.Module(app)` 没被加进 `[]plugin.Module`，启动看不到 |
| 忘记加 `cfg.Module` | module 在列表里但 Init/Register 跳过 |
| 错误码和别的模块撞了 | 查 [resp/errors.go](framework/pkg/resp/errors.go) 选空段 |
| 资源码没 seed,Permission.P 直接写字符串 | 可以工作,但失去 IDE 自动补全 |
| 还在用 `db.Get()` / `bootx.Pool()` | 这俩已删，改用 `app.DB` 显式注入 |
| RLS 没建,跨租户泄漏 | `ALTER TABLE xxx ENABLE ROW LEVEL SECURITY` + `FORCE` |
| 没加 `is_deleted = FALSE` filter | 删除的数据会混进 List |
| 唯一索引不是 partial index | 删除后无法重建 |
| 事务里需要 ctx 自动拿 tx | 用 `db.GetQuerier(ctx, pool)` 让 ctx 找 tx |
| `super_admin` 平台角色没 bypass 你的中间件 | 确认你的 Require 在 [framework/pkg/middleware/auth.go](framework/pkg/middleware/auth.go) `requireWithSpecs` 里有 `if uc.IsSuperAdmin() { c.Next(); return }` 短路 |

## 12. 下一步

| 你想... | 看 |
|---|---|
| 看所有可用中间件 | [architecture.md#中间件链](architecture.md#5-中间件链) |
| 理解 RBAC | [permissions.md](permissions.md) |
| 部署你的新模块 | [deployment.md](deployment.md) |
| 写测试 | [developing.md#8-测试](#8-测试-step-8) |