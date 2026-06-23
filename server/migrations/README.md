# 数据库迁移

> XinFramework 的 PostgreSQL schema 通过 `framework/pkg/migrate` 包的 `migrate.Run(pool, "migrations")` 在启动时按文件名升序自动应用。本目录放所有 DDL + 初始 seed。

## 1. 目录结构

```
migrations/
├── init_schema.sql      # 全部 schema（34 张表 + 索引 + RLS）—— 0023.3 终态
├── init_seed.sql        # 全部种子数据（admin/角色/菜单/权限码/字典/配置）
├── asset.sql            # 附件业务表（独立模块）
├── cms.sql              # CMS 业务表（独立模块）
├── flag.sql             # 头像/相框业务表（独立模块）
└── README.md            # 本文件
```

**为什么分两个 init 文件**：
- `init_schema.sql` —— 改 schema 的频率低（一次成型后基本稳定）
- `init_seed.sql` —— 改 seed 的频率高（调默认角色、菜单、字典时）
- dev 阶段重置时分开跑：先 schema 后 seed

**为什么 3 个业务文件保留**：
- asset / cms / flag 是**独立业务模块**，跟核心 schema 生命周期不同
- 合并会失去模块边界（flag 加张表要重跑 init_schema，不合理）
- 保留独立 .sql 让"业务表变更"和"核心表变更"互不干扰

## 2. 0023.3 终态表清单（init_schema.sql）

按数据域分组，**全部 tenant 域表启用 RLS**：

| 域 | 表 | RLS | 备注 |
|---|---|---|---|
| **平台与租户公共** | `tenants`, `accounts`, `auth_sessions` | ❌ | 跨域共享登录凭证 |
| **租户域** | `tenant_organizations`, `tenant_users`, `tenant_roles` | ✅ | 全部带 `tenant_id` |
| | `tenant_role_data_scopes`, `tenant_user_roles` | ✅ | |
| | `tenant_menus`, `tenant_permissions` | ✅ | 0023.3 物理拆出（menus → tenant_menus + sys_menus） |
| | `tenant_role_menus`, `tenant_role_resources` | ✅ | |
| | `tenant_user_seq` | ✅ | 租户用户序号 |
| **平台域** | `sys_users`, `sys_orgs`, `sys_roles`, `sys_menus` | ❌ | platform 是单租户概念，靠 API 层 `RequirePlatformRole(super_admin)` 守护 |
| | `sys_permissions`, `sys_user_roles`, `sys_role_menus`, `sys_role_permissions` | ❌ | |
| **业务支撑** | `subscriptions`, `usage_records`, `db_logs`, `routes` | ✅ | 按租户分账 |
| | `plans` | ❌ | 全局套餐表 |
| **字典 / 配置** | `dicts`, `dict_items`, `config_categories`, `config_items` | ✅ | 保留 `tenant_id=0` 短路：平台级跨租户共享 |
| | `dict_visibility`, `config_visibility` | ❌ | 可见性矩阵 |

**0023.3 关键表 rename（已 drop 的旧表）**：

| 旧 | 新 |
|---|---|
| `users` | `tenant_users` |
| `roles` | `tenant_roles` |
| `organizations` | `tenant_organizations` |
| `user_roles` | `tenant_user_roles` |
| `role_menus` | `tenant_role_menus` |
| `role_resources` | `tenant_role_resources` |
| `role_data_scopes` | `tenant_role_data_scopes` |
| `resources` | `tenant_permissions` |
| `menus`（`scope='tenant'` 部分） | `tenant_menus`（无 `scope` 字段） |
| `menus`（`scope='platform'` 部分） | `sys_menus` |
| `account_roles` | **drop**（由 `sys_user_roles` + `sys_roles` 替代） |

init_schema.sql 末尾的 `DO $$ ... $$` 块会校验：
- **所有 34 张目标表都建好**（缺一张就 RAISE EXCEPTION）
- **9 张旧表全部 drop**（遗留任一张就 RAISE EXCEPTION）

dev 库重置后可以 `\d` 看到所有 tenant_ / sys_ 表，没有任何 users/roles/menus 等旧表名残留。

## 3. 迁移机制

### 3.1 运行时机

```
framework.Serve(cfg, app, rt, modules)
  └─ migrate.Run(app.DB, "migrations")         // 启动时跑
       ├─ 扫 ./migrations/*.sql，按文件名升序
       ├─ 跳过 _schema_migrations 表里已记录的版本
       └─ 在事务里跑未应用的版本（事务开头 SET LOCAL row_security = off）
```

### 3.2 版本跟踪

`_schema_migrations` 表记录已应用的迁移：

```sql
CREATE TABLE IF NOT EXISTS _schema_migrations (
    version    VARCHAR(255) PRIMARY KEY,    -- 文件名（如 init_schema.sql）
    applied_at TIMESTAMPTZ DEFAULT NOW()
);
```

`migrate.Run` 用文件名做主键。**修改已部署的文件名 = 重跑迁移**（除非要承担风险）。

### 3.3 事务保证

每个 SQL 文件在 `db.RunInTx` 里执行，开头关 RLS：

```go
tx.Exec("SET LOCAL row_security = off")  // 关闭 RLS，让迁移本身能改任何表
tx.Exec(sql)                              // 跑该文件的全部 SQL
```

任何 DDL 失败 = 整个文件回滚 = `_schema_migrations` 不记录。

## 4. dev 阶段重置流程

