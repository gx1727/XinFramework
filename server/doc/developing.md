# 开发指南

> 本文件描述如何在 XinFramework 中新增业务模块、修改既有模块、调试常见问题。
> 模块清单见 [modules.md](./modules.md)；架构与约定见 [architecture.md](./architecture.md)。

---

## 1. 新增一个业务模块（8 步流程）

以 `apps/tenant/foo` 为例——一个租户域的 CRUD 模块。

### 第 1 步：规划

- **数据域**：tenant / platform / shared？
- **必带列**：`tenant_id` / `is_deleted` / `created_at` / `updated_at` / `created_by` / `updated_by`
- **错误码段**：申请 `resp.CodeFoo` 段（如 15000）
- **资源常量**：`permission.ResFoo`
- **权限码**：至少 `foo:list` / `foo:create` / `foo:update` / `foo:delete`

### 第 2 步：建包结构

```
apps/tenant/foo/
├── module.go        # Module(app) → *plugin.BaseModule
├── routes.go        # Register(tenant *gin.RouterGroup, h *Handler)
├── handler.go       # gin handler（请求解析 + 调 service）
├── service.go       # 业务编排（mapRepoError + 跨模块调用 + 审计）
├── repository.go    # 数据访问（Querier）
├── model.go         # 实体 struct（对应 DB 表）
├── types.go         # 请求/响应 DTO + 业务魔数
└── errors.go        # 业务错误码（从 resp.CodeFoo 段申请）
```

### 第 3 步：申请错误码

`errors.go`：

```go
package foo

import "gx1727.com/xin/framework/pkg/resp"

const (
    CodeFoo = 15000
)

var (
    ErrFooNotFound = resp.Err(CodeFoo+1, "foo 不存在")
    ErrFooDuplicate = resp.Err(CodeFoo+2, "foo 已存在")
    ErrFooInUse = resp.Err(CodeFoo+3, "foo 被引用，无法删除")
)

// 同时在 framework/pkg/resp/errors.go 追加段常量
// const CodeFoo = 15000  // foo: 15001-15999
```

### 第 4 步：定义资源常量

`framework/pkg/permission/constants.go` 追加：

```go
const ResFoo = "foo"
```

### 第 5 步：写 DB 迁移

`migrations/foo.sql`（如需独立文件），或追加到 `init_schema.sql`（不推荐）：

```sql
CREATE TABLE IF NOT EXISTS tenant_foos (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,
    code       VARCHAR(32) NOT NULL,
    name       VARCHAR(64) NOT NULL,
    status     SMALLINT    DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by BIGINT,
    updated_by BIGINT,
    is_deleted BOOLEAN     DEFAULT FALSE
);

CREATE UNIQUE INDEX uk_tenant_foos_code
    ON tenant_foos (tenant_id, code) WHERE is_deleted = FALSE;

-- RLS policy
ALTER TABLE tenant_foos ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_foos;
CREATE POLICY tenant_isolation_policy ON tenant_foos USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);
```

### 第 6 步：写代码

#### 6.1 `model.go`

```go
package foo

import "time"

type Foo struct {
    ID        uint64    `json:"id"`
    TenantID  uint64    `json:"tenant_id"`
    Code      string    `json:"code"`
    Name      string    `json:"name"`
    Status    int       `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    CreatedBy uint64    `json:"created_by"`
    UpdatedBy uint64    `json:"updated_by"`
}
```

#### 6.2 `types.go`

```go
package foo

const (
    StatusActive   = 1
    StatusInactive = 2
)

type ListRequest struct {
    Keyword string `form:"keyword"`
    Page    int    `form:"page"`
    Size    int    `form:"size"`
}

type CreateRequest struct {
    Code   string `json:"code" binding:"required"`
    Name   string `json:"name" binding:"required"`
    Status int    `json:"status"`
}

type UpdateRequest struct {
    Name   *string `json:"name"`
    Status *int    `json:"status"`
}
```

#### 6.3 `repository.go`

```go
package foo

import (
    "context"
    "errors"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "gx1727.com/xin/framework/pkg/db"
)

