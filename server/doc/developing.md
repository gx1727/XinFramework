# 开发指南

> 如何新增一个业务模块（app）。

## 1. 选位置

XinFramework 把业务模块分成三层：

| 位置 | 含义 | 例子 |
| --- | --- | --- |
| `apps/boot/<name>/` | 框架启动期必须 | auth / tenant |
| `apps/rbac/<name>/` | RBAC 标准件（Phase 3 待落地） | user / role / menu |
| `apps/reference/<name>/` | 参考实现，可被 fork 替换 | dict / asset / weixin |
| `apps/<name>/` | 业务专属（你的自定义 app） | cms / flag / order |

不放到 `framework/internal/module/`，那里正在清空（Phase 3）。

## 2. 最小骨架

```bash
mkdir -p apps/order/doc
```

文件清单：

```
apps/order/
├── module.go       # 模块入口（必须）
├── handler.go      # HTTP handler
├── service.go      # 业务逻辑
├── repository.go   # DB 访问
├── types.go        # DTO / struct
├── errors.go       # 业务错误
├── routes.go       # 路由注册
├── doc/api.md      # 模块文档
└── config.go       # （可选）模块私有配置
```

### 2.1 module.go

```go
package order

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/db"
    "gx1727.com/xin/framework/pkg/plugin"
)

type module struct{}

// Module 返回该模块的 plugin.Module 实例。
// 名称 "order" 必须与 config.yaml 的 module: 列表一致。
func Module() plugin.Module {
    return plugin.NewModule("order", func(public, protected *gin.RouterGroup) {
        repo := NewRepository(db.Get())
        svc := NewService(repo)
        h := NewHandler(svc)
        Register(protected, h)
    })
}

// init 在包加载时自动注册到 plugin.Apps()。
// 不需要 main.go 显式调用——但 main.go 必须有 side-effect import。
func init() {
    plugin.Register(Module())
}
```

### 2.2 types.go

```go
package order

type Order struct {
    ID         uint    `json:"id"`
    TenantID   uint    `json:"tenant_id"`
    Code       string  `json:"code"`
    Amount     float64 `json:"amount"`
    Status     int     `json:"status"`
    CreatedAt  string  `json:"created_at"`
    UpdatedAt  string  `json:"updated_at"`
}

type CreateRequest struct {
    Code    string  `json:"code"    binding:"required,max=64"`
    Amount  float64 `json:"amount"  binding:"required,gt=0"`
}

type UpdateRequest struct {
    Amount *float64 `json:"amount"`
    Status *int     `json:"status"`
}

type ListResponse struct {
    List  []Order `json:"list"`
    Total int64   `json:"total"`
    Page  int     `json:"page"`
    Size  int     `json:"size"`
}
```

### 2.3 errors.go

```go
package order

import "errors"

var (
    ErrOrderNotFound      = errors.New("order not found")
    ErrOrderCodeExists    = errors.New("order code already exists")
    ErrOrderStatusInvalid = errors.New("order status invalid")
)
```

### 2.4 repository.go

