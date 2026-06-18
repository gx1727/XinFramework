-- ============================================
-- XinFramework 数据库初始化脚本
-- Framework 核心表
-- 版本: PostgreSQL 14+
-- ============================================

SET
client_encoding = 'UTF8';

-- 启用扩展
CREATE
EXTENSION IF NOT EXISTS ltree;
CREATE
EXTENSION IF NOT EXISTS pg_trgm;

-- ============================================
-- 核心表
-- ============================================

-- 1. tenants (租户表)
CREATE TABLE IF NOT EXISTS tenants
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code       VARCHAR(50)  NOT NULL,
    name       VARCHAR(100) NOT NULL,
    status     SMALLINT    DEFAULT 1,
    contact    VARCHAR(50),
    phone      VARCHAR(20),
    email      VARCHAR(100),
    province   VARCHAR(32),
    city       VARCHAR(32),
    area       VARCHAR(32),
    address    VARCHAR(255),
    config     JSONB,
    dashboard  VARCHAR(64),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by BIGINT,
    updated_by BIGINT,
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_tenants_code ON tenants (code) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_tenants_config_gin ON tenants USING GIN (config);

-- 2. accounts (全局账号表)
CREATE TABLE IF NOT EXISTS accounts
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    phone      VARCHAR(20),
    email      VARCHAR(100),
    password   VARCHAR(255),
    username   VARCHAR(64),
    real_name  VARCHAR(64),
    avatar     VARCHAR(512),
    status     SMALLINT    DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_accounts_phone ON accounts (phone) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_accounts_email ON accounts (email) WHERE is_deleted = FALSE;