var (
    ErrFooNotFoundDB = errors.New("foo not found")
    ErrFooDuplicateDB = errors.New("foo code already exists")
)

type Repository struct {
    db *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
    return &Repository{db: pool}
}

func (r *Repository) GetByID(ctx context.Context, tenantID, id uint64) (*Foo, error) {
    q, err := db.GetQuerier(ctx, r.db)
    if err != nil { return nil, err }
    
    var f Foo
    err = q.QueryRow(ctx, `
        SELECT id, tenant_id, code, name, status,
               created_at, updated_at, created_by, updated_by
        FROM tenant_foos
        WHERE id = $1 AND tenant_id = $2 AND is_deleted = FALSE
    `, id, tenantID).Scan(
        &f.ID, &f.TenantID, &f.Code, &f.Name, &f.Status,
        &f.CreatedAt, &f.UpdatedAt, &f.CreatedBy, &f.UpdatedBy,
    )
    
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, ErrFooNotFoundDB
    }
    if err != nil { return nil, err }
    return &f, nil
}

// List / Create / Update / Delete 略
```

#### 6.4 `service.go`

```go
package foo

import (
    "context"
    "errors"
    "gx1727.com/xin/framework/pkg/audit"
    "gx1727.com/xin/framework/pkg/db"
    "gx1727.com/xin/framework/pkg/resp"
    "gx1727.com/xin/framework/pkg/xincontext"
)

type Service struct {
    pool *pgxpool.Pool
    repo *Repository
}

func NewService(pool *pgxpool.Pool, repo *Repository) *Service {
    return &Service{pool: pool, repo: repo}
}

func (s *Service) mapRepoError(err error) error {
    if errors.Is(err, ErrFooNotFoundDB) {
        return ErrFooNotFound
    }
    if errors.Is(err, ErrFooDuplicateDB) {
        return ErrFooDuplicate
    }
    return err
}

func (s *Service) GetByID(ctx context.Context, tenantID, id uint64) (*Foo, error) {
    foo, err := s.repo.GetByID(ctx, tenantID, id)
    if err != nil {
        return nil, s.mapRepoError(err)
    }
    return foo, nil
}

func (s *Service) Create(ctx context.Context, tenantID, creatorID uint64, req CreateRequest) (*Foo, error) {
    var f *Foo
    err := db.RunInTenantTx(ctx, s.pool, uint(tenantID), func(ctx context.Context) error {
        var err error
        f, err = s.repo.Create(ctx, tenantID, creatorID, req)
        return err
    })
    if err != nil {
        return nil, s.mapRepoError(err)
    }
    
    audit.Log(ctx, s.pool, audit.Entry{
        Action:    "foo:create",
        TableName: "tenant_foos",
        RecordID:  f.ID,
        NewData:   map[string]any{"id": f.ID, "code": f.Code, "name": f.Name},
    })
    return f, nil
}
```

#### 6.5 `handler.go`

```go
package foo

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/resp"
    "gx1727.com/xin/framework/pkg/xincontext"
)

type Handler struct {
    svc *Service
}

func NewHandler(svc *Service) *Handler {
    return &Handler{svc: svc}
}

func (h *Handler) Get(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil { resp.BadRequest(c, "无效的ID参数"); return }
    
    xc := xincontext.New(c)
    foo, err := h.svc.GetByID(c.Request.Context(), xc.TenantID, id)
    if err != nil { resp.HandleError(c, err); return }
    
    resp.Success(c, foo)
}
```

#### 6.6 `routes.go`

```go
package foo

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/middleware"
    "gx1727.com/xin/framework/pkg/permission"
)

func Register(tenant *gin.RouterGroup, h *Handler) {
    g := tenant.Group("/foos")
    g.GET("", middleware.Require(permission.P(permission.ResFoo, permission.ActList)), h.List)
    g.POST("", middleware.Require(permission.P(permission.ResFoo, permission.ActCreate)), h.Create)
    g.GET("/:id", middleware.Require(permission.P(permission.ResFoo, permission.ActList)), h.Get)
    g.PUT("/:id", middleware.Require(permission.P(permission.ResFoo, permission.ActUpdate)), h.Update)
    g.DELETE("/:id", middleware.Require(permission.P(permission.ResFoo, permission.ActDelete)), h.Delete)
}
```

#### 6.7 `module.go`

```go
package foo

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/appx"
    "gx1727.com/xin/framework/pkg/plugin"
)

