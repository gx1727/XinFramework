# XinFramework 模块开发指南

本文档详细说明如何在 XinFramework 中添加**内部模块**（内置模块）和**外部插件**（apps 模块）。

---

## 一、架构概览

XinFramework 采用**插件化架构**，支持两种类型的模块：

| 类型 | 位置 | 特点 | 示例 |
|------|------|------|------|
| **内部模块** | `framework/internal/module/*` | 框架核心功能，通过配置文件启用 | auth, user, tenant, menu, dict, role, etc. |
| **外部插件** | `apps/*` | 独立业务模块，可插拔部署 | cms, flag |

### 模块加载流程

```
framework.Run(cfg)
  → boot.Init(cfg)                    # 初始化 logger, db, cache, session, permService
  → initModules(cfg)                  # 初始化所有模块
    ├─ 内置模块: cfg.Module 列表中的模块
    └─ 外部插件: plugin.Apps() 注册的模块
  → setupRouter(app)                  # 配置路由和中间件
    ├─ 注册全局中间件 (Recovery, RequestID, CORS, Logger)
    └─ registerModules(srv.Engine)
      ├─ 内置模块: m.Register(public, protected)
      └─ 外部插件: m.Register(public, protected)
  → srv.Start(addr)                   # 启动 HTTP 服务器
```

---

## 二、添加内部模块（Built-in Module）

内部模块是框架的核心组成部分，位于 `framework/internal/module/` 目录下。

### 2.1 目录结构规范

以 `dict` 模块为例（简单 CRUD 模块）：

```
framework/internal/module/dict/
├── module.go          # 模块定义入口（必需）
├── handler.go         # HTTP 处理器（必需）
├── routes.go          # 路由注册（必需）
├── repository.go      # 数据访问层（可选，根据复杂度决定）
├── service.go         # 业务逻辑层（可选，复杂业务需要）
├── model.go           # 数据模型（可选）
├── types.go           # 类型定义（可选）
└── errors.go          # 错误定义（可选）
```

### 2.2 核心文件实现

#### 步骤 1: 创建模块入口 `module.go`

```go
package yourmodule

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回模块的完整定义
func Module() plugin.Module {
	return plugin.NewModule("yourmodule", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		// 1. 初始化 Repository（如果需要数据库访问）
		repo := NewYourRepository(db.Get())
		
		// 2. 初始化 Service（如果业务逻辑复杂）
		svc := NewService(repo)
		
		// 3. 创建 Handler
		h := NewHandler(svc)
		
		// 4. 注册路由
		Register(public, protected, h)
	})
}
```

**关键点：**
- 使用 `plugin.NewModule(name, registerFunc)` 创建模块
- `name` 是模块唯一标识，必须与配置文件中的名称一致
- `registerFunc` 接收 `public` 和 `protected` 两个路由组
- 依赖注入顺序：Repository → Service → Handler → Routes

#### 步骤 2: 实现路由注册 `routes.go`

```go
package yourmodule

import (
	"github.com/gin-gonic/gin"
)

// Register 注册模块的所有路由
func Register(public *gin.RouterGroup, protected *gin.RouterGroup, h *Handler) {
	// 公开路由（无需认证）
	public.GET("/yourmodule/public-endpoint", h.PublicEndpoint)
	
	// 受保护路由（需要认证 + 权限检查）
	protected.GET("/yourmodule/items", h.ListItems)
	protected.GET("/yourmodule/items/:id", h.GetItem)
	protected.POST("/yourmodule/items", h.CreateItem)
	protected.PUT("/yourmodule/items/:id", h.UpdateItem)
	protected.DELETE("/yourmodule/items/:id", h.DeleteItem)
}
```

**路由分组规则：**
- `public`: 登录、注册等无需认证的接口
- `protected`: 需要 JWT 认证和权限检查的接口
- 可使用中间件细化权限：`middleware.RequireAuthenticated()`, `middleware.RequirePermission("resource:action")`

#### 步骤 3: 实现 HTTP 处理器 `handler.go`

