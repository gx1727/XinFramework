# 数据库设计

> 当前 30+ 张表,核心表在 `migrations/framework.sql`,业务表分散在 `cms.sql` / `dict.sql` / `flag.sql` / `asset.sql`。

## 1. 扩展

迁移脚本默认装这两个 PG 扩展:

```sql
CREATE EXTENSION IF NOT EXISTS ltree;      -- 路径/树形存储
CREATE EXTENSION IF NOT EXISTS pg_trgm;    -- trigram 模糊匹配
```

需要 PG ≥ 14,扩展需要 superuser 权限装一次。

## 2. 迁移机制

启动时由 `framework/pkg/migrate.Run("migrations")` 执行:

- 扫 `./migrations/*.sql` 按文件名排序
- 在 `schema_migrations` 表里记录已执行版本
- 重复执行跳过

**注意**:迁移是**幂等的**(所有 `CREATE TABLE IF NOT EXISTS` / `CREATE INDEX IF NOT EXISTS`)。但 `ALTER TABLE` 类操作要自己保证幂等。

## 3. 核心表总览

### 3.1 ER 关系

```
tenants                       accounts ─── account_auths
   │                          │   │   │
   │                          │   │   └── account_roles  ← 平台角色(super_admin)
   │                          │   │
   │                          │   └── user_codes  ← 验证码
   │                          │
   ├── organizations          │
   │      ↑ (parent_id)       │
   │                          │
   ├── users ─── user_roles ───┤
   │      │     │              │
   │      │     └── roles      │
   │      │           │        │
   │      │           ├── role_menus
   │      │           │      │
   │      │           │      └── menus ─── resources (按 menu_id)
   │      │           │
   │      │           └── role_resources
   │      │
   │      └──(creator_id)
   │
   ├── file_assets            ← 所有 module 通用附件
   ├── dicts ─── dict_items
   ├── frames / avatars ...   ← flag 业务
   └── posts                  ← cms 业务
```

### 3.2 表清单

#### 平台级表(不受 RLS)

| 表 | 用途 | 关键字段 |
|---|---|---|
| `tenants` | 租户 | `code`(唯一)、`status`、`config jsonb` |
| `accounts` | 全局账号 | `username/phone/email`(各自唯一)、`password hash` |
| `account_auths` | 第三方授权 | `provider`、`openid`、`account_id` |
| `account_roles` | 平台角色 | `account_id`、`role`(如 `super_admin`) |
| `user_codes` | 验证码 | `account_id`、`code`、`expire_at` |

#### 租户级表(受 RLS)

| 表 | 用途 | 关键字段 |
|---|---|---|
| `organizations` | 组织架构 | `parent_id`、`ancestors`、`code`(租户内唯一) |
| `users` | 租户用户 | `tenant_id`、`account_id`、`code`、`org_id` |
| `user_roles` | 用户-角色 | `tenant_id`、`user_id`、`role_id` |
| `roles` | 角色 | `tenant_id`、`code`、`data_scope` |
| `role_menus` | 角色-菜单 | `role_id`、`menu_id` |
| `role_resources` | 角色-资源 | `role_id`、`resource_id`、`effect` |
| `menus` | 菜单 | `parent_id`、`code`、`path` |
| `resources` | 资源(按钮/API) | `code`、`action`、`menu_id` |
| `file_assets` | 文件 | `url`、`size`、`mime`、`owner_id` |
| `dicts` | 字典 | `code`、`name` |
| `dict_items` | 字典项 | `dict_id`、`code`、`name`、`parent_id` |
| `frames` / `frame_categories` / `spaces` / `avatars` / `avatar_categories` | flag 业务 | 见 [flag.sql](../migrations/flag.sql) |
| `posts` | cms 业务 | 见 [cms.sql](../migrations/cms.sql) |

## 4. 行级安全(RLS)

**多租户隔离通过 `db.RunInTenantTx(ctx, pool, tenantID, fn)` 实现**:把 `SET LOCAL app.tenant_id = <id>` 注入事务,然后查询触发表上定义的 RLS 策略。

### 4.1 一个 RLS 例子(users 表)

```sql
-- migrations/framework.sql 包含
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE users FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON users
    USING (tenant_id::text = current_setting('app.tenant_id', true));
```

查询 `SELECT * FROM users` 在没设置 `app.tenant_id` 时会返回 0 行。

### 4.2 在 Go 里套租户上下文

```go
err := db.RunInTenantTx(ctx, db.Get(), claims.TenantID, func(txCtx context.Context) error {
    // txCtx 里 SET LOCAL app.tenant_id = claims.TenantID
    // 这里的 SELECT/INSERT/UPDATE 自动受 RLS 限制
    return s.repo.GetByID(txCtx, userID)
})
```

**为什么用 `txCtx`?** 因为 `db.GetQuerier(ctx)` 优先返回 `ctx` 上的 tx,所以整个回调共享一个事务。

### 4.3 不受 RLS 的表

以下表的查询**不要**套 `RunInTenantTx`:

| 表 | 原因 |
|---|---|
| `accounts` | 全局唯一,登录时不知道 tenant_id |
| `account_auths` | 第三方授权也是全局维度 |
| `account_roles` | 平台角色,跨租户 |
| `tenants` | 平台管理,需要跨租户查询 |

## 5. 软删除

所有业务表都有 `is_deleted BOOLEAN DEFAULT FALSE` + `created_at` / `updated_at` / `created_by` / `updated_by`。

**约定**:

