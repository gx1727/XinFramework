# XinFramework

<div align="center">

**轻量的 Go SaaS 基础框架** — 不用 ORM，手写 SQL 的企业级开发框架

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow?style=flat-square)](LICENSE)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat-square&logo=postgresql)](https://www.postgresql.org/)

</div>

---

## ✨ 为什么选择 XinFramework

| 特性 | 说明 |
|------|------|
| 🚫 **无 ORM** | 直接使用 `pgx/v5`，完全掌控 SQL |
| 🔐 **安全认证** | Argon2id 密码加密 + JWT + Session 双层校验 |
| 🏢 **多租户** | 完整的租户隔离，支持行级安全策略 (RLS) |
| 📝 **轻量日志** | 按天分割，自动归档，支持模块分离 |
| 🌐 **跨平台** | Windows / Linux / macOS 全平台兼容 |
| 🧩 **插件架构** | 业务模块热插拔，按需启用 |
| ⚡ **高性能** | pgx 连接池 + Redis 连接池优化 |

---

## 🛠️ 技术栈

| 领域 | 技术 | 版本 |
|------|------|------|
| Web 框架 | [Gin](https://github.com/gin-gonic/gin) | v1.12.0 |
| 数据库驱动 | [pgx/v5](https://github.com/jackc/pgx) | v5.9.0 |
| Redis 客户端 | [go-redis/redis](https://github.com/go-redis/redis) | v8.11.5 |
| JWT | [golang-jwt/jwt](https://github.com/golang-jwt/jwt) | v5.2.2 |
| 密码加密 | Argon2id | golang.org/x/crypto |
| 配置解析 | yaml.v3 | gopkg.in/yaml.v3 |

---

## 📂 项目结构

```
xin/
├── cmd/xin/                     # 程序入口
│   └── main.go                   # 启动入口、插件注册
│
├── framework/                    # 框架核心
│   ├── framework.go              # Run() 主函数
│   ├── cmd.go                    # 命令控制 (start/stop/restart)
│   ├── signal.go                  # 信号处理 (优雅关闭)
│   │
│   ├── pkg/                      # 公共包
│   │   ├── config/               # 配置加载 (YAML + env)
│   │   ├── db/                    # PostgreSQL (pgx) + 租户会话
│   │   ├── cache/                # Redis 客户端
│   │   ├── logger/               # 日志 (按天分割)
│   │   ├── session/              # Session 管理 (Redis/DB)
│   │   ├── jwt/                  # Token 工具
│   │   ├── migrate/              # SQL 迁移
│   │   ├── plugin/               # 插件注册机制
│   │   └── resp/                 # 统一响应封装
│   │
│   ├── internal/
│   │   ├── core/                 # 核心组件
│   │   │   ├── boot/             # 初始化流程
│   │   │   ├── server/           # HTTP Server + 优雅关闭
│   │   │   ├── middleware/        # 中间件栈
│   │   │   └── context/          # 请求上下文 (租户/用户)
│   │   │
│   │   └── module/               # 内置模块
│   │       ├── user/             # 用户认证
│   │       ├── tenant/           # 租户管理
│   │       ├── system/           # 系统模块
│   │       └── weixin/           # 微信模块
│   │
│   └── api/v1/                   # API 路由注册
│
├── apps/                         # 外部插件 (可扩展)
│   └── cms/
│
├── config/                       # 配置文件
│   ├── config.yaml
│   ├── config.dev.yaml
│   └── config.prod.yaml
│
└── migrations/                   # SQL 迁移脚本
```

---

## 🚀 快速开始

### 前置要求

- Go 1.21+
- PostgreSQL 15+
- Redis (可选)

### 1. 克隆项目

```bash
git clone https://github.com/gx1727/XinFramework.git
cd XinFramework
```

### 2. 配置环境

```bash
cp .env.example .env
# 编辑 .env 配置数据库等信息
```

### 3. 初始化数据库

```bash
psql -U postgres -d xin_db -f migrations/001_init.sql
```

### 4. 启动服务

```bash
# 开发模式
go run ./cmd/xin run

# 编译运行
go build -o xin ./cmd/xin
./xin start
```

服务默认监听：`0.0.0.0:8080`

---

## 📋 服务管理命令

| 命令 | 说明 |
|------|------|
| `run` | 前台运行 |
| `start` | 后台启动 |
| `stop` | 停止服务 |
| `restart` | 重启服务 |
| `reload` | 热加载配置 (Unix) |
| `status` | 查看运行状态 |

---

## 🌐 API 路由

### 公开路由

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/health` | 健康检查 |
| POST | `/api/v1/login` | 用户登录 |
| POST | `/api/v1/register` | 用户注册 |
| POST | `/api/v1/logout` | 用户登出 |
| POST | `/api/v1/refresh` | 刷新 Token |

### 租户管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/tenants` | 获取租户列表 |
| POST | `/api/v1/tenants` | 创建租户 |
| PUT | `/api/v1/tenants/:id` | 更新租户 |
| DELETE | `/api/v1/tenants/:id` | 删除租户 |

---

## 🏢 多租户

配置项：

```yaml
saas:
  mode: shared  # shared | schema | database
```

请求时传递租户 ID：

```bash
curl -H "X-Tenant-ID: tenant_001" http://localhost:8080/api/v1/users
```

实现机制：通过 PostgreSQL `SET app.tenant_id = $1` 设置会话变量，配合行级安全策略 (RLS) 实现租户隔离。

---

## 🔐 认证机制

```
用户登录
    ↓
验证账号密码 (Argon2id)
    ↓
生成 Session (Redis 优先, DB 兜底)
    ↓
签发 JWT (含 sid)
    ↓
后续请求携带 JWT
    ↓
中间件校验 sid 有效性
    ↓
登出: 撤销 Session
```

---

## 🧩 插件开发

### 1. 创建插件

```go
// apps/myplugin/myplugin.go
package myplugin

import (
    "github.com/gin-gonic/gin"
    "gx1727.com/xin/framework/pkg/plugin"
)

type Module struct{}

func (m *Module) Name() string           { return "myplugin" }
func (m *Module) Init() error            { return nil }
func (m *Module) Register(public, protected *gin.RouterGroup) {
    protected.GET("/data", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "myplugin"})
    })
}

var _ plugin.Module = (*Module)(nil)
```

### 2. 注册插件

```go
// cmd/xin/main.go
if cfg.AppEnabled("myplugin") {
    framework.RegisterModule(&myplugin.Module{})
}
```

### 3. 启用插件

```yaml
# config.yaml
apps:
  - myplugin
```

---

## 📖 文档

- [开发指南](doc/developer-guide.md) — 框架使用详解
- [数据库规范](doc/database-conventions.md) — 表设计规范
- [API 调试示例](doc/api.http) — HTTP 调试文件

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

## 📄 License

MIT License
