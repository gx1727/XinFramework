# Phase 0023 — 平台域 / 租户域 数据分域（已完成）

> 目标：把原来"共用 + scope 字段"混居的 RBAC 表，物理拆成平台域（`sys_*`）和租户域（`tenant_*`）两套独立 schema，分别由 `RunInPlatformTx` / `RunInTenantTx` 双套事务上下文治理。
>
> **状态：全部完成（已落地）**

---

## 1. 背景与现状

### 1.1 已有的"分域骨架"

- `framework/pkg/db/db.go` 已经实现 `RunInPlatformTx`（设 `app.bypass_rls=on`）和 `RunInTenantTx`（设 `app.tenant_id`）双套事务上下文
- 14 张租户域表已经启用 RLS
- `accounts` 作为"全局账号"独立存在
- `apps/platform/{tenants,sys_menu,sys_user,sys_role,sys_permission}` 已是平台域包

### 1.2 解决了的"混居"问题

| 层 | 混什么 | 终态 |
|---|---|---|
| **表命名** | `menus / resources / roles / organizations / users` 同时承载 platform 和 tenant 域，靠 `scope` 字段区分 | 物理拆为 `tenant_*` 和 `sys_*` 两套表，9 张旧表已全部重命名或 drop |
| **表设计** | 平台域没有独立表 | 8 张 `sys_*` 表已落地（无 tenant_id，无 RLS） |
| **Go 包** | 租户域用 `apps/rbac/*` | 已 rename 为 `apps/tenant/*` |
| **特殊租户** | `tenant_id=1` 的 `bootstrap` 既是 admin 居住地、又是新租户克隆源 | admin 居住身份已迁到 `sys_users` |

### 1.3 关键设计决策

> "`sys_users` 应该和 `tenant_users` 对齐"

- `sys_users` 字段对齐 `tenant_users`（org_id / code / real_name / nickname / avatar / status / created_* / updated_*）
- 唯一差异：`sys_users` 没有 `tenant_id`
- 三层抽象：
  - `accounts` — 全局登录凭证（不分域，跨域共享）
  - `sys_*` — 平台域身份实体（无 tenant_id，不启用 RLS）
  - `tenant_*` — 租户域身份实体（带 tenant_id，启用 RLS）

---

## 2. 目标架构（已落地）

### 2.1 平台域（不启用 RLS）

| 表 | 来源 | 说明 |
|---|---|---|
| `sys_users` | 新建（对齐 tenant_users） | 平台管理员身份 |
| `sys_orgs` | 新建（对齐 organizations） | 平台组织 |
| `sys_roles` | 新建（对齐 roles） | 平台角色 |
| `sys_menus` | 新建（从 menus 抽出 scope='platform'） | 平台菜单 |
| `sys_permissions` | 新建（从 resources 抽出 scope='platform'） | 平台权限码 |
| `sys_user_roles` | 新建（替代 `account_roles`） | 平台用户-角色关联 |
| `sys_role_menus` | 新建（对齐 role_menus） | 平台角色-菜单 |
| `sys_role_permissions` | 新建（对齐 role_resources） | 平台角色-权限码 |

### 2.2 租户域（启用 RLS）

| 表（终态） | 原名 | RLS |
|---|---|---|
| `tenant_users` | `users` | ✅ |
| `tenant_roles` | `roles` | ✅ |
| `tenant_user_roles` | `user_roles` | ✅ |
| `tenant_role_menus` | `role_menus` | ✅ |
| `tenant_role_resources` | `role_resources` | ✅ |
| `tenant_menus` | `menus WHERE scope='tenant'` | ✅ |
| `tenant_permissions` | `resources` | ✅ |
| `tenant_organizations` | `organizations` | ✅ |
| `tenant_role_data_scopes` | `role_data_scopes` | ✅ |
| `tenant_user_seq` | `tenant_user_seq` | ✅ |

### 2.3 不动的表

| 表 | 归属 | 说明 |
|---|---|---|
| `tenants` | platform | 不需 RLS |
| `dicts / dict_items` | 混合 | `tenant_id=0` 是 platform，`>0` 是 tenant，已有 RLS |
| `config_categories / config_items` | 混合 | 同上 |
| `auth_sessions` | platform | 全局 session（account_id） |
| `subscriptions / plans / usage_records / db_logs / routes` | tenant | 保留 |

---

## 3. Go 抽象层级（已落地）

### 3.1 三层结构

```
framework/pkg/identity          ← 跨域基类（User/Role/Menu/Permission）
   ↑ embedded by
framework/pkg/tenant/auth       ← 租户域 contracts
   ↑ used by
apps/platform/sys_*             ← 平台域模块（5 个）
apps/tenant/*                   ← 租户域模块（6 个）
```

### 3.2 平台域包对照

| 现状 | 终态 | 状态 |
|---|---|---|
| `apps/platform/sys_menu` | - | ✅ 已落地 |
| `apps/platform/tenants` | - | ✅ 已落地（管的是 tenants 表） |
| `apps/platform/sys_user` | - | ✅ 已落地 |
| `apps/platform/sys_role` | - | ✅ 已落地 |
| `apps/platform/sys_permission` | - | ✅ 已落地 |
| `apps/platform/sys_org` | - | ✅ 已落地（仅 model/types/errors/repository，不挂路由） |

### 3.3 抽象基类（`framework/pkg/identity/identity.go`）

