# XinFramework

> 多租户 SaaS 后台框架。前端 React + 后端 Go + 内置 RBAC / 权限 / 数据范围。

XinFramework 是一个**面向多租户 SaaS 后台**的可扩展框架，包含：

- **后端**（[server/](file:///d:\work\xin\XinFramework\server)）：Go 1.25 + Gin + pgx + PostgreSQL，
  - 框架核心：`gx1727.com/xin/framework`（必装：登录/多租户/RBAC/迁移/中间件）
  - 业务模块：`gx1727.com/xin/apps`（选装：cms/flag/message 等，由 framework 统一注册）
  - 历史重构记录：[architecture.md](file:///d:\work\xin\XinFramework\server\doc\architecture.md)（Phase 1-3b 已完成）
- **前端**（[UI/](file:///d:\work\xin\XinFramework\UI)）：React 18 + TypeScript + Vite + shadcn/ui，
  - 自研 Schema 驱动的 DynamicForm / DynamicTable
  - 顶部错误条 + 用户主动切换 mock 数据（替代旧的静默兜底）

## 仓库结构

```
XinFramework/
├── server/                 # Go 后端
│   ├── cmd/xin/            # 入口
│   ├── framework/          # 框架核心（gx1727.com/xin/framework）
│   ├── apps/               # 业务模块（gx1727.com/xin/apps）
│   │   ├── boot/{auth,tenant}/                # 必装：登录 + 多租户
│   │   ├── rbac/{user,role,menu,resource,permission,organization}/  # 必装：RBAC
│   │   ├── reference/{asset,dict,weixin}/     # 选装：附件 / 字典 / 微信
│   │   ├── system/                           # 选装：运维管理 + /health
│   │   ├── cms/                              # 选装：内容管理
│   │   └── flag/                             # 选装：头像框
│   ├── config/             # 配置
│   ├── migrations/         # SQL 迁移
│   ├── doc/                # 后端文档
│   └── go.work             # 多 module 编排
│
├── UI/                     # React 前端
│   └── src/
│       ├── api/            # API 客户端
│       ├── components/     # shadcn/ui + 自研组件
│       ├── locales/        # i18n（zh-CN 为类型源头）
│       ├── pages/          # 每个模块一个页面
│       └── stores/         # zustand
│
└── README.md               # 本文件
```

## 快速开始

```bash
# 1. 启动 PostgreSQL（任选一种）
docker run --name xin-pg -e POSTGRES_PASSWORD=dev -p 5432:5432 -d postgres:16

# 2. 启动后端
cd server
go work sync
go run ./cmd/xin run

# 3. 启动前端
cd ../UI
npm install
npm run dev

# 4. 打开浏览器
# 前端：http://localhost:5173
# 后端：http://localhost:8080/api/v1
```

首次启动会执行迁移并要求初始化超级管理员：

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| tenant_code | 是 | 默认租户编码 |
| tenant_name | 是 | 默认租户名称 |
| admin_account | 是 | 超级管理员账号 |
| admin_password | 是 | 超级管理员密码 |

详见 [server/doc/quickstart.md](file:///d:\work\xin\XinFramework\server\doc\quickstart.md)。

## 核心能力

- **多租户隔离**：每个表带 `tenant_id`，所有 SQL 强制 tenant 过滤
- **RBAC + 数据范围**：用户 → 角色 → 资源权限码；角色同时携带数据范围（全部 / 部门 / 本人 等）
- **平台角色**：跨租户特权（如 `super_admin`），独立于租户内 RBAC
- **统一响应**：所有 API 返回 `{ code, msg, data }`，code=0 为成功
- **插件化模块**：内置模块（boot/rbac）与外部 app（reference/cms/flag 等）走同一注册路径，可按 config 启停

## 文档

### 后端（[server/doc/](file:///d:\work\xin\XinFramework\server\doc)）

| 文件 | 内容 |
| --- | --- |
| [architecture.md](file:///d:\work\xin\XinFramework\server\doc\architecture.md) | 架构总览、Go module 切分、Phase 1-3b 重构方案 |
| [quickstart.md](file:///d:\work\xin\XinFramework\server\doc\quickstart.md) | 安装、配置、首次启动 |
| [modules.md](file:///d:\work\xin\XinFramework\server\doc\modules.md) | 内置模块 + apps 列表、依赖关系、注册流程 |
| [api.md](file:///d:\work\xin\XinFramework\server\doc\api.md) | HTTP API 参考（认证、用户、租户、RBAC） |
| [database.md](file:///d:\work\xin\XinFramework\server\doc\database.md) | 表结构、迁移、命名约定 |
| [permissions.md](file:///d:\work\xin\XinFramework\server\doc\permissions.md) | RBAC、数据范围、平台角色 |
| [developing.md](file:///d:\work\xin\XinFramework\server\doc\developing.md) | 如何新增一个 app / 模块 |
| [deployment.md](file:///d:\work\xin\XinFramework\server\doc\deployment.md) | 编译、systemd、监控 |
| [AGENTS.md](file:///d:\work\xin\XinFramework\server\AGENTS.md) | 给 agent / AI 协作者看的高密度参考 |

### 前端（[UI/](file:///d:\work\xin\XinFramework\UI)）

| 文件 | 内容 |
| --- | --- |
| [UI/README.md](file:///d:\work\xin\XinFramework\UI\README.md) | 前端快速上手 |
| [UI/AGENTS.md](file:///d:\work\xin\XinFramework\UI\AGENTS.md) | 前端设计约定（i18n、Schema、mock 约定） |

### App 文档

- [apps/flag/doc/api.md](file:///d:\work\xin\XinFramework\server\apps\flag\doc\api.md) — Flag App（头像框生成器）API

## 版本与维护

- Go 1.25.0 / Node 20+
- 数据库：PostgreSQL 16
- 许可证：见 [LICENSE](file:///d:\work\xin\XinFramework\server\LICENSE)