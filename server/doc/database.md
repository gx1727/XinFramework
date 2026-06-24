# Database Design

> Phase 0023 final state: **32 tables**. Core tables in `migrations/init_schema.sql`; seed data in `migrations/init_seed.sql`; business modules (asset / cms / flag) keep their own .sql.
>
> Doc version: 2026-06-23 (Phase 0023 platform/tenant domain split complete)

## 1. Extensions

Migration scripts install these PG extensions by default:
```sql
CREATE EXTENSION IF NOT EXISTS ltree;      -- path / tree storage
CREATE EXTENSION IF NOT EXISTS pg_trgm;    -- trigram fuzzy match
```

Requires PG >= 14. Extensions need superuser permission once.

## 2. Migration Mechanism

At startup, `framework/pkg/migrate.Run("migrations")`:
- Scans `./migrations/*.sql` in filename order
- Records executed versions in `_schema_migrations` table
- Skips already-executed files

**Note**: migrations are **idempotent** (all use `CREATE TABLE IF NOT EXISTS` / `CREATE INDEX IF NOT EXISTS` / `DROP POLICY IF EXISTS` + `CREATE POLICY`).
```bash
ls migrations/
# asset.sql  cms.sql  flag.sql  init_schema.sql  init_seed.sql
```

### 2.1 File Layering

Dev phase: all DDL centralized in `init_schema.sql`, all seed in `init_seed.sql`.
- Schema changes are rare (stable after first cut)
- Seed changes are common (default roles, menus, dicts)
- Dev reset: schema first, then seed

**Business modules** (asset / cms / flag) keep their own .sql because they have different lifecycles.

## 3. Core Tables (Phase 0023 Final)

### 3.1 Three Data Domains

| Domain | Prefix | tenant_id | RLS | Module Location |
|---|---|---|---|---|
| **Platform** | `sys_*` | NO | NO | `apps/platform/sys_*` |
| **Tenant** | `tenant_*` | YES | YES | `apps/tenant/*` |
| **Shared** | `accounts / tenants / auth_sessions` | -- | -- | `apps/boot/auth` etc |

**Key invariants**:
- Platform SQL **must** go through `db.RunInPlatformTx` (sets `app.bypass_rls='on'`)
- Tenant SQL **must** go through `db.RunInTenantTx` (sets `app.tenant_id`)
- Login path: `accounts.id` -> `sys_users.account_id` -> `sys_user_roles.role_id` -> `sys_roles.code = 'super_admin'`
- `menus` table physically split into `tenant_menus` (no scope field, tenant only) and `sys_menus` (platform)
- `resources` -> `tenant_permissions` (with RLS); platform permissions are `sys_permissions`

### 3.2 ER Diagram (by Domain)

#### Shared Layer

```
tenants                              accounts
  |                                    |
  |                                    | account_id
  |                                    v
  |                                +-----------+-----------+
  |                                | sys_*     | tenant_*  |
  |                                | platform  | tenant    |
  |                                +-----------+-----------+
  |
auth_sessions -- account_id -- accounts
```

#### Platform Domain (no RLS, force RunInPlatformTx)

```
sys_users -- account_id -- accounts (shared)
  |
  +-- sys_user_roles -- sys_roles (incl super_admin)
       |
       +-- sys_role_menus -- sys_menus
       +-- sys_role_permissions -- sys_permissions

sys_orgs (parent_id recursive, for platform orgs)
```

#### Tenant Domain (RLS on, force RunInTenantTx)

```
tenant_users -- account_id -- accounts (shared)
  | -- tenant_id -- tenants (shared)
  |
  +-- tenant_user_roles -- tenant_roles
       |
       +-- tenant_role_data_scopes (per-dept)
       +-- tenant_role_menus -- tenant_menus
       +-- tenant_role_resources -- tenant_permissions

tenant_organizations (parent_id recursive + ancestors materialized path)
tenant_user_seq (tenant user sequence)
```

#### Business Support (per-tenant billing, RLS on)

```
subscriptions -- tenant_id -- tenants
plans (global, no RLS)
usage_records -- tenant_id
db_logs -- tenant_id (audit)
routes -- tenant_id (per-tenant routing)
```

#### Dictionary / Config (keep `tenant_id = 0` short-circuit for platform-level sharing)

```
dicts -- dict_items
  |          |
  |          +-- platform_item_id (tenant overrides platform item)
  +-- dict_visibility (platform dict visibility per tenant)

config_categories -- config_items
  |          |
  |          +-- platform_item_id (tenant overrides platform item)
  +-- config_visibility (platform config visibility per tenant)
```

### 3.3 Table List (32 tables)

#### Shared Layer (3 tables, no RLS)

