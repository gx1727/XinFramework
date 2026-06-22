# Phase 0023 — 平台域 / 租户域 数据分域

> 目标：把现在"共用 + scope 字段"混居的 RBAC 表，物理拆成平台域（`sys_*`）和租户域（`tenant_*`）两套独立 schema，分别由 `RunInPlatformTx` / `RunInTenantTx` 双套事务上下文治理。

---

## 1. 背景与现状

### 1.1 已有的"分域骨架"（比想象的好）

- `framework/pkg/db/db.go` 已经实现 `RunInPlatformTx`（设 `app.bypass_rls=on`）和 `RunInTenantTx`（设 `app.tenant_id`）双套事务上下文
- 8 张表已经启用 RLS（`users / roles / role_data_scopes / user_roles / organizations / tenant_user_seq / role_menus / role_resources / dicts / dict_items`）
- `accounts + account_roles` 已经作为"全局账号"独立存在（`framework_account_roles_independent.sql`）
- `apps/platform/{tenant,menu}` 已经是 platform 域包

### 1.2 真正的"混居"问题

| 层 | 混什么 | 证据 |
|---|---|---|
| **表命名** | `menus / resources / roles / organizations / users` 同时承载 platform 和 tenant 域，靠 `scope` 字段或 `tenant_id=1` 区分 | `framework.sql` L160 `menus.scope` 字段、L185/187 两个唯一索引 |
| **表设计** | 平台域的 `sys_user / sys_role / sys_menu / sys_org / sys_permission` **完全没有独立表** | `framework/pkg/rbac/*.go` 全部带 `TenantID uint` |
| **Go 包** | 租户域用 `apps/rbac/*`（rbac 是模式名，不是域） | `apps/rbac/{user,role,menu,organization,permission,resource}/` 6 个子包 |
| **特殊租户** | `tenant_id=1` 的 `bootstrap` 既是 admin 居住地、又是新租户克隆源 | `framework.sql` L487、tenant.sql 全段 |

### 1.3 关键设计决策

> "`sys_users` 应该和 `tenant_users` 对齐"（用户原话）

- `sys_users` 字段对齐 `tenant_users`（org_id / code / real_name / nickname / avatar / status / created_* / updated_*）
- 唯一差异：`sys_users` 没有 `tenant_id`
- 三层抽象：
  - `accounts` — 全局登录凭证（不分域，跨域共享）
  - `sys_*` — 平台域身份实体（无 tenant_id，不启用 RLS）
  - `tenant_*` — 租户域身份实体（带 tenant_id，启用 RLS）

---

## 2. 目标架构

### 2.1 平台域（不启用 RLS）

| 表 | 来源 | 说明 |
|---|---|---|
| `sys_users` | 新建（对齐 tenant_users） | 平台管理员身份 |
| `sys_orgs` | 新建（对齐 organizations） | 平台组织（先建表，业务后补） |
| `sys_roles` | 新建（对齐 roles） | 平台角色 |
| `sys_menus` | 新建（对齐 menus, scope='platform' 抽出） | 平台菜单 |
| `sys_permissions` | 新建（对齐 resources, scope='platform' 抽出） | 平台权限码 |
| `sys_user_roles` | 新建（替代 `account_roles`） | 平台用户-角色关联 |
| `sys_role_menus` | 新建（对齐 role_menus） | 平台角色-菜单 |
| `sys_role_permissions` | 新建（对齐 role_resources） | 平台角色-权限码 |

### 2.2 租户域（启用 RLS，结构基本不变）

| 表（终态） | 当前 | RLS |
|---|---|---|
| `tenant_users` | `users` | ✅ |
| `tenant_roles` | `roles` | ✅ |
| `tenant_user_roles` | `user_roles` | ✅ |
| `tenant_role_menus` | `role_menus` | ✅ |
| `tenant_role_resources` | `role_resources` | ✅ |
| `tenant_menus` | `menus WHERE scope='tenant'` | ❌ 待加 |
| `tenant_resources` | `resources` | ❌ 待加 |
| `tenant_organizations` | `organizations` | ✅ |
| `tenant_role_data_scopes` | `role_data_scopes` | ✅ |
| `tenant_user_seq` | `tenant_user_seq` | ✅ |

> Phase 0023.0 不重命名租户域表（避免破坏现有代码）；重命名推迟到 0023.3。

### 2.3 不动的表

| 表 | 归属 | 说明 |
|---|---|---|
| `tenants` | platform | 不需 RLS |
| `dicts / dict_items` | 混合 | `tenant_id=0` 是 platform，`>0` 是 tenant，已有 RLS |
| `config_categories / config_items` | 混合 | 同上 |
| `auth_sessions` | platform | 全局 session（account_id） |
| `subscriptions / plans / usage_records / db_logs / routes` | tenant | 保留 |

---

## 3. Go 抽象层级

### 3.1 三层结构