func Module(app *appx.App) plugin.Module {
    return &plugin.BaseModule{
        NameStr: "foo",
        InitFn: func(_ plugin.Reader, w plugin.Writer) error {
            // 如需对外暴露 Repository，写到 AppContext
            // w.SetFooRepo(NewRepository(app.DB))
            return nil
        },
        RegFn: func(_ plugin.Reader, _ *gin.RouterGroup,
                     tenant *gin.RouterGroup, _ *gin.RouterGroup) {
            pool := app.DB
            h := NewHandler(NewService(pool, NewRepository(pool)))
            Register(tenant, h)
        },
    }
}
```

### 第 7 步：注册到 main.go

`server/cmd/xin/main.go` 的 `modules := []plugin.Module{...}` 列表追加：

```go
import "gx1727.com/xin/apps/tenant/foo"

modules := []plugin.Module{
    // ... 既有
    foo.Module(app),  // ← 新增
}
```

### 第 8 步：写种子数据（可选）

如果模块需要 menu / permission seed，追加到 `migrations/init_seed.sql` 或独立 `migrations/foo.sql`：

```sql
-- foo 菜单
INSERT INTO tenant_menus (tenant_id, code, name, path, icon, sort, parent_id, ancestors, visible, enabled)
SELECT (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE),
       'foo', 'Foo 管理', '/foos', 'BoxIcon', 99, 0, '', TRUE, TRUE
ON CONFLICT (tenant_id, code) WHERE tenant_id <> 0 AND is_deleted = FALSE DO NOTHING;

-- foo 权限码
INSERT INTO tenant_permissions (tenant_id, menu_id, code, name, action, description, sort, status)
SELECT (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE),
       (SELECT id FROM tenant_menus WHERE code = 'foo' AND tenant_id = ...),
       'foo:list', '查询 Foo', 'list', '查询 Foo 列表', 1, 1
ON CONFLICT (tenant_id, code) WHERE ... DO NOTHING;