```go
type User struct {
    ID, AccountID uint
    OrgID *uint
    Code, RealName, Nickname, Avatar string
    Status int8
    CreatedAt, UpdatedAt time.Time
}
```

平台域和租户域分别嵌入此基类。

---

## 4. 实施阶段回顾

### Phase 0023.0 — 骨架落地 ✅

| 任务 | 产出 |
|---|---|
| 0.1 SQL 拆分 | `migrations/platform_split.sql` 8 张表 DDL |
| 0.2 跨域基类 | `framework/pkg/identity/identity.go` |
| 0.3 平台域 contracts | `framework/pkg/platformauth/platformauth.go` |
| 0.4 平台域模块骨架 | `apps/platform/sys_{user,role,menu,permission,org}/` 35 个 Go 文件 |

### Phase 0023.1 — 数据迁移 + 双写 ✅

| 任务 | 产出 |
|---|---|
| 1.1-1.5 数据迁移 SQL | seed sys_* 表 |
| 1.6 双写 trigger | 已实现 |
| 1.7 RLS 补到 `tenant_menus / tenant_permissions` | 已实现 |

### Phase 0023.2 — Go 平台域包接入 ✅

| 任务 | 产出 |
|---|---|
| 2.1 `apps/boot/auth` 读 `sys_users` | 登录路径已切换 |
| 2.2 `RequirePlatformRole` 查 `sys_user_roles` | 中间件已切换 |
| 2.3 `account_roles` 表保留 + drop | 已 drop |
| 2.4 权限包迁入 | 已完成 |

### Phase 0023.3 — 租户域包重命名 ✅

| 任务 | 产出 |
|---|---|
| 3.1 `apps/rbac/*` → `apps/tenant/*` | 6 个子包已迁移 |
| 3.2 `framework/pkg/rbac` → `framework/pkg/tenant/auth` | 已迁移 |
| 3.3 `identity.User` 嵌入 | 已完成 |
| 3.4-3.6 表重命名 | 7 张表已 rename |
| 3.7 `resources` → `tenant_permissions` | 已完成 |

### Phase 0023.4 — 业务层切流 ✅

| 任务 | 产出 |
|---|---|
| 4.1 first_install 改读 `sys_*` | 已切换 |
| 4.2 `apps/platform/menu` 标记弃用 | 已迁到 `sys_menu` |
| 4.3 关闭双写 trigger | 已关闭 |
| 4.4-4.5 清理旧表数据 | 已完成 |

### Phase 0023.5 — 清理 ✅

| 任务 | 产出 |
|---|---|
| 5.1 drop `account_roles` 表 | ✅ |
| 5.2 删旧包 | ✅ |
| 5.3 文档更新 | ✅ |

---

## 5. 关键不变量

> 任何后续修改如果违反这些约束，code review 必须拒掉。

1. **平台域 sys_* 不启用 RLS**——靠 API 层 `RequirePlatformRole(super_admin)` 守卫
2. **租户域 tenant_* 启用 RLS**——所有 SQL 走 `db.RunInTenantTx`
3. **`accounts` 全局共享**——sys_users 和 tenant_users 都通过 `account_id` 外键引用
4. **`super_admin` 语义不变**——存在 `sys_user_roles` 中，role.code = 'super_admin'
5. **`sys_users` 字段对齐 `tenant_users`**——任何字段增减都同步两表

---

## 6. 风险缓解记录

| 风险 | 缓解措施 | 结果 |
|---|---|---|
| 平台/租户域字段发散 | `identity.User` 基类强约束 | ✅ 无发散 |
| 双写 trigger 漏数据 | trigger 在事务内 + checksum 校验 | ✅ 数据一致 |
| `super_admin` 鉴权漏改导致越权 | Phase 0023.2 双读 1 周 | ✅ 无越权 |
| 大表重命名锁表 | 分批 + 维护窗口 | ✅ 完成 |
| 兼容层变"鬼代码" | Phase 0023.5 必做 | ✅ 已清理 |

---

## 7. 已落地文件清单

| 路径 | 状态 |
|---|---|
| `migrations/init_schema.sql`（含 34 张表） | ✅ |
| `migrations/init_seed.sql` | ✅ |
| `framework/pkg/identity/identity.go` | ✅ |
| `framework/pkg/tenant/auth/`（原 rbac） | ✅ |
| `apps/platform/sys_user/` | ✅ |
| `apps/platform/sys_role/` | ✅ |
| `apps/platform/sys_menu/` | ✅ |
| `apps/platform/sys_permission/` | ✅ |
| `apps/platform/sys_org/` | ✅ |
| `apps/tenant/{user,role,menu,organization,permission,resource}/` | ✅ |
| `apps/platform/tenants/`（含 first_install.go） | ✅ |

---

## 8. 错误码段位（终态）

| 段位 | 模块 |
|---|---|
| 15000-15099 | sys_menu |
| 15100-15199 | sys_user |
| 15200-15299 | sys_role |
| 15300-15399 | 预留 |
| 15400-15499 | sys_permission |
| 15500-15599 | sys_org |

---

## 9. 相关文档

- `doc/architecture.md` — 总架构（含 0023 阶段说明）
- `doc/database.md` — RLS 设计、表结构
- `doc/modules.md` — 19 个 module 清单
- `migrations/README.md` — 迁移规范
