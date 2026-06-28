# XinFramework

> 多租户 SaaS 后台框架。前端 React + 后端 Go + 内置 RBAC / 权限 / 数据范围。

XinFramework 是一个**面向多租户 SaaS 后台**的可扩展框架。当前共 **19 个模块**，全部跑在单一 Go module `gx1727.com/xin` 下（Phase 0023 平台/租户域分域完成）。

- **后端**（[server/](server/)）：Go 1.24 + Gin + pgx + PostgreSQL，
  - 框架核心：`gx1727.com/xin/framework`（必装：登录 / 多租户 / RBAC / 迁移 / 中间件）
  - 业务模块：`gx1727.com/xin/apps`（选装：cms / flag / message 等，由 framework 统一注册）
- **前端**（[UI/](UI/)）：React 19 + TypeScript + Vite + shadcn/ui，
  - 自研 Schema 驱动的 DynamicForm / DynamicTable
  - 顶部错误条 + 用户主动切换 mock 数据（替代旧的静默兜底）

## 仓库结构

```
XinFramework/
├── server/                       # Go 后端
│   ├── cmd/xin/                  # 入口（4 步显式 Build）
│   ├── .env                      # 首次启动 bootstrap token
│   ├── config/                   # YAML 配置
│   ├── migrations/               # SQL 迁移
│   ├── framework/                # 框架核心
│   │   ├── framework.go          # Boot() / Serve()
│   │   ├── internal/             # boot / middleware / server / ext_impl / service
│   │   └── pkg/                  # appx / audit / auth / authz / cache / config /
│   │                             # context / db / dict / extapi / jwt / logger /
│   │                             # middleware / migrate / model / permission /
│   │                             # plugin / rbac / resp / session / storage / tenant
│   └── apps/                     # 业务模块（与 framework 同 module）
│       ├── boot/auth/            # 登录（alwaysOn）
│       ├── platform/             # tenants / sys_user / sys_role / sys_menu / sys_permission
│       ├── tenant/               # user / role / menu / organization / permission / resource
│       ├── reference/            # asset / config / dict / weixin
│       ├── system/               # health / cache 运维
│       ├── cms/                  # 示例 CMS（extapi 模式）
│       └── flag/                 # 头像框 / 空间 / 头像
│
├── UI/                           # React 前端
│   └── src/
│       ├── api/                  # API 客户端（common.ts + 各模块 Api）
│       ├── components/           # shadcn/ui + 自研 schema 组件
│       ├── locales/              # i18n（zh-CN 为类型源头）
│       ├── pages/                # 每个模块一个页面
│       ├── stores/               # zustand（authStore / menuStore / configStore / permissionStore）
│       ├── types/schema.ts       # FormSchema / TableSchema
│       └── App.tsx               # 路由（scope-based guard）
│
├── server/scripts/strip_bom.py   # BOM 检测 / 剥离工具（含 --check CI 模式）
└── README.md                     # 本文件
```

## 快速开始

```bash
# 1. 启动 PostgreSQL
docker run --name xin-pg -e POSTGRES_PASSWORD=dev -p 5432:5432 -d postgres:16

# 2. 配置 bootstrap 凭据（首次启动必须）
cd server
cp .env.example .env  # 编辑 .env 

# 3. 启动后端
go run ./cmd/xin run
# → 首次启动会自动跑迁移 + bootstrap

# 4. 启动前端
cd ../UI
npm install
npm run dev

# 5. 验证
curl http://localhost:8087/api/v1/health
# → {"code":0,"msg":"ok","data":{"status":"ok"}}
```

- 前端 dev：`http://localhost:5241`
- 后端 API：`http://localhost:8087/api/v1`

首次启动（生产空库）通过环境变量注入初始 super_admin（详见 [server/doc/quickstart.md](server/doc/quickstart.md) §6）。

## 核心能力

- **多租户隔离**：每个业务表带 `tenant_id` + RLS `FORCE ROW LEVEL SECURITY`，所有 SQL 强制 tenant 过滤
- **RBAC + 数据范围**：用户 → 角色 → 资源权限码（`user:list` / `flag:create` 等）；角色同时携带数据范围（全部 / 部门 / 本人 / 自定义 等 5 种）
- **平台角色**：跨租户特权（`super_admin`），独立于租户内 RBAC，自动 bypass 所有 spec
- **统一响应**：所有 API 返回 `{ code, msg, data }`，`code=0` 为成功，按区段管理错误码
- **插件化模块**：内置模块（boot / rbac）与外部 app（cms / flag）走同一 `Module(app)` 工厂注册，可按 `cfg.Module` 白名单启停
- **JSONB 安全**：所有 JSONB 列在 SQL 里显式 `::jsonb` cast（避免 pgx 把 `string`/`[]byte` 当 text/bytea 发）

