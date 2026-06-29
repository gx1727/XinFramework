-- ============================================
-- init_schema.sql : 全部数据库 schema（开发期统一入口）
-- ============================================
-- 目的：
--   开发阶段（仍可 DROP DATABASE）把全部表 / 索引 / RLS 集中到一个文件。
--   业务模块（asset / cms / flag）保留独立 .sql，因为它们生命周期不同。
--
-- 文件分层（按数据域）：
--   1. 平台与租户公共表（tenants / accounts / auth_sessions）
--   2. 租户域 tenant_* 表（Phase 0023.3 终态）
--   3. 平台域 sys_* 表（Phase 0023+ 拆分）
--   4. 业务支撑表（订阅 / 套餐 / 用量 / 日志）
--   5. 字典 / 配置 / 可见性
--   6. 路由
--   7. RLS 策略
--   8. 完整性校验
--
-- 0023.3 关键变化（vs init_schema.sql 重写前）：
--   - users → tenant_users
--   - roles → tenant_roles
--   - organizations → tenant_organizations
--   - user_roles → tenant_user_roles
--   - role_menus → tenant_role_menus
--   - role_resources → tenant_role_resources
--   - role_data_scopes → tenant_role_data_scopes
--   - resources → tenant_permissions
--   - menus 表拆分：tenant_menus（租户域） + sys_menus（平台域）
--     tenant_menus 不再有 scope 字段（sys_menus 承担 platform scope）
--   - account_roles 表 DROP（由 sys_user_roles 替代，boot/auth 走 sys_*）
--
-- 不做什么：
--   - 不写 seed（见 init_seed.sql）
--   - 不写 admin 账号初始化（见 init_seed.sql 头部）
--   - 业务模块（asset/cms/flag）的表保留在各自 .sql
--   - 不做版本迁移（开发期随改随重跑）
--
-- 重置方式（dev）：
--   DROP DATABASE xin_dev; CREATE DATABASE xin_dev;
--   psql -d xin_dev -f init_schema.sql
--   psql -d xin_dev -f init_seed.sql
--   psql -d xin_dev -f asset.sql
--   psql -d xin_dev -f cms.sql
--   psql -d xin_dev -f flag.sql
-- ============================================

SET client_encoding = 'UTF8';
CREATE EXTENSION IF NOT EXISTS ltree;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ============================================
-- 1. 平台与租户公共表
-- ============================================
-- tenants / accounts / auth_sessions 不分域，跨域共享。
-- 平台超管走 sys_*，账号登录走 accounts + auth_sessions。

-- 1.1 tenants 租户表
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
COMMENT ON TABLE tenants IS '租户主表（不分域，跨域共享）';

-- 1.2 accounts 全局账号表（不分域，跨域共享）
-- accounts 是登录凭证层。sys_users（platform 域）和 tenant_users（tenant 域）
-- 都通过 account_id 外键引用 accounts。
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
COMMENT ON TABLE accounts IS '全局账号表（登录凭证，跨域共享）';

-- 1.3 auth_sessions 会话表（account_id，全局）
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
COMMENT ON TABLE auth_sessions IS '会话表（按 account_id，跨域共享）';

-- ============================================
-- 2. 租户域 tenant_* 表
-- ============================================
-- 全部带 tenant_id，启用 RLS（在第 7 节统一开启）。
-- 0023.3 物理表名清理：
--   users → tenant_users
--   roles → tenant_roles
--   organizations → tenant_organizations
--   user_roles → tenant_user_roles
--   role_menus → tenant_role_menus
--   role_resources → tenant_role_resources
--   role_data_scopes → tenant_role_data_scopes
--   resources → tenant_permissions（同时加 RLS；无 scope 字段）
--   menus（仅 tenant scope）→ tenant_menus（drop scope 字段）

