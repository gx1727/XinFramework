# 数据库表设计规范

## 1. 基本约定

- 数据库：PostgreSQL 14+
- 编码：`SET client_encoding = 'UTF8';`
- 主键：`id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY`（禁止手动赋值）
- 表名：蛇形复数名词（如 `users`、`dict_items`、`ai_documents`）
- 列名：蛇形命名（如 `tenant_id`、`created_at`、`is_deleted`）
- 所有表和列必须写 `COMMENT ON`

## 2. 必备字段

每张表必须包含以下字段（除非有明确理由省略）：

```sql
id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
created_at TIMESTAMPTZ DEFAULT NOW(),
updated_at TIMESTAMPTZ DEFAULT NOW(),
is_deleted BOOLEAN     DEFAULT FALSE
```

业务表（有归属人的表）额外加：

```sql
created_by BIGINT,
updated_by BIGINT,
```

## 3. 多租户字段

所有租户级数据表必须包含 `tenant_id`：

```sql
tenant_id BIGINT NOT NULL,
```

例外（不加 `tenant_id`）：
- `accounts` — 全局账号表
- `plans` — 全局套餐表
- `db_logs` — 审计日志（`tenant_id` 可选）

单租户项目：`tenant_id` 统一为 `0`，不影响功能。

## 4. 索引规范

### 4.1 租户索引

所有含 `tenant_id` 的表必须建：

```sql
CREATE INDEX idx_{表名简写}_tenant ON {表名} (tenant_id) WHERE is_deleted = FALSE;
```

### 4.2 唯一索引

业务唯一约束必须使用**部分唯一索引**（排除已删除记录）：

```sql
CREATE UNIQUE INDEX uk_{表名}_{列名} ON {表名} (tenant_id, code) WHERE is_deleted = FALSE;
```

可空列的唯一索引需额外排除 NULL：

```sql
CREATE UNIQUE INDEX uk_{表名}_{列名} ON {表名} (phone) WHERE is_deleted = FALSE AND phone IS NOT NULL;
```

### 4.3 索引命名

| 类型 | 前缀 | 示例 |
| --- | --- | --- |
| 唯一索引 | `uk_` | `uk_users_code` |
| 普通索引 | `idx_` | `idx_users_tenant` |
| GIN 索引 | `idx_` + `_gin` | `idx_tenants_config_gin` |

### 4.4 JSONB 列

用于查询的 JSONB 列建 GIN 索引：

```sql
CREATE INDEX idx_{表名}_{列名}_gin ON {表名} USING GIN ({列名});
```

## 5. 数据类型规范

| 场景 | 类型 | 示例 |
| --- | --- | --- |
| 主键 | `BIGINT GENERATED ALWAYS AS IDENTITY` | `id` |
| 外键关联 | `BIGINT` / `BIGINT NOT NULL` | `tenant_id`, `user_id` |
| 短文本 | `VARCHAR(n)` | `code VARCHAR(32)`, `name VARCHAR(64)` |
| 长文本 | `TEXT` | `content TEXT` |
| 状态/枚举 | `SMALLINT` | `status SMALLINT DEFAULT 1` |
| 布尔标记 | `BOOLEAN DEFAULT FALSE` | `is_deleted BOOLEAN` |
| 时间 | `TIMESTAMPTZ`（带时区） | `created_at TIMESTAMPTZ DEFAULT NOW()` |
| 金额 | `DECIMAL(10,2)` | `price DECIMAL(10,2) DEFAULT 0` |
| 扩展数据 | `JSONB` | `config JSONB`, `extend JSONB` |
| IP 地址 | `INET` | `client_ip INET` |

禁止使用 `TIMESTAMP`（不带时区）。

## 6. RLS（行级安全）

### 6.1 架构原则

**应用层为主，RLS 为纵深防御**。RLS 仅在关键表上启用，防止应用层漏写 `WHERE tenant_id` 时的数据泄漏，或运维直接操作 DB 时的越权。

### 6.2 RLS 关键表（保留策略）

