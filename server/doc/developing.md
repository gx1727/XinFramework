# 开发指南：新增一个业务模块

> 用具体例子带你走完新增模块的全流程。两个模板：
>
> - **业务模块**（租户域）：`feedback`（用户反馈）
> - **平台管理模块**（`/platform/*` 域）：`platform_dict`（平台字典管理）

## 0. 前提：理解现有架构

读这些文档后再继续：

1. [architecture.md](architecture.md) — 了解 AppContext / Module 接口 / Init / Register 流程
2. [modules.md](modules.md) — 看现有模块的结构
3. [database.md](database.md) — 了解 RLS / 软删除 / 索引 / JSONB 约定

## 1. 标准 8 步流程

```
1. SQL 迁移           migrations/<feature>.sql
2. 公共接口定义       framework/pkg/<scope>/<feature>.go（可选）
3. 业务模块           apps/<scope>/<feature>/{handler,service,repository,model,module,routes}.go
4. 错误码            apps/<scope>/<feature>/errors.go
5. 在 main.go 注册    cmd/xin/main.go
6. 在 cfg.Module 启   config/config.yaml
7. 资源码 seed（可选）migrations/<feature>.sql 末尾 INSERT INTO tenant_permissions
8. 单元测试（可选）   apps/<scope>/<feature>/*_test.go
```

下面先讲**业务模块**（Step A），再讲**平台管理模块**差异（Step B）。

---

# Step A：业务模块模板（租户域）

## 2. SQL 迁移（Step 1）

新建 `migrations/feedback.sql`：

```sql
CREATE TABLE IF NOT EXISTS feedbacks
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,
    creator_id BIGINT      NOT NULL,
    title      VARCHAR(128) NOT NULL,
    content    TEXT         NOT NULL,
    status     SMALLINT    DEFAULT 1,
    reply      TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);

ALTER TABLE feedbacks ENABLE ROW LEVEL SECURITY;
ALTER TABLE feedbacks FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON feedbacks
    USING (tenant_id::text = current_setting('app.tenant_id', true));

CREATE INDEX IF NOT EXISTS idx_feedbacks_tenant_status ON feedbacks (tenant_id, status)
    WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_feedbacks_creator ON feedbacks (creator_id)
    WHERE is_deleted = FALSE;

INSERT INTO tenant_permissions (code, action, name, menu_id, status, is_deleted)
VALUES
    ('feedback', 'list',   '查看反馈', NULL, 1, FALSE),
    ('feedback', 'create', '提交反馈', NULL, 1, FALSE),
    ('feedback', 'update', '处理反馈', NULL, 1, FALSE),
    ('feedback', 'delete', '删除反馈', NULL, 1, FALSE)
ON CONFLICT (code, action) DO NOTHING;
```

如果表里有 JSONB 列，SQL 写 `value = $N::jsonb`（见 [database.md §9.2](database.md#92-pgx-jsonb-写入必须-jsonb-cast)）。

## 3. 公共接口定义（Step 2，可选）

仅当要給其他模块消费时才需要：

```go
// framework/pkg/tenant/auth/feedback.go
package auth

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

`AppContext.Reader/Writer` 各加一行。**编译会引导你完成所有必要的接线**。

## 4. 业务模块文件（Step 3）

新建 `apps/feedback/`，7 个文件：

```
apps/feedback/
├── errors.go
├── model.go
├── repository.go
├── service.go
├── handler.go
├── routes.go
├── module.go
└── types.go（可选，Request/Response 结构）
```

### 4.1 errors.go

```go
package feedback

import "gx1727.com/xin/framework/pkg/resp"

var (
    ErrNotFound      = resp.Err(15001, "反馈不存在")
    ErrTitleEmpty    = resp.Err(15002, "标题不能为空")
    ErrContentEmpty  = resp.Err(15003, "内容不能为空")
    ErrStatusInvalid = resp.Err(15004, "状态值无效")
)
```

错误码走空段。可用区段见 [architecture.md §7.1](architecture.md#71-错误码分段管理)。

### 4.2 model.go / repository.go / service.go / handler.go / routes.go

参考现有 `apps/tenant/user/` 或 `apps/reference/dict/` 的同名文件，模式完全一致。

### 4.3 module.go

```go
package feedback

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/appx"
    "gx1727.com/xin/framework/pkg/plugin"
)

func Module(app *appx.App) plugin.Module {
    return &plugin.BaseModule{
        NameStr: "feedback",
        InitFn: func(_ plugin.Reader, _ plugin.Writer) error {
            return nil
        },
        // Phase 0022：RegFn 接收三组 RouterGroup
        //   public      → /api/v1/public/*           （公开）
        //   tenant      → /api/v1/<resource>         （业务域，必须登录 + tenant_id；无 /t 前缀）
        //   protected   → /api/v1/platform/*         （平台域，必须 super_admin）
        RegFn: func(_ plugin.Reader, public, tenant, protected *gin.RouterGroup) {
            pool := app.DB
            svc := NewService(pool)
            h := NewHandler(svc)
            // 业务模块挂到 tenant group（受 RequireTenantContext 中间件保护）
            Register(tenant, h)
        },
    }
}
```

## 5. 在 main.go 注册（Step 5）

```go
import "gx1727.com/xin/apps/feedback"