```go
package yourmodule

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/context"
	"gx1727.com/xin/framework/pkg/resp"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListItems(c *gin.Context) {
	// 1. 解析请求参数
	var req listRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	
	// 2. 获取用户上下文（租户 ID、用户 ID 等）
	xc := context.New(c)
	
	// 3. 调用业务逻辑
	items, err := h.svc.ListItems(c.Request.Context(), xc.GetTenantID(), req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	
	// 4. 返回统一响应
	resp.Success(c, items)
}

func (h *Handler) CreateItem(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	
	xc := context.New(c)
	item, err := h.svc.CreateItem(c.Request.Context(), xc.GetTenantID(), xc.GetUserID(), req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	
	resp.Success(c, item)
}
```

**Handler 设计原则：**
- **薄层设计**：只做参数校验、上下文提取、错误处理，不包含业务逻辑
- **统一响应**：始终使用 `resp.Success()` / `resp.HandleError()` / `resp.BadRequest()` 等
- **上下文传递**：通过 `context.New(c)` 获取 `XinContext`（包含 TenantID, UserID, SessionID, Role）

#### 步骤 4: 实现数据访问层 `repository.go`（可选）

```go
package yourmodule

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type YourRepository struct {
	db *pgxpool.Pool
}

func NewYourRepository(db *pgxpool.Pool) *YourRepository {
	return &YourRepository{db: db}
}

// 多租户查询必须使用 GetTenantQuerier
func (r *YourRepository) ListItems(ctx context.Context, tenantID uint, page, pageSize int) ([]Item, error) {
	q := db.GetTenantQuerier(r.db, tenantID)
	
	rows, err := q.Query(ctx, 
		`SELECT id, name, description FROM your_items 
		 WHERE is_deleted = FALSE 
		 ORDER BY created_at DESC 
		 LIMIT $1 OFFSET $2`,
		pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Description); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	
	return items, nil
}
```

**Repository 关键规范：**
- **多租户隔离**：必须使用 `db.GetTenantQuerier(db, tenantID)` 而非 `db.GetQuerier(db)`
- **事务安全**：多租户查询建议开启事务
- **软删除过滤**：所有查询必须包含 `WHERE is_deleted = FALSE`

#### 步骤 5: 实现业务逻辑层 `service.go`（可选）

```go
package yourmodule

import (
	"context"
)

type Service struct {
	repo *YourRepository
}

func NewService(repo *YourRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListItems(ctx context.Context, tenantID uint, req listRequest) ([]Item, error) {
	// 业务逻辑：参数验证、数据转换、缓存处理等
	return s.repo.ListItems(ctx, tenantID, req.Page, req.PageSize)
}
```

**何时需要 Service 层？**
- 需要协调多个 Repository
- 需要复杂的业务规则验证
- 需要缓存、消息队列等外部服务集成
- **简单 CRUD 可直接在 Handler 中调用 Repository**

### 2.3 注册模块到框架

#### 步骤 6: 在 `framework/framework.go` 中导入并注册

```go
// 1. 导入模块包
import (
	yourModule "gx1727.com/xin/framework/internal/module/yourmodule"
	// ... 其他模块
)

// 2. 添加到 builtinMap
var builtinMap = map[string]plugin.Module{
	"asset":        assetModule.Module(),
	"auth":         authModule.Module(),
	// ... 其他模块
	"yourmodule":   yourModule.Module(),  // 新增
}
```

#### 步骤 7: 在配置文件中启用模块

编辑 `config/config.yaml`：

```yaml
module:
  - auth
  - tenant
  - user
  - yourmodule  # 新增模块名称
```

或通过环境变量启用：

```bash
export XIN_MODULE=auth,tenant,user,yourmodule
```

### 2.4 完整示例清单

创建新内部模块的完整文件清单：

```
✓ framework/internal/module/yourmodule/
  ├── module.go          # 模块定义（必需）
  ├── routes.go          # 路由注册（必需）
  ├── handler.go         # HTTP 处理器（必需）
  ├── repository.go      # 数据访问（推荐）
  ├── service.go         # 业务逻辑（可选）
  ├── model.go           # 数据模型（可选）
  ├── types.go           # 类型定义（可选）
  └── errors.go          # 错误定义（可选）

✓ framework/framework.go
  - 导入模块包
  - 添加到 builtinMap

✓ config/config.yaml
  - 在 module 列表中添加模块名称

✓ migrations/
  - 添加数据库迁移 SQL（如需要）
```

---