```
framework/pkg/identity          ← 跨域基类（User/Org/Role/Menu/Permission）
   ↑ embedded by
framework/pkg/platformauth      ← 平台域 contracts（包 UserRepository 等）
framework/pkg/tenant/auth       ← 租户域 contracts（[Phase 0023.3 引入]）
   ↑ used by
apps/platform/sys_*             ← 平台域模块（5 个）
apps/tenant/*   [Phase 0023.3]  ← 租户域模块（取代 apps/rbac/*）
```

### 3.2 平台域包对照

| 现状 | 终态 | 状态 |
|---|---|---|
| `apps/platform/menu` | `apps/platform/sys_menu` | Phase 0023.0 全部新建 |
| `apps/platform/tenant` | `apps/platform/sys_tenant` | **保留**（管的是 tenants 表，与 sys_* 不同概念） |
| `apps/platform/sys_user` | - | Phase 0023.0 新建 |
| `apps/platform/sys_role` | - | Phase 0023.0 新建 |
| `apps/platform/sys_permission` | - | Phase 0023.0 新建 |
| `apps/platform/sys_org` | - | Phase 0023.0 新建（仅 model/types/errors/repository，不挂路由） |

> `apps/platform/menu` 和 `apps/platform/sys_menu` 是**两个并存**的模块，分别对应 `menus WHERE scope='platform'` 和 `sys_menus` 表。Phase 0023.4 切换后 `apps/platform/menu` 弃用。

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

平台域 `platformauth.User` 嵌入 `identity.User`，**没有** `TenantID` 字段。
租户域 `tenant.User` 嵌入 `identity.User` + 加 `TenantID uint`。

---

## 4. 实施分阶段

### Phase 0023.0 — 骨架落地（**当前阶段**）

| 任务 | 产出 | 风险 |
|---|---|---|
| 0.1 SQL 拆分 | `migrations/platform_split.sql` 8 张表 DDL | 低（只 CREATE，不动旧表） |
| 0.2 跨域基类 | `framework/pkg/identity/identity.go` | 低 |
| 0.3 平台域 contracts | `framework/pkg/platformauth/platformauth.go` | 低 |
| 0.4 平台域模块骨架 | `apps/platform/sys_{user,role,menu,permission,org}/` 35 个 Go 文件 | 低（**新代码，未挂路由不影响生产**） |

**关键不变量**：
- sys_* 表**不启用 RLS**——platform 是单租户概念
- 所有 sys_* DB 操作**强制**走 `db.RunInPlatformTx`
- sys_orgs Phase 0023.0 不挂路由（YAGNI）

### Phase 0023.1 — 数据迁移 + 双写（1-2 周）

| 任务 | 产出 | 风险 |
|---|---|---|
| 1.1 seed `sys_users` ← `accounts` | 数据迁移 SQL | 中 |
| 1.2 seed `sys_roles` ← `account_roles` | 数据迁移 SQL | 中 |
| 1.3 seed `sys_menus` ← `menus WHERE scope='platform'` | 数据迁移 SQL | 中 |
| 1.4 seed `sys_permissions` ← `resources WHERE scope='platform'` | 数据迁移 SQL | 中 |
| 1.5 seed `sys_user_roles / sys_role_menus / sys_role_permissions` | 数据迁移 SQL | 中 |
| 1.6 双写 trigger（旧表 → 新表） | trigger SQL | **高**：失败要事务回滚 |
| 1.7 RLS 补到 `tenant_menus / tenant_resources` | 2 个 policy | 中 |
| 1.8 完整性校验（`pg_policies` 检查 8 张表都有 RLS） | 校验 SQL | 低 |

### Phase 0023.2 — Go 平台域包接入（1-2 周）

| 任务 | 产出 | 风险 |
|---|---|---|
| 2.1 `apps/boot/auth` 读 `sys_users` 替代 `accounts` | 代码 | **高**：登录路径 |
| 2.2 `RequirePlatformRole` 查 `sys_user_roles.sys_roles.code` | 中间件 | **高**：鉴权是安全边界 |
| 2.3 `account_roles` 表保留 1 周做兼容读，然后 drop | 代码 + SQL | 中 |
| 2.4 `framework/pkg/permission/platform_role.go` 迁入 `framework/pkg/platformauth/` | rename | 低 |
| 2.5 dev/staging 跑 2 周回归 | 验证 | — |

### Phase 0023.3 — 租户域包重命名（2-3 周）

| 任务 | 产出 | 风险 |
|---|---|---|
| 3.1 `apps/rbac/*` → `apps/tenant/*`（6 个子包） | 包重构 | **中**：100+ import 路径 |
| 3.2 `framework/pkg/rbac` → `framework/pkg/tenant/auth` | 包重构 | 中 |
| 3.3 `apps/tenant/auth.User` 等用 `identity.User` 嵌入 | 代码 | 中 |
| 3.4 旧表 `users → tenant_users`（重命名 + 索引重建） | SQL | **高**：大表重命名 |
| 3.5 旧表 `roles / organizations / user_roles / role_menus / role_resources / role_data_scopes` 同步重命名 | SQL | **高** |
| 3.6 `resources` → `tenant_permissions`（同时加 RLS） | SQL | **高** |

