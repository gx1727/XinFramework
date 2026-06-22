-- ============================================
-- platform_split.sql : 平台域 / 租户域 数据分域
-- ============================================
-- 目的：
--   把现在共用 + scope 字段区分的"混居"表，拆成两个域各自独立的表。
--   平台域（sys_*）不启用 RLS；租户域（tenant_*）启用 RLS（已有）。
--   关键设计：sys_users 字段对齐 tenant_users，引入 accounts 作为全局登录凭证。
--
-- 三层抽象：
--   accounts  — 全局登录凭证（不分域，跨域共享）
--   sys_*     — 平台域身份实体（无 tenant_id）
--   tenant_*  — 租户域身份实体（带 tenant_id，启用 RLS）
--
-- 依赖：
--   - framework.sql 必先跑（提供 accounts / tenants）
--   - framework_account_roles_independent.sql 必先跑（提供 account_roles 现状）
--
-- 兼容性：
--   - 本文件只做"创建新表 + 注释"；不动旧表（users/roles/menus/resources/organizations）
--   - 旧表数据的迁移、双写 trigger 在后续增量迁移里做
--   - 全部 DDL 幂等（IF NOT EXISTS）
-- ============================================

-- ============================================
-- 1. sys_users（对齐 tenant_users / 原 users）
-- ============================================
-- 关键差异：没有 tenant_id（platform 是单租户概念）
-- account_id 关联全局 accounts（一个账号一个 platform 身份）
-- org_id 关联 sys_orgs
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
COMMENT ON TABLE sys_users IS '平台域用户身份表（对齐 tenant_users）';

-- ============================================
-- 2. sys_orgs（对齐 organizations）
-- ============================================
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
COMMENT ON TABLE sys_orgs IS '平台域组织表（对齐 organizations）';

-- ============================================
-- 3. sys_roles（对齐 roles）
-- ============================================
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
COMMENT ON TABLE sys_roles IS '平台域角色表（对齐 roles）';

-- ============================================
-- 4. sys_menus（对齐 menus，剥离 scope='platform' 行）
-- ============================================
-- 没有 tenant_id（platform 是单租户）
-- 没有 scope 字段（与 tenant_menus 物理分离，不再需要 scope 标记）
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
COMMENT ON TABLE sys_menus IS '平台域菜单表（对齐 menus，剥离 platform scope）';

-- ============================================
-- 5. sys_permissions（对齐 resources，剥离 scope='platform' 行）
-- ============================================
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
COMMENT ON TABLE sys_permissions IS '平台域权限码表（对齐 resources，剥离 platform scope）';

-- ============================================
-- 6. sys_user_roles（替代 account_roles）
-- ============================================
-- account_roles 是过渡产物：(account_id, role string) 白名单
-- sys_user_roles 是终态：(user_id, role_id) FK，关联 sys_users + sys_roles
-- 注：seed 时同步给 admin 账号创建 sys_user，再绑 super_admin 角色
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
COMMENT ON TABLE sys_user_roles IS '平台域用户-角色关联（终态，替代 account_roles）';

-- ============================================
-- 7. sys_role_menus（对齐 role_menus）
-- ============================================
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
COMMENT ON TABLE sys_role_menus IS '平台域角色-菜单关联（对齐 role_menus）';

-- ============================================
-- 8. sys_role_permissions（对齐 role_resources）
-- ============================================
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
COMMENT ON TABLE sys_role_permissions IS '平台域角色-权限码关联（对齐 role_resources）';

-- ============================================
-- 9. RLS 决策注释（重要）
-- ============================================
-- 平台域 sys_* 全部不启用 RLS。
--   原因：platform 是单租户概念，sys_users 等表没有 tenant_id 字段，
--         所有 super_admin 账号都允许看所有平台数据（否则 super_admin 没意义）。
--   替代安全措施：API 层用 RequirePlatformRole(super_admin) 中间件 + db.RunInPlatformTx。
-- 租户域 tenant_* 启用 RLS（在 tenant_split.sql 里做，物理重命名 + 索引重建）。
-- ============================================

-- ============================================
-- 10. 完整性校验（迁移后应全绿）
-- ============================================
DO $$
DECLARE
    expected_tables TEXT[] := ARRAY[
        'sys_users', 'sys_orgs', 'sys_roles', 'sys_menus', 'sys_permissions',
        'sys_user_roles', 'sys_role_menus', 'sys_role_permissions'
    ];
    t TEXT;
    missing TEXT := '';
BEGIN
    FOREACH t IN ARRAY expected_tables LOOP
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.tables
            WHERE table_schema = 'public' AND table_name = t
        ) THEN
            missing := missing || t || ', ';
        END IF;
    END LOOP;
    IF missing <> '' THEN
        RAISE EXCEPTION 'platform_split 校验失败：缺失表 %', missing;
    END IF;
    RAISE NOTICE 'platform_split 校验通过：8 张平台域表已建';
END $$;
