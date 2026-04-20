# XinFramework

轻量的 Go SaaS 基础框架，当前已具备：
- 配置管理（YAML + `.env` 覆盖）
- PostgreSQL + Redis 基础设施
- JWT 登录鉴权 + Session 校验（Redis 优先，DB 兜底）
- 多租户中间件（可开关）
- 按天分割日志（`logs/YYYY-MM-DD.log`）
- 跨平台启动/信号兼容（Windows / Unix）

## 技术栈

- Web：Gin
- ORM：GORM（PostgreSQL 驱动）
- Cache：go-redis/v8
- Auth：JWT（`github.com/golang-jwt/jwt/v5`）
- Config：YAML + 环境变量覆盖

## 项目结构（当前实际）

```text
xin/
├── cmd/server/                   # 启动入口、命令控制、信号处理
│   ├── main.go
│   ├── cmd.go
│   ├── signal.go
│   ├── cmd_compat_unix.go
│   ├── cmd_compat_windows.go
│   ├── signal_compat_unix.go
│   └── signal_compat_windows.go
│
├── config/                       # 仅配置文件（无 Go 代码）
│   ├── config.yaml
│   ├── config.dev.yaml
│   └── config.prod.yaml
│
├── api/v1/
│   └── register.go               # v1 路由注册（薄路由层）
│
├── internal/
│   ├── core/
│   │   ├── boot/boot.go          # 初始化与优雅关闭
│   │   ├── context/context.go
│   │   ├── middleware/middleware.go
│   │   └── server/server.go
│   ├── infra/
│   │   ├── db/db.go
│   │   ├── cache/cache.go
│   │   ├── logger/logger.go
│   │   └── session/session.go    # 会话存储（Redis/DB）
│   └── module/
│       ├── auth/                 # 登录模块（已拆分）
│       │   ├── handler.go
│       │   ├── service.go
│       │   ├── password.go
│       │   └── types.go
│       └── user/model.go
│
├── pkg/
│   ├── config/config.go          # 配置加载与 env 覆盖
│   ├── jwt/jwt.go
│   └── resp/resp.go
│
├── migrations/001_init.sql
├── .env.example
└── go.mod
```

## 快速开始

### 1. 准备依赖

- PostgreSQL（必需）
- Redis（可选，见配置 `redis.enabled`）

### 2. 配置

复制并修改环境变量（可选）：

```bash
cp .env.example .env
```

主配置文件：
- `config/config.yaml`

> 运行时优先级：环境变量 > `.env` > `config.yaml`

### 3. 初始化数据库

执行：
- `migrations/001_init.sql`

> 脚本包含：多租户/RLS、认证相关表、`auth_sessions`（Session DB 兜底）

### 4. 启动

```bash
go run ./cmd/server
```

或：

```bash
go run ./cmd/server run
```

服务默认监听：`0.0.0.0:8080`

## 运行命令

`cmd/server/main.go` 支持：

- `run`：前台运行
- `start`：后台启动
- `stop`：停止
- `restart`：重启
- `reload`：热加载信号（Windows 下不支持）
- `status`：状态检查
- `hot-restart`：热重启流程

## API（当前）

公开路由：

- `GET /api/v1/health`
- `POST /api/v1/login`
- `POST /api/v1/logout`

受保护示例路由（占位）：

- `GET /api/v1/users`
- `POST /api/v1/users`
- `PUT /api/v1/users/:id`
- `DELETE /api/v1/users/:id`

## 登录与会话机制

1. `/login` 校验账号（`username/phone/email`）与密码  
2. 生成 `session_id`  
3. Session 存储策略：
   - Redis 可用：写 Redis（TTL）
   - Redis 不可用：写 `auth_sessions` 表（DB 兜底）
4. 签发 JWT（Claims 含 `sid`）  
5. `Auth` 中间件每次请求校验 `sid` 是否有效  
6. `/logout` 撤销会话（Redis 删 key / DB 删记录）

## 多租户模式

配置项：
- `saas.mode`

行为：
- 空值：单租户模式（不注入 tenant）
- 非空：启用租户注入（从 `X-Tenant-ID` 读取）

数据库侧 RLS 已支持“未设置 tenant 时放行、设置后按 tenant 过滤”的策略。

## 业务域开关（Domain）

业务代码按目录放在：
- `internal/module/system`
- `internal/module/cms`
- `internal/module/weixin`
- 其他业务建议继续按 `internal/module/<domain>` 增加

配置项：
- `domain`（YAML 数组）
- `XIN_DOMAIN`（环境变量，逗号分隔）

允许值仅：
- `system`
- `cms`
- `weixin`

启动时会严格校验，出现其他值会直接启动失败。  
路由注册按开关生效：未启用的业务域不会注册路由，因此不可访问。

## 关键配置项

### 数据库连接池

- `database.max_open_conns`
- `database.max_idle_conns`
- `database.conn_max_lifetime_sec`
- `database.conn_max_idle_time_sec`

### Redis 开关与连接池

- `redis.enabled`
- `redis.required`
- `redis.pool_size`
- `redis.min_idle_conns`
- `redis.pool_timeout_sec`
- `redis.idle_timeout_sec`
- `redis.max_conn_age_sec`

### 日志

- `log.dir`（默认 `logs`）
- `log.level`（`debug|info|warn|error`）

## 优雅关闭

退出流程：
1. HTTP Server Shutdown
2. `boot.Shutdown()`
3. 关闭 Redis
4. 关闭 DB 连接池
5. 关闭日志文件句柄

## 跨平台兼容

`cmd/server` 已使用 build tags 做平台兼容：

- Unix：`*_unix.go`
- Windows：`*_windows.go`

无需额外配置，`go build` 会自动选择对应文件。

## 开发说明

- 数据库设计规范：`doc/database-conventions.md`
- API 调试示例：`doc/api.http`
- 其他文档：`doc/developer-guide.md`、`doc/handbook.md`

## 编译

```bash
go build ./...
```
