# XinFramework 开发者指南

## 1. 项目结构

```
XinFramework/
├── framework/                      # 核心框架 (gx1727.com/xin/framework)
│   ├── framework.go               # 框架入口、模块注册、路由设置
│   ├── cmd.go                     # 命令行处理 (start/stop/restart等)
│   ├── signal.go                  # 信号处理、优雅关闭
│   ├── pkg/                        # 公共包（可被外部apps引用）
│   │   ├── config/                 # YAML配置系统 + 环境变量覆盖
│   │   ├── db/                    # pgx/v5/pgxpool + 事务封装
│   │   ├── cache/                 # Redis客户端
│   │   ├── session/               # Session管理接口 + Redis/DB实现
│   │   ├── jwt/                   # Token生成/验证
│   │   ├── context/               # XinContext/UserContext上下文
│   │   ├── permission/            # 权限类型、接口、缓存
│   │   ├── plugin/                # 模块接口和注册表
│   │   ├── resp/                  # 统一响应格式
│   │   ├── migrate/              # SQL迁移运行器
│   │   ├── model/                 # 公共模型定义
│   │   ├── storage/              # 文件存储（local/COS）
│   │   └── dict/                 # 数据字典访问
│   └── internal/
│       ├── core/
│       │   ├── boot/              # 应用初始化 (boot.Init)
│       │   ├── server/            # HTTP服务器（优雅关闭）
│       │   └── middleware/        # 中间件（CORS/Logger/Recovery/Tenant/Auth）
│       ├── module/               # 12个内置模块
│       │   ├── auth/             # 登录/登出/注册
│       │   ├── user/             # 用户管理
│       │   ├── tenant/           # 租户管理
│       │   ├── role/             # 角色管理
│       │   ├── permission/      # 权限分配
│       │   ├── menu/            # 菜单管理（ltree）
│       │   ├── organization/    # 组织机构（ltree）
│       │   ├── resource/        # 按钮权限
│       │   ├── dict/            # 数据字典
│       │   ├── asset/           # 文件存储
│       │   ├── system/          # 健康检查
│       │   └── weixin/          # 微信登录（存根）
│       └── service/              # PermissionService
├── apps/                          # 外部业务插件
│   ├── cms/                       # CMS插件（扁平化架构）
│   └── flag/                      # Flag插件（头像/相框）
├── cmd/xin/                       # 程序入口点
├── config/                        # 配置文件
└── migrations/                   # SQL迁移文件
```

---

## 2. 配置系统

### 2.1 配置结构

配置文件：`config/config.yaml`

```yaml
app:
  name: xin
  env: dev
  host: 0.0.0.0
  port: 8080

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: xin
  sslmode: disable
  max_open_conns: 25
  max_idle_conns: 5

redis:
  host: 127.0.0.1
  port: 6379
  enabled: true
  required: false

jwt:
  secret: your-secret-key
  expire: 3600
  refresh_expire: 86400

storage:
  provider: local
  local_dir: ./uploads
  local_base_url: http://localhost:8080/uploads

log:
  dir: logs
  level: info

cors:
  enabled: true
  allow_origins: ["*"]
  allow_methods: "GET,POST,PUT,DELETE,OPTIONS"
  allow_headers: "*"

module:
  - auth
  - user
  - tenant
  - menu
  - role
  - organization
  - resource
  - permission
  - dict
  - asset
  - system
  - weixin

apps:
  - cms
  - flag
```

### 2.2 配置加载顺序

| 优先级 | 来源 | 说明 |
|--------|------|------|
| 1 | 环境变量 | `XIN_APP_PORT=9999` 覆盖任何配置 |
| 2 | `.env` 文件 | 项目根目录的 `.env` 文件 |
| 3 | `config.yaml` | YAML 配置文件 |
| 4 | 默认值 | 代码中的硬编码默认值 |

### 2.3 环境变量映射

| 配置项 | 环境变量前缀 | 示例 |
|--------|-------------|------|
| App | `XIN_APP_` | `XIN_APP_PORT=9999` |
| Database | `XIN_DB_` | `XIN_DB_HOST=localhost` |
| Redis | `XIN_REDIS_` | `XIN_REDIS_HOST=127.0.0.1` |
| JWT | `XIN_JWT_` | `XIN_JWT_SECRET=xxx` |
| 模块配置 | `XIN_<NAME>_` | `XIN_FLAG_xxx` |

### 2.4 添加新配置项

