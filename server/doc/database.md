# 数据库

> 表结构、迁移、命名约定。

## 总览

数据库：PostgreSQL 16+。

主要迁移文件：

| 文件 | 范围 |
| --- | --- |
| `migrations/framework.sql` | 平台 + 框架核心表（accounts, tenants, users, roles, menus, resources, organizations, dicts, attachments, sessions, platform_roles, role_resources, data_scopes, role_menus） |
| `migrations/cms.sql` | CMS 业务表 |
| `migrations/flag.sql` | Flag 业务表 |

迁移执行顺序：**按文件名排序**。框架自带 [framework/pkg/migrate](file:///d:\work\xin\XinFramework\server\framework\pkg\migrate) 按序跑 `.sql` 文件。

## 命名约定

| 类型 | 规则 | 示例 |
| --- | --- | --- |
| 表名 | 复数 snake_case | `users`, `role_resources`, `account_auths` |
| 主键 | `id BIGSERIAL` 或 `id BIGINT` | `id` |
| 外键 | `<单数表名>_id` | `user_id`, `tenant_id`, `role_id` |
| 时间戳 | `created_at TIMESTAMPTZ`, `updated_at TIMESTAMPTZ` | —— |
| 软删 | `is_deleted BOOLEAN DEFAULT FALSE`, `deleted_at TIMESTAMPTZ` | —— |
| 操作人 | `created_by BIGINT`, `updated_by BIGINT` | —— |
| 枚举 | `SMALLINT`（不要用 PG ENUM 类型） | `status SMALLINT` |
| 金额 | `NUMERIC(20, 4)` | `amount NUMERIC(20,4)` |
| 字符集 | UTF-8 | —— |
| 排序 | `sort INT DEFAULT 0` | —— |

## 多租户

**所有租户内业务表必带** `tenant_id BIGINT NOT NULL`。

```sql
CREATE TABLE users (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT NOT NULL,
    account_id  BIGINT NOT NULL,
    ...
);

CREATE INDEX idx_users_tenant ON users(tenant_id, is_deleted, created_at DESC);
```

查询模板：

```sql
SELECT * FROM users
WHERE tenant_id = $1
  AND is_deleted = FALSE
  AND ($2 = '' OR code ILIKE '%' || $2 || '%')
ORDER BY id DESC
LIMIT $3 OFFSET $4;
```

`is_deleted = FALSE` 必须显式过滤。

## 软删

```sql
UPDATE users SET is_deleted = TRUE, deleted_at = NOW() WHERE id = $1;
```

而非 `DELETE FROM users`。前端列表永远查 `is_deleted = FALSE`。

## 时间戳

统一 `TIMESTAMPTZ`（带时区），Go 端用 `time.Time` 解析。**不要用** `TIMESTAMP`（不带时区）。

## 主键策略

`BIGSERIAL`（自增）。预留 64 位容量，对租户数 / 用户数 / 数据量都够。

如果将来要分库分表，再考虑雪花算法或 UUID。但当前阶段一律 `BIGSERIAL`。

## JSON 字段

存配置 / 扩展字段时用 `JSONB`：

```sql
config JSONB NOT NULL DEFAULT '{}'::jsonb
```

查询用 `config->>'key'`：

```sql
SELECT * FROM tenants WHERE config->>'theme' = 'dark';
```

## 索引约定

- **复合索引**：高频查询 `tenant_id, is_deleted, <order_field>` 一起建
- **外键**：必加索引
- **unique 索引**：业务唯一键（如 `users(tenant_id, code)`）

```sql
CREATE UNIQUE INDEX uk_users_tenant_code ON users(tenant_id, code)
WHERE is_deleted = FALSE;
```

注意 `UNIQUE` 加 `WHERE is_deleted = FALSE` 是关键——避免软删后能复用 code。

## 表清单（框架）

### accounts — 全局账号

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | BIGSERIAL PK | |
| username | VARCHAR(64) UNIQUE | 登录账号 |
| phone | VARCHAR(20) | 可空，唯一 |
| email | VARCHAR(128) | 可空，唯一 |
| real_name | VARCHAR(64) | |
| password_hash | TEXT | argon2id |
| status | SMALLINT | 1=active 0=disabled |
| last_login_at | TIMESTAMPTZ | |
| is_deleted | BOOLEAN | |
| created_at/updated_at | TIMESTAMPTZ | |

### account_auths — 第三方授权

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | BIGSERIAL PK | |
| tenant_id | BIGINT | |
| account_id | BIGINT FK accounts | |
| type | VARCHAR(16) | wechat / qq / weibo / wxxcx |
| openid | VARCHAR(128) | |
| unionid | VARCHAR(128) | |
| session_key | TEXT | |
| created_at/updated_at | TIMESTAMPTZ | |

UNIQUE `(tenant_id, type, openid)`。

### tenants — 租户

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | BIGSERIAL PK | |
| code | VARCHAR(64) UNIQUE | 租户编码 |
| name | VARCHAR(128) | |
| status | SMALLINT | 1=active 0=disabled |
| contact / phone / email | VARCHAR | 联系人信息 |
| province / city / area / address | VARCHAR | 地址 |
| config | JSONB | 租户级配置 |
| dashboard | TEXT | 仪表盘 JSON |
| is_deleted | BOOLEAN | |
| created_at/updated_at | TIMESTAMPTZ | |

### users — 租户内用户

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | BIGSERIAL PK | |
| tenant_id | BIGINT | |
| account_id | BIGINT FK accounts | 关联全局账号 |
| org_id | BIGINT FK organizations NULL | |
| code | VARCHAR(64) | 租户内工号 |
| nickname / real_name / avatar / phone / email | VARCHAR | |
| status | SMALLINT | |
| is_deleted | BOOLEAN | |

UNIQUE `(tenant_id, code) WHERE is_deleted = FALSE`。

### organizations — 组织

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | BIGSERIAL PK | |
| tenant_id | BIGINT | |
| parent_id | BIGINT | 0 = 根 |
| name | VARCHAR(128) | |
| path | LTREE 或 TEXT | 物化路径 |
| sort | INT | |
| is_deleted | BOOLEAN | |

### roles — 角色

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | BIGSERIAL PK | |
| tenant_id | BIGINT | |
| code | VARCHAR(64) | admin / user / ... |
| name | VARCHAR(128) | 显示名 |
| data_scope_type | SMALLINT | 1=All 2=Custom 3=Dept 4=DeptAndBelow 5=Self |
| data_scope_org_ids | BIGINT[] | 仅 type=Custom |
| is_builtin | BOOLEAN | 内置角色（不可删） |
| is_deleted | BOOLEAN | |

### role_menus — 角色-菜单绑定

`role_id`, `menu_id` 联合主键。

### role_resources — 角色-资源绑定

`role_id`, `resource_id` 联合主键。

### menus — 菜单

`id, tenant_id, parent_id, code, title_i18n_key, path, icon, sort, is_deleted`。

### resources — 资源（API / 按钮）

`id, tenant_id, code (unique), type (api/button), name, method, path`。

### dicts — 字典分类

`id, code (unique), name, description, is_deleted`。

### dict_items — 字典项

`id, dict_id, code, label, value, sort, is_deleted`。

UNIQUE `(dict_id, code) WHERE is_deleted = FALSE`。

### attachments — 附件

`id, tenant_id, owner_id, filename, content_type, size, storage_key, is_deleted`。

### sessions — 会话

`id (uuid), user_id, tenant_id, role, expires_at`。会话存储由 `framework/pkg/session` 抽象（默认内存，可换 Redis）。

### platform_roles — 平台角色

`account_id, role` 联合主键。常用值：`super_admin`。

## 事务

Go 端用 `framework/pkg/db.WithTx`：

```go
import "gx1727.com/xin/framework/pkg/db"

err := db.WithTx(ctx, func(tx pgx.Tx) error {
    _, err := tx.Exec(ctx, "INSERT INTO ...", ...)
    if err != nil { return err }
    _, err = tx.Exec(ctx, "UPDATE ...", ...)
    return err
})
```

事务里通过 `db.GetQuerier(ctx)` 拿 tx，对 repository 无侵入。

## 迁移

```go
// framework/pkg/migrate/migrate.go
func Run(dir string) error {
    files, _ := filepath.Glob(filepath.Join(dir, "*.sql"))
    sort.Strings(files)
    for _, f := range files {
        data, _ := os.ReadFile(f)
        // 切分 ; 逐条执行
    }
}
```

新加 SQL：

1. 在 `migrations/` 加 `<timestamp>_<name>.sql`
2. 启动时自动跑

**注意**：

- 不要破坏已有表结构（用 `ALTER TABLE ADD COLUMN ... DEFAULT NULL`）
- 删字段前先 `ALTER TABLE DROP COLUMN IF EXISTS`
- 大批量 UPDATE 分批提交，避免长事务

## 备份与恢复

```bash
# 备份
pg_dump -U xin -h localhost xin > xin_backup_$(date +%Y%m%d).sql

# 恢复
psql -U xin -h localhost xin < xin_backup_20260615.sql
```

生产建议：

- 每日全量备份 + WAL 归档（pg_basebackup）
- 异地存储备份文件
- 定期演练恢复