## 模块清单

| Name | 类型 | 数据表 | 说明 |
|---|---|---|---|
| `auth` | alwaysOn | accounts / auth_sessions | 登录 / 注册 / JWT / 多身份 |
| `tenants` | alwaysOn | tenants | 租户管理（必须 super_admin，apps/platform/tenants） |
| `system` | alwaysOn | — | /health + 运维 cache |
| `user` | optOut | tenant_users / tenant_user_roles | 租户内用户 CRUD（apps/tenant/user） |
| `role` | optOut | tenant_roles / tenant_role_data_scopes / tenant_user_roles / tenant_role_menus / tenant_role_resources | 角色 + 数据范围（apps/tenant/role） |
| `menu` | optOut | tenant_menus / tenant_role_menus | 租户菜单树（apps/tenant/menu，平台菜单见 sys_menu） |
| `organization` | optOut | tenant_organizations | 租户组织架构（递归 CTE + 物化路径，apps/tenant/organization） |
| `permission` | optOut | tenant_role_resources | 租户角色-权限码分配（apps/tenant/permission） |
| `resource` | optOut | tenant_permissions | 租户权限码 CRUD（apps/tenant/resource） |
| `asset` | optOut | file_assets | 文件上传（local / COS） |
| `dict` | optOut | dicts / dict_items / dict_visibility | 数据字典（平台 + 租户二级，apps/reference/dict） |
| `config` | optOut | config_categories / config_items / config_visibility | 租户配置中心（apps/reference/config） |
| `weixin` | optional | — | 微信小程序登录（apps/reference/weixin） |
| `sys_user` | optional | sys_users / sys_orgs / sys_user_roles | 平台域用户身份（0023+） |
| `sys_role` | optional | sys_roles / sys_user_roles | 平台域角色（含 super_admin，0023+） |
| `sys_menu` | optional | sys_menus / sys_role_menus | 平台域菜单（0023+） |
| `sys_permission` | optional | sys_permissions / sys_role_permissions | 平台域权限码（0023+） |
| `cms` | optional | posts | 示例 CMS（extapi 模式，apps/cms） |
| `flag` | optional | frames / spaces / avatars ... | 头像框生成器（apps/flag） |

详见 [server/doc/modules.md](server/doc/modules.md)。

## 文档

### 后端（[server/doc/](server/doc)）

| 文件 | 内容 |
| --- | --- |
| [quickstart.md](server/doc/quickstart.md) | 安装、配置、首次启动 |
| [architecture.md](server/doc/architecture.md) | 架构总览、AppContext、模块生命周期 |
| [modules.md](server/doc/modules.md) | 19 个 module 清单和路由 |
| [api.md](server/doc/api.md) | HTTP API 参考 |
| [database.md](server/doc/database.md) | 表结构、RLS、JSONB、迁移 |
| [permissions.md](server/doc/permissions.md) | RBAC + DataScope + 平台角色 |
| [developing.md](server/doc/developing.md) | 新增模块的 8 步流程 |
| [deployment.md](server/doc/deployment.md) | 编译、systemd、Docker |
| [AGENTS.md](server/AGENTS.md) | 给 AI agent 协作者的高密度参考 |

### 前端（[UI/](UI)）

| 文件 | 内容 |
| --- | --- |
| [UI/README.md](UI/README.md) | 前端快速上手 |
| [UI/AGENTS.md](UI/AGENTS.md) | 前端设计约定（文案 / Schema / mock 切换） |

### 工具

- [server/scripts/strip_bom.py](server/scripts/strip_bom.py) — BOM 检测/剥离（`--check` 用于 CI gate）

### App 文档

- [server/apps/flag/doc/api.md](server/apps/flag/doc/api.md) — Flag App（头像框生成器）API

## 版本与维护

- Go 1.24+ / Node 20+
- 数据库：PostgreSQL 14+（需要 `ltree` / `pg_trgm` 扩展）
- 文件编码：**所有源文件 UTF-8 无 BOM**（PowerShell 默认 GBK，写入用 [UI/AGENTS.md §5.1](UI/AGENTS.md#51-编码最重要) 的方法绕过）
