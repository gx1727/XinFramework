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

1. 加载 `config/config.yaml`
2. 加载 `.env` 文件（如果存在）
3. 环境变量覆盖（优先级最高）

---

## 2. 日志系统

### 2.1 日志初始化

日志在 `boot.Init()` 中初始化：

```go
logger.Init(cfg.Log.Dir, cfg.Log.Level)
```

### 2.2 日志级别

| 级别 | 值 | 说明 |
|:-----|:--:|:-----|
| DEBUG | 0 | 调试信息 |
| INFO | 1 | 一般信息（默认） |
| WARN | 2 | 警告 |
| ERROR | 3 | 错误 |

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

---

## 3. 添加新模块

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

---

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

---

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

```go
err := db.Transaction(func(tx *gorm.DB) error {
    // 创建用户
    if err := tx.Create(&user).Error; err != nil {
        return err
    }

    // 创建用户角色关联
    if err := tx.Create(&userRole).Error; err != nil {
        return err
    }

    return nil
})
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

---

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

---

## 7. JWT 使用

### 7.1 生成 Token

```go
import "gx1727.com/xin/pkg/jwt"

token, err := jwt.Generate(&cfg.JWT, userID, tenantID, role)
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

---

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

---

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