| Table | Purpose | Key fields |
|---|---|---|
| `tenants` | Tenant registry | `code` (unique), `status`, `config JSONB` |
| `accounts` | Global accounts | `phone` / `email` (unique), `password` (argon2id hash) |
| `auth_sessions` | Sessions | `account_id`, `token` (unique), `expires_at` |

> **Note**: `account_auths` / `account_roles` / `user_codes` are **dropped** (Phase 0023.1).
> - 3rd party auth (wechat / oauth) not yet implemented. Weixin login uses `code2Session` to bind `account_id` directly.
> - Platform roles (super_admin) now use `sys_users + sys_user_roles + sys_roles`.

#### Platform Domain sys_* (8 tables, no RLS)

| Table | Purpose | Key fields |
|---|---|---|
| `sys_users` | Platform user identity (aligned with `tenant_users`) | `account_id` (FK accounts), `code`, `org_id` |
| `sys_orgs` | Platform org (parent_id recursive) | `parent_id`, `ancestors`, `code` (unique) |
| `sys_roles` | Platform roles | `code` (unique, e.g. `super_admin`), `data_scope` |
| `sys_menus` | Platform menus | `code` (unique), `parent_id`, `ancestors` |
| `sys_permissions` | Platform permission codes | `code` (unique), `menu_id`, `action` |
| `sys_user_roles` | Platform user-role (replaces `account_roles`) | `user_id`, `role_id` (soft delete `is_deleted`) |
| `sys_role_menus` | Platform role-menu | `role_id`, `menu_id` |
| `sys_role_permissions` | Platform role-permission | `role_id`, `permission_id`, `effect` |

#### Tenant Domain tenant_* (10 tables, RLS on)

| Table | Purpose | Key fields |
|---|---|---|
| `tenant_organizations` | Org structure | `tenant_id`, `parent_id`, `ancestors`, `code` (unique per tenant) |
| `tenant_users` | Tenant users | `tenant_id`, `account_id`, `code`, `org_id` |
| `tenant_user_roles` | Tenant user-role | `tenant_id`, `user_id`, `role_id` |
| `tenant_roles` | Tenant roles | `tenant_id`, `code` (unique per tenant), `data_scope JSONB` |
| `tenant_role_data_scopes` | Tenant role data scope | `tenant_id`, `role_id`, `org_id` |
| `tenant_role_menus` | Tenant role-menu | `tenant_id`, `role_id`, `menu_id` |
| `tenant_role_resources` | Tenant role-permission | `tenant_id`, `role_id`, `permission_id`, `effect` |
| `tenant_menus` | Tenant menus (no `scope` field) | `tenant_id`, `code` (unique per tenant), `parent_id`, `ancestors` |
| `tenant_permissions` | Tenant permission codes (was `resources`) | `tenant_id`, `code` (unique per tenant), `menu_id`, `action` |
| `tenant_user_seq` | Tenant user sequence | `tenant_id`, `seq` |

#### Business Support (4 tables + 1 global, per-tenant billing)

| Table | Purpose | RLS |
|---|---|---|
| `subscriptions` | Subscriptions | YES |
| `usage_records` | Usage records | YES |
| `db_logs` | Audit log | YES |
| `routes` | Routes (per tenant) | YES |
| `plans` | Plans (global) | NO |

#### Dictionary / Config / Visibility (6 tables, keep `tenant_id = 0` short-circuit)

| Table | Purpose | RLS |
|---|---|---|
| `dicts` | Dictionary master | YES (`tenant_id = 0` short-circuit) |
| `dict_items` | Dictionary items | YES (`tenant_id = 0` short-circuit) |
| `dict_visibility` | Dictionary visibility matrix | NO |
| `config_categories` | Config groups | YES (`tenant_id = 0` short-circuit) |
| `config_items` | Config items | YES (`tenant_id = 0` short-circuit) |
| `config_visibility` | Config visibility matrix | NO |

> Business module tables: `file_assets` (asset), `posts` (cms), `frames / spaces / avatars` etc (flag) -- see their own `migrations/*.sql`.

## 4. Row-Level Security (RLS)

**Multi-tenant isolation via `db.RunInTenantTx(ctx, pool, tenantID, fn)`**: injects `SET LOCAL app.tenant_id = <id>` into transaction, then RLS policies on the table filter rows.

### 4.1 RLS Example (`tenant_users` table)

```sql
ALTER TABLE tenant_users ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_users;
CREATE POLICY tenant_isolation_policy ON tenant_users USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);
```

`SELECT * FROM tenant_users` returns 0 rows when `app.tenant_id` is unset and `bypass_rls=off`.

### 4.2 Go: Tenant Context

```go
err := db.RunInTenantTx(ctx, s.pool, uc.TenantID, func(txCtx context.Context) error {
    q, _ := db.GetQuerier(txCtx, s.pool)
    return s.repo.GetByID(txCtx, userID)
})
```

