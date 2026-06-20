-- ============================================
-- 平台租户管理（apps/boot/tenant）seed
-- 后端 CRUD 已存在 (apps/boot/tenant/)；本文件补齐
-- 菜单/资源/平台角色 seed，让 super_admin 能登录后访问 /tenants。
--
-- 关键概念（与 framework.sql 的 admin 不同）：
--   - admin (roles.id=1)        : 租户级角色，存 roles 表，tenant-scoped
--   - super_admin (account_roles.role) : 平台级角色，存 account_roles 表，cross-tenant
--   - 当前 user 想要访问 /tenants/* 必须同时满足：
--       1) 账号在 account_roles 表有 role='super_admin'
--       2) 账号绑定的租户级 admin 角色拥有 tenant:* 资源 + tenants 菜单
-- ============================================

-- ============================================
-- 1) 平台角色：给种子账号加 super_admin
-- 兼容多种可能的账号名（admin / root / superadmin）
-- 只挑平台层"主"账号（不挑所有 admin 账号）
-- ============================================
INSERT INTO account_roles (account_id, role)
SELECT a.id, 'super_admin'
FROM accounts a
WHERE a.username IN ('admin', 'root', 'superadmin')
  AND a.is_deleted = FALSE
ON CONFLICT (account_id, role) DO NOTHING;

-- ============================================
-- 2) 菜单：顶级"租户管理"菜单（不挂在 system 下）
-- id 用 100 起步，避免与 framework.sql 的 1~72 撞
-- 顶级菜单 parent_id=0，ancestors 留空
-- ============================================
INSERT INTO menus (id, tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
    OVERRIDING SYSTEM VALUE
VALUES (100, 1, 'tenants', '租户管理', '跨租户平台管理', '', '/tenants', 'Building2Icon', 0, 0, '', TRUE, TRUE)
ON CONFLICT (tenant_id, code) WHERE scope = 'tenant' AND is_deleted = FALSE DO NOTHING;

-- 序列号兜底（保持 1000 倍数租户偏移量）
SELECT setval('menus_id_seq', GREATEST(
    (SELECT COALESCE(MAX(id), 0) FROM menus),
    1000
), true);

-- ============================================
-- 3) 资源：tenant:list/get/create/update/delete/purge/status
-- 注：routes.go 已注册并使用这些 code，这里只是把它们 seed 到 resources 表
-- 这样 RequirePlatformRole + Require(tenant:xxx) 才能正常放行
-- ============================================
INSERT INTO resources (tenant_id, menu_id, code, name, action, description, sort, status)
SELECT 1,
       (SELECT id FROM menus WHERE code = 'tenants' AND tenant_id = 1 AND is_deleted = FALSE),
       r.code, r.name, r.action, r.description, r.sort, 1
FROM (VALUES
    ('tenant:list',   '查询租户',   'list',   '查询所有租户列表',  1),
    ('tenant:get',    '查看租户',   'get',    '查看单个租户详情',  2),
    ('tenant:create', '创建租户',   'create', '新建租户并触发首装', 3),
    ('tenant:update', '更新租户',   'update', '更新租户档案信息',  4),
    ('tenant:delete', '删除租户',   'delete', '软删/硬删租户',     5)
) AS r(code, name, action, description, sort)
ON CONFLICT (tenant_id, code) WHERE is_deleted = FALSE DO NOTHING;

-- 序列号兜底
SELECT setval('resources_id_seq', GREATEST(
    (SELECT COALESCE(MAX(id), 0) FROM resources),
    1000
), true);

-- ============================================
-- 4) 角色绑定：admin 角色 (roles.id=1) → tenants 菜单 + tenant:* 资源
-- 注：用户说"root 是 super_admin"——但 seed 里没有 root 账号。
-- 这里默认给 admin 账号（已 seed 的）绑 tenant 权限。
-- 实际登录的 super_admin 应该是这个 admin 账号。
-- ============================================

-- 4a) admin 角色绑定 tenants 菜单
INSERT INTO role_menus (tenant_id, role_id, menu_id)
SELECT 1, 1, m.id
FROM menus m
WHERE m.code = 'tenants' AND m.tenant_id = 1 AND m.is_deleted = FALSE
ON CONFLICT DO NOTHING;

-- 4b) admin 角色绑定 tenant:* 资源
INSERT INTO role_resources (tenant_id, role_id, resource_id, effect)
SELECT 1, 1, r.id, 1
FROM resources r
WHERE r.tenant_id = 1
  AND r.code IN ('tenant:list', 'tenant:get', 'tenant:create', 'tenant:update', 'tenant:delete')
  AND r.is_deleted = FALSE
ON CONFLICT DO NOTHING;

-- ============================================
-- 5) 平台配置菜单（顶级，与 tenants 平级）
-- bootstrap 租户 (tenant_id=1) 的 config menu（id=101 已被 framework.sql 占用）
-- icon=SlidersHorizontalIcon（与 system 的 SettingsIcon 区分）
-- admin 角色通过 * 通配资源 + role_menus 自动有权限（不需要单独 bind config:* 资源）
-- ============================================
INSERT INTO menus (id, tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
    OVERRIDING SYSTEM VALUE
VALUES (102, 1, 'config', '配置管理', '系统配置项管理', '', '/settings', 'SlidersHorizontalIcon', 1, 0, '', TRUE, TRUE)
ON CONFLICT (tenant_id, code) WHERE scope = 'tenant' AND is_deleted = FALSE DO NOTHING;

-- 5a) admin 角色绑定 config 菜单（确保侧边栏可见）
INSERT INTO role_menus (tenant_id, role_id, menu_id)
SELECT 1, 1, m.id
FROM menus m
WHERE m.code = 'config' AND m.tenant_id = 1 AND m.is_deleted = FALSE
ON CONFLICT DO NOTHING;