**步骤 1**：在 `framework/pkg/config/config.go` 的 `Config` 结构体中添加字段

```go
type Config struct {
    App      AppConfig      `yaml:"app"`
    Database DatabaseConfig `yaml:"database"`
    // ... 现有字段
    SMS      SMSConfig      `yaml:"sms"` // 新增
}
```

**步骤 2**：定义新的配置结构体

```go
type SMSConfig struct {
    Provider   string `yaml:"provider"`
    AccessKey  string `yaml:"access_key"`
    SecretKey  string `yaml:"secret_key"`
    SignName   string `yaml:"sign_name"`
}
```

**步骤 3**：添加环境变量覆盖（可选）

在 `overrideWithEnv()` 函数中添加：

```go
envStr("SMS_PROVIDER", &c.SMS.Provider)
envStr("SMS_ACCESS_KEY", &c.SMS.AccessKey)
```

**步骤 4**：在 `config.yaml` 中添加配置

```yaml
sms:
  provider: aliyun
  access_key: your-access-key
  secret_key: your-secret-key
  sign_name: XinFramework
```

---

## 3. 模块系统

### 3.1 Module 接口

所有模块（内置和外部）都实现 `plugin.Module` 接口：

```go
type Module interface {
    Name() string                    // 模块名称
    Init() error                    // 初始化（可选）
    Register(public, protected *gin.RouterGroup) // 注册路由
    Shutdown() error                // 关闭时清理（可选）
}
```

### 3.2 创建模块

#### 方式一：使用 plugin.NewModule（简单）

```go
package mymodule

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/plugin"
    "gx1727.com/xin/framework/pkg/db"
)

func Module() plugin.Module {
    return plugin.NewModule("mymodule", func(public, protected *gin.RouterGroup) {
        repo := NewRepository()
        h := NewHandler(repo)

        // 注册公开路由（可选认证）
        public.GET("/mymodule/public", h.PublicHandler)

        // 注册受保护路由（需要认证）
        protected.GET("/mymodule/data", h.DataHandler)
    })
}
```

#### 方式二：自定义模块结构（支持 Init/Shutdown）

```go
type module struct {
    name string
}

func (m *module) Name() string     { return m.name }
func (m *module) Init() error      { /* 初始化逻辑 */ return nil }
func (m *module) Shutdown() error  { /* 清理逻辑 */ return nil }

func (m *module) Register(public, protected *gin.RouterGroup) {
    repo := NewRepository()
    h := NewHandler(repo)
    Register(h, public, protected)
}

func Module() plugin.Module {
    return &module{name: "mymodule"}
}
```

### 3.3 注册模块

**内置模块**（在 `framework/framework.go` 中注册）：

```go
var builtinMap = map[string]plugin.Module{
    "auth":       authModule.Module(),
    "user":       userModule.Module(),
    "tenant":     tenantModule.Module(),
    // ...
}
```

**外部插件**（在 `apps/*/module.go` 中注册）：

```go
// apps/cms/module.go
func init() {
    plugin.Register(cms.Module())
}

// apps/cms/module.go
func Module() plugin.Module {
    return &module{name: "cms"}
}
```

### 3.4 启用模块

在 `config/config.yaml` 中配置：

```yaml
module:
  - auth
  - user
  - tenant
  # ...

apps:
  - cms
  - flag
```

---

## 4. 数据库访问

### 4.1 核心接口

```go
// framework/pkg/db/db.go

// Querier 数据库查询接口
type Querier interface {
    Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
    Query(ctx context.Context, sql string, optionsAndArgs ...any) (pgx.Rows, error)
    QueryRow(ctx context.Context, sql string, optionsAndArgs ...any) pgx.Row
}

// GetQuerier 从上下文获取查询器（自动适配事务）
func GetQuerier(ctx context.Context) (Querier, error)

// RunInTenantTx 在租户上下文中执行事务
func RunInTenantTx(ctx context.Context, pool *pgxpool.Pool, tenantID uint, fn func(ctx context.Context) error) error

// WithTx 将事务注入上下文
func WithTx(ctx context.Context, tx pgx.Tx) context.Context
```

### 4.2 Repository 模式