```go
package order

import (
    "context"
    "errors"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
    Create(ctx context.Context, tenantID uint, req CreateRequest) (*Order, error)
    GetByID(ctx context.Context, tenantID, id uint) (*Order, error)
    List(ctx context.Context, tenantID uint, page, size int, keyword string) ([]Order, int64, error)
    Update(ctx context.Context, tenantID, id uint, req UpdateRequest) error
    Delete(ctx context.Context, tenantID, id uint) error
}

type pgRepository struct {
    db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
    return &pgRepository{db: db}
}

func (r *pgRepository) Create(ctx context.Context, tenantID uint, req CreateRequest) (*Order, error) {
    var o Order
    err := r.db.QueryRow(ctx, `
        INSERT INTO orders (tenant_id, code, amount, status, created_at, updated_at)
        VALUES ($1, $2, $3, 1, NOW(), NOW())
        RETURNING id, tenant_id, code, amount, status, created_at, updated_at
    `, tenantID, req.Code, req.Amount).Scan(
        &o.ID, &o.TenantID, &o.Code, &o.Amount, &o.Status, &o.CreatedAt, &o.UpdatedAt,
    )
    if err != nil {
        if isUniqueViolation(err) {
            return nil, ErrOrderCodeExists
        }
        return nil, err
    }
    return &o, nil
}

func (r *pgRepository) GetByID(ctx context.Context, tenantID, id uint) (*Order, error) {
    var o Order
    err := r.db.QueryRow(ctx, `
        SELECT id, tenant_id, code, amount, status, created_at, updated_at
        FROM orders
        WHERE tenant_id = $1 AND id = $2 AND is_deleted = FALSE
    `, tenantID, id).Scan(
        &o.ID, &o.TenantID, &o.Code, &o.Amount, &o.Status, &o.CreatedAt, &o.UpdatedAt,
    )
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, ErrOrderNotFound
    }
    return &o, err
}

func (r *pgRepository) List(ctx context.Context, tenantID uint, page, size int, keyword string) ([]Order, int64, error) {
    if page < 1 {
        page = 1
    }
    if size < 1 || size > 200 {
        size = 20
    }
    offset := (page - 1) * size

    var total int64
    err := r.db.QueryRow(ctx, `
        SELECT COUNT(*) FROM orders
        WHERE tenant_id = $1 AND is_deleted = FALSE
          AND ($2 = '' OR code ILIKE '%' || $2 || '%')
    `, tenantID, keyword).Scan(&total)
    if err != nil {
        return nil, 0, err
    }

    rows, err := r.db.Query(ctx, `
        SELECT id, tenant_id, code, amount, status, created_at, updated_at
        FROM orders
        WHERE tenant_id = $1 AND is_deleted = FALSE
          AND ($2 = '' OR code ILIKE '%' || $2 || '%')
        ORDER BY id DESC
        LIMIT $3 OFFSET $4
    `, tenantID, keyword, size, offset)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    list := []Order{}
    for rows.Next() {
        var o Order
        if err := rows.Scan(&o.ID, &o.TenantID, &o.Code, &o.Amount, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
            return nil, 0, err
        }
        list = append(list, o)
    }
    return list, total, rows.Err()
}

func (r *pgRepository) Update(ctx context.Context, tenantID, id uint, req UpdateRequest) error {
    // 用 COALESCE 做局部更新
    _, err := r.db.Exec(ctx, `
        UPDATE orders SET
            amount    = COALESCE($3, amount),
            status    = COALESCE($4, status),
            updated_at = NOW()
        WHERE tenant_id = $1 AND id = $2 AND is_deleted = FALSE
    `, tenantID, id, req.Amount, req.Status)
    return err
}

func (r *pgRepository) Delete(ctx context.Context, tenantID, id uint) error {
    _, err := r.db.Exec(ctx, `
        UPDATE orders SET is_deleted = TRUE, deleted_at = NOW()
        WHERE tenant_id = $1 AND id = $2
    `, tenantID, id)
    return err
}

func isUniqueViolation(err error) bool {
    // pgx 的 PgError.Code == "23505"
    var pgErr *pgconn.PgError
    return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
```

### 2.5 service.go

```go
package order

import (
    "context"

    "gx1727.com/xin/framework/pkg/audit"
    xincontext "gx1727.com/xin/framework/pkg/context"
)

type Service struct {
    repo Repository
}

func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*Order, error) {
    xc := xincontext.FromContext(ctx)
    return s.repo.Create(ctx, xc.TenantID, req)
}

func (s *Service) Get(ctx context.Context, id uint) (*Order, error) {
    xc := xincontext.FromContext(ctx)
    o, err := s.repo.GetByID(ctx, xc.TenantID, id)
    if err != nil {
        return nil, err
    }
    audit.WithContext(ctx, "order.get", id)
    return o, nil
}

func (s *Service) List(ctx context.Context, page, size int, keyword string) (*ListResponse, error) {
    xc := xincontext.FromContext(ctx)
    list, total, err := s.repo.List(ctx, xc.TenantID, page, size, keyword)
    if err != nil {
        return nil, err
    }
    return &ListResponse{List: list, Total: total, Page: page, Size: size}, nil
}

func (s *Service) Update(ctx context.Context, id uint, req UpdateRequest) error {
    xc := xincontext.FromContext(ctx)
    audit.WithContext(ctx, "order.update", id)
    return s.repo.Update(ctx, xc.TenantID, id, req)
}

func (s *Service) Delete(ctx context.Context, id uint) error {
    xc := xincontext.FromContext(ctx)
    audit.WithContext(ctx, "order.delete", id)
    return s.repo.Delete(ctx, xc.TenantID, id)
}
```