### Phase 0023.4 — 业务层切流（1-2 周）

| 任务 | 产出 | 风险 |
|---|---|---|
| 4.1 `apps/boot/tenant` first_install 改读 `sys_*` | 代码 | 中 |
| 4.2 `apps/platform/menu` 标记弃用，迁到 `apps/platform/sys_menu` | 代码 | **高**：路由变更 |
| 4.3 关闭 Phase 0023.1 的双写 trigger | trigger drop | **高** |
| 4.4 旧 `menus WHERE scope='platform'` 行清空，`menus.scope` 字段 drop | SQL | **高** |
| 4.5 旧 `resources` scope='platform' 行清空 | SQL | 中 |

### Phase 0023.5 — 清理（3-5 天）

| 任务 | 产出 | 风险 |
|---|---|---|
| 5.1 drop `account_roles` 表 | SQL | 低 |
| 5.2 `framework/pkg/permission/platform_role.go` 删除 | rm | 低 |
| 5.3 `framework/pkg/rbac` 目录删除（如未在 0023.3 完成） | rm | 低 |
| 5.4 文档更新：`doc/architecture.md` Phase 表补 0023 | doc | 低 |

---

## 5. 关键不变量

> 任何后续修改如果违反这些约束，code review 必须拒掉。

1. **平台域 sys_* 不启用 RLS**——靠 API 层 `RequirePlatformRole(super_admin)` 守卫
2. **租户域 tenant_* 启用 RLS**——所有 SQL 走 `db.RunInTenantTx`
3. **`accounts` 全局共享**——sys_users 和 tenant_users 都通过 `account_id` 外键引用
4. **`super_admin` 语义不变**——存在 `sys_user_roles` 中，role.code = 'super_admin'
5. **`sys_users` 字段对齐 `tenant_users`**——任何字段增减都同步两表

---

## 6. 风险矩阵

| 风险 | 影响 | 概率 | 缓解 |
|---|---|---|---|
| 平台/租户域字段发散 | 高 | 中 | `identity.User/Role/...` 基类强约束 |
| 双写 trigger 漏数据 | 高 | 中 | trigger 在事务内 + checksum 校验脚本 |
| `super_admin` 鉴权漏改导致越权 | **极高** | 中 | Phase 0023.2 双读 1 周；audit_log 监控 |
| 大表重命名（users → tenant_users）锁表 | 高 | 中 | 分批 + 维护窗口 |
| 兼容层忘清理变成"鬼代码" | 中 | 高 | Phase 0023.5 必做 |
| `bootstrap` 租户的 admin 要搬到 `sys_users` 还是双轨 | 中 | 高 | 决策见 §7 |

---

## 7. 待拍板问题

1. **bootstrap 租户的 admin 用户**：要搬到 `sys_users` 还是双轨制？
2. **租户域包重命名时机**：现在就做（推荐）还是等 0023.1 稳了再做？
3. **`sys_role_data_scopes` 建不建**？（YAGNI：先不建）
4. **`audit_log / db_logs` 要不要拆 platform/tenant**？（YAGNI：先不分）
5. **双写 trigger 保留多久**？（建议至少 2 周 dev/staging 观察期）

---

## 8. 已落地文件清单（Phase 0023.0）

| 路径 | 状态 |
|---|---|
| `migrations/platform_split.sql` | ✅ |
| `framework/pkg/identity/identity.go` | ✅ |
| `framework/pkg/platformauth/platformauth.go` | ✅ |
| `apps/platform/sys_user/{model,types,errors,repository,service,handler,routes,module}.go` | ✅ 8 |
| `apps/platform/sys_role/{model,types,errors,repository,service,handler,routes,module,itoa}.go` | ✅ 9 |
| `apps/platform/sys_menu/{model,types,errors,repository,service,handler,routes,module}.go` | ✅ 7 |
| `apps/platform/sys_permission/{model,types,errors,repository,service,handler,routes,module}.go` | ✅ 7 |
| `apps/platform/sys_org/{model,types,errors,repository}.go` | ✅ 4（**不挂路由**） |

合计 **35 个文件**。

---

## 9. 错误码段位

| 段位 | 模块 |
|---|---|
| 15000-15099 | platform_menu（旧） |
| 15100-15199 | sys_user |
| 15200-15299 | sys_role |
| 15300-15399 | sys_menu |
| 15400-15499 | sys_permission |
| 15500-15599 | sys_org |

---

## 10. 相关文档

- `doc/architecture.md` — 总架构（0022b Phase C 之后是 0023）
- `doc/database.md` — RLS 设计（待补 tenant_menus/tenant_resources 段位）
- `doc/modules.md` — 16 个 module 清单（0023+ 会变 21 个）
- `migrations/README.md` — 迁移规范