```go
// 示例：apps/flag/repository.go
package flag

import (
    "context"
    "gx1727.com/xin/framework/pkg/db"
)

type Repository struct{}

func NewRepository() *Repository {
    return &Repository{}
}

func (r *Repository) GetByID(ctx context.Context, id uint) (*Avatar, error) {
    q, err := db.GetQuerier(ctx)
    if err != nil {
        return nil, err
    }

    var avatar Avatar
    err = q.QueryRow(ctx,
        "SELECT id, name, url FROM flag_avatars WHERE is_deleted = FALSE AND id = $1",
        id,
    ).Scan(&avatar.ID, &avatar.Name, &avatar.URL)
    return &avatar, err
}

func (r *Repository) Create(ctx context.Context, a *Avatar) error {
    q, err := db.GetQuerier(ctx)
    if err != nil {
        return err
    }

    _, err = q.Exec(ctx,
        "INSERT INTO flag_avatars (name, url, tenant_id) VALUES ($1, $2, $3)",
        a.Name, a.URL, a.TenantID,
    )
    return err
}
```

### 4.3 事务控制

**Handler 层负责事务管理**：

```go
func (h *Handler) GetAvatar(c *gin.Context) {
    xc := xincontext.New(c)
    var avatar *Avatar

    // 使用 RunInTenantTx 提供租户上下文（触发 RLS）
    err := db.RunInTenantTx(c.Request.Context(), db.Get(), xc.GetTenantID(), func(ctx context.Context) error {
        var err error
        avatar, err = h.repo.GetByID(ctx, avatarID)
        return err
    })

    if err != nil {
        resp.ServerError(c, err.Error())
        return
    }

    resp.Success(c, avatar)
}
```

### 4.4 关键规则

| 层级 | 职责 | 禁止 |
|------|------|------|
| **Repository** | 数据访问，使用 `db.GetQuerier(ctx)` | 禁止直接使用 `db.Get()` |
| **Handler/Service** | 业务逻辑，事务控制，使用 `db.RunInTenantTx` | 禁止在 Repository 层开始事务 |

---

## 5. 中间件

### 5.1 中间件链顺序

```
1. Recovery()     — panic recovery，最先执行
2. RequestID()    — 请求ID生成/传播
3. CORS()         — 跨域资源共享 + OPTIONS预检
4. Logger()       — 请求日志（依赖 RequestID）
5. [Tenant]       — 租户上下文（可选）
6. [protected] → Auth() — JWT+Session验证
```

### 5.2 Auth 中间件变体

| 中间件 | Token验证 | XinContext | UserContext | 用途 |
|--------|-----------|------------|------------|------|
| `Auth()` | 必需 | 注入 | 懒加载 | 受保护路由 |
| `AuthLite()` | 必需 | 注入 | 不加载 | 轻量认证场景 |
| `OptionalAuth()` | 可选 | 有则注入 | 有则懒加载 | 公共接口 |

### 5.3 使用示例

```go
func setupRouter(app *boot.App) {
    srv := app.Server
    cfg := app.Config

    srv.Engine.Use(middleware.Recovery())
    srv.Engine.Use(middleware.RequestID())
    srv.Engine.Use(middleware.CORS(&cfg.CORS))
    srv.Engine.Use(middleware.Logger())

    v1 := r.Group("/api/v1")
    public := v1.Group("")
    public.Use(middleware.OptionalAuth(&cfg.JWT, app.SessionMgr, app.PermService))

    protected := v1.Group("")
    protected.Use(middleware.Auth(&cfg.JWT, app.SessionMgr, app.PermService))

    // 注册模块路由
    // ...
}
```

---

## 6. 上下文系统

### 6.1 XinContext（轻量级身份）

```go
// framework/pkg/context/context.go

type XinContext struct {
    TenantID  uint
    UserID    uint
    SessionID string
    Role      string
}

func New(c *gin.Context) *XinContext
func (x *XinContext) GetUserID() uint
func (x *XinContext) GetTenantID() uint
func (x *XinContext) GetSessionID() string
func (x *XinContext) GetRole() string
```

### 6.2 UserContext（RBAC + DataScope）

```go
type UserContext struct {
    *XinContext
    OrgID       int64
    Roles       []string
    Permissions map[string]bool
    DataScope   permission.DataScope
}

// 懒加载：首次访问时从数据库加载
func (x *XinContext) MustNewUserContext(ctx context.Context) *UserContext
```

### 6.3 使用示例

```go
func MyHandler(c *gin.Context) {
    xc := xincontext.New(c)

    // 获取基本信息
    userID := xc.GetUserID()
    tenantID := xc.GetTenantID()

    // 需要权限时，获取完整上下文
    uc := xc.MustNewUserContext(c.Request.Context())
    if !uc.Permissions["post:create"] {
        resp.Forbidden(c, "权限不足")
        return
    }
}
```