-- admin 角色绑定
INSERT INTO tenant_role_resources (tenant_id, role_id, permission_id)
SELECT 1, 1, p.id FROM tenant_permissions p
WHERE p.code IN ('foo:list', 'foo:create', 'foo:update', 'foo:delete')
ON CONFLICT (role_id, permission_id) WHERE is_deleted = FALSE DO NOTHING;
```

---

## 2. 修改既有模块

### 2.1 加新接口

1. 在 `routes.go` 加路由 + permission spec
2. 在 `handler.go` 加 handler
3. 在 `service.go` 加业务方法
4. 在 `repository.go` 加数据访问
5. 在 `permissions.md` / `modules.md` 文档里补一笔

### 2.2 加新表

1. 创建迁移 `00XX_add_xxx.sql`
2. 写 DDL（带 `IF NOT EXISTS` / `ADD COLUMN IF NOT EXISTS`）
3. **不**改旧迁移文件
4. 如启用 RLS，加 `ENABLE ROW LEVEL SECURITY` + `CREATE POLICY`

### 2.3 改错误码段

**禁止**：已经发布的错误码段不要改含义（前端已基于 code 处理）。

新加的模块申请**新段位**（见 `framework/pkg/resp/errors.go`）。

### 2.4 改 Repository 返回类型

- 用新类型：加 `GetByIDV2` 而非改 `GetByID`（避免破坏既有调用方）
- 删除旧类型：先 deprecate 一个版本，下个版本再删

---

## 3. 调试常见问题

### 3.1 启动失败：循环依赖

```
import cycle not allowed
```

→ `framework` 不能 import `apps`。`apps/xxx` 之间也避免互相 import。共享走 `plugin.AppContext`。

### 3.2 业务 Handler 拿不到 tenant

```
Context.TenantID == 0
```

→ 检查路由是否挂到 `tenant` group（不是 `public`），并确保请求带了有效 token。

### 3.3 RLS 拒所有数据

```
no rows returned
```

→ 检查：
- `db.RunInTenantTx` 是否设置正确 tenant_id
- 当前 SQL 是否在该表上有 RLS policy
- 数据是否在另一个 tenant_id 下

### 3.4 权限 403

```json
{ "code": 4001, "msg": "无权限访问", "data": null }
```

→ 检查：
- 路由是否挂了 `middleware.Require(spec)`
- 用户角色是否绑定了对应 `tenant_role_resources` 记录
- Spec 拼写（`user:list` 还是 `users:list`？）

### 3.5 JSONB 写入失败

```
invalid input syntax for type json
```

→ SQL 必须 `::jsonb` cast：
```sql
INSERT INTO tenants (config) VALUES ($1::jsonb)
```

### 3.6 前端 zustand 状态没持久化

→ 检查 `persist` 配置的 `name` 和 `partialize`：
```ts
persist((set, get) => ({ ... }), {
    name: "auth-storage",
    partialize: (state) => ({ ... })
})
```

### 3.7 前端 mock 兜底静默

→ 已废弃！必须改成：
```ts
try { ... }
catch (err) { setError(err.message) }
// mock 仅在 useMockFallback=true 时使用
```

---

## 4. 编码规范

### 4.1 Go

- 包名：单数小写（`user` 不是 `users`）
- 错误变量：命名 sentinel（`ErrFooNotFoundDB`），禁止 `errors.New("xxx not found")`
- 错误处理：DB 层命名 sentinel；Service 层 `mapRepoError` 翻译；Handler 层 `resp.HandleError`
- 不要在 Repository 内新建 context
- 不要用 `init()` 注册模块
- 不要手写 `c.JSON(...)`——走 `resp.Success` / `resp.HandleError` / `resp.Paginate`
- 不要绕过 `db.RunInTenantTx` 写租户域表
- 不要直接 import `pgxpool.Pool` 之外的 DB 类型
- 不要自己判断 HTTP 状态码

### 4.2 TypeScript / React

- 文件：PascalCase 组件、camelCase 工具
- 状态：复杂用 `zustand` store，简单用 `useState`
- 文案：先加 `zh-CN.ts`，再用 `t.xxx.yyy`
- 编码：UTF-8 无 BOM（PowerShell 默认 GBK 会破坏中文）
- tsc 检查：`.\node_modules\.bin\tsc --noEmit` 0 错误才算完成

### 4.3 SQL

- DDL：`IF NOT EXISTS` / `ADD COLUMN IF NOT EXISTS`
- Seed：`ON CONFLICT DO NOTHING` / `ON CONFLICT (...) DO UPDATE`
- 索引：`WHERE is_deleted = FALSE` 谓词
- JSONB：显式 `::jsonb` cast
- RLS：先 `DROP POLICY IF EXISTS` 再 `CREATE POLICY`

---

## 5. 测试

当前框架**未集成**单元测试套件，但可以这样加：

```go
// apps/tenant/foo/foo_test.go
package foo

import (
    "context"
    "testing"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/stretchr/testify/assert"
)