## 三、添加外部插件（External Plugin）

外部插件位于 `apps/` 目录下，是独立的 Go 模块，可单独构建和部署。

### 3.1 目录结构规范

以 `cms` 模块为例：

```
apps/cms/
├── go.mod               # 独立的 Go 模块定义（必需）
├── go.sum               # 依赖锁定文件（自动生成）
├── module.go            # 插件模块定义（必需）
├── routes.go            # 路由注册（必需）
├── handler.go           # HTTP 处理器（必需）
├── types.go             # 类型定义（可选）
└── doc/
    └── api.md           # API 文档（推荐）
```

### 3.2 核心文件实现

#### 步骤 1: 初始化 Go 模块

在 `apps/cms/` 目录下执行：

```bash
cd apps/cms
go mod init gx1727.com/xin/apps/cms
go mod tidy
```

**`go.mod` 示例：**

```go
module gx1727.com/xin/apps/cms

go 1.21

require (
	gx1727.com/xin/framework v0.0.0
	github.com/gin-gonic/gin v1.9.1
)

replace gx1727.com/xin/framework => ../../framework
```

**关键点：**
- 使用 `replace` 指令指向本地 framework 路径
- 模块路径遵循 `gx1727.com/xin/apps/<name>` 规范

#### 步骤 2: 创建模块定义 `module.go`

有两种实现方式：

**方式 A：手动实现 plugin.Module 接口（推荐）**

```go
package cms

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
)

type module struct {
	name string
}

func (m *module) Name() string    { return m.name }
func (m *module) Init() error     { return nil }
func (m *module) Shutdown() error { return nil }

func (m *module) Register(public *gin.RouterGroup, protected *gin.RouterGroup) {
	// 1. 初始化 Repository（如果需要）
	// InitRepositories(db.Get())
	
	// 2. 创建 Handler
	h := NewHandler()
	
	// 3. 注册路由
	Register(h, public, protected)
}

func Module() plugin.Module {
	return &module{name: "cms"}
}
```

**方式 B：使用 plugin.NewModule 辅助函数**

```go
package cms

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
)

func Module() plugin.Module {
	return plugin.NewModule("cms", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		h := NewHandler()
		Register(h, public, protected)
	})
}
```

**两种方式的区别：**
- **方式 A**：更灵活，可实现自定义的 `Init()` 和 `Shutdown()` 逻辑
- **方式 B**：更简洁，适合无特殊初始化需求的模块

#### 步骤 3: 实现路由注册 `routes.go`

```go
package cms

import (
	"github.com/gin-gonic/gin"
)

func Register(h *Handler, public *gin.RouterGroup, protected *gin.RouterGroup) {
	// 公开路由
	public.GET("/cms/ping", h.Ping)
	
	// 受保护路由
	protected.GET("/cms/me", h.GetCurrentUser)
	protected.GET("/cms/posts", h.ListPosts)
	protected.GET("/cms/posts/:id", h.GetPost)
	protected.POST("/cms/posts", h.CreatePost)
	protected.PUT("/cms/posts/:id", h.UpdatePost)
	protected.DELETE("/cms/posts/:id", h.DeletePost)
}
```

#### 步骤 4: 实现 HTTP 处理器 `handler.go`

```go
package cms

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/context"
	"gx1727.com/xin/framework/pkg/resp"
)

type Handler struct {
	// 如果需要访问 Repository，在此声明
	// postRepo *PostRepository
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Ping(c *gin.Context) {
	resp.Success(c, gin.H{"message": "pong"})
}

func (h *Handler) ListPosts(c *gin.Context) {
	xc := context.New(c)
	
	// 示例：直接返回模拟数据
	resp.Success(c, gin.H{
		"tenant_id": xc.GetTenantID(),
		"user_id":   xc.GetUserID(),
		"posts":     []string{},
	})
}
```

**外部插件的特殊性：**
- 可以访问 framework 的公共包（`pkg/*`）
- **不能**直接访问 `framework/internal/*`（internal 包对外部不可见）
- 如需数据库访问，通过 `db.Get()` 获取连接池
- 如需 Session 管理，通过框架提供的接口间接使用

#### 步骤 5: 在插件中注册到框架

创建 `init.go` 或在 `module.go` 中添加 init 函数：

