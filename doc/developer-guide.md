# XinFramework 开发者指南

## 1. 配置系统

### 1.1 配置结构

配置文件：`config/config.yaml`

```yaml
app:
  name: xin-framework
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

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

jwt:
  secret: your-secret-key
  expire: 3600
  refresh_expire: 86400

saas:
  mode: shared

log:
  dir: logs
  level: info
```

### 1.2 添加新配置项

**步骤 1**：在 `pkg/config/config.go` 的 `Config` 结构体中添加字段

```go
type Config struct {
    App      AppConfig      `yaml:"app"`
    Database DatabaseConfig `yaml:"database"`
    Redis    RedisConfig    `yaml:"redis"`
    JWT      JWTConfig      `yaml:"jwt"`
    Saas     SaasConfig     `yaml:"saas"`
    Log      LogConfig      `yaml:"log"`
    // 新增：SMS 配置
    SMS      SMSConfig      `yaml:"sms"`
}
```

**步骤 2**：定义新的配置结构体

```go
type SMSConfig struct {
    Provider   string `yaml:"provider"`   // aliyun, tencent
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

### 1.3 配置加载顺序

1. 加载 `config/config.yaml`（框架公共配置）
2. 加载 `.env` 文件（如果存在）
3. 环境变量覆盖（优先级最高）
4. 加载各业务模块独立配置 `config/modules/<name>.yaml`

### 1.4 业务模块独立配置

每个业务模块可以拥有自己的配置文件，物理隔离、互不干扰。

#### 1.4.1 文件结构

```
config/
├── config.yaml              # 框架公共配置（app/database/redis/jwt/log/domain）
├── config.dev.yaml
├── config.prod.yaml
└── modules/                 # 业务模块配置目录
    ├── auth.yaml            # auth 模块专属配置
    ├── cms.yaml             # cms 模块专属配置
    └── weixin.yaml          # weixin 模块专属配置
```

#### 1.4.2 在模块中定义配置结构体

以 auth 模块为例，`internal/module/auth/config.go`：

```go
package auth

import "gx1727.com/xin/pkg/config"

type AuthConfig struct {
    MaxLoginAttempts      int    `yaml:"max_login_attempts"`
    LockDurationSec       int    `yaml:"lock_duration_sec"`
    PasswordPolicy        string `yaml:"password_policy"`
    TokenExpireSec        int    `yaml:"token_expire_sec"`
    RefreshTokenExpireSec int    `yaml:"refresh_token_expire_sec"`
}

var moduleCfg *AuthConfig

func Cfg() *AuthConfig {
    return moduleCfg
}

func InitConfig() error {
    moduleCfg = &AuthConfig{
        MaxLoginAttempts:      5,
        LockDurationSec:       300,
        PasswordPolicy:        "standard",
        TokenExpireSec:        3600,
        RefreshTokenExpireSec: 86400,
    }
    return config.LoadModule("auth", moduleCfg)
}
```

#### 1.4.3 配置文件示例

`config/modules/auth.yaml`：

```yaml
max_login_attempts: 5
lock_duration_sec: 300
password_policy: standard
token_expire_sec: 3600
refresh_token_expire_sec: 86400
```

#### 1.4.4 环境变量覆盖

模块配置同样支持环境变量覆盖，规则为 `XIN_<模块名大写>_<字段名大写>`：

```bash
XIN_AUTH_MAX_LOGIN_ATTEMPTS=10
XIN_AUTH_LOCK_DURATION_SEC=600
XIN_AUTH_PASSWORD_POLICY=strict
```

#### 1.4.5 在 boot.Init 中注册

在 `internal/core/boot/boot.go` 的 `loadModuleConfigs` 函数中添加新模块：

```go
func loadModuleConfigs(cfg *config.Config) error {
    if err := auth.InitConfig(); err != nil {
        return err
    }
    // 新增模块时在此添加：
    // if cfg.DomainEnabled("cms") {
    //     if err := cms.InitConfig(); err != nil {
    //         return err
    //     }
    // }
    return nil
}
```

#### 1.4.6 设计原则

| 原则 | 说明 |
|------|------|
| 文件物理隔离 | 每个模块一个 YAML，改 auth 配置不影响 cms |
| 结构体默认值 | `InitConfig` 中设置合理默认值，配置文件不存在也不报错 |
| 环境变量覆盖 | 框架用 `XIN_DB_*`，模块用 `XIN_AUTH_*`、`XIN_CMS_*`，互不冲突 |
| 按需加载 | 可结合 `cfg.DomainEnabled()` 按域开关决定是否加载 |

***

## 2. 日志系统

### 2.1 日志初始化

日志在 `boot.Init()` 中初始化：

```go
logger.Init(cfg.Log.Dir, cfg.Log.Level)
```

### 2.2 日志级别

| 级别    |  值  | 说明       |
| :---- | :-: | :------- |
| DEBUG |  0  | 调试信息     |
| INFO  |  1  | 一般信息（默认） |
| WARN  |  2  | 警告       |
| ERROR |  3  | 错误       |

### 2.3 日志函数

```go
// 格式化输出
logger.Debugf("用户登录: userID=%d", userID)
logger.Infof("请求处理完成: path=%s duration=%dms", path, duration)
logger.Warnf("配置缺失: key=%s", key)
logger.Errorf("数据库错误: %v", err)

