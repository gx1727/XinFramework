# 数据库设计

> 本文件描述 XinFramework 的 PostgreSQL schema 设计：表结构、RLS、JSONB、迁移、约束。
> 完整 SQL 见 `server/migrations/`。

---

## 1. 扩展依赖

启动期自动创建：

```sql
CREATE EXTENSION IF NOT EXISTS ltree;     -- 物化路径（组织树）
CREATE EXTENSION IF NOT EXISTS pg_trgm;   -- 三元组索引（模糊搜索）
```

---

## 2. 数据域分层

| 域 | 标识 | tenant_id | RLS | 业务模块 |
|---|---|---|---|---|
| **sys 域** | `sys_*` | ❌ 无 | ❌ 不启用 | `apps/sys/*` |
| **租户域** | `tenant_*` | ✅ 必填 | ✅ 全部启用 | `apps/tenant/*` |
| **共享层** | `accounts` / `tenants` / `auth_sessions` | ❌ | ❌ | `apps/boot/auth` + `apps/sys/tenants` |
| **字典/配置** | `dicts` / `config_*` | ✅ 必填（=0 表示平台级） | ✅ 启用 | `apps/reference/dict` + `apps/reference/config` |
| **业务支撑** | `subscriptions` / `usage_records` / `db_logs` / `routes` / `plans` | ✅ 必填 | ✅ 启用 | 对应模块 |
| **业务模块** | `assets` / `cms_*` / `flag_*` / `messages` | ✅ 或 ❌（看模块） | 视模块而定 | `apps/reference/asset` / `apps/cms` / `apps/flag` / `apps/tenant/message` |

---

## 3. RLS 策略

### 3.1 租户域 policy 模板

```sql
ALTER TABLE tenant_xxx ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON tenant_xxx USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);
```

含义：
- 当 `app.tenant_id` 设置为某 tenant_id 时，只能看到该 tenant 的行
- 当 `app.bypass_rls='on'` 时（`RunInSysTx`），可以跨租户访问（sys 域）

### 3.2 字典/配置类（带 tenant_id=0 短路）

```sql
CREATE POLICY tenant_isolation_policy ON dicts USING (
    tenant_id = 0                                          -- 平台级（跨租户共享）
    OR tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);
```

`tenant_id=0` 表示平台级（如 gender / user_status 等通用字典），所有租户共享。

### 3.3 平台域（不启用 RLS）

sys 域表（`sys_*`）无 `tenant_id`，不启用 RLS。安全靠 API 层 `RequireSysRole("super_admin")` + `db.RunInSysTx` 守护。

### 3.4 应用层配套

```go
// 业务域事务
db.RunInTenantTx(ctx, pool, tenantID, func(ctx context.Context) error {
    // ctx 中所有 SQL 自动走本租户 RLS
    return svc.GetByID(ctx, id)
})

// sys 域事务
db.RunInPlatformTx(ctx, pool, func(ctx context.Context) error {
    // ctx 中 app.bypass_rls='on'，可跨租户访问
    return svc.CrossTenantQuery(ctx)
})
```

---

## 4. 字段规范

### 4.1 通用列

所有业务表必备：

```sql
created_at  TIMESTAMPTZ DEFAULT NOW(),
updated_at  TIMESTAMPTZ DEFAULT NOW(),
created_by  BIGINT,
updated_by  BIGINT,
is_deleted  BOOLEAN     DEFAULT FALSE
```

### 4.2 软删除 + 谓词索引

```sql
-- 索引加 WHERE is_deleted = FALSE 谓词
CREATE INDEX idx_tenant_users_tenant ON tenant_users (tenant_id) WHERE is_deleted = FALSE;

-- 唯一约束同理
CREATE UNIQUE INDEX uk_tenant_users_account_tenant
    ON tenant_users (account_id, tenant_id) WHERE is_deleted = FALSE;
```

**优势**：删除记录后唯一约束立即释放，未删除记录不浪费索引空间。

### 4.3 JSONB 显式 cast