```go
package cms

import (
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	// 自动注册到插件系统
	plugin.Register(Module())
}
```

**或者在主程序中显式注册**（见步骤 7）。

### 3.3 在主程序中集成插件

#### 步骤 6: 在主程序 `cmd/xin/main.go` 中导入插件

```go
package main

import (
	"log"
	
	_ "gx1727.com/xin/apps/cms"  // 导入 CMS 插件（触发 init）
	_ "gx1727.com/xin/apps/flag" // 导入 Flag 插件
	
	"gx1727.com/xin/framework"
	"gx1727.com/xin/framework/pkg/config"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}
	
	framework.Run(cfg)
}
```

**关键点：**
- 使用 `_` 空白导入，仅触发 `init()` 函数
- 插件的 `init()` 会调用 `plugin.Register()` 自动注册

#### 步骤 7: 在配置文件中启用插件

编辑 `config/config.yaml`：

```yaml
apps:
  - cms
  - flag
```

或通过环境变量：

```bash
export XIN_APPS=cms,flag
```

### 3.4 完整示例清单

创建新外部插件的完整文件清单：

```
✓ apps/yourplugin/
  ├── go.mod               # Go 模块定义（必需）
  ├── go.sum               # 依赖锁定（自动生成）
  ├── module.go            # 插件模块定义（必需）
  ├── routes.go            # 路由注册（必需）
  ├── handler.go           # HTTP 处理器（必需）
  ├── init.go              # 自动注册（可选，推荐）
  ├── types.go             # 类型定义（可选）
  └── doc/
      └── api.md           # API 文档（推荐）

✓ cmd/xin/main.go
  - 导入插件包: _ "gx1727.com/xin/apps/yourplugin"

✓ config/config.yaml
  - 在 apps 列表中添加插件名称

✓ go.work（根目录）
  - 确保包含新的 apps 模块路径

✓ migrations/
  - 添加插件专用的数据库迁移 SQL（如需要）
```

---

## 四、内部模块 vs 外部插件对比

| 特性 | 内部模块 | 外部插件 |
|------|----------|----------|
| **位置** | `framework/internal/module/*` | `apps/*` |
| **可见性** | 可访问所有 internal 包 | 只能访问 pkg 公共包 |
| **依赖注入** | 直接在 module.go 中注入 | 通过 pkg 接口间接访问 |
| **构建方式** | 随框架一起构建 | 可单独构建，也可集成到主程序 |
| **适用场景** | 框架核心功能（auth, user, tenant） | 业务扩展功能（cms, flag, 自定义业务） |
| **耦合度** | 高耦合，紧密集成 | 低耦合，可插拔 |
| **测试独立性** | 需依赖框架整体环境 | 可独立测试（mock framework 接口） |
| **发布策略** | 随框架版本发布 | 可独立版本管理 |

---

## 五、最佳实践

### 5.1 模块设计原则

1. **单一职责**：每个模块只负责一个业务领域
2. **依赖倒置**：高层模块不依赖低层模块的具体实现
3. **接口分离**：暴露最小化的公共接口
4. **YAGNI 原则**：不要过度设计，简单 CRUD 不需要 Service 层

### 5.2 命名规范

- **模块名称**：小写字母 + 下划线（如 `user_management`）或纯小写（如 `user`）
- **包名**：与目录名一致，简短有意义
- **Handler 方法**：动词 + 名词（如 `ListUsers`, `CreatePost`）
- **Repository 方法**：语义清晰（如 `FindByID`, `ListByTenant`）

### 5.3 错误处理

```go
// 定义模块专用错误
var (
	ErrItemNotFound = errors.New("item not found")
	ErrInvalidInput = errors.New("invalid input")
)

// Handler 中统一处理
if err != nil {
	resp.HandleError(c, err)  // 自动映射 HTTP 状态码
	return
}
```

### 5.4 多租户安全

```go
// ✅ 正确：使用 GetTenantQuerier
q := db.GetTenantQuerier(pool, tenantID)

// ❌ 错误：直接使用 GetQuerier（可能导致跨租户数据泄露）
q := db.GetQuerier(pool)
```

### 5.5 路由前缀规范

- 内部模块：`/api/v1/{module_name}/*`
- 外部插件：`/api/v1/{plugin_name}/*`
- 避免路由冲突：不同模块使用不同的路径前缀