-- 2.1 tenant_organizations 租户组织
CREATE TABLE IF NOT EXISTS tenant_organizations
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id   BIGINT      NOT NULL,
    parent_id   BIGINT,
    code        VARCHAR(32) NOT NULL,
    name        VARCHAR(64) NOT NULL,
    type        VARCHAR(32),
    description VARCHAR(255),
    admin_code  VARCHAR(64),
    ancestors   TEXT        DEFAULT '',
    sort        INT         DEFAULT 0,
    status      SMALLINT    DEFAULT 1,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    created_by  BIGINT,
    updated_by  BIGINT,
    is_deleted  BOOLEAN     DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_tenant_org_tenant ON tenant_organizations (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_tenant_org_parent ON tenant_organizations (parent_id) WHERE is_deleted = FALSE;
COMMENT ON TABLE tenant_organizations IS '租户域组织表';

-- 2.2 tenant_users 租户用户
CREATE TABLE IF NOT EXISTS tenant_users
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,
    account_id BIGINT      NOT NULL,
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
CREATE INDEX IF NOT EXISTS idx_tenant_users_tenant ON tenant_users (tenant_id) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_tenant_users_account_tenant
    ON tenant_users (account_id, tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_tenant_users_org ON tenant_users (org_id) WHERE is_deleted = FALSE;
COMMENT ON TABLE tenant_users IS '租户域用户表';

-- 2.3 tenant_roles 租户角色
CREATE TABLE IF NOT EXISTS tenant_roles
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
CREATE INDEX IF NOT EXISTS idx_tenant_roles_tenant ON tenant_roles (tenant_id) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_tenant_roles_code ON tenant_roles (tenant_id, code) WHERE is_deleted = FALSE;
COMMENT ON TABLE tenant_roles IS '租户域角色表';

-- 2.4 tenant_role_data_scopes 租户角色数据范围
CREATE TABLE IF NOT EXISTS tenant_role_data_scopes
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    role_id    BIGINT NOT NULL,
    org_id     BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_tenant_rds_role ON tenant_role_data_scopes (role_id) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_tenant_rds_unique ON tenant_role_data_scopes (role_id, org_id) WHERE is_deleted = FALSE;
COMMENT ON TABLE tenant_role_data_scopes IS '租户域角色数据范围表';

-- 2.5 tenant_user_roles 租户用户-角色
CREATE TABLE IF NOT EXISTS tenant_user_roles
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    user_id    BIGINT NOT NULL,
    role_id    BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_tenant_ur_user ON tenant_user_roles (user_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_tenant_ur_role ON tenant_user_roles (role_id) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_tenant_ur_unique ON tenant_user_roles (user_id, role_id) WHERE is_deleted = FALSE;
COMMENT ON TABLE tenant_user_roles IS '租户域用户-角色关联表';

-- 2.6 tenant_menus 租户菜单（仅 tenant scope，不再有 scope 字段）
-- 历史：menus 表用 scope='platform' / scope='tenant' 区分；0023.3 拆出 sys_menus
-- 后，tenant_menus 只承载租户域菜单，scope 字段 drop。
CREATE TABLE IF NOT EXISTS tenant_menus
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
CREATE UNIQUE INDEX IF NOT EXISTS uk_tenant_menu_code
    ON tenant_menus (tenant_id, code) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_tenant_menus_tenant ON tenant_menus (tenant_id) WHERE is_deleted = FALSE;
COMMENT ON TABLE tenant_menus IS '租户域菜单表（仅租户 scope）';

-- 2.7 tenant_permissions 租户权限码（原 resources）
-- 0023.3：resources → tenant_permissions 同时加 RLS。action 字段保留（read/list/create/update/delete）。
CREATE TABLE IF NOT EXISTS tenant_permissions
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
CREATE UNIQUE INDEX IF NOT EXISTS uk_tenant_permissions_code
    ON tenant_permissions (tenant_id, code) WHERE is_deleted = FALSE;
COMMENT ON TABLE tenant_permissions IS '租户域权限码表（原 resources，加 RLS）';

-- 2.8 tenant_role_menus 租户角色-菜单
CREATE TABLE IF NOT EXISTS tenant_role_menus
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    role_id    BIGINT NOT NULL,
    menu_id    BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_tenant_role_menu ON tenant_role_menus (role_id, menu_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_tenant_role_menus_tenant ON tenant_role_menus (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_tenant_role_menus_role ON tenant_role_menus (role_id) WHERE is_deleted = FALSE;
COMMENT ON TABLE tenant_role_menus IS '租户域角色-菜单关联表';

-- 2.9 tenant_role_resources 租户角色-权限码
-- 关联表名沿用 role_resources → tenant_role_resources（与原表名同后缀）。
-- 指向的权限码表是 tenant_permissions。
CREATE TABLE IF NOT EXISTS tenant_role_resources
(
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id     BIGINT NOT NULL,
    role_id       BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    effect        SMALLINT    DEFAULT 1,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW(),
    is_deleted    BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_tenant_role_resource
    ON tenant_role_resources (role_id, permission_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_tenant_role_resources_tenant ON tenant_role_resources (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_tenant_role_resources_role ON tenant_role_resources (role_id) WHERE is_deleted = FALSE;
COMMENT ON TABLE tenant_role_resources IS '租户域角色-权限码关联表（指向 tenant_permissions）';

-- 2.10 tenant_user_seq 租户用户序号
CREATE TABLE IF NOT EXISTS tenant_user_seq
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    seq        BIGINT      DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_tenant_seq ON tenant_user_seq (tenant_id);
COMMENT ON TABLE tenant_user_seq IS '租户域用户序号表';

-- ============================================
-- 3. 平台域 sys_* 表
-- ============================================
-- 全部不带 tenant_id（platform 是单租户概念）。
-- 不启用 RLS，靠 API 层 RequirePlatformRole(super_admin) + db.RunInPlatformTx 守护。

-- 3.1 sys_users 平台用户身份（对齐 tenant_users）
CREATE TABLE IF NOT EXISTS sys_users
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    account_id BIGINT      NOT NULL,
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
CREATE INDEX IF NOT EXISTS idx_sys_users_account
    ON sys_users (account_id) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_users_code
    ON sys_users (code) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_sys_users_org
    ON sys_users (org_id) WHERE is_deleted = FALSE;
COMMENT ON TABLE sys_users IS '平台域用户身份表（对齐 tenant_users，无 tenant_id）';

-- 3.2 sys_orgs 平台组织
CREATE TABLE IF NOT EXISTS sys_orgs
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    parent_id   BIGINT,
    code        VARCHAR(32) NOT NULL,
    name        VARCHAR(64) NOT NULL,
    type        VARCHAR(32),
    description VARCHAR(255),
    admin_code  VARCHAR(64),
    ancestors   TEXT             DEFAULT '',
    sort        INT              DEFAULT 0,
    status      SMALLINT         DEFAULT 1,
    created_at  TIMESTAMPTZ      DEFAULT NOW(),
    updated_at  TIMESTAMPTZ      DEFAULT NOW(),
    created_by  BIGINT,
    updated_by  BIGINT,
    is_deleted  BOOLEAN          DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_sys_orgs_parent
    ON sys_orgs (parent_id) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_orgs_code
    ON sys_orgs (code) WHERE is_deleted = FALSE;
COMMENT ON TABLE sys_orgs IS '平台域组织表（对齐 tenant_organizations）';

-- 3.3 sys_roles 平台角色
CREATE TABLE IF NOT EXISTS sys_roles
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    org_id      BIGINT,
    code        VARCHAR(32)  NOT NULL,
    name        VARCHAR(32)  NOT NULL,
    description VARCHAR(256),
    data_scope  SMALLINT     NOT NULL DEFAULT 1,
    extend      JSONB,
    is_default  BOOLEAN               DEFAULT FALSE,
    sort        INT                   DEFAULT 0,
    status      SMALLINT              DEFAULT 1,
    created_at  TIMESTAMPTZ           DEFAULT NOW(),
    updated_at  TIMESTAMPTZ           DEFAULT NOW(),
    created_by  BIGINT,
    updated_by  BIGINT,
    is_deleted  BOOLEAN               DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_sys_roles_org
    ON sys_roles (org_id) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_roles_code
    ON sys_roles (code) WHERE is_deleted = FALSE;
COMMENT ON TABLE sys_roles IS '平台域角色表（对齐 tenant_roles）';

-- 3.4 sys_menus 平台菜单
CREATE TABLE IF NOT EXISTS sys_menus
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code       VARCHAR(64),
    name       VARCHAR(64)  NOT NULL,
    subtitle   VARCHAR(128),
    url        VARCHAR(255),
    path       VARCHAR(255),
    icon       VARCHAR(64),
    sort       INT              DEFAULT 0,
    parent_id  BIGINT,
    ancestors  TEXT             DEFAULT '',
    visible    BOOLEAN          DEFAULT TRUE,
    enabled    BOOLEAN          DEFAULT TRUE,
    created_at TIMESTAMPTZ      DEFAULT NOW(),
    updated_at TIMESTAMPTZ      DEFAULT NOW(),
    created_by BIGINT,
    updated_by BIGINT,
    is_deleted BOOLEAN          DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_menus_code
    ON sys_menus (code) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_sys_menus_parent
    ON sys_menus (parent_id) WHERE is_deleted = FALSE;
COMMENT ON TABLE sys_menus IS '平台域菜单表（替代 menus WHERE scope=platform）';

-- 3.5 sys_permissions 平台权限码
CREATE TABLE IF NOT EXISTS sys_permissions
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    menu_id     BIGINT,
    code        VARCHAR(64)  NOT NULL,
    name        VARCHAR(64)  NOT NULL,
    action      VARCHAR(32)  DEFAULT 'read',
    description VARCHAR(512),
    sort        INT             DEFAULT 0,
    status      SMALLINT        DEFAULT 1,
    created_at  TIMESTAMPTZ     DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     DEFAULT NOW(),
    created_by  BIGINT,
    updated_by  BIGINT,
    is_deleted  BOOLEAN         DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_permissions_code
    ON sys_permissions (code) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_sys_permissions_menu
    ON sys_permissions (menu_id) WHERE is_deleted = FALSE;
COMMENT ON TABLE sys_permissions IS '平台域权限码表（替代 resources WHERE scope=platform）';

-- 3.6 sys_user_roles 平台用户-角色
-- 0023.3 终态：替代 account_roles 角色映射（account_roles 已 drop）。
-- boot/auth 登录路径：accounts → sys_users → sys_user_roles → sys_roles.code
CREATE TABLE IF NOT EXISTS sys_user_roles
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    BIGINT  NOT NULL,
    role_id    BIGINT  NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_sys_user_roles_user
    ON sys_user_roles (user_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_sys_user_roles_role
    ON sys_user_roles (role_id) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_user_roles
    ON sys_user_roles (user_id, role_id) WHERE is_deleted = FALSE;
COMMENT ON TABLE sys_user_roles IS '平台域用户-角色关联（终态，替代已 drop 的 account_roles）';

-- 3.7 sys_role_menus 平台角色-菜单
CREATE TABLE IF NOT EXISTS sys_role_menus
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    role_id    BIGINT  NOT NULL,
    menu_id    BIGINT  NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_role_menus
    ON sys_role_menus (role_id, menu_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_sys_role_menus_role
    ON sys_role_menus (role_id) WHERE is_deleted = FALSE;
COMMENT ON TABLE sys_role_menus IS '平台域角色-菜单关联';

-- 3.8 sys_role_permissions 平台角色-权限码
CREATE TABLE IF NOT EXISTS sys_role_permissions
(
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    role_id       BIGINT  NOT NULL,
    permission_id BIGINT  NOT NULL,
    effect        SMALLINT     DEFAULT 1,
    created_at    TIMESTAMPTZ  DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  DEFAULT NOW(),
    is_deleted    BOOLEAN      DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_role_permissions
    ON sys_role_permissions (role_id, permission_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_sys_role_permissions_role
    ON sys_role_permissions (role_id) WHERE is_deleted = FALSE;
COMMENT ON TABLE sys_role_permissions IS '平台域角色-权限码关联';

-- ============================================
-- 4. 业务支撑表（订阅 / 套餐 / 用量 / 日志）
-- ============================================
-- 全部带 tenant_id（按租户分账），启用 RLS。

-- 4.1 subscriptions 订阅
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
COMMENT ON TABLE subscriptions IS '订阅表（按租户）';

-- 4.2 plans 套餐
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
COMMENT ON TABLE plans IS '套餐表（全局）';

-- 4.3 usage_records 用量记录
CREATE TABLE IF NOT EXISTS usage_records
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,
    metric     VARCHAR(32) NOT NULL,
    value      BIGINT      DEFAULT 0,
    period     VARCHAR(20),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_usage_tenant ON usage_records (tenant_id);
COMMENT ON TABLE usage_records IS '用量记录表（按租户）';

-- 4.4 db_logs 数据库变更日志
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
COMMENT ON TABLE db_logs IS '数据库变更日志表（按租户）';

-- ============================================
-- 5. 字典 / 配置 / 可见性
-- ============================================
-- 字典与配置同构：tenant_id=0 是平台级（跨租户共享），>0 是租户级。
-- is_override + platform_item_id 支持租户覆盖平台项。
-- dict_visibility / config_visibility 矩阵控平台对租户的可见性。

-- 5.1 dicts 字典主表
CREATE TABLE IF NOT EXISTS dicts
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,
    code       VARCHAR(32) NOT NULL,
    name       VARCHAR(64) NOT NULL,
    scope      VARCHAR(16) NOT NULL DEFAULT 'tenant',
    visibility VARCHAR(16) NOT NULL DEFAULT 'all',
    status     SMALLINT    DEFAULT 1,
    sort       INT         DEFAULT 0,
    extend     JSONB       DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_dict_code_platform
    ON dicts (code) WHERE tenant_id = 0 AND is_deleted = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_dict_code_tenant
    ON dicts (tenant_id, code) WHERE tenant_id <> 0 AND is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_dicts_tenant ON dicts (tenant_id);
CREATE INDEX IF NOT EXISTS idx_dicts_scope ON dicts (scope) WHERE is_deleted = FALSE;
COMMENT ON TABLE dicts IS '字典主表（tenant_id=0 平台级，>0 租户级）';

-- 5.2 dict_items 字典项
CREATE TABLE IF NOT EXISTS dict_items
(
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id        BIGINT       NOT NULL,
    dict_id          BIGINT       NOT NULL,
    code             VARCHAR(64)  NOT NULL,
    name             VARCHAR(128) NOT NULL,
    scope            VARCHAR(16)  NOT NULL DEFAULT 'tenant',
    platform_item_id BIGINT,
    is_override      BOOLEAN      NOT NULL DEFAULT FALSE,
    sort             INT          DEFAULT 0,
    status           SMALLINT     DEFAULT 1,
    extend           JSONB        DEFAULT '{}'::jsonb,
    created_at       TIMESTAMPTZ  DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  DEFAULT NOW(),
    is_deleted       BOOLEAN      DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_dict_item_platform
    ON dict_items (dict_id, code) WHERE tenant_id = 0 AND is_deleted = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_dict_item_tenant
    ON dict_items (tenant_id, dict_id, code)
    WHERE tenant_id <> 0 AND is_deleted = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_dict_item_override
    ON dict_items (tenant_id, platform_item_id)
    WHERE is_override = TRUE AND is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_dict_items_dict ON dict_items (dict_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_dict_items_tenant ON dict_items (tenant_id);
CREATE INDEX IF NOT EXISTS idx_dict_items_platform_ref
    ON dict_items (dict_id, id) WHERE tenant_id = 0 AND is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_dict_items_scope
    ON dict_items (scope) WHERE is_deleted = FALSE;
COMMENT ON TABLE dict_items IS '字典项表（tenant_id=0 平台级，>0 租户级）';

-- 5.3 dict_visibility 字典可见性矩阵
CREATE TABLE IF NOT EXISTS dict_visibility
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    dict_id    BIGINT      NOT NULL,
    tenant_id  BIGINT      NOT NULL,
    access     VARCHAR(16) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT uk_dict_visibility UNIQUE (dict_id, tenant_id)
);
CREATE INDEX IF NOT EXISTS idx_dict_visibility_tenant ON dict_visibility (tenant_id);
CREATE INDEX IF NOT EXISTS idx_dict_visibility_dict   ON dict_visibility (dict_id);
COMMENT ON TABLE dict_visibility IS '字典可见性矩阵';

-- 5.4 config_categories 配置分组
CREATE TABLE IF NOT EXISTS config_categories
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id   BIGINT       NOT NULL,
    code        VARCHAR(64)  NOT NULL,
    name        VARCHAR(64)  NOT NULL,
    description VARCHAR(255),
    icon        VARCHAR(64),
    sort        INT          DEFAULT 0,
    scope       VARCHAR(16)  NOT NULL DEFAULT 'tenant',
    is_system   BOOLEAN      DEFAULT FALSE,
    is_public   BOOLEAN      DEFAULT FALSE,
    visibility  VARCHAR(16)  NOT NULL DEFAULT 'all',
    status      SMALLINT     DEFAULT 1,
    extend      JSONB,
    created_at  TIMESTAMPTZ  DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  DEFAULT NOW(),
    is_deleted  BOOLEAN      DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_config_group_code_platform
    ON config_categories (code) WHERE tenant_id = 0 AND is_deleted = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_config_group_code_tenant
    ON config_categories (tenant_id, code) WHERE tenant_id <> 0 AND is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_config_groups_tenant ON config_categories (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_config_categories_scope
    ON config_categories (scope) WHERE is_deleted = FALSE;
COMMENT ON TABLE config_categories IS '配置分组表';

-- 5.5 config_items 配置项
CREATE TABLE IF NOT EXISTS config_items
(
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id        BIGINT       NOT NULL,
    category_id      BIGINT       NOT NULL,
    key              VARCHAR(128) NOT NULL,
    value            JSONB        DEFAULT NULL,
    default_value    JSONB        DEFAULT NULL,
    type             VARCHAR(32)  NOT NULL,
    label            VARCHAR(128),
    description      VARCHAR(512),
    options          JSONB        DEFAULT NULL,
    validation       JSONB        DEFAULT NULL,
    platform_item_id BIGINT,
    is_override      BOOLEAN      NOT NULL DEFAULT FALSE,
    sort             INT          DEFAULT 0,
    is_public        BOOLEAN      DEFAULT FALSE,
    is_readonly      BOOLEAN      DEFAULT FALSE,
    is_system        BOOLEAN      DEFAULT FALSE,
    status           SMALLINT     DEFAULT 1,
    created_at       TIMESTAMPTZ  DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  DEFAULT NOW(),
    is_deleted       BOOLEAN      DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_config_item_key_platform
    ON config_items (category_id, key) WHERE tenant_id = 0 AND is_deleted = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_config_item_key_tenant
    ON config_items (tenant_id, category_id, key)
    WHERE tenant_id <> 0 AND is_deleted = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_config_item_override
    ON config_items (tenant_id, platform_item_id)
    WHERE is_override = TRUE AND is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_config_items_category ON config_items (category_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_config_items_tenant ON config_items (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_config_items_platform_ref
    ON config_items (category_id, id) WHERE tenant_id = 0 AND is_deleted = FALSE;
COMMENT ON TABLE config_items IS '配置项表';

-- 5.6 config_visibility 配置可见性矩阵
CREATE TABLE IF NOT EXISTS config_visibility
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    category_id BIGINT      NOT NULL,
    tenant_id   BIGINT      NOT NULL,
    access      VARCHAR(16) NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT uk_config_visibility UNIQUE (category_id, tenant_id)
);
CREATE INDEX IF NOT EXISTS idx_config_visibility_tenant ON config_visibility (tenant_id);
CREATE INDEX IF NOT EXISTS idx_config_visibility_category  ON config_visibility (category_id);
COMMENT ON TABLE config_visibility IS '配置可见性矩阵';

-- ============================================
-- 6. 路由
-- ============================================
-- routes 携带 tenant_id（按租户配置路由），启用 RLS。
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
COMMENT ON TABLE routes IS '路由表（按租户）';

-- ============================================
-- 7. RLS 策略（行级安全 — 纵深防御层）
-- ============================================
-- 7.1 租户域 tenant_* 全部启用 RLS
--     policy 表达式：
--       tenant_id 匹配 app.tenant_id  OR  app.bypass_rls = 'on'
--     RunInPlatformTx 设置 app.bypass_rls='on' 后能跨租户访问，
--     RunInTenantTx 仍只看本租户数据。

ALTER TABLE tenant_organizations     ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_users             ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_roles             ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_role_data_scopes  ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_user_roles        ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_menus             ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_permissions       ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_role_menus        ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_role_resources    ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_user_seq          ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_organizations;
CREATE POLICY tenant_isolation_policy ON tenant_organizations USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_users;
CREATE POLICY tenant_isolation_policy ON tenant_users USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_roles;
CREATE POLICY tenant_isolation_policy ON tenant_roles USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_role_data_scopes;
CREATE POLICY tenant_isolation_policy ON tenant_role_data_scopes USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_user_roles;
CREATE POLICY tenant_isolation_policy ON tenant_user_roles USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_menus;
CREATE POLICY tenant_isolation_policy ON tenant_menus USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_permissions;
CREATE POLICY tenant_isolation_policy ON tenant_permissions USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_role_menus;
CREATE POLICY tenant_isolation_policy ON tenant_role_menus USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_role_resources;
CREATE POLICY tenant_isolation_policy ON tenant_role_resources USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_user_seq;
CREATE POLICY tenant_isolation_policy ON tenant_user_seq USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

-- 7.2 业务支撑表（按租户分账）
ALTER TABLE subscriptions  ENABLE ROW LEVEL SECURITY;
ALTER TABLE usage_records   ENABLE ROW LEVEL SECURITY;
ALTER TABLE db_logs         ENABLE ROW LEVEL SECURITY;
ALTER TABLE routes          ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON subscriptions;
CREATE POLICY tenant_isolation_policy ON subscriptions USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

DROP POLICY IF EXISTS tenant_isolation_policy ON usage_records;
CREATE POLICY tenant_isolation_policy ON usage_records USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

DROP POLICY IF EXISTS tenant_isolation_policy ON db_logs;
CREATE POLICY tenant_isolation_policy ON db_logs USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

DROP POLICY IF EXISTS tenant_isolation_policy ON routes;
CREATE POLICY tenant_isolation_policy ON routes USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

-- 7.3 字典与配置（保留 tenant_id=0 短路：系统级字典/配置跨租户共享）
ALTER TABLE dicts             ENABLE ROW LEVEL SECURITY;
ALTER TABLE dict_items        ENABLE ROW LEVEL SECURITY;
ALTER TABLE config_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE config_items      ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON dicts;
CREATE POLICY tenant_isolation_policy ON dicts USING (
    tenant_id = 0
    OR tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

DROP POLICY IF EXISTS tenant_isolation_policy ON dict_items;
CREATE POLICY tenant_isolation_policy ON dict_items USING (
    tenant_id = 0
    OR tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

DROP POLICY IF EXISTS tenant_isolation_policy ON config_categories;
CREATE POLICY tenant_isolation_policy ON config_categories USING (
    tenant_id = 0
    OR tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

DROP POLICY IF EXISTS tenant_isolation_policy ON config_items;
CREATE POLICY tenant_isolation_policy ON config_items USING (
    tenant_id = 0
    OR tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

-- ============================================
-- 8. 完整性校验（dev 阶段，确保所有表都建好）
-- ============================================

DO $$
DECLARE
    expected_tables TEXT[] := ARRAY[
        -- 平台与租户公共
        'tenants', 'accounts', 'auth_sessions',
        -- 租户域 tenant_*（0023.3 终态）
        'tenant_organizations', 'tenant_users', 'tenant_roles',
        'tenant_role_data_scopes', 'tenant_user_roles',
        'tenant_menus', 'tenant_permissions',
        'tenant_role_menus', 'tenant_role_resources', 'tenant_user_seq',
        -- 平台域 sys_*
        'sys_users', 'sys_orgs', 'sys_roles', 'sys_menus', 'sys_permissions',
        'sys_user_roles', 'sys_role_menus', 'sys_role_permissions',
        -- 业务支撑
        'subscriptions', 'plans', 'usage_records', 'db_logs',
        -- 字典与配置
        'dicts', 'dict_items', 'dict_visibility',
        'config_categories', 'config_items', 'config_visibility',
        -- 路由
        'routes'
    ];
    t TEXT;
    missing TEXT := '';
    obsolete TEXT := '';
    obsolete_tables TEXT[] := ARRAY[
        'users', 'roles', 'organizations', 'user_roles', 'role_menus',
        'role_resources', 'role_data_scopes', 'resources', 'account_roles'
    ];
BEGIN
    FOREACH t IN ARRAY expected_tables LOOP
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.tables
            WHERE table_schema = 'public' AND table_name = t
        ) THEN
            missing := missing || t || ', ';
        END IF;
    END LOOP;
    FOREACH t IN ARRAY obsolete_tables LOOP
        IF EXISTS (
            SELECT 1 FROM information_schema.tables
            WHERE table_schema = 'public' AND table_name = t
        ) THEN
            obsolete := obsolete || t || ', ';
        END IF;
    END LOOP;
    IF missing <> '' THEN
        RAISE EXCEPTION 'init_schema 校验失败：缺失表 %', missing;
    END IF;
    IF obsolete <> '' THEN
        RAISE EXCEPTION 'init_schema 校验失败：遗留旧表 %（0023.3 已 drop/rename）', obsolete;
    END IF;
    RAISE NOTICE 'init_schema 校验通过：% 张表已建，旧表已清', array_length(expected_tables, 1);
END $$;