```sql
-- 错误：pgx 会把 string 当 text 发，触发 type mismatch
INSERT INTO tenants (config) VALUES ('{"a": 1}');

-- 正确：显式 ::jsonb cast
INSERT INTO tenants (config) VALUES ('{"a": 1}'::jsonb);
```

**所有 JSONB 列在 SQL 中必须显式 `::jsonb` cast**（pgx v5 默认不会自动推断）。这是 `framework/pkg/db/db.go` 注释里特别强调的"PG JSONB 安全"约定。

### 4.4 ID 生成

主键用 PostgreSQL `BIGINT GENERATED ALWAYS AS IDENTITY`：

```sql
CREATE TABLE tenant_users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    ...
);
```

**优势**：
- 跨节点唯一
- 顺序自增
- 序列可重置（`SELECT setval('tenant_users_id_seq', N, true)`）

---

## 5. 主要表结构

### 5.1 共享层

#### 5.1.1 `tenants` 租户表

```sql
CREATE TABLE tenants (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code       VARCHAR(50)  NOT NULL,           -- 租户编码（唯一）
    name       VARCHAR(100) NOT NULL,
    status     SMALLINT DEFAULT 1,
    contact    VARCHAR(50),
    phone      VARCHAR(20),
    email      VARCHAR(100),
    province   VARCHAR(32),
    city       VARCHAR(32),
    area       VARCHAR(32),
    address    VARCHAR(255),
    config     JSONB,                          -- 租户级配置
    dashboard  VARCHAR(64),                    -- 默认 dashboard 路径
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by BIGINT,
    updated_by BIGINT,
    is_deleted BOOLEAN     DEFAULT FALSE
);
```

#### 5.1.2 `accounts` 全局账号

```sql
CREATE TABLE accounts (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    phone      VARCHAR(20),
    email      VARCHAR(100),
    password   VARCHAR(255),                  -- argon2id 哈希
    username   VARCHAR(64),
    real_name  VARCHAR(64),
    avatar     VARCHAR(512),
    status     SMALLINT    DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
```

#### 5.1.3 `auth_sessions` 会话表

```sql
CREATE TABLE auth_sessions (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    account_id BIGINT       NOT NULL,
    token      VARCHAR(255) NOT NULL,
    ip         VARCHAR(64),
    user_agent VARCHAR(255),
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### 5.2 租户域

#### 5.2.1 `tenant_organizations` 组织（递归 + 物化路径）

```sql
CREATE TABLE tenant_organizations (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id   BIGINT      NOT NULL,
    parent_id   BIGINT,                       -- 父组织 id
    code        VARCHAR(32) NOT NULL,
    name        VARCHAR(64) NOT NULL,
    type        VARCHAR(32),                  -- company/department/team
    description VARCHAR(255),
    admin_code  VARCHAR(64),                  -- 组织负责人 user_code
    ancestors   TEXT        DEFAULT '',       -- 物化路径（如 "0.1.5"）
    sort        INT         DEFAULT 0,
    status      SMALLINT    DEFAULT 1,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    created_by  BIGINT,
    updated_by  BIGINT,
    is_deleted  BOOLEAN     DEFAULT FALSE
);
```

#### 5.2.2 `tenant_users` 租户用户

```sql
CREATE TABLE tenant_users (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,
    account_id BIGINT      NOT NULL,          -- FK accounts.id
    code       VARCHAR(32),
    org_id     BIGINT,                        -- FK tenant_organizations.id
    real_name  VARCHAR(64),
    nickname   VARCHAR(64),
    avatar     VARCHAR(512),
    status     SMALLINT    DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by BIGINT,
    updated_by BIGINT,
    is_deleted BOOLEAN     DEFAULT FALSE
);

-- 唯一：一个账号在同租户只能有 1 个 user
CREATE UNIQUE INDEX uk_tenant_users_account_tenant
    ON tenant_users (account_id, tenant_id) WHERE is_deleted = FALSE;