---

## 六、常见问题

### Q1: 如何选择内部模块还是外部插件？

**选择内部模块：**
- 框架核心功能（认证、授权、用户管理）
- 所有租户都需要的基础功能
- 需要深度集成框架内部组件

**选择外部插件：**
- 特定业务场景的功能
- 可能独立部署或禁用的功能
- 第三方集成功能

### Q2: 外部插件如何访问数据库？

```go
import "gx1727.com/xin/framework/pkg/db"

func (m *module) Register(public, protected *gin.RouterGroup) {
	pool := db.Get()  // 获取全局数据库连接池
	repo := NewRepository(pool)
	// ...
}
```

### Q3: 如何实现模块间的依赖？

**内部模块之间：**
```go
import "gx1727.com/xin/framework/internal/module/tenant"

func Module() plugin.Module {
	return plugin.NewModule("yourmodule", func(public, protected *gin.RouterGroup) {
		tenantRepo := tenant.NewTenantRepository(db.Get())
		// 使用 tenantRepo
	})
}
```

**外部插件依赖内部模块：**
- 通过 framework 提供的公共接口（pkg/*）
- 避免直接 import internal 包

### Q4: 如何调试新模块？

1. 在模块代码中添加日志：
   ```go
   import "gx1727.com/xin/framework/pkg/logger"
   
   logger.Info("module initialized")
   ```

2. 查看日志文件：
   ```bash
   tail -f logs/2026-05-31.log
   ```

3. 测试 API 端点：
   ```bash
   curl http://localhost:8080/api/v1/yourmodule/ping
   ```

### Q5: 模块初始化失败怎么办？

检查以下几点：
1. 模块名称是否在配置文件中正确配置
2. `builtinMap` 中是否正确注册（内部模块）
3. `plugin.Register()` 是否被调用（外部插件）
4. 依赖的包是否正确导入
5. 查看启动日志中的错误信息

---

## 七、快速开始模板

### 内部模块模板

```bash
# 1. 创建目录
mkdir -p framework/internal/module/yourmodule

# 2. 创建 module.go
cat > framework/internal/module/yourmodule/module.go << 'EOF'
package yourmodule

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
)

func Module() plugin.Module {
	return plugin.NewModule("yourmodule", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		h := NewHandler(NewRepository(db.Get()))
		Register(public, protected, h)
	})
}
EOF

# 3. 创建 routes.go, handler.go, repository.go
# ...（参考上文示例）

# 4. 在 framework.go 中注册
# 5. 在 config.yaml 中启用
```

### 外部插件模板

```bash
# 1. 创建目录
mkdir -p apps/yourplugin

# 2. 初始化 Go 模块
cd apps/yourplugin
go mod init gx1727.com/xin/apps/yourplugin
go mod tidy

# 3. 创建 module.go
cat > module.go << 'EOF'
package yourplugin

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
)

func Module() plugin.Module {
	return plugin.NewModule("yourplugin", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		h := NewHandler()
		Register(h, public, protected)
	})
}

func init() {
	plugin.Register(Module())
}
EOF

# 4. 创建 routes.go, handler.go
# 5. 在 cmd/xin/main.go 中导入
# 6. 在 config.yaml 中启用
```

---

## 八、总结

| 步骤 | 内部模块 | 外部插件 |
|------|----------|----------|
| 1. 创建目录 | `framework/internal/module/name/` | `apps/name/` |
| 2. 实现核心文件 | module.go, routes.go, handler.go | module.go, routes.go, handler.go, go.mod |
| 3. 注册到框架 | 修改 `framework.go` 的 `builtinMap` | 在 `main.go` 中导入或 `init()` 自动注册 |
| 4. 配置文件启用 | `config.yaml` 的 `module` 列表 | `config.yaml` 的 `apps` 列表 |
| 5. 数据库迁移 | `migrations/xxx.sql` | `migrations/xxx.sql` |

**核心要点：**
- 内部模块适合框架核心功能，外部插件适合业务扩展
- 遵循统一的模块接口 `plugin.Module`
- 严格区分 `public` 和 `protected` 路由
- 多租户查询必须使用 `GetTenantQuerier`
- 保持模块间松耦合，避免循环依赖