---

## 7. 权限系统

### 7.1 权限格式

```
"resource_code:action"
例如：
  "user:list"    — 查看用户列表
  "user:create"  — 创建用户
  "user:update"  — 修改用户
  "user:delete"  — 删除用户
  "*:*"          — 超级管理员（所有权限）
```

### 7.2 DataScope（数据范围）

| 值 | 名称 | 说明 |
|----|------|------|
| 1 | DataScopeAll | 所有数据（租户内） |
| 2 | DataScopeCustom | 自定义机构范围 |
| 3 | DataScopeDept | 本部门数据 |
| 4 | DataScopeDeptAndBelow | 本部门及下级数据 |
| 5 | DataScopeSelf | 仅本人数据 |

### 7.3 权限检查

**中间件方式**：

```go
router.GET("/users",
    middleware.RequirePermission("user", "list"),
    handler.ListUsers,
)
```

**代码方式**：

```go
func MyHandler(c *gin.Context) {
    xc := xincontext.New(c)
    uc := xc.MustNewUserContext(c.Request.Context())

    if !uc.HasPermission("post:create") {
        resp.Forbidden(c, "权限不足")
        return
    }
    // ...
}
```

---

## 8. 响应格式

### 8.1 统一响应结构

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

### 8.2 响应函数

| 函数 | HTTP状态 | 业务码 | 用途 |
|------|---------|--------|------|
| `resp.Success(c, data)` | 200 | 0 | 成功响应 |
| `resp.Error(c, code, msg)` | 200 | 自定义 | 业务错误 |
| `resp.Unauthorized(c, msg)` | 401 | 401 | 未认证 |
| `resp.Forbidden(c, msg)` | 403 | 403 | 无权限 |
| `resp.BadRequest(c, msg)` | 400 | 400 | 参数错误 |
| `resp.NotFound(c, msg)` | 404 | 404 | 资源不存在 |
| `resp.ServerError(c, msg)` | 500 | 500 | 系统错误 |
| `resp.Paginate(c, total, list)` | 200 | 0 | 分页列表 |

### 8.3 使用示例

```go
// 成功
resp.Success(c, map[string]interface{}{
    "id": 1,
    "name": "test",
})

// 失败
resp.Error(c, 1001, "用户名或密码错误")

// 分页
resp.Paginate(c, 100, []User{...})
```

---

## 9. 启动与关闭

### 9.1 启动流程

```
main() [cmd/xin/main.go]
  → config.Load("config/config.yaml")
  → framework.Run(cfg)

framework.Run(cfg)
  → boot.Init(cfg)                  # logger → db → cache → session → PermService
  → initModules(cfg)                # 内置模块.Init() + 插件.Init()
  → runMigrations()                 # migrate.Run("migrations")
  → setupRouter(app)               # 中间件链 + 路由注册
  → srv.Start(addr)                # HTTP 服务器启动
  → waitForSignal(srv, app)        # 信号处理 → 优雅关闭
```

### 9.2 命令行接口

| 命令 | 功能 |
|------|------|
| `./xin run` | 前台运行 |
| `./xin start` | 守护进程模式（写入 PID） |
| `./xin stop` | 优雅停止（SIGTERM，30s 超时） |
| `./xin restart` | 重启 |
| `./xin reload` | 热重载（SIGUSR1） |
| `./xin hot-restart` | 零宕机重启 |
| `./xin status` | 查看状态 |
| `./xin help` | 帮助 |

---

## 10. 常见问题

### Q: 如何添加一个新的内置模块？

1. 在 `framework/internal/module/` 下创建目录
2. 实现 `module.go`（Module 函数）和 `routes.go`（Register 函数）
3. 在 `framework/framework.go` 的 `builtinMap` 中注册

### Q: 如何在 Apps 模块中访问 Framework 的表？

使用 `db.GetQuerier(ctx)` 获取查询器，它会自动适配当前的事务上下文（包括 RLS 所需的 tenant_id 设置）。

### Q: 权限检查失败的原因？

1. 检查是否使用了正确的 Auth 中间件
2. 检查 UserContext 是否正确加载
3. 检查数据库中 `permissions` 表的数据

### Q: RLS 策略不生效？

1. 确保在 `db.RunInTenantTx` 内执行操作
2. 检查 `set_config('app.tenant_id', ...)` 是否正确执行
3. 检查表是否正确启用了 RLS