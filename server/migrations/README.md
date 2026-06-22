# 数据库迁移

> XinFramework 的 PostgreSQL schema 通过 `framework/pkg/migrate` 包的 `migrate.Run(pool, "migrations")` 在启动时按文件名升序自动应用。本目录放所有 DDL + 初始 seed。

## 1. 迁移机制

### 1.1 运行时机

```
framework.Serve(cfg, app, rt, modules)
  └─ migrate.Run(app.DB, "migrations")         // 启动时跑（cmd/xin/main.go 之前已经写过）
       ├─ 扫 ./migrations/*.sql，按文件名升序
       ├─ 跳过 _schema_migrations 表里已记录的版本
       └─ 在事务里跑未应用的版本（事务开头 SET LOCAL row_security = off）
```

### 1.2 版本跟踪

`_schema_migrations` 表记录已应用的迁移：

```sql
CREATE TABLE IF NOT EXISTS _schema_migrations (
    version    VARCHAR(255) PRIMARY KEY,    -- 文件名（如 framework.sql）
    applied_at TIMESTAMPTZ DEFAULT NOW()
);
```

`migrate.Run` 用文件名做主键。**修改已部署的文件名 = 重跑迁移**（除非要承担风险）。

### 1.3 事务保证

每个 SQL 文件在 `db.RunInTx` 里执行，开头关 RLS：

```go
tx.Exec("SET LOCAL row_security = off")  // 关闭 RLS，让迁移本身能改任何表
tx.Exec(sql)                              // 跑该文件的全部 SQL
```

任何 DDL 失败 = 整个文件回滚 = `_schema_migrations` 不记录。

## 2. 命名约定

### 2.1 当前规则（按模块命名）

| 文件 | 职责 | 所属模块 |
|---|---|---|
| `framework.sql` | 框架核心（22 张表 + RLS + seed） | framework |
| `tenant.sql` | 租户管理 | apps/boot/tenant |
| `asset.sql` | 附件 | apps/reference/asset |
| `config.sql` | 配置中心 | apps/reference/config |
| `config_alignment.sql` | 配置中心约束对齐（ALTER，非新表） | apps/reference/config |
| `dict.sql` | 字典 | apps/reference/dict |
| `flag.sql` | 头像框 + 头像 | apps/flag |
| `cms.sql` | CMS（示例业务） | apps/cms |
| `framework_account_roles_independent.sql` | 平台账号独立（account_roles CHECK 约束 + users 索引重设 + admin seed 清理）| framework |

### 2.2 推荐：增量迁移用日期前缀（未来）

新文件建议用：

```
YYYY_MM_DD_NNN_<short-description>.sql
```

例：

```
2026_06_22_001_account_roles_constraint.sql
2026_07_05_001_add_xxx_index.sql
```

- `YYYY_MM_DD`：日期
- `NNN`：当天序号（001, 002, ...）
- `<short-description>`：说明改了什么（kebab-case）

**为什么用日期前缀**：

- 字母序不稳定：`config.sql` < `config_alignment.sql` 靠的是 `.` < `_`，但 `config_v2.sql` < `config_alignment.sql` 又不行——日期前缀永远稳定
- 已部署环境的 `_schema_migrations` 表是文件名主键——日期前缀一目了然"什么时候加的"

老文件（无日期前缀）保持不动，避免破坏已部署环境的迁移记录。

### 2.3 不该做的事

- ❌ 改已部署文件的文件名（除非接受重跑风险）
- ❌ 改已部署文件的内容（除非变更幂等，如 `ADD COLUMN IF NOT EXISTS`）
- ❌ 在新文件里 ALTER 已存在于老文件的表（除非有依赖顺序约束）

## 3. 编写规范

### 3.1 DDL 模板

每个文件头部加说明：

```sql
-- ============================================
-- <文件名>: <一句话说明>
-- ============================================
-- 目的：<这一版做什么>
-- 依赖：<依赖于哪些前置迁移>
-- 兼容性：<是否幂等 / 是否破坏性 / 是否需要迁移期>
-- ============================================
```

### 3.2 幂等性

所有 DDL 必须幂等（迁移可能重跑）：

```sql
-- ✅ 推荐
CREATE TABLE IF NOT EXISTS xxx (...);
CREATE INDEX IF NOT EXISTS idx_xxx ON ...;
ALTER TABLE xxx ADD COLUMN IF NOT EXISTS yyy INT;
DROP CONSTRAINT IF EXISTS xxx_yyy;

-- ❌ 避免
CREATE TABLE xxx (...);                  -- 没有 IF NOT EXISTS
CREATE INDEX idx_xxx ON ...;              -- 没有 IF NOT EXISTS
ALTER TABLE xxx ADD COLUMN yyy INT;       -- 没有 IF NOT EXISTS
```

### 3.3 DDL 与 seed 分离（推荐）

- **DDL 文件**：只放 `CREATE TABLE / INDEX / POLICY / ...`
- **Seed 文件**：只放 `INSERT ...`（初始数据）
- 优点：DDL 失败时 seed 不会被部分应用

当前 `framework.sql` 还在混用（DDL + seed 一锅炖），属于历史遗留——不强制拆但**新增文件**建议分开。

### 3.4 软删除约定

所有业务表必须有：

```sql
is_deleted BOOLEAN DEFAULT FALSE

-- 所有索引加 WHERE 谓词
CREATE INDEX idx_xxx ON yyy (tenant_id) WHERE is_deleted = FALSE;
```

## 4. 新增迁移步骤

1. **命名**：用日期前缀（参见 §2.2）
2. **写 SQL**：在 `migrations/` 新建文件，内容遵循 §3 规范
3. **本地验证**：
   ```bash
   # 跑迁移（启动时自动跑）
   go run ./cmd/xin run &
   sleep 3
   curl http://localhost:8087/api/v1/health
   
   # 检查 _schema_migrations
   psql -h localhost -U xin -d xin_dev -c "SELECT * FROM _schema_migrations ORDER BY version;"
   ```
4. **看大小**：单个迁移文件 ≤ 5KB 是好的；> 10KB 考虑拆分
5. **不要 commit 数据库**：`migrations/` 是唯一 schema 真相源

## 5. 全量部署（首次安装 / CI 重建）

跑 `server/scripts/schema.sql`（如果存在）作为单一入口。否则按字母序跑 `migrations/*.sql`。

> 注：`server/scripts/schema.sql` 还未生成（TODO），生成时注意：
> - 输出文件**不放**在 `migrations/` 里（避免被 migrate.Run 当作未应用版本）
> - 在文件头部加注释："auto-generated, do not edit"

## 6. 排错速查

| 现象 | 排查路径 |
|---|---|
| 启动 panic: `migrations failed` | 看具体哪个文件 SQL 出错；`git log migrations/<file>.sql` 看是不是被改坏 |
| `relation "xxx" already exists` | 文件名改了导致重跑；恢复旧文件名或 ALTER 改成 `IF NOT EXISTS` |
| `column "xxx" already exists` | 同上，ALTER 加 `IF NOT EXISTS` |
| RLS 拒绝迁移 | 迁移本身已经在事务里 `SET LOCAL row_security = off`，不该出现；可能事务被嵌套 |
| `permission denied for table _schema_migrations` | 数据库账号权限不足；用 superuser 跑迁移 |