### 2.6 handler.go

```go
package order

import (
    "errors"

    "github.com/gin-gonic/gin"

    "gx1727.com/xin/framework/pkg/resp"
)

type Handler struct {
    svc *Service
}

func NewHandler(svc *Service) *Handler {
    return &Handler{svc: svc}
}

func (h *Handler) Create(c *gin.Context) {
    var req CreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        resp.Fail(c, resp.ErrBadRequest, err.Error())
        return
    }
    o, err := h.svc.Create(c.Request.Context(), req)
    if err != nil {
        if errors.Is(err, ErrOrderCodeExists) {
            resp.Fail(c, resp.ErrConflict, "订单编码已存在")
            return
        }
        resp.Fail(c, resp.ErrInternal, err.Error())
        return
    }
    resp.OK(c, o)
}

func (h *Handler) Get(c *gin.Context) {
    id := parseID(c)
    o, err := h.svc.Get(c.Request.Context(), id)
    if err != nil {
        if errors.Is(err, ErrOrderNotFound) {
            resp.Fail(c, resp.ErrNotFound, "订单不存在")
            return
        }
        resp.Fail(c, resp.ErrInternal, err.Error())
        return
    }
    resp.OK(c, o)
}

func (h *Handler) List(c *gin.Context) {
    page := atoiDefault(c.Query("page"), 1)
    size := atoiDefault(c.Query("size"), 20)
    keyword := c.Query("keyword")

    list, err := h.svc.List(c.Request.Context(), page, size, keyword)
    if err != nil {
        resp.Fail(c, resp.ErrInternal, err.Error())
        return
    }
    resp.OK(c, list)
}

func (h *Handler) Update(c *gin.Context) {
    id := parseID(c)
    var req UpdateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        resp.Fail(c, resp.ErrBadRequest, err.Error())
        return
    }
    if err := h.svc.Update(c.Request.Context(), id, req); err != nil {
        if errors.Is(err, ErrOrderNotFound) {
            resp.Fail(c, resp.ErrNotFound, "订单不存在")
            return
        }
        resp.Fail(c, resp.ErrInternal, err.Error())
        return
    }
    resp.OK(c, nil)
}

func (h *Handler) Delete(c *gin.Context) {
    id := parseID(c)
    if err := h.svc.Delete(c.Request.Context(), id); err != nil {
        resp.Fail(c, resp.ErrInternal, err.Error())
        return
    }
    resp.OK(c, nil)
}
```

### 2.7 routes.go

```go
package order

import (
    "github.com/gin-gonic/gin"

    "gx1727.com/xin/framework/pkg/middleware"
    "gx1727.com/xin/framework/pkg/permission"
)

func Register(protected *gin.RouterGroup, h *Handler) {
    orders := protected.Group("/orders")
    orders.GET("", middleware.Require(permission.P(permission.ResOrder, permission.ActList)), h.List)
    orders.GET("/:id", middleware.Require(permission.P(permission.ResOrder, permission.ActList)), h.Get)
    orders.POST("", middleware.Require(permission.P(permission.ResOrder, permission.ActCreate)), h.Create)
    orders.PUT("/:id", middleware.Require(permission.P(permission.ResOrder, permission.ActUpdate)), h.Update)
    orders.DELETE("/:id", middleware.Require(permission.P(permission.ResOrder, permission.ActDelete)), h.Delete)
}
```

需要在 `framework/pkg/permission/constants.go` 加：

```go
const (
    ResOrder = "order"
)
```

并在 `migrations/framework.sql` 插入对应资源码。

## 3. 注册到主程序

```go
// cmd/xin/main.go
import (
    "gx1727.com/xin/framework"
    _ "gx1727.com/xin/framework" // 触发 builtin_modules.go
    _ "gx1727.com/xin/apps/boot/auth"
    _ "gx1727.com/xin/apps/boot/tenant"
    _ "gx1727.com/xin/apps/order"  // ← 新加
    _ "gx1727.com/xin/apps/cms"
    _ "gx1727.com/xin/apps/flag"
)
```

## 4. 加入配置

```yaml
# config/config.yaml
module:
  - auth
  - tenant
  - ...
  - order  # ← 新加
  - cms
  - flag
```

## 5. 加 SQL 迁移