以下 7 个表启用 RLS 策略：

| 表 | 保留原因 |
|---|---|
| `users` | 用户数据最敏感，漏写 WHERE tenant_id 风险高 |
| `roles` | 权限体系核心，RLS 兜底 |
| `permissions` | 权限配置表 |
| `role_data_scopes` | 数据范围配置表 |
| `user_roles` | 用户-角色关联 |
| `organizations` | 组织树，层级深容易遗漏 |
| `tenant_user_seq` | 租户下用户序号生成 |

**标准 RLS 策略**：
```sql
ALTER TABLE {表名} ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON {表名}
    USING (
        tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    );
```

### 6.3 应用层管控表（无 RLS）

以下表不启用 RLS，依赖应用层显式 `WHERE tenant_id` + 索引隔离：

| 表 | 原因 |
|---|---|
| `menus` | 应用层控制力度强 |
| `resources` | 按钮权限，菜单关联查询不会跨租户 |
| `routes` | API 路由配置，按租户查询即可 |
| `dicts` / `dict_items` | 字典数据相对静态 |
| `subscriptions` / `usage_records` | 账单相关，按租户查询即可 |
| `ai_documents` / `attachments` | 文件/文档相对独立 |

### 6.4 安全默认值

**未设置 `app.tenant_id` 时，拒绝访问所有 RLS 保护的表。**

**⚠️ 架构红线**：
1. **绝对禁止在 RLS 中加入业务逻辑**：不要在 RLS 策略中写 `is_deleted = FALSE` 或状态判断。如果在 RLS 中加入 `is_deleted = FALSE`，当执行软删除（`UPDATE xxx SET is_deleted = TRUE`）时，新行将无法通过 RLS 校验，导致 `new row violates row-level security policy` 报错。
2. 软删除、状态过滤等纯业务逻辑，必须由 Repository 层的 SQL `WHERE` 条件来保证。
3. `tenants` 这类特殊的全局表，由于自身的主键是 `id`，且需要允许系统管理员（`tenant_id = 0`）访问所有租户，其 RLS 策略应写为：
```sql
CREATE POLICY tenant_isolation_policy ON tenants
    USING (
        current_setting('app.tenant_id', true) = '0'
        OR id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    );
```

## 7. 软删除

统一使用 `is_deleted BOOLEAN DEFAULT FALSE`，不物理删除数据。

查询时默认加条件：

```sql
WHERE is_deleted = FALSE
```

## 8. COMMENT 规范

### 8.1 表注释

```sql
COMMENT ON TABLE {表名} IS '{中文名} - {简要说明}';
```

示例：`COMMENT ON TABLE users IS '租户用户表 - 租户内的用户信息';`

### 8.2 列注释

每个列都必须有 COMMENT：

```sql
COMMENT ON COLUMN {表名}.{列名} IS '{说明}';
```

枚举值/状态值在注释中说明取值范围：

```sql
COMMENT ON COLUMN users.status IS '用户状态：0-禁用，1-启用';
```

## 9. 新建表 Checklist

创建新表时按以下清单逐项检查：

- [ ] `id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY`
- [ ] `tenant_id BIGINT NOT NULL`（租户级表）
- [ ] `created_at TIMESTAMPTZ DEFAULT NOW()`
- [ ] `updated_at TIMESTAMPTZ DEFAULT NOW()`
- [ ] `created_by BIGINT`（业务表）
- [ ] `updated_by BIGINT`（业务表）
- [ ] `is_deleted BOOLEAN DEFAULT FALSE`
- [ ] 租户索引：`CREATE INDEX idx_xxx_tenant ON xxx (tenant_id) WHERE is_deleted = FALSE;`
- [ ] 唯一索引使用部分索引：`WHERE is_deleted = FALSE`
- [ ] RLS 启用 + 隔离策略
- [ ] `COMMENT ON TABLE`
- [ ] 每列都有 `COMMENT ON COLUMN`
- [ ] 无 `TIMESTAMP`（用 `TIMESTAMPTZ`）
