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
DROP TABLE IF EXISTS tenants;
CREATE TABLE tenants
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
CREATE UNIQUE INDEX uk_tenants_code ON tenants (code) WHERE is_deleted = FALSE;
CREATE INDEX idx_tenants_config_gin ON tenants USING GIN (config);

-- 2. accounts (全局账号表)
DROP TABLE IF EXISTS accounts;
CREATE TABLE accounts
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
CREATE UNIQUE INDEX uk_accounts_phone ON accounts (phone) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX uk_accounts_email ON accounts (email) WHERE is_deleted = FALSE;

-- 3. organizations (组织表)
DROP TABLE IF EXISTS organizations;
CREATE TABLE organizations
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,
    parent_id  BIGINT,
    code       VARCHAR(32) NOT NULL,
    name       VARCHAR(64) NOT NULL,
    ancestors  TEXT        DEFAULT '',
    sort       INT         DEFAULT 0,
    status     SMALLINT    DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by BIGINT,
    updated_by BIGINT,
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE INDEX idx_org_tenant ON organizations (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_org_parent ON organizations (parent_id) WHERE is_deleted = FALSE;

-- 4. users (用户表)
DROP TABLE IF EXISTS users;
CREATE TABLE users
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    account_id BIGINT NOT NULL,
    code       VARCHAR(32),
    org_id     BIGINT,
    real_name  VARCHAR(64),
    avatar     VARCHAR(512),
    status     SMALLINT    DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by BIGINT,
    updated_by BIGINT,
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE INDEX idx_users_tenant ON users (tenant_id) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX uk_users_account ON users (account_id) WHERE is_deleted = FALSE;

-- 5. roles (角色表)
DROP TABLE IF EXISTS roles;
CREATE TABLE roles
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
CREATE INDEX idx_roles_tenant ON roles (tenant_id) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX uk_roles_code ON roles (tenant_id, code) WHERE is_deleted = FALSE;

-- 6. role_data_scopes (角色数据范围表)
DROP TABLE IF EXISTS role_data_scopes;
CREATE TABLE role_data_scopes
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    role_id    BIGINT NOT NULL,
    org_id     BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE INDEX idx_rds_role ON role_data_scopes (role_id) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX uk_rds_unique ON role_data_scopes (role_id, org_id) WHERE is_deleted = FALSE;

-- 7. user_roles (用户角色关联表)
DROP TABLE IF EXISTS user_roles;
CREATE TABLE user_roles
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    user_id    BIGINT NOT NULL,
    role_id    BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE INDEX idx_ur_user ON user_roles (user_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_ur_role ON user_roles (role_id) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX uk_ur_unique ON user_roles (user_id, role_id) WHERE is_deleted = FALSE;

-- 8. menus (菜单表)
DROP TABLE IF EXISTS menus;
CREATE TABLE menus
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
CREATE UNIQUE INDEX uk_menu_code ON menus (tenant_id, code) WHERE is_deleted = FALSE;
CREATE INDEX idx_menu_tenant ON menus (tenant_id) WHERE is_deleted = FALSE;

-- 9. resources (资源表)
DROP TABLE IF EXISTS resources;
CREATE TABLE resources
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
CREATE UNIQUE INDEX uk_resource_code ON resources (tenant_id, code) WHERE is_deleted = FALSE;

-- 10. routes (路由表)
DROP TABLE IF EXISTS routes;
CREATE TABLE routes
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
CREATE INDEX idx_routes_tenant ON routes (tenant_id) WHERE is_deleted = FALSE;

-- 12. dicts (字典表)
DROP TABLE IF EXISTS dicts;
CREATE TABLE dicts
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,
    code       VARCHAR(32) NOT NULL,
    name       VARCHAR(64) NOT NULL,
    status     SMALLINT    DEFAULT 1,
    sort       INT         DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX uk_dict_code ON dicts (tenant_id, code) WHERE is_deleted = FALSE;

-- 13. dict_items (字典项表)
DROP TABLE IF EXISTS dict_items;
CREATE TABLE dict_items
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,
    dict_id    BIGINT      NOT NULL,
    label      VARCHAR(64) NOT NULL,
    value      VARCHAR(128),
    sort       INT         DEFAULT 0,
    status     SMALLINT    DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE INDEX idx_dict_items_dict ON dict_items (dict_id) WHERE is_deleted = FALSE;

-- 14. db_logs (数据库日志表)
DROP TABLE IF EXISTS db_logs;
CREATE TABLE db_logs
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
CREATE INDEX idx_db_logs_tenant ON db_logs (tenant_id);
CREATE INDEX idx_db_logs_user ON db_logs (user_id);

-- 15. subscriptions (订阅表)
DROP TABLE IF EXISTS subscriptions;
CREATE TABLE subscriptions
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
CREATE INDEX idx_subs_tenant ON subscriptions (tenant_id) WHERE is_deleted = FALSE;

-- 16. plans (套餐表)
DROP TABLE IF EXISTS plans;
CREATE TABLE plans
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
CREATE UNIQUE INDEX uk_plan_code ON plans (code) WHERE is_deleted = FALSE;

-- 17. usage_records (用量记录表)
DROP TABLE IF EXISTS usage_records;
CREATE TABLE usage_records
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,
    metric     VARCHAR(32) NOT NULL,
    value      BIGINT      DEFAULT 0,
    period VARCHAR (20),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_usage_tenant ON usage_records (tenant_id);

-- 18. ai_documents (AI文档表)
DROP TABLE IF EXISTS ai_documents;
CREATE TABLE ai_documents
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    title      VARCHAR(255),
    content    TEXT,
    status     SMALLINT    DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE INDEX idx_ai_doc_tenant ON ai_documents (tenant_id) WHERE is_deleted = FALSE;

-- 19. attachments (附件表)
DROP TABLE IF EXISTS attachments;
CREATE TABLE attachments
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    user_id    BIGINT,
    file_name  VARCHAR(255),
    file_ext   VARCHAR(32),
    mime_type  VARCHAR(64),
    file_size  BIGINT,
    storage    VARCHAR(32),
    object_key VARCHAR(255),
    url        VARCHAR(512),
    hash       VARCHAR(64),
    status     SMALLINT    DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE INDEX idx_attachments_tenant ON attachments (tenant_id) WHERE is_deleted = FALSE;

-- 20. auth_sessions (会话表)
DROP TABLE IF EXISTS auth_sessions;
CREATE TABLE auth_sessions
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    account_id BIGINT       NOT NULL,
    token      VARCHAR(255) NOT NULL,
    ip         VARCHAR(64),
    user_agent VARCHAR(255),
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_auth_session_token ON auth_sessions (token);

-- 21. tenant_user_seq (租户用户序号表)
DROP TABLE IF EXISTS tenant_user_seq;
CREATE TABLE tenant_user_seq
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    seq        BIGINT      DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE UNIQUE INDEX uk_tenant_seq ON tenant_user_seq (tenant_id);

-- 22. account_roles (平台角色表)
DROP TABLE IF EXISTS account_roles;
CREATE TABLE account_roles
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    account_id BIGINT      NOT NULL,
    role       VARCHAR(32) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE UNIQUE INDEX uk_account_role ON account_roles (account_id, role);

-- ============================================
-- role_menus 和 role_resources (重构后的角色权限表)
-- ============================================
DROP TABLE IF EXISTS role_menus;
CREATE TABLE role_menus
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    role_id    BIGINT NOT NULL,
    menu_id    BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX uk_role_menu ON role_menus (role_id, menu_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_role_menus_tenant ON role_menus (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_role_menus_role ON role_menus (role_id) WHERE is_deleted = FALSE;

DROP TABLE IF EXISTS role_resources;
CREATE TABLE role_resources
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
CREATE UNIQUE INDEX uk_role_resource ON role_resources (role_id, resource_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_role_resources_tenant ON role_resources (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_role_resources_role ON role_resources (role_id) WHERE is_deleted = FALSE;

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

-- users
DROP
POLICY IF EXISTS tenant_isolation_policy ON users;
CREATE
POLICY tenant_isolation_policy ON users USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT);

-- roles
DROP
POLICY IF EXISTS tenant_isolation_policy ON roles;
CREATE
POLICY tenant_isolation_policy ON roles USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT);

-- role_data_scopes
DROP
POLICY IF EXISTS tenant_isolation_policy ON role_data_scopes;
CREATE
POLICY tenant_isolation_policy ON role_data_scopes USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT);

-- user_roles
DROP
POLICY IF EXISTS tenant_isolation_policy ON user_roles;
CREATE
POLICY tenant_isolation_policy ON user_roles USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT);

-- organizations
DROP
POLICY IF EXISTS tenant_isolation_policy ON organizations;
CREATE
POLICY tenant_isolation_policy ON organizations USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT);

-- tenant_user_seq
DROP
POLICY IF EXISTS tenant_isolation_policy ON tenant_user_seq;
CREATE
POLICY tenant_isolation_policy ON tenant_user_seq USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT);