**Platform management** uses `db.RunInPlatformTx(ctx, pool, fn)` -- sets `app.bypass_rls='on'`, bypasses RLS, can read/write across tenants.

```go
err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
    return s.sysUserRepo.Create(txCtx, su)
})
```

### 4.3 RLS Policy Templates

#### Tenant Domain (10 tables + 4 business = 14 tables)

```sql
ALTER TABLE tenant_xxx ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_xxx;
CREATE POLICY tenant_isolation_policy ON tenant_xxx USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);
```

#### Dictionary / Config (`tenant_id = 0` is platform-level, cross-tenant)

```sql
ALTER TABLE xxx ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON xxx;
CREATE POLICY tenant_isolation_policy ON xxx USING (
    tenant_id = 0
    OR tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);
```

### 4.4 Tables NOT Affected by RLS

These tables should NOT be queried with `RunInTenantTx`:

| Table | Reason |
|---|---|
| `accounts` | Global unique, login doesn't know tenant_id |
| `tenants` | Platform mgmt needs cross-tenant query |
| `auth_sessions` | By account_id, cross-domain |
| `sys_*` (8 tables) | Platform domain, no RLS by design |
| `plans` | Global plans |
| `dict_visibility` / `config_visibility` | Platform visibility matrix, super_admin managed |
| Business module tables (`file_assets` / `posts` / `flag_*`) | See their own .sql |

## 5. Soft Delete

All business tables have `is_deleted BOOLEAN DEFAULT FALSE` + `created_at` / `updated_at` / `created_by` / `updated_by`.
**Convention**:
- Queries default to `WHERE is_deleted = FALSE`
- Unique indexes are partial: `UNIQUE INDEX ... WHERE is_deleted = FALSE`
- "Delete" actually means `UPDATE ... SET is_deleted = TRUE`, data preserved
- "Hard delete" (physical `DELETE`) only for `purge` operations (e.g. `POST /api/v1/platform/tenants/:id/purge`)

```sql
CREATE UNIQUE INDEX uk_tenant_users_account_tenant
    ON tenant_users (account_id, tenant_id) WHERE is_deleted = FALSE;
```

## 6. Index Strategy

Every table has at least:

| Index | Field | Purpose |
|---|---|---|
| PK | `id` | `BIGINT GENERATED ALWAYS AS IDENTITY` |
| `created_at` | default `idx_xxx_created_at` | Sorting / incremental sync |
| `is_deleted` partial | combined with other unique | Soft delete + unique |

High-frequency query fields have dedicated `idx_*`:
```sql
CREATE INDEX idx_tenant_users_org ON tenant_users (org_id) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX uk_tenant_users_account_tenant
    ON tenant_users (account_id, tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_tenant_org_tenant ON tenant_organizations (tenant_id) WHERE is_deleted = FALSE;
```

## 7. Materialized Path

`tenant_organizations` uses ltree-style `ancestors TEXT` field for materialized path:

```
ancestors = ""                   <-- top-level
ancestors = "/3/"                <-- child of parent_id=3
ancestors = "/3/7/"              <-- child of parent_id=7, which is child of parent_id=3
```

Quick lookup of all ancestors of a node:

```sql
SELECT * FROM tenant_organizations
WHERE id = ANY(string_to_array(trim(ancestors, '/'), '/')::bigint[]);
```

Quick lookup of all descendants of a node:

```sql
SELECT * FROM tenant_organizations WHERE ancestors LIKE '/3/%';
```

## 8. Timezone

All `TIMESTAMPTZ DEFAULT NOW()` -- PostgreSQL stores UTC internally.
Production recommendations:
- DB server TZ = UTC
- App server TZ = Asia/Shanghai
- All cross-TZ logic handled in app layer

## 9. JSONB Fields (Important)

### 9.1 Current JSONB Columns (11, post Phase 0023)

| Table | Field | Purpose |
|---|---|---|
| `db_logs` | `old_data` | Audit: pre-change snapshot |
| `db_logs` | `new_data` | Audit: post-change snapshot |
| `tenants` | `config` | Tenant extended config (reserved) |
| `config_items` | `value` | Config item current value |
| `config_items` | `default_value` | Config item default value |
| `config_items` | `options` | select/radio options |
| `config_items` | `validation` | Validation rules |
| `dicts` | `extend` | Dictionary extended fields |
| `dict_items` | `extend` | Dictionary item extended fields |
| `tenant_roles` | `extend` | Role extended fields (0023+) |
| `sys_roles` | `extend` | Platform role extended fields (0023+) |

Business module JSONB fields see their own .sql (e.g. `flag_frames.template_config`).