```bash
cat > migrations/order.sql <<'EOF'
CREATE TABLE orders (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT NOT NULL,
    code        VARCHAR(64) NOT NULL,
    amount      NUMERIC(20,4) NOT NULL DEFAULT 0,
    status      SMALLINT NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted  BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_orders_tenant ON orders(tenant_id, is_deleted, id DESC);
CREATE UNIQUE INDEX uk_orders_tenant_code ON orders(tenant_id, code) WHERE is_deleted = FALSE;

INSERT INTO resources (code, type, name, method, path) VALUES
    ('order:list',   'api', '订单列表', 'GET',    '/api/v1/orders'),
    ('order:get',    'api', '订单详情', 'GET',    '/api/v1/orders/:id'),
    ('order:create', 'api', '新建订单', 'POST',   '/api/v1/orders'),
    ('order:update', 'api', '更新订单', 'PUT',    '/api/v1/orders/:id'),
    ('order:delete', 'api', '删除订单', 'DELETE', '/api/v1/orders/:id');
EOF
```

启动时 `framework/pkg/migrate` 自动按文件名排序执行。

## 6. 前端配合

1. `UI/src/api/client.ts` 加 `orderApi = { list, get, create, update, delete }`
2. `UI/src/locales/zh-CN.ts` 加 `pages.order.*` 块（**先加**，作为 `LocaleKeys` 类型源头）
3. `UI/src/locales/en-US.ts` 同步
4. `UI/src/App.tsx` 加 `lazy(() => import("@/pages/Order"))` + 路由
5. 写 `UI/src/pages/Order.tsx`（参考 `UI/src/pages/Users.tsx`）

详见 [UI/AGENTS.md](file:///d:\work\xin\XinFramework\UI\AGENTS.md)。

## 7. 验证

```bash
# 后端编译
cd server
go build ./...
go vet ./...

# 启动
go run ./cmd/xin run

# 测试 API
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"tenant_code":"default","account":"admin","password":"xxx"}'

curl -X POST http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"code":"o001","amount":99.9}'
```

## 8. 模块自检清单

- [ ] 包名与目录名一致
- [ ] `Module()` 函数返回 `plugin.Module`
- [ ] `init()` 调用 `plugin.Register(Module())`
- [ ] `Register(public, protected, h)` 注册路由时加 `middleware.Require`
- [ ] Repository 强制按 `tenant_id` 过滤
- [ ] 用 `is_deleted = FALSE` 软删过滤
- [ ] Service 从 `xincontext.FromContext(ctx)` 拿 tenantID
- [ ] Handler 用 `resp.OK/Fail` 返回
- [ ] 业务错误在 `errors.go` 定义并在 handler 用 `errors.Is`
- [ ] 关键操作调 `audit.WithContext`
- [ ] SQL 迁移文件加好（表 + 索引 + resources 记录）
- [ ] `cmd/xin/main.go` 加 side-effect import
- [ ] `config.yaml` 的 `module:` 列表加模块名

---

## 9. 进阶

### 9.1 跨模块依赖

详见 [architecture.md §4](file:///d:\work\xin\XinFramework\server\doc\architecture.md#4-跨模块依赖规则)。

如果你的模块需要用其他模块的 repository，通过 framework/pkg 的注册钩子。

### 9.2 自定义配置

```go
// apps/order/config.go
package order

type Config struct {
    MaxAmount float64 `yaml:"max_amount"`
}

func (c *Config) Default() *Config {
    return &Config{MaxAmount: 100000}
}
```

`config.Get()` 读全局配置（YAML root），按 yaml 路径取。

### 9.3 异步任务

```go
import "github.com/hibiken/asynq"

func (s *Service) HeavyTask(ctx context.Context, id uint) error {
    payload, _ := json.Marshal(map[string]any{"order_id": id})
    _, err := asynqClient.Enqueue(asynq.NewTask("order:heavy", payload))
    return err
}
```

框架未集成 asynq，需要自行接入。

### 9.4 中间件

模块级中间件（作用于该模块的所有路由）：

```go
func Register(protected *gin.RouterGroup, h *Handler) {
    orders := protected.Group("/orders")
    orders.Use(myMiddleware())  // ← 模块级
    orders.GET("", ..., h.List)
}
```

公共中间件（全局）放 `framework/internal/core/middleware`。