-- role_menus
DROP
POLICY IF EXISTS tenant_isolation_policy ON role_menus;
CREATE
POLICY tenant_isolation_policy ON role_menus USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT);

-- role_resources
DROP
POLICY IF EXISTS tenant_isolation_policy ON role_resources;
CREATE
POLICY tenant_isolation_policy ON role_resources USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT);

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
VALUES (1, 'admin', '管理员', '系统管理员', 5, FALSE, 1, 1, 0, 0),
       (1, 'user', '普通用户', '普通用户', 4, TRUE, 2, 1, 0, 0);

-- 用户角色
INSERT INTO user_roles (tenant_id, user_id, role_id)
VALUES (1, 1, 1);

-- 菜单
INSERT INTO menus (id, tenant_id, code, name, path, icon, sort, parent_id, ancestors, visible, enabled)
    OVERRIDING SYSTEM VALUE
VALUES (1, 1, 'dashboard', '仪表盘', '/dashboard', 'LayoutDashboardIcon', 1, 0, '1', TRUE, TRUE),
       (2, 1, 'analytics', '数据分析', '/analytics', 'ChartBarIcon', 2, 0, '2', TRUE, TRUE),
       (3, 1, 'projects', '项目管理', '/projects', 'FolderIcon', 3, 0, '3', TRUE, TRUE),
       (4, 1, 'team', '团队管理', '/team', 'UsersIcon', 4, 0, '4', TRUE, TRUE),
       (5, 1, 'system', '系统管理', '/system', 'SettingsIcon', 5, 0, '5', TRUE, TRUE),
       (6, 1, 'frames', '相框管理', '/frames', 'FrameIcon', 6, 0, '6', TRUE, TRUE),
       (7, 1, 'avatars', '头像管理', '/avatars', 'ImageIcon', 7, 0, '7', TRUE, TRUE);