- 查询默认带 `WHERE is_deleted = FALSE`
- 唯一索引都是部分索引:`UNIQUE INDEX ... WHERE is_deleted = FALSE`
- "删除" 实际是 `UPDATE ... SET is_deleted = TRUE`,数据保留
- "硬删"(物理 DELETE)只用于 `purge` 类操作(如 `POST /tenants/:id/purge`)

例子:

```sql
CREATE UNIQUE INDEX uk_users_account ON users (tenant_id, account_id)
    WHERE is_deleted = FALSE;
```

这保证同一 account 在同一租户内只能有一个 user 行,但删除后可以重建。

## 6. 索引策略

每个表都至少有以下索引:

| 索引 | 字段 | 说明 |
|---|---|---|
| 主键 | `id` | `BIGINT GENERATED ALWAYS AS IDENTITY` |
| `created_at` | 默认 `idx_xxx_created_at` | 排序 / 增量同步 |
| `is_deleted` 部分索引 | 配合其他唯一性 | 软删除 + 唯一 |

高频查询字段都有专门的 `idx_*`,例如:

```sql
CREATE INDEX idx_users_tenant_org ON users (tenant_id, org_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_users_code_trgm ON users USING gin (code gin_trgm_ops);
CREATE INDEX idx_org_tenant_parent ON organizations (tenant_id, parent_id) WHERE is_deleted = FALSE;
```

模糊搜索用 `gin_trgm_ops`(由 `pg_trgm` 扩展提供),对中文支持有限,生产环境建议加全文索引或外置 ES。

## 7. 物化路径

`organizations` 表用 ltree 风格的 `ancestors TEXT` 字段做物化路径:

```
ancestors = ""                   ← 顶级
ancestors = "/3/"                ← parent_id=3 的子
ancestors = "/3/7/"              ← parent_id=7,parent_id=3
```

快速查某节点的所有祖先:

```sql
SELECT * FROM organizations
WHERE id = ANY(string_to_array(trim(ancestors, '/'), '/')::bigint[]);
```

快速查某节点的所有后代:

```sql
SELECT * FROM organizations WHERE ancestors LIKE '/3/%';
```

## 8. 时区

所有 `TIMESTAMPTZ DEFAULT NOW()` —— PostgreSQL 内部用 UTC 存储,Go 端用 `time.Time` 自动按本地时区渲染。

生产建议:

- DB server TZ = UTC
- 应用 server TZ = Asia/Shanghai
- 所有跨时区逻辑在应用层处理

## 9. JSONB 字段

部分表有 `JSONB` 字段存半结构化数据:

| 表 | 字段 | 用途 |
|---|---|---|
| `tenants.config` | jsonb | 租户级配置(主题、限额、特性开关) |
| 各种业务表 | `extra` / `metadata` | 业务扩展属性 |

可以用 GIN 索引:

```sql
CREATE INDEX idx_tenants_config_gin ON tenants USING GIN (config);
```

## 10. 迁移操作清单

新加表/字段的流程:

```bash
# 1. 改 SQL 文件
vi migrations/framework.sql     # 加 CREATE TABLE IF NOT EXISTS xxx

# 2. 在仓库根目录跑
psql -h localhost -U xin_user -d xin -f migrations/framework.sql

# 3. 提交 SQL + Go 实体(apps/.../model.go)
git add migrations/ apps/
git commit -m "feat(db): add xxx table"

# 4. 部署后 xin restart 会自动跑未执行的迁移
```

**重要**:永远不要直接改已经上线的迁移脚本。要"反向"加字段就写新脚本做 `ALTER TABLE`,**不要**改 `CREATE TABLE`。

## 11. 数据完整性约束

### 11.1 FK 关系

```sql
-- 示例:users -> tenants + accounts
ALTER TABLE users
    ADD CONSTRAINT fk_users_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_users_account FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE RESTRICT;
```

`tenant ON DELETE CASCADE`:租户硬删 → users 全清。
`account ON DELETE RESTRICT`:账号不能被删,只能软删 + 禁用。

### 11.2 Check 约束

```sql
ALTER TABLE accounts
    ADD CONSTRAINT chk_accounts_status CHECK (status IN (0, 1));
```

## 12. 备份与恢复

不在框架范围内,但典型做法:

```bash
# 全量备份
pg_dump -h localhost -U xin_user -d xin > backup.sql

# 恢复
psql -h localhost -U xin_user -d xin < backup.sql
```

生产建议:

- WAL 归档 + 时间点恢复(`archive_mode = on`)
- 异地副本(`streaming replication`)
- 每天全量备份 + 持续增量

## 13. 性能调优参考

| 表大小 | 建议 |
|---|---|
| < 100 万行 | 无需分区 |
| 100 万 - 1 亿 | 按 `tenant_id` 范围分区 |
| > 1 亿 | 按 `tenant_id` 哈希分区 + 定期归档 |

常见热点表(users / accounts / resources)走 `tenant_id + status` 联合索引即可。flag 业务(avatars / frames)需要按 `creator_id` 索引,因为 DataScopeSelf 大量用 `WHERE creator_id = $1`。

## 14. 监控

关键指标:

- `pg_stat_user_tables`:各表 `seq_scan` vs `idx_scan` 比例(`> 0.1` 提示索引缺失)
- `pg_stat_user_indexes`:索引使用频率(`idx_scan = 0` 是无用索引)
- `pg_locks`:锁等待
- `pg_stat_activity`:长事务(`state = 'active' AND query_start < now() - interval '1 min'`)

具体 SQL 见 [PostgreSQL 官方文档](https://www.postgresql.org/docs/current/monitoring-stats.html)。