// 原始内容输出
logger.Debug("收到请求")
logger.Info("任务完成")
```

### 2.4 日志输出

- 输出到标准输出（stdout）
- 输出到按天分割的文件：`{log.dir}/2026-01-01.log`

### 2.5 添加新日志场景

```go
import "gx1727.com/xin/internal/infra/logger"

// 在业务代码中使用
func DoSomething() {
    logger.Infof("执行操作: param=%s", param)
    // ...
    if err != nil {
        logger.Errorf("操作失败: %v", err)
    }
}
```

***

## 3. 添加新模块

### 3.0 业务域与分层规范（结论）

为保证可维护性和可扩展性，项目采用“先分域，再分层”的规范。

#### 3.0.1 业务代码放置规范

- 业务代码统一放在：`internal/module/<domain>/`
- 当前标准域：`system`、`cms`、`weixin`
- 每个域独立维护自己的路由、业务、数据访问，不允许跨域直接操作对方数据表

#### 3.0.2 配置开关规范（Domain）

在配置中启用业务域：

```yaml
domain:
  - system
  - cms
  - weixin
```

或使用环境变量：

```bash
XIN_DOMAIN=system,cms,weixin
```

启动时只允许以上三个值，出现其他值会直接启动失败。未启用域不会注册路由，因此接口不可访问。

#### 3.0.3 分层规范（推荐 C-S-M，复杂域扩展到 H-S-R-M）

每个业务域至少包含：

- `routes.go`：路由注册（按 domain 开关启用）
- `handler.go`：Controller/Handler 层（参数校验、错误映射、响应组装）
- `service.go`：业务规则层（流程编排、领域逻辑）
- `model.go`/`types.go`：数据结构（持久化模型、DTO/VO）

复杂业务建议增加：

- `repo.go`：数据访问层（Repository）
- `validator.go`：输入校验规则
- `policy.go`：权限策略

落地原则：

- Handler 不写复杂业务与 SQL
- Service 不依赖 Gin 上下文
- Repo 不处理业务规则

#### 3.0.4 跨模块依赖规范（允许 / 禁止清单）

为避免模块之间耦合失控，跨模块依赖遵循“单向依赖 + 最小契约 + 不共享 ORM 细节”。

允许：

- 模块 A 调用模块 B 的 **Service 能力**（导出的函数/方法），例如 `auth -> user` 调用 `user.ResolveLoginIdentity(...)`
- 跨模块传递 **最小必要的契约结构**（DTO/VO），优先放在被依赖模块内；只有确实被多个模块复用时，才抽到一个共享位置（例如 `internal/shared/*`）
- 共享基础设施能力统一走 `internal/infra/*`（db/cache/logger/session）和 `pkg/*`（config/jwt/resp）

禁止：

- 双向/循环依赖（A import B，同时 B import A）
- 直接 import 另一个模块的持久化模型并复用（例如直接复用对方的 `model.go` 里的 GORM struct）
- 直接跨模块访问对方的数据表（在自己模块里写 SQL 去操作另一个模块的核心表），应通过对方模块暴露的能力完成

### 3.1 目录结构

以用户模块为例：

```
internal/module/user/
├── handler.go    # HTTP 处理层
├── service.go    # 业务逻辑层
├── repo.go       # 数据访问层
└── model.go      # 数据模型
```

### 3.2 定义模型

`internal/module/user/model.go`：

```go
package user

import "time"

type User struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    TenantID  uint      `gorm:"index" json:"tenant_id"`
    Username  string    `gorm:"size:64;uniqueIndex" json:"username"`
    Password  string    `gorm:"size:255" json:"-"`
    Email     string    `gorm:"size:128" json:"email"`
    Status    int8      `gorm:"default:1" json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string {
    return "users"
}
```

### 3.3 实现 Repository

`internal/module/user/repo.go`：

```go
package user

import "gorm.io/gorm"

type Repo struct {
    db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
    return &Repo{db: db}
}

func (r *Repo) Create(user *User) error {
    return r.db.Create(user).Error
}

func (r *Repo) GetByID(id uint) (*User, error) {
    var user User
    err := r.db.Where("id = ? AND is_deleted = FALSE", id).First(&user).Error
    return &user, err
}

func (r *Repo) List(tenantID uint, offset, limit int) ([]User, int64, error) {
    var users []User
    var total int64

    query := r.db.Model(&User{}).Where("tenant_id = ? AND is_deleted = FALSE", tenantID)
    query.Count(&total)
    query.Offset(offset).Limit(limit).Find(&users)

    return users, total, nil
}
```

### 3.4 实现 Service

`internal/module/user/service.go`：

```go
package user

import "gx1727.com/xin/pkg/resp"

type Service struct {
    repo *Repo
}

func NewService(repo *Repo) *Service {
    return &Service{repo: repo}
}

func (s *Service) Create(tenantID uint, username, email, password string) error {
    user := &User{
        TenantID: tenantID,
        Username: username,
        Email:    email,
        Password: password, // 实际应加密
    }
    return s.repo.Create(user)
}

func (s *Service) GetUser(id uint) (*User, error) {
    return s.repo.GetByID(id)
}
```

### 3.5 实现 Handler

`internal/module/user/handler.go`：

```go
package user

import (
    "strconv"

    "github.com/gin-gonic/gin"
    "gx1727.com/xin/internal/core/context"
    "gx1727.com/xin/pkg/resp"
)

type Handler struct {
    svc *Service
}

func NewHandler(svc *Service) *Handler {
    return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
    r.GET("/users", h.List)
    r.POST("/users", h.Create)
    r.GET("/users/:id", h.Get)
    r.PUT("/users/:id", h.Update)
    r.DELETE("/users/:id", h.Delete)
}

func (h *Handler) List(c *gin.Context) {
    ctx := context.New(c)
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

    offset := (page - 1) * pageSize
    users, total, err := h.svc.List(ctx.TenantID, offset, pageSize)
    if err != nil {
        resp.ServerError(c, "查询失败")
        return
    }

    resp.Paginate(c, total, users)
}

func (h *Handler) Create(c *gin.Context) {
    // 实现创建逻辑
}

func (h *Handler) Get(c *gin.Context) {
    // 实现获取单个逻辑
}

func (h *Handler) Update(c *gin.Context) {
    // 实现更新逻辑
}

func (h *Handler) Delete(c *gin.Context) {
    // 实现删除逻辑
}
```

### 3.6 注册路由

在 `api/v1/register.go` 中注册：

```go
package v1

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/internal/core/context"
    "gx1727.com/xin/internal/infra/db"
    "gx1727.com/xin/internal/module/user"
    userRepo "gx1727.com/xin/internal/module/user"
    userService "gx1727.com/xin/internal/module/user"
)

func RegisterRoutes(r *gin.Engine) {
    v1 := r.Group("/api/v1")
    {
        v1.GET("/health", healthCheck)

        // 用户模块
        userHandler := initUserHandler()
        auth := v1.Group("/users")
        auth.Use(middleware.Auth(&cfg.JWT))
        userHandler.RegisterRoutes(auth)
    }
}

func initUserHandler() *userHandler.Handler {
    db := db.Get()
    repo := userRepo.NewRepo(db)
    svc := userService.NewService(repo)
    return userHandler.NewHandler(svc)
}
```

***

## 4. 中间件开发

### 4.1 中间件模板

```go
package middleware

import (
    "github.com/gin-gonic/gin"
)

func MyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 前置处理

        c.Next() // 执行业务逻辑

        // 后置处理
    }
}
```

### 4.2 获取请求数据

```go
func MyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 请求头
        header := c.GetHeader("X-Custom-Header")

        // Query 参数
        query := c.Query("key")

        // Path 参数
        pathParam := c.Param("id")

        // Form 参数
        form := c.PostForm("field")

        // JSON Body
        var body map[string]interface{}
        c.ShouldBindJSON(&body)

        c.Next()
    }
}
```

### 4.3 设置响应数据

```go
func MyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 在 context 中设置值，后续 handler 可以获取
        c.Set("request_id", uuid.New().String())

        c.Next()

        // 后置处理可以访问响应
        // 注意：此时响应已经写入客户端
    }
}
```

### 4.4 终止请求

```go
func MyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if someCondition {
            resp.Unauthorized(c, "不符合条件")
            c.Abort() // 阻止后续 handler 执行
            return
        }

        c.Next()
    }
}
```

***

## 5. 数据库操作

### 5.1 获取数据库连接

```go
import "gx1727.com/xin/internal/infra/db"

func SomeFunc() {
    db := db.Get() // 获取全局 db 实例
    // ...
}
```

### 5.2 基础 CRUD

```go
// 创建
db.Create(&user)

// 查询单个
db.Where("id = ?", id).First(&user)

// 条件查询
db.Where("tenant_id = ? AND status = ?", tenantID, 1).Find(&users)

// 更新
db.Model(&user).Updates(map[string]interface{}{
    "email": "new@example.com",
})

// 删除（软删除）
db.Delete(&user) // UPDATE users SET is_deleted = TRUE WHERE id = ?
```

### 5.3 事务

#### 5.3.1 事务边界规范

| 层 | 职责 | 禁止 |
|---|---|---|
| **Handler** | 不感知事务 | 禁止调用 Begin/Commit/Rollback |
| **Service** | 定义事务边界，调用 `d.Transaction(...)` | 禁止直接写 SQL |
| **Repo** | 接受 `*gorm.DB`（可能是 tx）执行操作 | 禁止自己开事务，禁止跨表操作 |

核心原则：**事务边界由 Service 层控制，Repo 层负责单表操作，Handler 层完全不感知事务。**

事务的本质是"一组业务操作必须原子完成"——这是业务规则，不是数据访问细节。Repo 只负责"对一张表做 CRUD"，不知道自己的操作要和其他操作在同一个事务里；Service 知道。

#### 5.3.2 Repo 层：接受 `*gorm.DB` 参数

Repo 的每个方法接受 `*gorm.DB` 参数，由 Service 决定传入普通连接还是事务连接：

```go
// repo.go
package user

import "gorm.io/gorm"

type Repo struct{}

func (r *Repo) Create(d *gorm.DB, user *User) error {
    return d.Create(user).Error
}

func (r *Repo) GetByID(d *gorm.DB, id uint) (*User, error) {
    var u User
    err := d.Where("id = ? AND is_deleted = FALSE", id).First(&u).Error
    return &u, err
}

func (r *Repo) AssignRole(d *gorm.DB, userID, roleID uint) error {
    return d.Create(&UserRole{UserID: userID, RoleID: roleID}).Error
}
```

#### 5.3.3 Service 层：定义事务边界

**不需要事务的简单场景**（单表操作）——直接传 `db.Get()`：

```go
func (s *Service) GetUser(id uint) (*User, error) {
    return s.repo.GetByID(db.Get(), id)
}
```

**需要事务的场景**（多表操作必须原子完成）——用 `d.Transaction(...)` 闭包：

```go
func (s *Service) CreateWithRole(user *User, roleID uint) error {
    d := db.Get()

    return d.Transaction(func(tx *gorm.DB) error {
        // 第 1 步：创建用户（tx 保证在同一事务内）
        if err := s.repo.Create(tx, user); err != nil {
            return err // 返回 error 自动 Rollback
        }

        // 第 2 步：分配角色
        if err := s.repo.AssignRole(tx, user.ID, roleID); err != nil {
            return err // 返回 error 自动 Rollback
        }

        return nil // 返回 nil 自动 Commit
    })
}
```

GORM 的 `Transaction` 闭包会自动处理 Commit（返回 nil）和 Rollback（返回 error），不需要手动调用。

#### 5.3.4 跨模块事务

当一个业务操作涉及多个模块的 Repo 时，事务仍在 Service 层控制：

```go
// order/service.go
func (s *Service) CreateOrder(order *Order, userID uint) error {
    d := db.Get()

    return d.Transaction(func(tx *gorm.DB) error {
        // 本模块 Repo
        if err := s.orderRepo.Create(tx, order); err != nil {
            return err
        }

        // 调用其他模块 Repo（通过对方模块暴露的能力）
        if err := user.UpdateUserStats(tx, userID); err != nil {
            return err
        }

        return nil
    })
}
```

被调用模块需要暴露接受 `*gorm.DB` 的函数，以支持外部传入事务连接。

#### 5.3.5 简单查询不需要事务

对于只读或单表操作，不需要包裹事务，直接用 `db.Get()` 即可：

```go
func (s *Service) List(tenantID uint, offset, limit int) ([]User, int64, error) {
    return s.repo.List(db.Get(), tenantID, offset, limit)
}
```

### 5.4 租户查询

在 `Tenant` 中间件处理后，后续查询自动带上租户过滤：

```go
// 内部已设置 SET app.tenant_id = ?，GORM 查询会自动过滤
users, _ := userRepo.List(tenantID, offset, limit)
```

手动设置租户查询：

```go
db := db.Get()
db.Exec("SET app.tenant_id = ?", tenantID)
defer db.Exec("RESET app.tenant_id")
```

***

## 6. 缓存操作

### 6.1 初始化

```go
import "gx1727.com/xin/internal/infra/cache"

cache.Init(&cfg.Redis)
```

### 6.2 使用缓存

```go
client := cache.Get()

// 设置值
client.Set(ctx, "key", "value", 0) // 0 = 永不过期
client.Set(ctx, "key", "value", time.Hour) // 1小时过期

// 获取值
val, err := client.Get(ctx, "key").Result()

// 删除
client.Del(ctx, "key")

// 存在性检查
exists := client.Exists(ctx, "key").Val()
```

***

## 7. JWT 使用

### 7.1 生成 Token

```go
import "gx1727.com/xin/pkg/jwt"

// token 内已包含 sid（SessionID），用于服务端会话校验
token, err := jwt.Generate(&cfg.JWT, userID, tenantID, role, sessionID)
```

### 7.2 在中间件中验证

框架已提供 `middleware.Auth()` 中间件，自动验证并设置 `XinContext.UserID`。

手动验证：

```go
token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    return []byte(cfg.JWT.Secret), nil
})
if err != nil {
    // Token 无效
}
claims := token.Claims.(jwt.MapClaims)
userID := uint(claims["user_id"].(float64))
```

***

## 8. 编译与运行

### 8.1 编译

```bash
# 开发环境编译
go build -o xin-server.exe ./cmd/server/main.go

# 生产环境编译（去除调试信息）
go build -ldflags="-s -w" -o xin-server.exe ./cmd/server/main.go
```

### 8.2 运行

```bash
# 设置环境变量（可选）
export APP_ENV=dev
export DB_HOST=localhost

# 运行
./xin-server.exe
```

### 8.3 热重载（开发）

推荐使用 `air`：

```bash
# 安装
go install github.com/air-verse/air@latest

# 运行
air
```

***

## 9. 目录结构参考

```
internal/
├── core/                    # 框架核心
│   ├── boot/               # 启动初始化
│   ├── server/             # Gin 封装
│   ├── middleware/         # 中间件
│   └── context/           # 上下文
├── module/                 # 业务模块
│   ├── user/              # 用户模块
│   ├── auth/              # 认证模块
│   └── saas/              # 多租户模块
└── infra/                  # 基础设施
    ├── db/                 # 数据库
    ├── cache/             # 缓存
    └── logger/            # 日志

pkg/                        # 可复用组件
├── config/                 # 配置
├── jwt/                    # JWT
└── resp/                   # 响应
```