-- 3. organizations (组织表)
CREATE TABLE IF NOT EXISTS organizations
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,
    parent_id  BIGINT,
    code       VARCHAR(32) NOT NULL,
    name       VARCHAR(64) NOT NULL,
    type        VARCHAR(32),
    description VARCHAR(255),
    admin_code  VARCHAR(64),
    ancestors  TEXT        DEFAULT '',
    sort       INT         DEFAULT 0,
    status     SMALLINT    DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by BIGINT,
    updated_by BIGINT,
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_org_tenant ON organizations (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_org_parent ON organizations (parent_id) WHERE is_deleted = FALSE;

-- 4. users (用户表)
CREATE TABLE IF NOT EXISTS users
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    account_id BIGINT NOT NULL,
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
CREATE INDEX IF NOT EXISTS idx_users_tenant ON users (tenant_id) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_users_account ON users (account_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_users_org ON users (org_id) WHERE is_deleted = FALSE;

-- 5. roles (角色表)
CREATE TABLE IF NOT EXISTS roles
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id   BIGINT      NOT NULL,
    org_id      BIGINT,
    code        VARCHAR(32),
    name        VARCHAR(32) NOT NULL,
    description VARCHAR(256),
    data_scope  SMALLINT    NOT NULL DEFAULT 1,
    extend      JSONB,
    is_default  BOOLEAN              DEFAULT FALSE,
    sort        INT                  DEFAULT 0,
    status      SMALLINT             DEFAULT 1,
    created_at  TIMESTAMPTZ          DEFAULT NOW(),
    updated_at  TIMESTAMPTZ          DEFAULT NOW(),
    created_by  BIGINT,
    updated_by  BIGINT,
    is_deleted  BOOLEAN              DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_roles_tenant ON roles (tenant_id) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_roles_code ON roles (tenant_id, code) WHERE is_deleted = FALSE;

-- 6. role_data_scopes (角色数据范围表)
CREATE TABLE IF NOT EXISTS role_data_scopes
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    role_id    BIGINT NOT NULL,
    org_id     BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_rds_role ON role_data_scopes (role_id) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_rds_unique ON role_data_scopes (role_id, org_id) WHERE is_deleted = FALSE;

-- 7. user_roles (用户角色关联表)
CREATE TABLE IF NOT EXISTS user_roles
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    user_id    BIGINT NOT NULL,
    role_id    BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_ur_user ON user_roles (user_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_ur_role ON user_roles (role_id) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_ur_unique ON user_roles (user_id, role_id) WHERE is_deleted = FALSE;

-- 8. menus (菜单表)
CREATE TABLE IF NOT EXISTS menus
(
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
CREATE UNIQUE INDEX IF NOT EXISTS uk_menu_code ON menus (tenant_id, code) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_menu_tenant ON menus (tenant_id) WHERE is_deleted = FALSE;

-- 9. resources (资源表)
CREATE TABLE IF NOT EXISTS resources
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id   BIGINT      NOT NULL,
    menu_id     BIGINT,
    code        VARCHAR(64) NOT NULL,
    name        VARCHAR(64) NOT NULL,
    action      VARCHAR(32) DEFAULT 'read',
    description VARCHAR(512),
    sort        INT         DEFAULT 0,
    status      SMALLINT    DEFAULT 1,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    created_by  BIGINT,
    updated_by  BIGINT,
    is_deleted  BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_resource_code ON resources (tenant_id, code) WHERE is_deleted = FALSE;

-- 10. routes (路由表)
CREATE TABLE IF NOT EXISTS routes
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,
    menu_id    BIGINT,
    code       VARCHAR(64) NOT NULL,
    name       VARCHAR(64),
    path       VARCHAR(255),
    method     VARCHAR(16),
    sort       INT         DEFAULT 0,
    status     SMALLINT    DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_routes_tenant ON routes (tenant_id) WHERE is_deleted = FALSE;

-- 14. db_logs (数据库日志表)
CREATE TABLE IF NOT EXISTS db_logs
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    user_id    BIGINT,
    action     VARCHAR(32),
    table_name VARCHAR(64),
    record_id  BIGINT,
    old_data   JSONB,
    new_data   JSONB,
    ip         VARCHAR(64),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_db_logs_tenant ON db_logs (tenant_id);
CREATE INDEX IF NOT EXISTS idx_db_logs_user ON db_logs (user_id);

-- 15. subscriptions (订阅表)
CREATE TABLE IF NOT EXISTS subscriptions
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    plan_id    BIGINT,
    status     SMALLINT    DEFAULT 1,
    start_at   TIMESTAMPTZ,
    end_at     TIMESTAMPTZ,
    auto_renew BOOLEAN     DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_subs_tenant ON subscriptions (tenant_id) WHERE is_deleted = FALSE;

-- 16. plans (套餐表)
CREATE TABLE IF NOT EXISTS plans
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        VARCHAR(64) NOT NULL,
    code        VARCHAR(32) NOT NULL,
    price       DECIMAL(10, 2) DEFAULT 0,
    period_days INT            DEFAULT 30,
    max_users   INT,
    max_storage BIGINT,
    features    JSONB,
    sort        INT            DEFAULT 0,
    status      SMALLINT       DEFAULT 1,
    created_at  TIMESTAMPTZ    DEFAULT NOW(),
    updated_at  TIMESTAMPTZ    DEFAULT NOW(),
    is_deleted  BOOLEAN        DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_plan_code ON plans (code) WHERE is_deleted = FALSE;

-- 17. usage_records (用量记录表)
CREATE TABLE IF NOT EXISTS usage_records
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,
    metric     VARCHAR(32) NOT NULL,
    value      BIGINT      DEFAULT 0,
    period VARCHAR (20),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_usage_tenant ON usage_records (tenant_id);

-- 19. attachments 已迁出至 migrations/asset.sql（apps/reference/asset 拥有）

-- 20. auth_sessions (会话表)
CREATE TABLE IF NOT EXISTS auth_sessions
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    account_id BIGINT       NOT NULL,
    token      VARCHAR(255) NOT NULL,
    ip         VARCHAR(64),
    user_agent VARCHAR(255),
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_auth_session_token ON auth_sessions (token);

-- 21. tenant_user_seq (租户用户序号表)
CREATE TABLE IF NOT EXISTS tenant_user_seq
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    seq        BIGINT      DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_tenant_seq ON tenant_user_seq (tenant_id);

-- 22. account_roles (平台角色表)
CREATE TABLE IF NOT EXISTS account_roles
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    account_id BIGINT      NOT NULL,
    role       VARCHAR(32) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_account_role ON account_roles (account_id, role);

-- ============================================
-- role_menus 和 role_resources (重构后的角色权限表)
-- ============================================
CREATE TABLE IF NOT EXISTS role_menus
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    role_id    BIGINT NOT NULL,
    menu_id    BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_role_menu ON role_menus (role_id, menu_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_role_menus_tenant ON role_menus (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_role_menus_role ON role_menus (role_id) WHERE is_deleted = FALSE;

CREATE TABLE IF NOT EXISTS role_resources
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id   BIGINT NOT NULL,
    role_id     BIGINT NOT NULL,
    resource_id BIGINT NOT NULL,
    effect      SMALLINT    DEFAULT 1,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    is_deleted  BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_role_resource ON role_resources (role_id, resource_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_role_resources_tenant ON role_resources (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_role_resources_role ON role_resources (role_id) WHERE is_deleted = FALSE;

-- ============================================
-- 🔐 RLS (行级安全) 策略 — 纵深防御层
-- 关键7表启用 RLS
-- ============================================
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE role_data_scopes ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_user_seq ENABLE ROW LEVEL SECURITY;
ALTER TABLE role_menus ENABLE ROW LEVEL SECURITY;
ALTER TABLE role_resources ENABLE ROW LEVEL SECURITY;
ALTER TABLE dicts ENABLE ROW LEVEL SECURITY;
ALTER TABLE dict_items ENABLE ROW LEVEL SECURITY;

-- RLS policy 表达式（bypass_rls-aware）：
--   tenant_id 匹配 app.tenant_id  OR  app.bypass_rls = 'on'
-- 这样 RunInPlatformTx 设置 app.bypass_rls='on' 后能真正跨租户访问，
-- 业务请求走 RunInTenantTx 仍然只看到自己租户的数据。

-- users
DROP
POLICY IF EXISTS tenant_isolation_policy ON users;
CREATE
POLICY tenant_isolation_policy ON users
USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

-- roles
DROP
POLICY IF EXISTS tenant_isolation_policy ON roles;
CREATE
POLICY tenant_isolation_policy ON roles
USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

-- role_data_scopes
DROP
POLICY IF EXISTS tenant_isolation_policy ON role_data_scopes;
CREATE
POLICY tenant_isolation_policy ON role_data_scopes
USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

-- user_roles
DROP
POLICY IF EXISTS tenant_isolation_policy ON user_roles;
CREATE
POLICY tenant_isolation_policy ON user_roles
USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

-- organizations
DROP
POLICY IF EXISTS tenant_isolation_policy ON organizations;
CREATE
POLICY tenant_isolation_policy ON organizations
USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

-- tenant_user_seq
DROP
POLICY IF EXISTS tenant_isolation_policy ON tenant_user_seq;
CREATE
POLICY tenant_isolation_policy ON tenant_user_seq
USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

-- role_menus
DROP
POLICY IF EXISTS tenant_isolation_policy ON role_menus;
CREATE
POLICY tenant_isolation_policy ON role_menus
USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

-- role_resources
DROP
POLICY IF EXISTS tenant_isolation_policy ON role_resources;
CREATE
POLICY tenant_isolation_policy ON role_resources
USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

-- dicts（保留 tenant_id=0 短路：系统级字典跨租户共享）
DROP
POLICY IF EXISTS tenant_isolation_policy ON dicts;
CREATE
POLICY tenant_isolation_policy ON dicts
USING (
    tenant_id = 0
    OR tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

-- dict_items（保留 tenant_id=0 短路）
DROP
POLICY IF EXISTS tenant_isolation_policy ON dict_items;
CREATE
POLICY tenant_isolation_policy ON dict_items
USING (
    tenant_id = 0
    OR tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

-- ============================================
-- 初始化数据
-- ============================================

-- 租户
INSERT INTO tenants (code, name, status, created_by, updated_by)
VALUES ('default', '默认租户', 1, 0, 0);

-- 账号 (password: admin123)
INSERT INTO accounts (phone, email, password, username, real_name, status)
VALUES ('13800138000', 'admin@example.com',
        '$argon2id$v=19$m=19456,t=2,p=1$nMpweyGYDB9dvGMQAkzcHw$Tfc9vn1or7d0KMg0h6aRYFuDMxZbuK2cO8o6VaOyBk4', 'admin',
        '系统管理员', 1);

-- 用户
INSERT INTO users (tenant_id, account_id, code, org_id, real_name, status, created_by, updated_by)
VALUES (1, 1, 'admin', NULL, '系统管理员', 1, 0, 0);

-- 角色
INSERT INTO roles (tenant_id, code, name, description, data_scope, is_default, sort, status, created_by, updated_by)
VALUES (1, 'admin', '管理员', '系统管理员', 1, FALSE, 1, 1, 0, 0),
       (1, 'user', '普通用户', '普通用户', 4, TRUE, 2, 1, 0, 0);

-- 用户角色
INSERT INTO user_roles (tenant_id, user_id, role_id)
VALUES (1, 1, 1);

-- 菜单（框架基础菜单；业务菜单如 frames/avatars 由各自模块的 migration 提供）
INSERT INTO menus (id, tenant_id, code, name, path, icon, sort, parent_id, ancestors, visible, enabled)
    OVERRIDING SYSTEM VALUE
VALUES (1, 1, 'dashboard', '仪表盘', '/dashboard', 'LayoutDashboardIcon', 1, 0, '1', TRUE, TRUE),
       (2, 1, 'analytics', '数据分析', '/analytics', 'ChartBarIcon', 2, 0, '2', TRUE, TRUE),
       (3, 1, 'projects', '项目管理', '/projects', 'FolderIcon', 3, 0, '3', TRUE, TRUE),
       (4, 1, 'team', '团队管理', '/team', 'UsersIcon', 4, 0, '4', TRUE, TRUE),
       (5, 1, 'system', '系统管理', '/system', 'SettingsIcon', 5, 0, '5', TRUE, TRUE);

INSERT INTO menus (id, tenant_id, code, name, path, icon, sort, parent_id, ancestors, visible, enabled)
    OVERRIDING SYSTEM VALUE
VALUES (51, 1, 'users', '用户管理', '/users', 'FileIcon', 1, 5, '5.51', TRUE, TRUE),
       (52, 1, 'roles', '角色管理', '/roles', 'ShieldIcon', 2, 5, '5.52', TRUE, TRUE),
       (53, 1, 'menus', '菜单管理', '/menus', 'MenuIcon', 3, 5, '5.53', TRUE, TRUE),
       (54, 1, 'resources', '资源管理', '/resources', 'ResourceIcon', 4, 5, '5.54', TRUE, TRUE),
       (55, 1, 'organizations', '组织管理', '/organizations', 'ResourceIcon', 5, 5, '5.55', TRUE, TRUE),
       (56, 1, 'dicts', '数据字典', '/dicts', 'BookIcon', 6, 5, '5.56', TRUE, TRUE);

SELECT setval('menus_id_seq', 300, true);

-- 资源 (系统管理菜单下的资源)
INSERT INTO resources (tenant_id, menu_id, code, name, action, description, sort, status)
VALUES (1, 54, 'resource:list', '查询资源', 'list', '查询资源列表', 1, 1),
       (1, 54, 'resource:get', '查看资源', 'get', '查看单个资源详情', 2, 1),
       (1, 54, 'resource:create', '创建资源', 'create', '新建资源', 3, 1),
       (1, 54, 'resource:update', '更新资源', 'update', '更新资源信息', 4, 1),
       (1, 54, 'resource:delete', '删除资源', 'delete', '删除资源', 5, 1);

-- 字典资源\uff08菜单 56\uff09
INSERT INTO resources (tenant_id, menu_id, code, name, action, description, sort, status)
VALUES (1, 56, 'dict:list', '查询字典', 'list', '查询字典列表', 1, 1),
       (1, 56, 'dict:get', '查看字典', 'get', '查看单个字典及字典项', 2, 1),
       (1, 56, 'dict:create', '创建字典', 'create', '新建字典', 3, 1),
       (1, 56, 'dict:update', '更新字典', 'update', '更新字典及字典项', 4, 1),
       (1, 56, 'dict:delete', '删除字典', 'delete', '删除字典及字典项', 5, 1);

SELECT setval('resources_id_seq', 100, true);

-- admin 角色绑定所有菜单 (role_menus)
INSERT INTO role_menus (tenant_id, role_id, menu_id)
SELECT 1, 1, id
FROM menus
WHERE is_deleted = FALSE;

-- 超级资源
INSERT INTO resources (tenant_id, code, name, action, description, status)
VALUES (1, '*', '超级管理员通配权限', '*', '拥有系统所有权限', 1);

-- admin 角色绑定超级资源 (role_resources)
INSERT INTO role_resources (tenant_id, role_id, resource_id, effect)
VALUES (1, 1, (SELECT id FROM resources WHERE code = '*'), 1);

-- 字典示例数据
INSERT INTO dicts (tenant_id, code, name, sort, status)
VALUES (0, 'gender', '性别', 1, 1),
       (0, 'user_status', '用户状态', 2, 1),
       (0, 'education', '学历', 3, 1);

INSERT INTO dict_items (tenant_id, dict_id, code, name, sort, status)
SELECT 0, d.id, x.code, x.name, x.sort, 1
FROM dicts d
JOIN (VALUES
  ('gender', 'male', '男', 1),
  ('gender', 'female', '女', 2),
  ('user_status', 'active', '启用', 1),
  ('user_status', 'disabled', '停用', 2),
  ('education', 'bachelor', '本科', 1),
  ('education', 'master', '硕士', 2),
  ('education', 'doctor', '博士', 3)
) AS x(dict_code, code, name, sort) ON x.dict_code = d.code;

-- ============================================
-- 📦 模板租户（__template__）—— 所有新建租户的首装源
-- ============================================
-- 设计意图：通过 API 创建租户时，tenant.Service.Create → first_install
-- 从本租户复制 menus / resources / dicts。后续运维只需在 __template__
-- 内追加一次，新租户自动获得。
--
-- 包含的表：menus / resources（含 *:*）/ dicts / dict_items
-- 不包含的表：users / organizations / user_roles / role_data_scopes / tenant_user_seq
-- （这些是租户级独有，每次首装独立创建）

-- 1) 模板租户本身
-- status=0：不参与业务；is_deleted=FALSE：必须存活的"源"
INSERT INTO tenants (code, name, status, created_by, updated_by)
VALUES ('__template__', '[系统] 租户模板', 0, 0, 0)
ON CONFLICT (code) WHERE is_deleted = FALSE DO NOTHING;

-- 2) 复制 menus：从 default 租户（用 code 查源，比硬编码 tenant_id=1 更稳健）
-- 2a) 根菜单（parent_id=0）
INSERT INTO menus (tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
SELECT (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE),
       code, name, subtitle, url, path, icon, sort, 0, '', visible, enabled
FROM menus
WHERE tenant_id = (SELECT id FROM tenants WHERE code = 'default' AND is_deleted = FALSE)
  AND parent_id = 0
  AND is_deleted = FALSE;

-- 2b) 子菜单：用 code 重新映射 parent_id 到新租户同 code 菜单
INSERT INTO menus (tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
SELECT (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE),
       m.code, m.name, m.subtitle, m.url, m.path, m.icon, m.sort,
       new_p.id, '', m.visible, m.enabled
FROM menus m
JOIN menus old_p ON old_p.id = m.parent_id
                 AND old_p.tenant_id = (SELECT id FROM tenants WHERE code = 'default' AND is_deleted = FALSE)
                 AND old_p.is_deleted = FALSE
JOIN menus new_p ON new_p.code = old_p.code
                AND new_p.tenant_id = (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE)
                AND new_p.is_deleted = FALSE
WHERE m.tenant_id = (SELECT id FROM tenants WHERE code = 'default' AND is_deleted = FALSE)
  AND m.parent_id > 0
  AND m.is_deleted = FALSE;

-- 2c) 重建 ancestors（用 parent_id::text 表达层级路径）
UPDATE menus SET ancestors = parent_id::text
WHERE tenant_id = (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE)
  AND parent_id > 0;

-- 3) 复制 resources：从 default 租户
-- 用 code 重新映射 menu_id（resources.menu_id 是 BIGINT，无显式 FK，跨租户 NULL 安全）
INSERT INTO resources (tenant_id, menu_id, code, name, action, description, sort, status)
SELECT (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE),
       new_m.id, r.code, r.name, r.action, r.description, r.sort, r.status
FROM resources r
LEFT JOIN menus new_m ON new_m.code = (
        SELECT code FROM menus WHERE id = r.menu_id
            AND tenant_id = (SELECT id FROM tenants WHERE code = 'default' AND is_deleted = FALSE)
            AND is_deleted = FALSE
    ) AND new_m.tenant_id = (
        SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE
    ) AND new_m.is_deleted = FALSE
WHERE r.tenant_id = (SELECT id FROM tenants WHERE code = 'default' AND is_deleted = FALSE)
  AND r.is_deleted = FALSE;

-- 4) 复制 dicts：从系统级（tenant_id=0）
INSERT INTO dicts (tenant_id, code, name, sort, status, extend)
SELECT (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE),
       code, name, sort, status, extend
FROM dicts
WHERE tenant_id = 0 AND is_deleted = FALSE;

-- 5) 复制 dict_items：用 code 重新映射 dict_id
INSERT INTO dict_items (tenant_id, dict_id, code, name, sort, status, extend)
SELECT (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE),
       new_d.id, di.code, di.name, di.sort, di.status, di.extend
FROM dict_items di
JOIN dicts old_d ON old_d.id = di.dict_id AND old_d.tenant_id = 0 AND old_d.is_deleted = FALSE
JOIN dicts new_d ON new_d.code = old_d.code
                AND new_d.tenant_id = (
                    SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE
                ) AND new_d.is_deleted = FALSE
WHERE di.tenant_id = 0 AND di.is_deleted = FALSE;

-- 6) 推 sequence（避免新租户 first_install 时 id 撞 default 1~300 区间）
SELECT setval('menus_id_seq', GREATEST(
    (SELECT COALESCE(MAX(id), 0) FROM menus),
    (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE) * 1000
), true);

SELECT setval('resources_id_seq', GREATEST(
    (SELECT COALESCE(MAX(id), 0) FROM resources),
    (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE) * 1000
), true);

SELECT setval('dicts_id_seq', GREATEST(
    (SELECT COALESCE(MAX(id), 0) FROM dicts),
    (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE) * 1000
), true);

SELECT setval('dict_items_id_seq', GREATEST(
    (SELECT COALESCE(MAX(id), 0) FROM dict_items),
    (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE) * 1000
), true);

-- ============================================
-- 🔧 Config 模块 seed（通用配置）
-- 依赖：config.sql 已建表（字母序 c < f，config.sql 先于本文件跑）
-- 4 个预置分组 + 19 个预置项 + 1 个菜单 (config) + 5 个资源 (config:*)
-- 新租户首装时由 apps/boot/tenant/first_install.go 从 __template__ 复制
-- ============================================

-- config_groups
INSERT INTO config_groups (tenant_id, code, name, description, icon, sort, is_system, is_public)
SELECT (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE),
       'site', '站点信息', '站点名称、Logo、版权等公开信息', 'GlobeIcon', 1, TRUE, TRUE
ON CONFLICT (tenant_id, code) WHERE is_deleted = FALSE DO NOTHING;

INSERT INTO config_groups (tenant_id, code, name, description, icon, sort, is_system, is_public)
SELECT (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE),
       'security', '安全策略', '密码强度、会话超时等安全相关配置', 'ShieldIcon', 2, TRUE, FALSE
ON CONFLICT (tenant_id, code) WHERE is_deleted = FALSE DO NOTHING;

INSERT INTO config_groups (tenant_id, code, name, description, icon, sort, is_system, is_public)
SELECT (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE),
       'email', '邮件服务', 'SMTP 邮件服务配置', 'MailIcon', 3, TRUE, FALSE
ON CONFLICT (tenant_id, code) WHERE is_deleted = FALSE DO NOTHING;

INSERT INTO config_groups (tenant_id, code, name, description, icon, sort, is_system, is_public)
SELECT (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE),
       'feature_flag', '功能开关', '系统级功能启用/禁用开关', 'ToggleLeftIcon', 4, TRUE, FALSE
ON CONFLICT (tenant_id, code) WHERE is_deleted = FALSE DO NOTHING;

-- site items
INSERT INTO config_items (tenant_id, group_id, key, value, default_value, type, label, description, sort, is_public, is_system)
SELECT
    (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE),
    (SELECT id FROM config_groups WHERE code = 'site' AND tenant_id = (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE) AND is_deleted = FALSE),
    s.key, s.value, s.default_value, s.type, s.label, s.description, s.sort, s.is_public, TRUE
FROM (VALUES
    ('site_name',         '"XinFramework"'::jsonb, '"XinFramework"'::jsonb, 'string', '站点名称', '显示在页面标题、登录页等位置', 1, TRUE),
    ('site_logo',         '""'::jsonb,              '""'::jsonb,             'image',  '站点 Logo', '建议 PNG/SVG，背景透明', 2, TRUE),
    ('site_favicon',      '""'::jsonb,              '""'::jsonb,             'image',  'Favicon',  '浏览器标签图标', 3, TRUE),
    ('site_copyright',    '""'::jsonb,              '""'::jsonb,             'string', '版权信息', '页面底部显示', 4, TRUE),
    ('site_icp',          '""'::jsonb,              '""'::jsonb,             'string', 'ICP 备案号', '中国大陆站点必填', 5, TRUE),
    ('site_locale_default', '"zh-CN"'::jsonb,       '"zh-CN"'::jsonb,        'select', '默认语言', 'zh-CN / en-US', 6, TRUE),
    ('login_background',  '""'::jsonb,              '""'::jsonb,             'image',  '登录页背景', '登录页右侧大图', 7, TRUE)
) AS s(key, value, default_value, type, label, description, sort, is_public)
ON CONFLICT (tenant_id, group_id, key) WHERE is_deleted = FALSE DO NOTHING;

-- security items
INSERT INTO config_items (tenant_id, group_id, key, value, default_value, type, label, description, validation, sort, is_public, is_system)
SELECT
    (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE),
    (SELECT id FROM config_groups WHERE code = 'security' AND tenant_id = (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE) AND is_deleted = FALSE),
    s.key, s.value, s.default_value, s.type, s.label, s.description, s.validation, s.sort, FALSE, TRUE
FROM (VALUES
    ('password_min_length',  '8'::jsonb,    '8'::jsonb,    'number', '密码最小长度', '新建/修改密码时校验',           '{"min":6,"max":32,"required":true}'::jsonb, 1),
    ('password_complexity',  '"standard"'::jsonb, '"standard"'::jsonb, 'select', '密码复杂度', 'low/standard/strong', '[{"label":"低(纯字母数字)","value":"low"},{"label":"标准(字母+数字)","value":"standard"},{"label":"强(字母+数字+符号)","value":"strong"}]'::jsonb, 2),
    ('session_timeout_min',  '30'::jsonb,   '30'::jsonb,   'number', '会话超时(分钟)', '空闲超过此时间强制下线',       '{"min":5,"max":1440,"required":true}'::jsonb, 3),
    ('max_login_attempts',   '5'::jsonb,    '5'::jsonb,    'number', '最大登录失败次数', '超过后锁定账户',                '{"min":1,"max":20,"required":true}'::jsonb, 4),
    ('lock_duration_min',    '5'::jsonb,    '5'::jsonb,    'number', '锁定时长(分钟)',   '失败次数超限后的锁定时长',       '{"min":1,"max":1440,"required":true}'::jsonb, 5)
) AS s(key, value, default_value, type, label, description, validation, sort)
ON CONFLICT (tenant_id, group_id, key) WHERE is_deleted = FALSE DO NOTHING;

-- email items
INSERT INTO config_items (tenant_id, group_id, key, value, default_value, type, label, description, sort, is_public, is_readonly, is_system)
SELECT
    (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE),
    (SELECT id FROM config_groups WHERE code = 'email' AND tenant_id = (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE) AND is_deleted = FALSE),
    s.key, s.value, s.default_value, s.type, s.label, s.description, s.sort, FALSE, s.is_readonly, TRUE
FROM (VALUES
    ('smtp_host',     '""'::jsonb,         '""'::jsonb,         'string',  'SMTP 主机',   '如 smtp.example.com',  1, FALSE),
    ('smtp_port',     '465'::jsonb,        '465'::jsonb,        'number',  'SMTP 端口',   '常用 25/465/587',     2, FALSE),
    ('smtp_user',     '""'::jsonb,         '""'::jsonb,         'string',  'SMTP 用户',   '通常为邮箱地址',       3, FALSE),
    ('smtp_password', '""'::jsonb,         '""'::jsonb,         'password','SMTP 密码',   '授权码或登录密码',     4, TRUE),
    ('smtp_from',     '""'::jsonb,         '""'::jsonb,         'string',  '发件人邮箱',  '邮件 From 头',         5, FALSE),
    ('smtp_use_tls',  'true'::jsonb,       'true'::jsonb,       'boolean', '启用 TLS',    '465 通常 TLS，587 STARTTLS', 6, FALSE)
) AS s(key, value, default_value, type, label, description, sort, is_readonly)
ON CONFLICT (tenant_id, group_id, key) WHERE is_deleted = FALSE DO NOTHING;

-- feature_flag items\
INSERT INTO config_items (tenant_id, group_id, key, value, default_value, type, label, description, sort, is_public, is_system)
SELECT
    (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE),
    (SELECT id FROM config_groups WHERE code = 'feature_flag' AND tenant_id = (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE) AND is_deleted = FALSE),
    s.key, s.value, s.default_value, s.type, s.label, s.description, s.sort, FALSE, TRUE
FROM (VALUES
    ('enable_registration', 'true'::jsonb, 'true'::jsonb, 'boolean', '开放注册', '允许外部用户自助注册', 1),
    ('enable_audit_log',    'true'::jsonb, 'true'::jsonb, 'boolean', '审计日志', '记录关键操作审计日志', 2)
) AS s(key, value, default_value, type, label, description, sort)
ON CONFLICT (tenant_id, group_id, key) WHERE is_deleted = FALSE DO NOTHING;

-- 菜单：系统管理 → 配置管理
-- 注意：parent_id 必须用子查询拿 system 菜单在 __template__ 里的实际 id
-- （__template__ 里的 system id 已经被 setval 推到几千，不是源 default 里的 5）
-- ancestors 留空，下面 UPDATE 段会统一重建为 parent_id::text
INSERT INTO menus (tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
SELECT (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE),
       'config', '配置管理', '系统配置项管理', '', '/settings', 'SettingsIcon', 0,
       (SELECT id FROM menus WHERE code = 'system' AND tenant_id = (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE) AND is_deleted = FALSE),
       '', TRUE, TRUE
ON CONFLICT (tenant_id, code) WHERE is_deleted = FALSE DO NOTHING;

-- 重建 config menu 的 ancestors（与上面 2c 段保持一致）
UPDATE menus SET ancestors = parent_id::text
WHERE tenant_id = (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE)
  AND code = 'config' AND parent_id > 0 AND is_deleted = FALSE;

-- 资源：config:list/get/create/update/delete
INSERT INTO resources (tenant_id, menu_id, code, name, action, description, sort, status)
SELECT (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE),
       (SELECT id FROM menus WHERE code = 'config' AND tenant_id = (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE) AND is_deleted = FALSE),
       r.code, r.name, r.action, r.description, r.sort, 1
FROM (VALUES
    ('config:list',   '查询配置', 'list',   '查询配置分组与项',  1),
    ('config:get',    '查看配置', 'get',    '查看分组/项详情',   2),
    ('config:create', '创建配置', 'create', '新建分组或项',       3),
    ('config:update', '更新配置', 'update', '更新分组或项',       4),
    ('config:delete', '删除配置', 'delete', '删除分组或项',       5)
) AS r(code, name, action, description, sort)
ON CONFLICT (tenant_id, code) WHERE is_deleted = FALSE DO NOTHING;

-- 序列号兜底（与上面 setval 段保持一致；保证后续 first_install 复制时 id 不冲突）
SELECT setval('config_groups_id_seq', GREATEST(
    (SELECT COALESCE(MAX(id), 0) FROM config_groups),
    (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE) * 1000
), true);

SELECT setval('config_items_id_seq', GREATEST(
    (SELECT COALESCE(MAX(id), 0) FROM config_items),
    (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE) * 1000
), true);

-- ============================================
-- 🖼️ flag 模块业务菜单 seed（写在 framework.sql 末尾是因为 flag.sql 字母序在 framework 之前，
-- 但 seed 需要 tenants/menus 表，故统一放 __template__ 段；first_install.go 全量复制到新租户）
-- ============================================

-- 顶级：相框管理
INSERT INTO menus (tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
SELECT (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE),
       'frames', '相框管理', '头像框与活动空间', '', '/frames', 'FrameIcon', 6, 0, '', TRUE, TRUE
ON CONFLICT (tenant_id, code) WHERE is_deleted = FALSE DO NOTHING;

-- 顶级：头像管理
INSERT INTO menus (tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
SELECT (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE),
       'avatars', '头像管理', '用户头像与分类', '', '/avatars', 'ImageIcon', 7, 0, '', TRUE, TRUE
ON CONFLICT (tenant_id, code) WHERE is_deleted = FALSE DO NOTHING;

-- 子菜单：相框列表、相框分类（parent = frames）
INSERT INTO menus (tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
SELECT (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE),
       s.code, s.name, s.subtitle, s.url, s.path, s.icon, s.sort,
       (SELECT id FROM menus WHERE code = 'frames' AND tenant_id = (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE) AND is_deleted = FALSE),
       '', TRUE, TRUE
FROM (VALUES
    ('frame-list',        '相框列表', '', '', '/frames',           'FileIcon',  1),
    ('frame-categories',  '相框分类', '', '', '/frame-categories', 'ListIcon',  2)
) AS s(code, name, subtitle, url, path, icon, sort)
ON CONFLICT (tenant_id, code) WHERE is_deleted = FALSE DO NOTHING;

-- 子菜单：头像列表、头像分类（parent = avatars）
INSERT INTO menus (tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
SELECT (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE),
       s.code, s.name, s.subtitle, s.url, s.path, s.icon, s.sort,
       (SELECT id FROM menus WHERE code = 'avatars' AND tenant_id = (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE) AND is_deleted = FALSE),
       '', TRUE, TRUE
FROM (VALUES
    ('avatar-list',        '头像列表', '', '', '/avatars',           'FileIcon',  1),
    ('avatar-categories',  '头像分类', '', '', '/avatar-categories', 'ListIcon',  2)
) AS s(code, name, subtitle, url, path, icon, sort)
ON CONFLICT (tenant_id, code) WHERE is_deleted = FALSE DO NOTHING;

-- 重建 frames/avatars 子菜单的 ancestors（与 first_install.go 2c 段一致）
UPDATE menus SET ancestors = parent_id::text
WHERE tenant_id = (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE)
  AND code IN ('frame-list', 'frame-categories', 'avatar-list', 'avatar-categories')
  AND parent_id > 0 AND is_deleted = FALSE;

-- 🔑 flag 资源 seed（__template__ 租户；first_install.go 会全量复制）
-- 让 flag 模块的菜单可被角色授权 / RBAC 校验
INSERT INTO resources (tenant_id, menu_id, code, name, action, description, sort, status)
SELECT (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE),
       (SELECT id FROM menus WHERE code = s.menu_code AND tenant_id = (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE) AND is_deleted = FALSE),
       s.code, s.name, s.action, s.description, s.sort, 1
FROM (VALUES
    ('flag:list',   '查询相框/头像', 'list',   '查询相框、头像、活动空间',  1, 'frames'),
    ('flag:get',    '查看详情',      'get',    '查看相框/头像/空间详情',   2, 'frames'),
    ('flag:create', '创建相框/头像', 'create', '创建相框/头像/活动空间',   3, 'frames'),
    ('flag:update', '更新相框/头像', 'update', '更新相框/头像/活动空间',   4, 'frames'),
    ('flag:delete', '删除相框/头像', 'delete', '软删相框/头像/活动空间',   5, 'frames'),
    ('flag:list',   '查询相框/头像', 'list',   '查询相框、头像、活动空间',  1, 'avatars'),
    ('flag:get',    '查看详情',      'get',    '查看相框/头像/空间详情',   2, 'avatars'),
    ('flag:create', '创建相框/头像', 'create', '创建相框/头像/活动空间',   3, 'avatars'),
    ('flag:update', '更新相框/头像', 'update', '更新相框/头像/活动空间',   4, 'avatars'),
    ('flag:delete', '删除相框/头像', 'delete', '软删相框/头像/活动空间',   5, 'avatars')
) AS s(code, name, action, description, sort, menu_code)
ON CONFLICT (tenant_id, code) WHERE is_deleted = FALSE DO NOTHING;