```bash
# 1. 删库（dev 阶段：先 DROP 再 CREATE，因为表结构是终态，不考虑迁移）
psql -h localhost -U postgres -c "DROP DATABASE xin_dev;"
psql -h localhost -U postgres -c "CREATE DATABASE xin_dev;"

# 2. 跑核心 schema + seed
psql -h localhost -U xin -d xin_dev -f migrations/init_schema.sql
psql -h localhost -U xin -d xin_dev -f migrations/init_seed.sql

# 3. 跑业务模块 schema
psql -h localhost -U xin -d xin_dev -f migrations/asset.sql
psql -h localhost -U xin -d xin_dev -f migrations/cms.sql
psql -h localhost -U xin -d xin_dev -f migrations/flag.sql

# 4. 启动 dev 服务
cd server && go run ./cmd/xin
```

或者启动服务时自动跑（只要 `_schema_migrations` 没记录过这 3 个 init 文件）：

```bash
go run ./cmd/xin    # migrate.Run 自动按字母序应用 init_schema.sql → init_seed.sql → asset.sql → ...
```

## 5. 新增迁移

### 5.1 改 seed（最常见）

直接编辑 `init_seed.sql`，**不要新建**日期前缀文件。开发期重置比增量迁移简单。

### 5.2 改 schema（少见）

如果只是新增列 / 索引：
- 编辑 `init_schema.sql` 加 `ALTER TABLE ... ADD COLUMN IF NOT EXISTS`
- 跑过的 dev 库手动跑这段 ALTER（dev 可丢）
- 新 dev 库通过 init_schema.sql 一遍跑下来

如果是新增 / 删除整张表：
- 编辑 `init_schema.sql` 加 `CREATE TABLE` / `DROP TABLE` 段
- 同步考虑业务模块是否受影响（如 `tenant_user_seq` 改字段可能影响 first_install.go）
- **同时**更新末尾 `expected_tables` 数组——缺一张就 RAISE EXCEPTION

### 5.3 业务模块新增表

**不要**塞进 `init_schema.sql`——在你的业务模块 .sql 里加：

```
init_schema.sql   # 不动
init_seed.sql     # 不动
asset.sql         # 你加 CREATE TABLE
```

业务模块 seed（如果有）也写在各自的 .sql 里。

### 5.4 字段命名约定

- 全部业务表必须有 `is_deleted BOOLEAN DEFAULT FALSE`
- 所有索引加 `WHERE is_deleted = FALSE` 谓词
- 租户域表必须有 `tenant_id BIGINT NOT NULL` + RLS policy
- 平台域表（`sys_*`）**不**带 `tenant_id`、**不**启用 RLS

### 5.5 RLS 写入保护

所有 tenant 域表的 RLS policy 必须用 `OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'` 短路，让 `db.RunInPlatformTx`（设 `app.bypass_rls='on'`）能跨租户读写。

写新表时复制已有 RLS policy 模板：

```sql
ALTER TABLE xxx ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON xxx;
CREATE POLICY tenant_isolation_policy ON xxx USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);
```

字典/配置类表（`tenant_id=0` 表示平台级）多加一条 `tenant_id = 0` 短路：

```sql
CREATE POLICY tenant_isolation_policy ON xxx USING (
    tenant_id = 0
    OR tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);
```

## 6. 编写规范

### 6.1 DDL 幂等性

所有 DDL 必须幂等（迁移可能重跑）：

```sql
-- ✅ 推荐
CREATE TABLE IF NOT EXISTS xxx (...);
CREATE INDEX IF NOT EXISTS idx_xxx ON ...;
ALTER TABLE xxx ADD COLUMN IF NOT EXISTS yyy INT;
DROP CONSTRAINT IF EXISTS xxx_yyy;

-- ❌ 避免
CREATE TABLE xxx (...);
CREATE INDEX idx_xxx ON ...;
ALTER TABLE xxx ADD COLUMN yyy INT;
```

### 6.2 Seed 幂等性

所有 INSERT 用 `ON CONFLICT DO NOTHING` 或 `ON CONFLICT (...) DO UPDATE`：

```sql
INSERT INTO accounts (phone, ...) VALUES (...) ON CONFLICT (phone) DO NOTHING;

INSERT INTO tenant_role_menus (tenant_id, role_id, menu_id)
SELECT ... ON CONFLICT (role_id, menu_id) WHERE is_deleted = FALSE DO NOTHING;
```

### 6.3 文件头部说明

每个文件头部加说明：

```sql
-- ============================================
-- <filename> : <一句话说明>
-- ============================================
-- 目的：<这一版做什么>
-- 依赖：<依赖于哪些前置迁移>
-- 兼容性：<是否幂等 / 是否破坏性 / 是否需要迁移期>
-- ============================================
```

## 7. 排错速查

| 现象 | 排查路径 |
|---|---|
| 启动 panic: `migrations failed` | 看具体哪个文件 SQL 出错；`git log migrations/<file>.sql` 看是不是被改坏 |
| `relation "xxx" already exists` | 文件名改了导致重跑；恢复旧文件名或 ALTER 改成 `IF NOT EXISTS` |
| `column "xxx" already exists` | 同上，ALTER 加 `IF NOT EXISTS` |
| RLS 拒绝迁移 | 迁移本身已经在事务里 `SET LOCAL row_security = off`，不该出现；可能事务被嵌套 |
| `permission denied for table _schema_migrations` | 数据库账号权限不足；用 superuser 跑迁移 |
| 启动后 `tenant_users` 等表找不到 | 跑过 dev 但 init_schema.sql 没全跑——删库重跑或手动补 `psql -f init_schema.sql` |
| `init_schema 校验失败：遗留旧表 users, roles, ...` | dev 库是从 0023.3 之前迁过来的；drop 库后用新 init_schema.sql 重跑 |