### 9.2 pgx JSONB Write Must Use `::jsonb` Cast

pgx encodes Go types to PG by default:

| Go type | PG type (default) | Write to JSONB |
|---|---|---|
| `string` | `text` | ERROR `42804` |
| `[]byte` | `bytea` | ERROR `42804` |

**Fix**: explicit `::jsonb` cast in SQL:
```go
// WRONG: _, err := q.Exec(ctx, `UPDATE t SET value = $1 WHERE id = $2`, valueJSON, id)

// RIGHT: direct ::jsonb
_, err := q.Exec(ctx, `UPDATE t SET value = $1::jsonb WHERE id = $2`, valueJSON, id)

// RIGHT: COALESCE scenario (patch / update)
valueJSON := toJSON(req.Value)
_, err := q.Exec(ctx, `UPDATE t SET value = COALESCE($1::jsonb, value) WHERE id = $2`,
                 valueJSON, id)
```

Error example:
```
ERROR: column "value" is of type jsonb but expression is of type text (SQLSTATE 42804)
```

### 9.3 GIN Index

```sql
CREATE INDEX idx_tenants_config_gin ON tenants USING GIN (config);
```

If writing to `tenants.config` later, remember the `::jsonb` cast from section 9.2.

## 10. Migration Workflow (Dev Phase)

```bash
# 1. Edit SQL files (dev phase: re-run freely)
vi migrations/init_schema.sql      # add CREATE TABLE IF NOT EXISTS xxx
vi migrations/init_seed.sql        # add initial seed

# 2. Run in repo root (dev DB can be lost)
psql -U xin_user -d xin -f migrations/init_schema.sql
psql -U xin_user -d xin -f migrations/init_seed.sql

# 3. Commit SQL + Go entities
git add migrations/ apps/<new-module>/
git commit -m "feat(db): add xxx table"

# 4. Deploy: xin restart runs unapplied migrations automatically
```

**Important**:
- Never modify already-deployed migration scripts
- Add fields via new alignment script: `ALTER TABLE ... ADD COLUMN IF NOT EXISTS`
- Don't use `CREATE TABLE` without `IF NOT EXISTS`

## 11. Data Integrity Constraints

### 11.1 FK Relationships

```sql
ALTER TABLE tenant_users
    ADD CONSTRAINT fk_tenant_users_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_tenant_users_account FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE RESTRICT;
```

- `tenant ON DELETE CASCADE`: hard-delete tenant -> clear all tenant_users
- `account ON DELETE RESTRICT`: account can only be soft-deleted + disabled, never hard-deleted

### 11.2 Check Constraints

```sql
ALTER TABLE accounts
    ADD CONSTRAINT chk_accounts_status CHECK (status IN (0, 1));
```

## 12. Backup and Recovery

Not in framework scope, but typical practice:
```bash
pg_dump -Fc -h db.host -U xin_user xin > backup_$(date +%F).dump
pg_restore -d xin backup_2026-06-23.dump
```

Production recommendations:
- WAL archive + point-in-time recovery (`archive_mode = on`)
- Off-site replica (`streaming replication`)
- Daily full backup + continuous increment

## 13. Performance Tuning

| Table size | Recommendation |
|---|---|
| < 1M rows | No partitioning needed |
| 1M - 100M | Range partition by `tenant_id` |
| > 100M | Hash partition by `tenant_id` + periodic archive |

Hot tables (`tenant_users` / `accounts` / `tenant_permissions`) need `tenant_id + status` composite index. flag business (`avatars` / `frames`) needs `creator_id` index because `DataScopeSelf` runs `WHERE creator_id = $1` heavily.

## 14. Monitoring

Key metrics:
- `pg_stat_user_tables`: each table's `seq_scan` vs `idx_scan` ratio (`> 0.1` suggests missing index)
- `pg_stat_user_indexes`: index usage frequency (`idx_scan = 0` means dead index)
- `pg_locks`: lock waits
- `pg_stat_activity`: long transactions (`state = 'active' AND query_start < now() - interval '1 min'`)

See [PostgreSQL official docs](https://www.postgresql.org/docs/current/monitoring-stats.html) for SQL.

## 15. Phase 0023 Final Validation

`init_schema.sql` ends with an integrity check block (dev phase):

```sql
DO $$ ... RAISE EXCEPTION 'init_schema validation failed: missing table %', missing; ... $$;
```

- All 32 target tables must be created (any missing -> RAISE EXCEPTION)
- 9 legacy tables must be dropped (any leftover: `users / roles / organizations / user_roles / role_menus / role_resources / role_data_scopes / resources / account_roles` -> RAISE EXCEPTION)

After dev DB reset, `\d` shows all `tenant_*` / `sys_*` tables with no legacy names left.
