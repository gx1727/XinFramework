# XinFramework

企业级 SaaS 基建框架（开箱可卖）

## 目标客户

- 小公司 / 外包团队
- 想快速做系统的老板
- 想低成本上线 SaaS 的人

## 整体架构

采用：模块化单体 + 可演进微服务

```
                    ┌──────────────┐
                    │  API Gateway │
                    └──────┬───────┘
                           │
         ┌─────────────────┼─────────────────┐
         │                 │                 │
    xin-auth          xin-core          xin-saas
   (认证中心)         (基础能力)        (多租户)

         │                 │                 │
    xin-admin         xin-billing         xin-ai
    (后台)            (计费)             (AI能力)

                           │
                    ┌──────▼──────┐
                    │  PostgreSQL │
                    │   + Redis   │
                    └─────────────┘
```

## 技术选型

**Web 框架**
- Gin（成熟、生态全，企业项目够用）

**ORM**
- GORM + SQL 混用
  - 简单 CRUD 用 GORM
  - 复杂查询用 SQL

**数据库**
- PostgreSQL（核心）
- Redis（缓存 / Session / 限流）

**消息队列**
- NATS（轻量）或 Kafka（重型）

**AI 能力**
- OpenAI API + 本地模型（Ollama）

## 项目结构

```
xin/
├── cmd/                          # 启动入口（多服务扩展点）
│   └── server/
│       └── main.go

├── configs/                      # 配置（多环境）
│   ├── config.yaml
│   ├── config.dev.yaml
│   ├── config.prod.yaml
│   └── config.go

├── api/                          # API 版本定义（对外协议层）
│   ├── v1/
│   │   ├── router.go
│   │   └── register.go
│   └── v2/                       # 未来扩展

├── internal/                     # 核心业务（禁止外部引用）
│   ├── core/                     # 框架核心
│   │   ├── server/               # Gin 封装
│   │   ├── middleware/           # 中间件（JWT/日志/租户）
│   │   ├── config/               # 配置加载
│   │   ├── boot/                 # 启动流程（依赖注入）
│   │   └── context/              # 上下文（用户/tenant）
│   ├── module/                   # 业务模块
│   │   ├── user/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── repo.go
│   │   │   └── model.go
│   │   ├── auth/                 # 登录/权限（RBAC）
│   │   ├── saas/                # 多租户（核心）
│   │   ├── billing/             # 计费系统
│   │   ├── ai/                  # AI 能力
│   │   ├── admin/               # 后台系统
│   │   └── wx/                  # 微信生态（登录/支付）
│   └── infra/                    # 基础设施
│       ├── db/                  # PostgreSQL
│       ├── cache/               # Redis
│       ├── mq/                  # 消息队列（NATS/Kafka）
│       └── logger/              # 日志系统

├── pkg/                          # 可复用组件（可开源）
│   ├── jwt/
│   ├── resp/                     # 统一返回
│   └── utils/

├── migrations/                   # 数据库迁移
│   └── *.sql

├── deploy/                       # 部署
│   ├── docker/
│   └── k8s/

└── scripts/                      # 脚本（初始化/部署）
```

## 核心模块

### 1. xin-auth（权限系统）

必须做到：
- JWT + Refresh Token
- RBAC（角色权限）
- 支持多租户隔离

### 2. xin-saas（多租户核心）

支持三种模式：
- 共享库 + tenant_id（基础版）
- Schema 隔离（进阶）
- 独立数据库（高端）

### 3. xin-admin（后台系统）

必须有：
- 用户管理
- 租户管理
- 权限管理
- 操作日志

### 4. xin-billing（变现模块）

必须做：
- 套餐系统（免费 / 付费）
- 使用量统计
- API 计费

### 5. xin-ai（差异化）

核心能力：
- 文档问答（RAG）
- AI 客服
- 自动生成内容

技术方案：
- Embedding（OpenAI / 本地）
- 向量存储（PostgreSQL + pgvector）

## 数据库设计

关键表：
- users
- roles
- permissions
- user_roles
- tenants
- tenant_users
- subscriptions
- plans
- usage_records
- ai_documents
- ai_embeddings

所有表必须带：
- tenant_id
- created_at
- updated_at

## 接口风格

```
GET    /api/v1/users
POST   /api/v1/users
PUT    /api/v1/users/:id
DELETE /api/v1/users/:id
```

统一返回：
```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

## 中间件

- Auth Middleware（JWT）
- Tenant Middleware（自动注入 tenant_id）
- Logger Middleware
- Rate Limit（限流）

## 编译

```
# 基础编译
go build -o xin-server.exe .\cmd\server\main.go


# 优化编译
go build -ldflags="-s -w" -o xin-server.exe .\cmd\server\main.go


```