modules := []plugin.Module{
    feedback.Module(app),
}
```

## 6. 在 cfg.Module 启用（Step 6）

```yaml
# config/config.yaml
module:
  - feedback
```

默认启用（optOut）则加到 `framework/pkg/config/config.go` 的 `optOutModules`。

## 7. 资源码 seed（Step 7）

见 Step 1 的 SQL 示例 + [`framework/pkg/permission/constants.go`](../framework/pkg/permission/constants.go) 加常量：

```go
const (
    ResFeedback = "feedback"
)
```

## 8. 测试（Step 8）

```go
package feedback

import (
    "context"
    "testing"
)

func TestService_List_NoRows(t *testing.T) {
    s := &Service{repo: NewRepository(nil)}
    _, _, err := s.List(context.Background(), 1, 20)
    if err == nil {
        t.Error("expected error from nil pool, got nil")
    }
}
```

---

# Step B：平台管理模块模板（`/platform/*` 域）

适用于 `sys_menu` / `tenants` / `platform_dict` 等。
**与业务模块的 4 个关键差异**：

1. 路径前缀统一 `/platform/<platform_resource>`
2. `adminGroup := protected.Group("/admin", RequirePlatformRole("super_admin"))` 分组级守卫
3. 数据表 `tenant_id = 0`（平台域）或独立的 `config_visibility` 类跨租户表
4. 写操作用 `db.RunInPlatformTx(ctx, pool, fn)` 跳过 RLS

## B.1 SQL 迁移

如果给现有表加平台维度（如给 `dicts` 加 platform scope），用 `*_alignment.sql`：

```sql
-- migrations/dict_alignment.sql
ALTER TABLE dicts ADD COLUMN IF NOT EXISTS scope VARCHAR(16) DEFAULT 'tenant';
ALTER TABLE dicts ADD COLUMN IF NOT EXISTS visibility VARCHAR(16) DEFAULT 'private';
-- scope: 'platform' | 'tenant'
-- visibility: 'public' | 'tenant_only' | 'hidden'

CREATE TABLE IF NOT EXISTS dict_visibility (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    dict_id     BIGINT      NOT NULL REFERENCES dicts(id) ON DELETE CASCADE,
    tenant_id   BIGINT      NOT NULL DEFAULT 0,
    access      VARCHAR(16) NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (dict_id, tenant_id)
);
```

如果是新表（平台独立），自由创建。

## B.2 业务模块文件

位置：`apps/platform/platform_dict/`（用 `apps/platform/` 命名空间，与业务模块区分）。

```
apps/platform/platform_dict/
├── errors.go
├── handler.go
├── routes.go
├── service.go
├── repository.go
├── model.go
├── module.go
└── types.go
```

### B.2.1 errors.go

错误码从 `resp/errors.go` 选空段。`sys_menu` 用 15001，`tenants` 复用 3xxx，新加的 `platform_dict` 用 17xxx：

```go
package platformdict

import "gx1727.com/xin/framework/pkg/resp"

var (
    ErrDictNotFound   = resp.Err(17001, "平台字典不存在")
    ErrDictCodeExists = resp.Err(17002, "平台字典 code 已存在")
    // ...
)
```

### B.2.2 routes.go（**关键差异**）

```go
package platformdict

import (
    "github.com/gin-gonic/gin"

    pkgmiddleware "gx1727.com/xin/framework/pkg/middleware"
    "gx1727.com/xin/framework/pkg/permission"
)

func Register(protected *gin.RouterGroup, h *Handler) {
    // 分组级 super_admin 守卫
    adminGroup := protected.Group("/admin",
        pkgmiddleware.RequirePlatformRole("super_admin"))

    g := adminGroup.Group("/platform-dicts")
    {
        // 平台 dict CRUD
        g.GET("", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActList)), h.List)
        g.GET("/:id", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActGet)), h.Get)
        g.POST("", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActCreate)), h.Create)
        g.PUT("/:id", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.Update)
        g.DELETE("/:id", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActDelete)), h.Delete)

        // visibility 子路由
        g.GET("/:id/visibility", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActList)), h.ListVisibility)
        g.POST("/:id/visibility", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.UpsertVisibility)
        g.DELETE("/:id/visibility/:tenant_id", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.DeleteVisibility)
    }
}
```

**双层守卫**：
- `adminGroup.Use(RequirePlatformRole("super_admin"))` — group 级短路所有非 super_admin
- 各路 `Require(ResDict.*)` — super_admin 内做资源码细分

> 单层守卫（只 super_admin）：如果模块不允许业务层 admin 越权，用单层（参考 `sys_menu`）。
> 双层守卫（super_admin + ResX）：如果 super_admin 也需细分权限，用双层（参考 `tenants`）。

### B.2.3 service.go（**关键差异**）

平台域写操作用 `db.RunInPlatformTx` 跳过 RLS：

```go
func (s *Service) CreateGroup(ctx context.Context, req CreateGroupReq) (*DictGroup, error) {
    var out *DictGroup
    err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
        var err error
        out, err = s.repo.CreatePlatformDict(txCtx, 0, CreateDictRepoReq{
            Code:      req.Code,
            Name:      req.Name,
            Scope:     "platform",
            Visibility: "public",
        })
        return err
    })
    if err != nil {
        return nil, err
    }
    return out, nil
}
```

> 关键常量：`platformTenantID = 0` —— 平台域数据 `tenant_id` 强制写 0，不接受外部传入。

### B.2.4 module.go

```go
package platformdict

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/appx"
    "gx1727.com/xin/framework/pkg/plugin"
)

func Module(app *appx.App) plugin.Module {
    return &plugin.BaseModule{
        NameStr: "platform_dict",
        InitFn: func(_ plugin.Reader, _ plugin.Writer) error {
            return nil
        },
        // Phase 0022：平台域挂到 protected group（/api/v1/platform-*）
        RegFn: func(_ plugin.Reader, public, tenant, protected *gin.RouterGroup) {
            pool := app.DB
            svc := NewService(pool, NewPostgresPlatformDictRepository(pool))
            h := NewHandler(svc)
            Register(protected, h)  // 平台域路由
        },
    }
}
```

## B.3 在 main.go 注册

```go
import "gx1727.com/xin/apps/platform/platformdict"

modules := []plugin.Module{
    platformdict.Module(app),
}
```

## B.4 错误码段

参考 [`architecture.md §7.1`](architecture.md#71-错误码分段管理) 选空段。已用段：

| 段 | 占用 |
|---|---|
| 1000-1999 | auth |
| 2000-2999 | user |
| 3000-3999 | tenant / platform_tenant |
| 4000-4999 | role |
| 5000-5999 | menu（sys_menu 复用） |
| 6000-7999 | org / permission |
| 8000-8999 | resource |
| 9000-9999 | asset |
| 10000-10999 | dict |
| 11000-11999 | system |
| 12000-12999 | weixin |
| 13000-13999 | flag |
| 14000-14999 | cms |
| **15000-15999** | sys_menu |
| **18000-18999** | config |
| **17xxx / 19xxx+** | 留给未来 platform_* / 新模块 |

---

# 通用步骤

## 9. 验收清单

```bash
go build ./...
go vet ./...
go test ./...
python scripts/strip_bom.py --check .

go run ./cmd/xin run &
sleep 3
curl http://localhost:8087/api/v1/health      # 必须 200
curl http://localhost:8087/api/v1/feedbacks   # 应该 200 或 403
```

启动日志应该看到：

```
2026/06/21 module feedback initialized
2026/06/21 module platform_dict initialized
```

## 10. 常见陷阱

| 陷阱 | 解决 |
|---|---|
| 忘了在 main.go 显式 import 模块 | `feedback.Module(app)` 没加进 `[]plugin.Module` |
| 忘了加 `cfg.Module` | module 在列表里但 Init/Register 跳过 |
| 错误码和别的模块撞了 | 查 `resp/errors.go` 选空段 |
| 资源码没 seed，`Permission.P` 直接写字符串 | 可以工作，但失去 IDE 自动补全 |
| 还在用 `db.Get()` / `bootx.Pool()` | 已删，改用 `app.DB` 显式注入 |
| RLS 没建，跨租户泄漏 | `ALTER TABLE xxx ENABLE ROW LEVEL SECURITY` + `FORCE` |
| 平台模块忘加 `RequirePlatformRole` | 路由在 `/platform/*` 域必须 super_admin 守卫 |
| 平台模块忘用 `db.RunInPlatformTx` | 平台域写操作会受 RLS 限制失败 |
| 平台模块 `tenant_id` 传错 | 强制 `platformTenantID = 0`，不接受外部传入 |
| `super_admin` 平台角色没 bypass 你的中间件 | 检查 `requireWithSpecs` 短路逻辑 |
| JSONB 列写入报 `42804` | SQL 加 `::jsonb` cast |
| 源文件编译报 `invalid BOM in the middle of the file` | 跑 `python scripts/strip_bom.py .` |
| gin 同路径下不同 param name 冲突 | `:id` 和 `:code` 不可在同一 segment 共存，统一用 `:id` |
| public / protected 同前缀 `/configs` 冲突 | public 路径改为 `/public/configs` |

## 11. 下一步

| 你想... | 看 |
|---|---|
| 看所有可用中间件 | [architecture.md §5](architecture.md#5-中间件链) |
| 理解 RBAC | [permissions.md](permissions.md) |
| 部署你的新模块 | [deployment.md](deployment.md) |
| 看完整 API | [api.md](api.md) |