INSERT INTO menus (id, tenant_id, code, name, path, icon, sort, parent_id, ancestors, visible, enabled)
    OVERRIDING SYSTEM VALUE
VALUES (51, 1, 'users', '用户管理', '/users', 'FileIcon', 1, 5, '5.51', TRUE, TRUE),
       (52, 1, 'roles', '角色管理', '/roles', 'ShieldIcon', 2, 5, '5.52', TRUE, TRUE),
       (53, 1, 'menus', '菜单管理', '/menus', 'MenuIcon', 3, 5, '5.53', TRUE, TRUE),
       (54, 1, 'resources', '资源管理', '/resources', 'ResourceIcon', 4, 5, '5.54', TRUE, TRUE),
       (61, 1, 'frame-list', '相框列表', '/frames', 'FileIcon', 1, 6, '6.61', TRUE, TRUE),
       (62, 1, 'frame-categories', '相框分类', '/frame-categories', 'ListIcon', 2, 6, '6.62', TRUE, TRUE),
       (71, 1, 'avatar-list', '头像列表', '/avatars', 'FileIcon', 1, 7, '7.71', TRUE, TRUE),
       (72, 1, 'avatar-categories', '头像分类', '/avatar-categories', 'ListIcon', 2, 7, '7.72', TRUE, TRUE);

SELECT setval('menus_id_seq', 300, true);

-- 资源 (系统管理菜单下的资源)
INSERT INTO resources (tenant_id, menu_id, code, name, action, description, sort, status)
VALUES (1, 54, 'resource:list', '查询资源', 'GET', '查询资源列表', 1, 1),
       (1, 54, 'resource:get', '查看资源', 'GET', '查看单个资源详情', 2, 1),
       (1, 54, 'resource:create', '创建资源', 'POST', '新建资源', 3, 1),
       (1, 54, 'resource:update', '更新资源', 'PUT', '更新资源信息', 4, 1),
       (1, 54, 'resource:delete', '删除资源', 'DELETE', '删除资源', 5, 1);

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