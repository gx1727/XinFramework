# XinFramework 开发者指南

## 1. 配置系统

### 1.1 配置结构

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

redis:
  host: 192.168.151.176
  port: 6379
  enabled: true
  required: false

jwt:
  secret: your-secret-key
  expire: 3600
  refresh_expire: 86400

saas:
  mode: shared

log:
  dir: logs
  level: info

module:
  - weixin  # 内置模块开关

apps:
  - cms     # 外部插件开关

user:
  max_login_attempts: 5
  lock_duration_sec: 300
  # ... 用户模块配置
```

### 1.2 添加新配置项

**步骤 1**：在 `framework/pkg/config/config.go` 的 `Config` 结构体中添加字段

```go
type Config struct {
    App      AppConfig      `yaml:"app"`
    Database DatabaseConfig `yaml:"database"`
    Redis    RedisConfig    `yaml:"redis"`
    JWT      JWTConfig      `yaml:"jwt"`
    Saas     SaasConfig     `yaml:"saas"`
    Log      LogConfig      `yaml:"log"`
    Modules  []string       `yaml:"module"`
    Apps     []string       `yaml:"apps"`
    User     UserConfig     `yaml:"user"`
    // 新增：SMS 配置
    SMS      SMSConfig      `yaml:"sms"`
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

### 1.3 配置加载顺序

1. 加载 `config/config.yaml`
2. 环境变量覆盖（优先级最高）

### 1.4 模块独立配置

#### 1.4.1 文件结构

```
config/
├── config.yaml              # 主配置
├── config.dev.yaml
├── config.prod.yaml
└── modules/                 # 业务模块配置目录（如果使用）
    ├── auth.yaml
    └── cms.yaml
```

#### 1.4.2 在模块中定义配置结构体

以 user 模块为例，`framework/internal/module/user/config.go`：

```go
package user

import "gx1727.com/xin/framework/pkg/config"

type UserConfig struct {
    MaxLoginAttempts      int    `yaml:"max_login_attempts"`
    LockDurationSec       int    `yaml:"lock_duration_sec"`
    PasswordPolicy        string `yaml:"password_policy"`
    TokenExpireSec        int    `yaml:"token_expire_sec"`
    RefreshTokenExpireSec int    `yaml:"refresh_token_expire_sec"`
}

var moduleCfg *UserConfig

func Cfg() *UserConfig {
    return moduleCfg
}

func InitConfig() error {
    moduleCfg = &UserConfig{
        MaxLoginAttempts:      5,
        LockDurationSec:       300,
        PasswordPolicy:        "standard",
        TokenExpireSec:        3600,
        RefreshTokenExpireSec: 86400,
    }
    return config.LoadModule("user", moduleCfg)
}
```

#### 1.4.3 环境变量覆盖

模块配置支持环境变量覆盖，规则为 `XIN_<模块名大写>_<字段名大写>`：

```bash
XIN_USER_MAX_LOGIN_ATTEMPTS=10
XIN_USER_LOCK_DURATION_SEC=600
```

***

## 2. 日志系统

### 2.1 日志初始化

日志在 `boot.Init()` 中初始化：

```go
logger.Init(cfg.Log.Dir, cfg.Log.Level)
```

### 2.2 日志级别

| 级别 | 值 | 说明 |
|------|---|------|
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
- 输出到按天分割的文件：`{log.dir}/2026-04-22.log`

### 2.5 模块自定义日志

使用 `logger.Module("<prefix>")` 获取模块日志器：

```go
package user

import "gx1727.com/xin/framework/pkg/logger"

var userLogger *logger.Logger

func InitConfig() error {
    userLogger = logger.Module("user")
    if userLogger != nil {
        userLogger.Infof("user module config loaded")
    } else {
        logger.Infof("user module config loaded")
    }
    return nil
}
```

文件命名规则：
- 全局日志器：`{log.dir}/YYYY-MM-DD.log`
- 模块日志器：`{log.dir}/user-YYYY-MM-DD.log`

***

## 3. 添加新模块

### 3.0 架构说明

模块分两类：
1. **内置模块** (`framework/internal/module/`) - 通过 `module:` 配置控制
2. **外部插件** (`apps/*`) - 调用 `framework.RegisterModule()` 注册

### 3.1 外部插件（推荐）

创建 `apps/mymodule/mymodule.go`：

```go
package mymodule

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/config"
    "gx1727.com/xin/framework/pkg/migrate"
    "gx1727.com/xin/framework/pkg/plugin"
)

type MyConfig struct {
    FeatureX bool `yaml:"feature_x"`
}

var moduleCfg *MyConfig

func Cfg() *MyConfig {
    return moduleCfg
}

func Register(public, protected *gin.RouterGroup) {
    protected.GET("/mymodule/ping", func(c *gin.Context) {
        // handler logic
    })
}

func Module() plugin.Module {
    return plugin.NewModuleWithOpts("mymodule", Register,
        plugin.WithInit(initModule),
        plugin.WithMigrate(migrateModule),
    )
}

func initModule() error {
    moduleCfg = &MyConfig{FeatureX: true}
    return config.LoadModule("mymodule", moduleCfg)
}

func migrateModule() error {
    return migrate.Run("apps/mymodule/migrations")
}
```

在 `cmd/xin/main.go` 中注册：

```go
func main() {
    cfg, _ := config.Load("config/config.yaml")

    if cfg.AppEnabled("mymodule") {
        framework.RegisterModule(mymodule.Module())
    }

    framework.Run(cfg)
}
```

### 3.2 内置模块

在 `framework/internal/module/` 下创建目录，包含：
- `routes.go` - 路由注册 + `Module()` 函数
- `handler.go` - HTTP 处理层
- `service.go` - 业务逻辑层（可选）
- `model.go` - 数据模型（可选）

示例：`framework/internal/module/mymodule/routes.go`

```go
package mymodule

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/plugin"
)

func Register(public, protected *gin.RouterGroup) {
    protected.GET("/mymodule/ping", func(c *gin.Context) {
        // ...
    })
}

func Module() plugin.Module {
    return plugin.NewModule("mymodule", Register)
}
```

内置模块默认全部启用，可在 `config.yaml` 的 `module:` 列表中控制。

### 3.3 分层规范

推荐 C-S（Handler-Service）两层，复杂业务可扩展到 H-S-R-M：

| 层 | 职责 | 禁止 |
|---|---|---|
| **Handler** | 参数校验、响应组装 | 写业务逻辑、写 SQL |
| **Service** | 业务规则、事务边界 | 依赖 Gin 上下文 |
| **Repo** | 数据访问（可选） | 跨表操作 |
| **Model** | 数据结构 | 业务逻辑 |

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
```

### 4.3 终止请求

```go
func MyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if someCondition {
            resp.Unauthorized(c, "不符合条件")
            c.Abort()
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
import "gx1727.com/xin/framework/pkg/db"

d := db.Get() // 获取全局 db 实例
```

### 5.2 事务规范

| 层 | 职责 | 禁止 |
|---|---|---|
| **Handler** | 不感知事务 | 调用 Begin/Commit/Rollback |
| **Service** | 定义事务边界 | 直接写 SQL |
| **Repo** | 接受 `*gorm.DB` 执行操作 | 自己开事务 |

**Service 层事务示例**：

```go
func (s *Service) CreateWithRole(user *User, roleID uint) error {
    d := db.Get()
    return d.Transaction(func(tx *gorm.DB) error {
        if err := s.repo.Create(tx, user); err != nil {
            return err
        }
        if err := s.repo.AssignRole(tx, user.ID, roleID); err != nil {
            return err
        }
        return nil
    })
}
```

### 5.3 租户查询

Tenant 中间件自动设置 `SET app.tenant_id = ?`，GORM 查询自动带上租户过滤。

手动设置：

```go
d := db.Get()
d.Exec("SET app.tenant_id = ?", tenantID)
defer d.Exec("RESET app.tenant_id")
```

***

## 6. 缓存操作

### 6.1 使用缓存

```go
import "gx1727.com/xin/framework/pkg/cache"

client := cache.Get()

// 设置值
client.Set(ctx, "key", "value", time.Hour)

// 获取值
val, err := client.Get(ctx, "key").Result()

// 删除
client.Del(ctx, "key")
```

Redis 配置中 `enabled=false` 时 `cache.Get()` 返回 nil。

***

## 7. JWT 使用

### 7.1 生成 Token

```go
import jwtpkg "gx1727.com/xin/framework/pkg/jwt"

// token 内包含 sid（SessionID）
token, err := jwtpkg.Generate(&cfg.JWT, userID, tenantID, role, sessionID)
```

### 7.2 验证

框架提供 `middleware.Auth()` 中间件自动验证并设置 `XinContext.UserID`。

***

## 8. 编译与运行

### 8.1 编译

```bash
# 构建脚本
./build.ps1          # Windows
./build.sh           # Linux

# 手动构建
go build -ldflags="-s -w" -o ./out/xin ./cmd/xin
```

### 8.2 运行

```bash
# 前台运行
go run ./cmd/xin run

# 守护进程模式
./out/xin start

# 查看状态
./out/xin status

# 停止
./out/xin stop
```

### 8.3 热重载

```bash
air  # 使用 go install github.com/air-verse/air@latest 安装
```

***

## 9. 目录结构参考

```
XinFramework/
├── apps/                      # 外部业务插件
│   └── cms/
│       ├── cms.go
│       └── migrations/
├── cmd/xin/                   # 入口点
│   └── main.go
├── config/
│   └── config.yaml
├── framework/                  # 核心框架
│   ├── framework.go
│   ├── cmd.go
│   ├── signal.go
│   ├── api/v1/
│   │   └── register.go        # 路由注册
│   ├── pkg/                   # 公共包
│   │   ├── config/
│   │   ├── db/
│   │   ├── cache/
│   │   ├── logger/
│   │   ├── session/
│   │   ├── jwt/
│   │   ├── migrate/
│   │   ├── plugin/
│   │   └── resp/
│   └── internal/
│       ├── core/
│       │   ├── boot/
│       │   ├── server/
│       │   ├── middleware/
│       │   └── context/
│       └── module/
│           ├── auth/          # 占位符
│           ├── user/
│           ├── system/
│           └── weixin/
└── migrations/
    └── framework/
```