```

#### 5.2.3 `tenant_roles` 角色

```sql
CREATE TABLE tenant_roles (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id   BIGINT      NOT NULL,
    org_id      BIGINT,
    code        VARCHAR(32),
    name        VARCHAR(32) NOT NULL,
    description VARCHAR(256),
    data_scope  SMALLINT    NOT NULL DEFAULT 1,   -- 数据范围 1-5
    extend      JSONB,                            -- 扩展配置
    is_default  BOOLEAN              DEFAULT FALSE,
    sort        INT                  DEFAULT 0,
    status      SMALLINT             DEFAULT 1,
    created_at  TIMESTAMPTZ          DEFAULT NOW(),
    updated_at  TIMESTAMPTZ          DEFAULT NOW(),
    created_by  BIGINT,
    updated_by  BIGINT,
    is_deleted  BOOLEAN              DEFAULT FALSE
);
```

#### 5.2.4 `tenant_role_data_scopes` 角色数据范围（自定义 org 列表）

```sql
CREATE TABLE tenant_role_data_scopes (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    role_id    BIGINT NOT NULL,
    org_id     BIGINT NOT NULL,        -- 自定义 org 列表的成员
    created_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
```

#### 5.2.5 `tenant_user_roles` 用户-角色

```sql
CREATE TABLE tenant_user_roles (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    user_id    BIGINT NOT NULL,
    role_id    BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
```

#### 5.2.6 `tenant_menus` 菜单

```sql
CREATE TABLE tenant_menus (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,
    code       VARCHAR(64),
    name       VARCHAR(64) NOT NULL,
    subtitle   VARCHAR(128),
    url        VARCHAR(255),
    path       VARCHAR(255),
    icon       VARCHAR(64),
    sort       INT         DEFAULT 0,
    parent_id  BIGINT,
    ancestors  TEXT        DEFAULT '',
    visible    BOOLEAN     DEFAULT TRUE,
    enabled    BOOLEAN     DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by BIGINT,
    updated_by BIGINT,
    is_deleted BOOLEAN     DEFAULT FALSE
);
```

#### 5.2.7 `tenant_permissions` 权限码（原 `resources`）

```sql
CREATE TABLE tenant_permissions (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id   BIGINT      NOT NULL,
    menu_id     BIGINT,
    code        VARCHAR(64) NOT NULL,        -- 如 "user:list"
    name        VARCHAR(64) NOT NULL,
    action      VARCHAR(32) DEFAULT 'read',  -- list/get/create/update/delete
    description VARCHAR(512),
    sort        INT         DEFAULT 0,
    status      SMALLINT    DEFAULT 1,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    created_by  BIGINT,
    updated_by  BIGINT,
    is_deleted  BOOLEAN     DEFAULT FALSE
);
```

#### 5.2.8 `tenant_role_menus` 角色-菜单

```sql
CREATE TABLE tenant_role_menus (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    role_id    BIGINT NOT NULL,
    menu_id    BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
```

#### 5.2.9 `tenant_role_resources` 角色-权限码

```sql
CREATE TABLE tenant_role_resources (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id     BIGINT NOT NULL,
    role_id       BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    effect        SMALLINT    DEFAULT 1,    -- 1=允许，2=拒绝
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW(),
    is_deleted    BOOLEAN     DEFAULT FALSE
);
```

### 5.3 平台域

#### 5.3.1 `sys_users` 平台用户身份

```sql
CREATE TABLE sys_users (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    account_id BIGINT      NOT NULL,         -- FK accounts.id
    code       VARCHAR(32),
    org_id     BIGINT,
    real_name  VARCHAR(64),
    nickname   VARCHAR(64),
    avatar     VARCHAR(512),
    status     SMALLINT    DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by BIGINT,
    updated_by BIGINT,
    is_deleted BOOLEAN     DEFAULT FALSE
);
```

#### 5.3.2 `sys_orgs` 平台组织

```sql
CREATE TABLE sys_orgs (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    parent_id   BIGINT,
    code        VARCHAR(32) NOT NULL,
    name        VARCHAR(64) NOT NULL,
    type        VARCHAR(32),
    description VARCHAR(255),
    admin_code  VARCHAR(64),
    ancestors   TEXT             DEFAULT '',
    sort        INT              DEFAULT 0,
    status      SMALLINT         DEFAULT 1,
    created_at  TIMESTAMPTZ      DEFAULT NOW(),
    updated_at  TIMESTAMPTZ      DEFAULT NOW(),
    created_by  BIGINT,
    updated_by  BIGINT,
    is_deleted  BOOLEAN          DEFAULT FALSE
);
```

#### 5.3.3 `sys_roles` 平台角色

```sql
CREATE TABLE sys_roles (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    org_id      BIGINT,
    code        VARCHAR(32)  NOT NULL,        -- 如 "super_admin"
    name        VARCHAR(32)  NOT NULL,
    description VARCHAR(256),
    data_scope  SMALLINT     NOT NULL DEFAULT 1,
    extend      JSONB,
    is_default  BOOLEAN               DEFAULT FALSE,
    sort        INT                   DEFAULT 0,
    status      SMALLINT              DEFAULT 1,
    created_at  TIMESTAMPTZ           DEFAULT NOW(),
    updated_at  TIMESTAMPTZ           DEFAULT NOW(),
    created_by  BIGINT,
    updated_by  BIGINT,
    is_deleted  BOOLEAN               DEFAULT FALSE
);
```

#### 5.3.4 `sys_menus` 平台菜单

```sql
CREATE TABLE sys_menus (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code       VARCHAR(64),
    name       VARCHAR(64)  NOT NULL,
    subtitle   VARCHAR(128),
    url        VARCHAR(255),
    path       VARCHAR(255),
    icon       VARCHAR(64),
    sort       INT              DEFAULT 0,
    parent_id  BIGINT,
    ancestors  TEXT             DEFAULT '',
    visible    BOOLEAN          DEFAULT TRUE,
    enabled    BOOLEAN          DEFAULT TRUE,
    created_at TIMESTAMPTZ      DEFAULT NOW(),
    updated_at TIMESTAMPTZ      DEFAULT NOW(),
    created_by BIGINT,
    updated_by BIGINT,
    is_deleted BOOLEAN          DEFAULT FALSE
);
```

#### 5.3.5 `sys_permissions` 平台权限码

```sql
CREATE TABLE sys_permissions (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    menu_id     BIGINT,
    code        VARCHAR(64)  NOT NULL,        -- 如 "platform:dicts:list"
    name        VARCHAR(64)  NOT NULL,
    action      VARCHAR(32)  DEFAULT 'read',
    description VARCHAR(512),
    sort        INT             DEFAULT 0,
    status      SMALLINT        DEFAULT 1,
    created_at  TIMESTAMPTZ     DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     DEFAULT NOW(),
    created_by  BIGINT,
    updated_by  BIGINT,
    is_deleted  BOOLEAN         DEFAULT FALSE
);
```

#### 5.3.6 `sys_user_roles` 平台用户-角色（替代旧 `account_roles`）

```sql
CREATE TABLE sys_user_roles (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    BIGINT  NOT NULL,
    role_id    BIGINT  NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
```

### 5.4 字典 / 配置

#### 5.4.1 `dicts` + `dict_items` + `dict_visibility`

```sql
CREATE TABLE dicts (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,          -- 0=平台级，>0=租户级
    code       VARCHAR(32) NOT NULL,
    name       VARCHAR(64) NOT NULL,
    visibility VARCHAR(16) NOT NULL DEFAULT 'all',  -- all / specified / hidden
    status     SMALLINT    DEFAULT 1,
    sort       INT         DEFAULT 0,
    extend     JSONB       DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);

CREATE TABLE dict_items (
    id               BIGINT       NOT NULL,
    tenant_id        BIGINT       NOT NULL,
    dict_id          BIGINT       NOT NULL,
    code             VARCHAR(64)  NOT NULL,
    name             VARCHAR(128) NOT NULL,
    platform_item_id BIGINT,                    -- 平台项 id（租户覆盖时记录）
    is_override      BOOLEAN      NOT NULL DEFAULT FALSE,
    sort             INT          DEFAULT 0,
    status           SMALLINT     DEFAULT 1,
    extend           JSONB        DEFAULT '{}'::jsonb,
    ...
);

-- 平台级唯一
CREATE UNIQUE INDEX uk_dict_code_platform
    ON dicts (code) WHERE tenant_id = 0 AND is_deleted = FALSE;
-- 租户级唯一
CREATE UNIQUE INDEX uk_dict_code_tenant
    ON dicts (tenant_id, code) WHERE tenant_id <> 0 AND is_deleted = FALSE;
```

**租户覆盖模式**：
- 平台级字典 `dict.code='gender'`，平台级 item `code='male'`
- 租户 A 想覆盖 male → 创建租户级 item，`is_override=true`，`platform_item_id=<platform male.id>`

#### 5.4.2 `config_categories` + `config_items` + `config_visibility`

与 dicts 同构，支持租户覆盖平台项。

```sql
CREATE TABLE config_items (
    id               BIGINT       NOT NULL,
    tenant_id        BIGINT       NOT NULL,    -- 0=平台级，>0=租户级
    category_id      BIGINT       NOT NULL,
    key              VARCHAR(128) NOT NULL,
    value            JSONB        DEFAULT NULL,
    default_value    JSONB        DEFAULT NULL,
    type             VARCHAR(32)  NOT NULL,    -- string/number/select/boolean/image/...
    label            VARCHAR(128),
    description      VARCHAR(512),
    options          JSONB        DEFAULT NULL,    -- select 选项
    validation       JSONB        DEFAULT NULL,    -- {min, max, required, ...}
    platform_item_id BIGINT,
    is_override      BOOLEAN      NOT NULL DEFAULT FALSE,
    sort             INT          DEFAULT 0,
    is_public        BOOLEAN      DEFAULT FALSE,   -- 是否可通过 /public/configs 读
    is_readonly      BOOLEAN      DEFAULT FALSE,
    is_system        BOOLEAN      DEFAULT FALSE,   -- 系统预置（不可删）
    status           SMALLINT     DEFAULT 1,
    ...
);
```

### 5.5 业务支撑

```sql
-- 订阅
CREATE TABLE subscriptions (
    id BIGINT, tenant_id BIGINT NOT NULL, plan_id BIGINT,
    status SMALLINT DEFAULT 1, start_at TIMESTAMPTZ, end_at TIMESTAMPTZ,
    auto_renew BOOLEAN DEFAULT TRUE, ...
);

-- 用量记录
CREATE TABLE usage_records (
    id BIGINT, tenant_id BIGINT NOT NULL, metric VARCHAR(32) NOT NULL,
    value BIGINT DEFAULT 0, period VARCHAR(20), ...
);

-- 数据库变更日志（审计）
CREATE TABLE db_logs (
    id BIGINT, tenant_id BIGINT NOT NULL, user_id BIGINT,
    action VARCHAR(32), table_name VARCHAR(64), record_id BIGINT,
    old_data JSONB, new_data JSONB, ip VARCHAR(64), created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### 5.6 路由表

```sql
CREATE TABLE routes (
    id BIGINT, tenant_id BIGINT NOT NULL, menu_id BIGINT,
    code VARCHAR(64) NOT NULL, name VARCHAR(64),
    path VARCHAR(255), method VARCHAR(16), ...
);
```

---

## 6. 业务模块独立 DDL

不在 `init_schema.sql` 里：

| 文件 | 内容 | 模块 |
|---|---|---|
| `asset.sql` | `assets` 表 | `apps/reference/asset/` |
| `cms.sql` | `posts` 表 | `apps/cms/` |
| `flag.sql` | `flag_*` 系列表 | `apps/flag/` |
| `message.sql` | `messages` 表 | `apps/tenant/message/` |

每张表在模块目录的 `migrations/<module>.sql`（如有）或直接 `migrations/<module>.sql`。

---

## 7. 种子数据（init_seed.sql）

`init_seed.sql` 提供：

| 概念 | 内容 |
|---|---|
| 平台账号 | `accounts.phone='13800138000'`, `username='admin'`, `password='admin123'`（argon2id 哈希） |
| bootstrap 租户 | `tenants.code='bootstrap'`（新租户克隆源） |
| 基础角色 | `tenant_roles.code='admin' / 'user'` |
| 基础菜单 | dashboard / analytics / projects / team / system（含 6 个子菜单） |
| 权限码 | `resource:*` / `dict:*` / `*`（通配） |
| 字典 | gender / user_status / education（平台 + bootstrap 副本） |
| 配置 | 4 个分组：site / security / email / feature_flag，19 个项 |
| 平台 sys_* | `sys_roles.code='super_admin'`，admin 账号的 sys_user + 角色绑定 |

**全部用 `ON CONFLICT DO NOTHING` / `ON CONFLICT (...) DO UPDATE`**，重复跑不会出错。

---

## 8. 迁移机制

`framework/pkg/migrate.Run(pool, "migrations")` 启动期执行：

```
1. CREATE TABLE IF NOT EXISTS _schema_migrations (version PK, applied_at)
2. 扫 ./migrations/*.sql，按文件名字母序
3. 跳过 _schema_migrations 已记录版本
4. 每条在事务里跑（SET LOCAL row_security = off）
5. 成功则 INSERT _schema_migrations
```

### 8.1 当前迁移文件

```
migrations/
├── init_schema.sql    # 34 张表 + 索引 + RLS + 末尾断言
├── init_seed.sql      # 种子数据
├── asset.sql          # assets 表
├── cms.sql            # posts 表
├── flag.sql           # flag_* 系列表
└── message.sql        # messages 表
```

### 8.2 新增迁移

1. 创建新文件 `0042_add_xxx.sql`（数字前缀保证顺序）
2. 写 DDL（带 `IF NOT EXISTS` / `ADD COLUMN IF NOT EXISTS`）
3. 写数据修复（带 `ON CONFLICT DO NOTHING`）
4. 不要修改已存在的旧文件（会破坏迁移记录）

### 8.3 完整性断言

`init_schema.sql` 末尾有断言块：

```sql
DO $$
DECLARE
    expected_tables TEXT[] := ARRAY[
        'tenants', 'accounts', 'auth_sessions',
        'tenant_organizations', 'tenant_users', 'tenant_roles', ...
    ];
    obsolete_tables TEXT[] := ARRAY[
        'users', 'roles', 'organizations', 'user_roles',
        'role_menus', 'role_resources', 'role_data_scopes',
        'resources', 'account_roles'  -- Phase 0023.3 已 drop/rename
    ];
BEGIN
    -- 检查每张 expected_tables 都存在
    -- 检查每张 obsolete_tables 都不存在
    -- 缺一即 RAISE EXCEPTION
END $$;
```

**作用**：dev 阶段强保证 schema 正确，遗留旧表（0023.3 已 drop）会立即报错。

---

## 9. 完整性约束清单

### 9.1 必填

| 字段 | 表 | 备注 |
|---|---|---|
| `tenant_id` | 租户域表 | NOT NULL，且参与 RLS |
| `is_deleted` | 业务表 | DEFAULT FALSE |
| `created_at` / `updated_at` | 业务表 | DEFAULT NOW() |
| `code` / `name` | 字典、配置、角色、菜单、权限码 | NOT NULL |

### 9.2 唯一约束

| 表 | 唯一索引 |
|---|---|
| `tenants` | `(code) WHERE is_deleted = FALSE` |
| `accounts` | `(phone) WHERE is_deleted = FALSE` + `(email) WHERE is_deleted = FALSE` |
| `tenant_users` | `(account_id, tenant_id) WHERE is_deleted = FALSE` |
| `tenant_roles` | `(tenant_id, code) WHERE is_deleted = FALSE` |
| `tenant_menus` | `(tenant_id, code) WHERE is_deleted = FALSE` |
| `tenant_permissions` | `(tenant_id, code) WHERE is_deleted = FALSE` |
| `sys_users` | `(code) WHERE is_deleted = FALSE` |
| `sys_roles` | `(code) WHERE is_deleted = FALSE` |
| `sys_menus` | `(code) WHERE is_deleted = FALSE` |
| `dicts` | `(code) WHERE tenant_id = 0` + `(tenant_id, code) WHERE tenant_id <> 0` |
| `config_categories` | 同上 |
| `config_items` | `(category_id, key) WHERE tenant_id = 0` + `(tenant_id, category_id, key) WHERE tenant_id <> 0` |

### 9.3 外键（应用层维护，DB 层无 FK）

**有意决策**：跨表外键在 DB 层**不**建（避免删除/迁移时的级联复杂性），由 Repository 显式 `Validate` + `Map` 保证。

例外：`account_id`（业务必填，无 account 不能存在 user），应用层校验。

---

## 10. 索引规范

### 10.1 必带谓词

```sql
-- 软删除谓词：让未删除记录的索引更小
CREATE INDEX idx_xxx ON xxx (col) WHERE is_deleted = FALSE;
```

### 10.2 JSONB GIN 索引

```sql
-- 字典 / 配置 / 租户 config 的 JSONB 查询
CREATE INDEX idx_tenants_config_gin ON tenants USING GIN (config);
```

### 10.3 组合索引

按常用查询模式：

```sql
CREATE INDEX idx_tenant_users_tenant ON tenant_users (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_tenant_users_org ON tenant_users (org_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_tenant_ur_user ON tenant_user_roles (user_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_tenant_ur_role ON tenant_user_roles (role_id) WHERE is_deleted = FALSE;
```

---

## 11. JSONB 用法

### 11.1 显式 cast

```sql
-- 写入
INSERT INTO tenants (config) VALUES ('{"theme": "dark"}'::jsonb);

-- 更新
UPDATE tenants SET config = config || '{"language": "zh-CN"}'::jsonb
WHERE id = 1;

-- 查询
SELECT * FROM tenants WHERE config->>'theme' = 'dark';
SELECT * FROM tenants WHERE config @> '{"theme": "dark"}'::jsonb;
```

### 11.2 GIN 索引

```sql
CREATE INDEX idx_tenants_config_gin ON tenants USING GIN (config);
-- 支持 @> 、 ? 、 ?& 、 ?| 操作符
```

### 11.3 应用层交互

Go 端 pgx 默认把 `string` 当 text、`[]byte` 当 bytea 发。所有 JSONB 字段在 SQL 中**必须**显式 `::jsonb` cast；Go 端用 `json.Marshal` 后传 `string` 或 `[]byte` 都可以（pgx 会接受 `[]byte` + cast）。

---

## 12. 性能与维护

### 12.1 常用维护 SQL

```sql
-- 查看表大小
SELECT relname, pg_size_pretty(pg_total_relation_size(relid))
FROM pg_stat_user_tables
ORDER BY pg_total_relation_size(relid) DESC;

-- 重建索引
REINDEX INDEX idx_tenant_users_tenant;

-- 真空（回收软删除行）
VACUUM ANALYZE tenant_users;

-- 序列重置
SELECT setval('tenant_users_id_seq', (SELECT MAX(id) FROM tenant_users), true);
```

### 12.2 连接池

```yaml
database:
  max_open_conns: 100
  max_idle_conns: 20
  conn_max_lifetime_sec: 300
  conn_max_idle_time_sec: 60
```

pgxpool 默认行为 + 上述配置。监控：`pool.Stat()`。

### 12.3 慢查询定位

```sql
-- 启用慢查询日志（postgresql.conf）
log_min_duration_statement = 500  # 500ms
```

应用层日志（`framework/framework.go:setupRouter` 的 Logger 中间件）记录每条请求耗时。