func TestRepository_GetByID(t *testing.T) {
    // 准备测试 DB
    pool, err := pgxpool.New(context.Background(), "postgres://test:test@localhost:5432/xin_test")
    require.NoError(t, err)
    defer pool.Close()
    
    repo := NewRepository(pool)
    
    // 跑测试
    foo, err := repo.GetByID(context.Background(), 1, 100)
    assert.NoError(t, err)
    assert.NotNil(t, foo)
}
```

**集成测试**：用 testcontainers-go 拉起 PG + Redis，跑端到端。

---

## 6. 性能分析

### 6.1 慢查询

启用 PostgreSQL 慢查询日志：

```bash
# postgresql.conf
log_min_duration_statement = 500
```

### 6.2 N+1 查询

Repository 里 `for _, id := range ids` 调 `GetByID(id)` 是经典 N+1。**用 `WHERE id IN (...)`** 一次查完。

### 6.3 索引使用

```sql
EXPLAIN ANALYZE SELECT * FROM tenant_users WHERE tenant_id = 1 AND is_deleted = FALSE;
```

确保走索引（`Index Scan using idx_tenant_users_tenant`）而非全表扫（`Seq Scan`）。

### 6.4 慢 API

中间件 `framework/framework.go:setupRouter` 的 Logger 记录每条请求耗时：

```
[INFO] [req-id-123] GET /api/v1/users | 200 | 23.5ms
```

可 grep `| 5[0-9]{2}ms` 找慢请求。

---

## 7. 提交流程

```bash
# 1. 本地验证
go build -o bin/xin ./cmd/xin
go test ./...                                    # 如有测试
cd UI && .\node_modules\.bin\tsc --noEmit
cd ..

# 2. BOM 检测（CI gate）
python server/scripts/strip_bom.py --check .

# 3. 提交
git add -A
git commit -m "feat: 新增 foo 模块"
```

---

## 8. 升级路径

### 8.1 框架升级

项目是单一 Go module，框架升级就是 `git pull` 拉最新代码。关注：

- `migrations/` 有新 SQL → 自动跑
- `framework/pkg/resp/errors.go` 加新段位 → 可能影响前端错误处理
- `apps/*/module.go` 加新 slot → 老模块可能不兼容
- `plugin.Module` 接口扩展 → 老模块需实现新方法（用 `BaseModule` 跳过）

### 8.2 业务模块升级

- 不要修改 `framework` 内部的代码（业务模块应通过 AppContext 槽位交互）
- 通过 PR 把通用代码移到 framework
- 重大重构分 phase 走（参考 Phase 0023 平台/租户域分域）

---

## 9. 进阶：横切关注点

### 9.1 审计（`audit`）

```go
audit.Log(ctx, pool, audit.Entry{
    Action:    "user:delete",
    TableName: "tenant_users",
    RecordID:  userID,
    OldData:   oldUser,
    NewData:   nil,
})
```

**失败不抛**——业务路径不应被审计写库失败打断。

### 9.2 缓存

`framework/pkg/cache.Get()` 返回 `*redis.Client`（可能为 nil）。

```go
rdb := cache.Get()
if rdb == nil { /* 降级处理 */ return nil }
err := rdb.Set(ctx, key, value, ttl).Err()
```

### 9.3 文件上传

```go
import "gx1727.com/xin/apps/reference/asset"

assetSvc := asset.NewFileService(storage, assetRepo)
url, err := assetSvc.Upload(ctx, file)
```

### 9.4 字典查询

```go
import dictpkg "gx1727.com/xin/framework/pkg/dict"

items, err := dictpkg.GetItems(ctx, "gender")  // 内存缓存
```

### 9.5 权限校验

```go
uc := xincontext.NewUserContext(c)
if !uc.HasPlatformRole(jwt.PlatformRoleSuperAdmin) {
    return ErrForbidden
}
```

### 9.6 配置读取

```go
import "gx1727.com/xin/apps/reference/config"

val, err := config.GetItem(ctx, "site", "site_name")
```

---

## 10. 排错清单

| 现象 | 检查 |
|---|---|
| 启动期卡住 | PG 连接；`migrations/` 可读 |
| 启动期 `migrate failed` | 看具体 SQL 错误（可能 JSONB cast 缺） |
| 登录返回 401 | 密码哈希是否与 `password.go` 兼容 |
| 业务 403 | 角色权限码绑定；spec 拼写 |
| 业务 404 | RLS policy；tenant_id 过滤 |
| 前端 401 持续 | refresh token 失效；clear tokens 跳登录 |
| 前端 fetch CORS | `cfg.cors.allow_origins` |
| 前端状态丢失 | `persist` name；`partialize` 字段 |
| 慢查询 | `EXPLAIN ANALYZE`；索引谓词 |
| 内存泄漏 | goroutine 泄漏；DB 连接未释放 |
