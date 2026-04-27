-- ============================================
-- MI7Soft-One 数据库初始化脚本 (PostgreSQL 商用版 v2.0)
-- 版本: PostgreSQL 14+
-- 特性: 部分唯一索引 | RLS行级安全 | JSONB+GIN | 审计溯源
-- 生成时间: 2026-04-26
-- 变更: 删除tenant_users | accounts字段NOT NULL | roles.scope_orgs独立表 | users去除冗余字段 | user_roles添加FK
-- ============================================

SET client_encoding = 'UTF8';

-- 启用 ltree 扩展以支持高效的树形结构存储和查询
CREATE EXTENSION IF NOT EXISTS ltree;

-- 启用 pg_trgm 扩展以支持 ILIKE 模糊搜索优化
CREATE EXTENSION IF NOT EXISTS pg_trgm;

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
DO $$
BEGIN
CREATE INDEX IF NOT EXISTS idx_tenants_name_trgm ON tenants USING gin (name gin_trgm_ops);
EXCEPTION WHEN insufficient_privilege THEN
    RAISE NOTICE 'Skipped idx_tenants_name_trgm: no privileges on tenants table';
END $$;
DO $$
BEGIN
CREATE INDEX IF NOT EXISTS idx_tenants_code_trgm ON tenants USING gin (code gin_trgm_ops);
EXCEPTION WHEN insufficient_privilege THEN
    RAISE NOTICE 'Skipped idx_tenants_code_trgm: no privileges on tenants table';
END $$;

COMMENT ON TABLE tenants IS '租户表 - SaaS多租户核心表';
COMMENT ON COLUMN tenants.id IS '租户ID';
COMMENT ON COLUMN tenants.code IS '租户编码，全局唯一';
COMMENT ON COLUMN tenants.name IS '租户名称';
COMMENT ON COLUMN tenants.status IS '租户状态：0-禁用，1-启用';
COMMENT ON COLUMN tenants.contact IS '联系人';
COMMENT ON COLUMN tenants.phone IS '联系电话';
COMMENT ON COLUMN tenants.email IS '联系邮箱';
COMMENT ON COLUMN tenants.province IS '省份';
COMMENT ON COLUMN tenants.city IS '城市';
COMMENT ON COLUMN tenants.area IS '区县';
COMMENT ON COLUMN tenants.address IS '详细地址';
COMMENT ON COLUMN tenants.config IS '租户配置信息（JSONB）';
COMMENT ON COLUMN tenants.dashboard IS '默认仪表盘';
COMMENT ON COLUMN tenants.created_at IS '创建时间';
COMMENT ON COLUMN tenants.updated_at IS '更新时间';
COMMENT ON COLUMN tenants.created_by IS '创建人ID';
COMMENT ON COLUMN tenants.updated_by IS '更新人ID';
COMMENT ON COLUMN tenants.is_deleted IS '逻辑删除标记';

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
    status     SMALLINT DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX uk_accounts_phone ON accounts (phone) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX uk_accounts_email ON accounts (email) WHERE is_deleted = FALSE;
CREATE INDEX idx_accounts_username ON accounts (username) WHERE is_deleted = FALSE;
CREATE INDEX idx_accounts_status ON accounts (status) WHERE is_deleted = FALSE;

COMMENT ON TABLE accounts IS '全局账号表 - 跨租户统一账号管理';
COMMENT ON COLUMN accounts.id IS '账号ID';
COMMENT ON COLUMN accounts.phone IS '手机号';
COMMENT ON COLUMN accounts.email IS '邮箱地址';
COMMENT ON COLUMN accounts.password IS '密码（加密存储）';
COMMENT ON COLUMN accounts.username IS '用户名';
COMMENT ON COLUMN accounts.real_name IS '真实姓名';
COMMENT ON COLUMN accounts.avatar IS '头像URL';
COMMENT ON COLUMN accounts.status IS '账号状态：0-禁用，1-启用';
COMMENT ON COLUMN accounts.created_at IS '创建时间';
COMMENT ON COLUMN accounts.updated_at IS '更新时间';
COMMENT ON COLUMN accounts.is_deleted IS '逻辑删除标记';

-- 3. account_auths (第三方授权表)
DROP TABLE IF EXISTS account_auths;
CREATE TABLE account_auths
(
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id    BIGINT      NOT NULL,
    account_id   BIGINT      NOT NULL,
    type         VARCHAR(32) NOT NULL,
    config       VARCHAR(32),
    openid       VARCHAR(64),
    unionid      VARCHAR(64),
    nickname     VARCHAR(256),
    avatar       VARCHAR(512),
    sex          SMALLINT,
    city         VARCHAR(64),
    province     VARCHAR(64),
    country      VARCHAR(64),
    subscribe    BOOLEAN     DEFAULT FALSE,
    subscribe_at TIMESTAMPTZ,
    session_key  VARCHAR(64),
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW(),
    is_deleted   BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX uk_auth_openid ON account_auths (tenant_id, type, openid) WHERE is_deleted = FALSE;
CREATE INDEX idx_auth_account ON account_auths (account_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_auth_unionid ON account_auths (unionid) WHERE is_deleted = FALSE AND unionid IS NOT NULL;

COMMENT ON TABLE account_auths IS '第三方授权表 - 微信、QQ等OAuth授权信息';
COMMENT ON COLUMN account_auths.id IS '授权记录ID';
COMMENT ON COLUMN account_auths.tenant_id IS '租户ID';
COMMENT ON COLUMN account_auths.account_id IS '关联账号ID';
COMMENT ON COLUMN account_auths.type IS '授权类型：wechat, qq, weibo等';
COMMENT ON COLUMN account_auths.openid IS '第三方OpenID';
COMMENT ON COLUMN account_auths.unionid IS '第三方UnionID';
COMMENT ON COLUMN account_auths.nickname IS '第三方昵称';
COMMENT ON COLUMN account_auths.avatar IS '头像URL';
COMMENT ON COLUMN account_auths.sex IS '性别：0-未知，1-男，2-女';
COMMENT ON COLUMN account_auths.subscribe IS '是否订阅公众号/服务号';
COMMENT ON COLUMN account_auths.subscribe_at IS '订阅时间';
COMMENT ON COLUMN account_auths.session_key IS '会话密钥（小程序）';

-- 4. organizations (机构表)
DROP TABLE IF EXISTS organizations;
CREATE TABLE organizations
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id   BIGINT       NOT NULL,
    code        VARCHAR(32)  NOT NULL,
    name        VARCHAR(100) NOT NULL,
    type        VARCHAR(32)  NOT NULL,
    description VARCHAR(512),
    admin_code  VARCHAR(32),
    parent_id   BIGINT,
    ancestors   ltree,
    sort        INT         DEFAULT 0,
    status      SMALLINT    DEFAULT 1,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    created_by  BIGINT,
    updated_by  BIGINT,
    is_deleted  BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX uk_org_code ON organizations (tenant_id, code) WHERE is_deleted = FALSE;
CREATE INDEX idx_org_tenant ON organizations (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_org_parent ON organizations (parent_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_org_ancestors_gist ON organizations USING GIST (ancestors);

COMMENT ON TABLE organizations IS '组织机构表 - 树形结构';
COMMENT ON COLUMN organizations.id IS '机构ID';
COMMENT ON COLUMN organizations.tenant_id IS '租户ID';
COMMENT ON COLUMN organizations.code IS '机构编码';
COMMENT ON COLUMN organizations.name IS '机构名称';
COMMENT ON COLUMN organizations.type IS '机构类型：department-部门，company-公司';
COMMENT ON COLUMN organizations.description IS '机构描述';
COMMENT ON COLUMN organizations.admin_code IS '管理员编码';
COMMENT ON COLUMN organizations.parent_id IS '父机构ID';
COMMENT ON COLUMN organizations.ancestors IS '祖先节点路径(ltree格式，如: 1.2.3)';
COMMENT ON COLUMN organizations.sort IS '排序号';
COMMENT ON COLUMN organizations.status IS '机构状态：0-禁用，1-启用';

-- 5. users (租户用户表) - 去除冗余字段
DROP TABLE IF EXISTS users;
CREATE TABLE users
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id   BIGINT      NOT NULL,
    account_id  BIGINT      NOT NULL,
    org_id      BIGINT,
    code        VARCHAR(32) NOT NULL,
    nickname    VARCHAR(100),
    real_name   VARCHAR(64),
    phone       VARCHAR(20),
    email       VARCHAR(100),
    avatar      VARCHAR(512),
    status      SMALLINT    DEFAULT 1,
    parent_code VARCHAR(32),
    ancestors   ltree,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    created_by  BIGINT,
    updated_by  BIGINT,
    is_deleted  BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX uk_user_code ON users (tenant_id, code) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX uk_user_tenant_account ON users (tenant_id, account_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_user_tenant ON users (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_user_org ON users (org_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_user_ancestors_gist ON users USING GIST (ancestors);
DO $$
BEGIN
CREATE INDEX IF NOT EXISTS idx_users_code_trgm ON users USING gin (code gin_trgm_ops);
EXCEPTION WHEN insufficient_privilege THEN
    RAISE NOTICE 'Skipped idx_users_code_trgm: no privileges on users table';
END $$;
DO $$
BEGIN
CREATE INDEX IF NOT EXISTS idx_users_real_name_trgm ON users USING gin (real_name gin_trgm_ops);
EXCEPTION WHEN insufficient_privilege THEN
    RAISE NOTICE 'Skipped idx_users_real_name_trgm: no privileges on users table';
END $$;
DO $$
BEGIN
CREATE INDEX IF NOT EXISTS idx_users_phone_trgm ON users USING gin (phone gin_trgm_ops);
EXCEPTION WHEN insufficient_privilege THEN
    RAISE NOTICE 'Skipped idx_users_phone_trgm: no privileges on users table';
END $$;
DO $$
BEGIN
CREATE INDEX IF NOT EXISTS idx_users_nickname_trgm ON users USING gin (nickname gin_trgm_ops);
EXCEPTION WHEN insufficient_privilege THEN
    RAISE NOTICE 'Skipped idx_users_nickname_trgm: no privileges on users table';
END $$;

COMMENT ON TABLE users IS '租户用户表 - 租户内的用户信息';
COMMENT ON COLUMN users.id IS '用户ID';
COMMENT ON COLUMN users.tenant_id IS '租户ID';
COMMENT ON COLUMN users.account_id IS '关联全局账号ID';
COMMENT ON COLUMN users.org_id IS '所属机构ID';
COMMENT ON COLUMN users.code IS '用户编码';
COMMENT ON COLUMN users.nickname IS '用户昵称/显示名（优先级高于real_name）';
COMMENT ON COLUMN users.real_name IS '真实姓名（冗余自accounts，仅用于租户内展示）';
COMMENT ON COLUMN users.phone IS '手机号（冗余自accounts）';
COMMENT ON COLUMN users.email IS '邮箱（冗余自accounts）';
COMMENT ON COLUMN users.avatar IS '头像URL（冗余自accounts）';
COMMENT ON COLUMN users.status IS '用户状态：0-禁用，1-启用';
COMMENT ON COLUMN users.parent_code IS '上级用户编码';
COMMENT ON COLUMN users.ancestors IS '祖先用户路径(ltree格式，使用code构建，如: root.admin.test)';

-- 6. roles (角色表) - 移除scope_orgs JSONB，改为独立表
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
CREATE UNIQUE INDEX uk_role_code ON roles (tenant_id, code) WHERE is_deleted = FALSE;
CREATE INDEX idx_role_tenant ON roles (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_role_org ON roles (org_id) WHERE is_deleted = FALSE;

COMMENT ON TABLE roles IS '角色表 - RBAC权限模型';
COMMENT ON COLUMN roles.id IS '角色ID';
COMMENT ON COLUMN roles.tenant_id IS '租户ID';
COMMENT ON COLUMN roles.org_id IS '所属机构ID';
COMMENT ON COLUMN roles.code IS '角色编码';
COMMENT ON COLUMN roles.name IS '角色名称';
COMMENT ON COLUMN roles.description IS '角色描述';
COMMENT ON COLUMN roles.data_scope IS '数据权限范围：1-全部，2-自定义，3-本部门，4-本部门及以下，5-仅本人';
COMMENT ON COLUMN roles.extend IS '扩展信息（JSONB）';
COMMENT ON COLUMN roles.is_default IS '是否默认角色';
COMMENT ON COLUMN roles.sort IS '排序号';
COMMENT ON COLUMN roles.status IS '角色状态：0-禁用，1-启用';

-- 6.1 role_data_scopes (角色数据权限表) - 替代原来的 scope_orgs JSONB
DROP TABLE IF EXISTS role_data_scopes;
CREATE TABLE role_data_scopes
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    role_id    BIGINT NOT NULL,
    org_id     BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE UNIQUE INDEX uk_role_data_scope ON role_data_scopes (tenant_id, role_id, org_id);
CREATE INDEX idx_rds_tenant ON role_data_scopes (tenant_id);
CREATE INDEX idx_rds_role ON role_data_scopes (role_id);
CREATE INDEX idx_rds_org ON role_data_scopes (org_id);

COMMENT ON TABLE role_data_scopes IS '角色数据权限表 - 替代scope_orgs JSONB，支持多机构权限';
COMMENT ON COLUMN role_data_scopes.id IS '记录ID';
COMMENT ON COLUMN role_data_scopes.tenant_id IS '租户ID';
COMMENT ON COLUMN role_data_scopes.role_id IS '角色ID';
COMMENT ON COLUMN role_data_scopes.org_id IS '授权机构ID';

-- 7. user_roles (用户角色关联表) - 已有FK约束
DROP TABLE IF EXISTS user_roles;
CREATE TABLE user_roles
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    user_id    BIGINT NOT NULL,
    role_id    BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX uk_user_role ON user_roles (user_id, role_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_ur_tenant ON user_roles (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_ur_role ON user_roles (role_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_user_roles_user_active
    ON user_roles (user_id) WHERE is_deleted = FALSE;


COMMENT ON TABLE user_roles IS '用户角色关联表 - 多对多关系';
COMMENT ON COLUMN user_roles.id IS '关联ID';
COMMENT ON COLUMN user_roles.tenant_id IS '租户ID';
COMMENT ON COLUMN user_roles.user_id IS '用户ID';
COMMENT ON COLUMN user_roles.role_id IS '角色ID';

-- 8. menus (菜单表)
DROP TABLE IF EXISTS menus;
CREATE TABLE menus
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,
    code       VARCHAR(64),
    name       VARCHAR(64) NOT NULL,
    subtitle   VARCHAR(64),
    url        VARCHAR(256),
    path       VARCHAR(256),
    icon       VARCHAR(64),
    sort       INT         DEFAULT 1024,
    parent_id  BIGINT,
    ancestors  ltree,
    visible    BOOLEAN     DEFAULT TRUE,
    enabled    BOOLEAN     DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted  BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX uk_menu_code ON menus (tenant_id, code) WHERE is_deleted = FALSE;
CREATE INDEX idx_menu_tenant ON menus (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_menu_ancestors_gist ON menus USING GIST (ancestors);
CREATE INDEX IF NOT EXISTS idx_menus_tenant_active
    ON menus (tenant_id) WHERE is_deleted = FALSE;

COMMENT ON TABLE menus IS '菜单表 - 前端导航菜单';
COMMENT ON COLUMN menus.id IS '菜单ID';
COMMENT ON COLUMN menus.tenant_id IS '租户ID';
COMMENT ON COLUMN menus.code IS '菜单编码';
COMMENT ON COLUMN menus.name IS '菜单名称';
COMMENT ON COLUMN menus.subtitle IS '副标题';
COMMENT ON COLUMN menus.url IS '菜单URL';
COMMENT ON COLUMN menus.path IS '路由路径';
COMMENT ON COLUMN menus.icon IS '图标';
COMMENT ON COLUMN menus.sort IS '排序号';
COMMENT ON COLUMN menus.parent_id IS '父菜单ID';
COMMENT ON COLUMN menus.ancestors IS '祖先节点路径(ltree格式，如: 1.2.3)';
COMMENT ON COLUMN menus.visible IS '是否显示';
COMMENT ON COLUMN menus.enabled IS '是否启用';

-- 9. resources (资源/按钮权限表)
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
    is_deleted  BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX uk_resource_code ON resources (tenant_id, code) WHERE is_deleted = FALSE;
CREATE INDEX idx_resource_tenant ON resources (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_resource_menu ON resources (menu_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_resource_action ON resources (action) WHERE is_deleted = FALSE;

COMMENT ON TABLE resources IS '资源/按钮权限表 - 细粒度权限控制';
COMMENT ON COLUMN resources.id IS '资源ID';
COMMENT ON COLUMN resources.tenant_id IS '租户ID';
COMMENT ON COLUMN resources.menu_id IS '所属菜单ID';
COMMENT ON COLUMN resources.code IS '资源编码';
COMMENT ON COLUMN resources.name IS '资源名称';
COMMENT ON COLUMN resources.action IS '操作类型：create/read/update/delete/list/export';
COMMENT ON COLUMN resources.description IS '资源描述';
COMMENT ON COLUMN resources.sort IS '排序号';
COMMENT ON COLUMN resources.status IS '资源状态：0-禁用，1-启用';

-- 10. routes (路由表)
DROP TABLE IF EXISTS routes;
CREATE TABLE routes
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id   BIGINT NOT NULL,
    code        VARCHAR(64),
    name        VARCHAR(64),
    method      VARCHAR(10) DEFAULT 'GET',
    url         VARCHAR(256),
    path        VARCHAR(256),
    description VARCHAR(512),
    auth_type   SMALLINT    DEFAULT 1,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    is_deleted  BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX uk_route_code ON routes (tenant_id, code) WHERE is_deleted = FALSE;
CREATE INDEX idx_route_tenant ON routes (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_route_method ON routes (method) WHERE is_deleted = FALSE;

COMMENT ON TABLE routes IS '路由表 - API路由权限控制';
COMMENT ON COLUMN routes.id IS '路由ID';
COMMENT ON COLUMN routes.tenant_id IS '租户ID';
COMMENT ON COLUMN routes.code IS '路由编码';
COMMENT ON COLUMN routes.name IS '路由名称';
COMMENT ON COLUMN routes.method IS '请求方法：GET/POST/PUT/DELETE';
COMMENT ON COLUMN routes.url IS '请求URL';
COMMENT ON COLUMN routes.path IS '路由路径';
COMMENT ON COLUMN routes.description IS '路由描述';
COMMENT ON COLUMN routes.auth_type IS '认证类型：1-需要认证，0-公开';

-- 11. permissions (权限关联表) - 重构为独立表结构
DROP TABLE IF EXISTS permissions;
CREATE TABLE permissions
(
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id     BIGINT      NOT NULL,
    role_id       BIGINT      NOT NULL,
    resource_type VARCHAR(20) NOT NULL,
    resource_id   BIGINT,
    resource_code VARCHAR(64),
    effect        SMALLINT    DEFAULT 1,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW(),
    is_deleted    BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX uk_permission_unique ON permissions (role_id, resource_type, resource_id, resource_code) WHERE is_deleted = FALSE;
CREATE INDEX idx_permission_tenant ON permissions (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_permission_role ON permissions (role_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_permission_resource ON permissions (resource_type, resource_id) WHERE is_deleted = FALSE;


COMMENT ON TABLE permissions IS '权限关联表 - 角色与资源/路由的关联';
COMMENT ON COLUMN permissions.id IS '权限ID';
COMMENT ON COLUMN permissions.tenant_id IS '租户ID';
COMMENT ON COLUMN permissions.role_id IS '角色ID';
COMMENT ON COLUMN permissions.resource_type IS '资源类型：menu-菜单，resource-按钮，route-路由';
COMMENT ON COLUMN permissions.resource_id IS '资源ID（可选，与resource_code二选一）';
COMMENT ON COLUMN permissions.resource_code IS '资源编码（可选，与resource_id二选一）';
COMMENT ON COLUMN permissions.effect IS '权限效果：1-允许，0-禁止';

-- 12. dicts (字典表)
DROP TABLE IF EXISTS dicts;
CREATE TABLE dicts
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,
    code       VARCHAR(32) NOT NULL,
    name       VARCHAR(64) NOT NULL,
    type       VARCHAR(32) DEFAULT 'system',
    extend     JSONB,
    sort       INT         DEFAULT 0,
    status     SMALLINT    DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX uk_dict_code ON dicts (tenant_id, code) WHERE is_deleted = FALSE;
CREATE INDEX idx_dict_tenant ON dicts (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_dict_type ON dicts (type) WHERE is_deleted = FALSE;

COMMENT ON TABLE dicts IS '字典表 - 系统数据字典';
COMMENT ON COLUMN dicts.id IS '字典ID';
COMMENT ON COLUMN dicts.tenant_id IS '租户ID';
COMMENT ON COLUMN dicts.code IS '字典编码';
COMMENT ON COLUMN dicts.name IS '字典名称';
COMMENT ON COLUMN dicts.type IS '字典类型：system-系统，custom-自定义';
COMMENT ON COLUMN dicts.extend IS '扩展信息（JSONB）';
COMMENT ON COLUMN dicts.sort IS '排序号';
COMMENT ON COLUMN dicts.status IS '状态：0-禁用，1-启用';

-- 13. dict_items (字典项表)
DROP TABLE IF EXISTS dict_items;
CREATE TABLE dict_items
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,
    dict_id    BIGINT      NOT NULL,
    code       VARCHAR(32) NOT NULL,
    name       VARCHAR(64) NOT NULL,
    sort       INT         DEFAULT 1,
    extend     JSONB,
    status     SMALLINT    DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX uk_dict_item_code ON dict_items (dict_id, code) WHERE is_deleted = FALSE;
CREATE INDEX idx_dict_item_dict ON dict_items (dict_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_dict_item_tenant ON dict_items (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_dict_item_sort ON dict_items (sort) WHERE is_deleted = FALSE;

COMMENT ON TABLE dict_items IS '字典项表 - 字典的具体选项';
COMMENT ON COLUMN dict_items.id IS '字典项ID';
COMMENT ON COLUMN dict_items.tenant_id IS '租户ID';
COMMENT ON COLUMN dict_items.dict_id IS '所属字典ID';
COMMENT ON COLUMN dict_items.code IS '字典项编码';
COMMENT ON COLUMN dict_items.name IS '字典项名称';
COMMENT ON COLUMN dict_items.sort IS '排序号';
COMMENT ON COLUMN dict_items.extend IS '扩展信息（JSONB）';
COMMENT ON COLUMN dict_items.status IS '状态：0-禁用，1-启用';

-- 14. db_logs (审计日志表) - 改进
DROP TABLE IF EXISTS db_logs;
CREATE TABLE db_logs
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id   BIGINT,
    op_type     VARCHAR(8) NOT NULL,
    user_code   VARCHAR(32),
    user_id     BIGINT,
    trace_id    VARCHAR(64),
    client_ip   INET,
    app_name    VARCHAR(32),
    class_name  VARCHAR(64),
    method_name VARCHAR(64),
    uri         VARCHAR(256),
    table_name  VARCHAR(64),
    detail      JSONB,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_db_logs_tenant ON db_logs (tenant_id);
CREATE INDEX idx_db_logs_created ON db_logs (created_at);
CREATE INDEX idx_db_logs_trace ON db_logs (trace_id);
CREATE INDEX idx_db_logs_user ON db_logs (user_id);
CREATE INDEX idx_db_logs_op_type ON db_logs (op_type);

COMMENT ON TABLE db_logs IS '审计日志表 - 数据库操作审计';
COMMENT ON COLUMN db_logs.id IS '日志ID';
COMMENT ON COLUMN db_logs.tenant_id IS '租户ID';
COMMENT ON COLUMN db_logs.op_type IS '操作类型：INSERT, UPDATE, DELETE';
COMMENT ON COLUMN db_logs.user_code IS '操作用户编码';
COMMENT ON COLUMN db_logs.user_id IS '操作用户ID';
COMMENT ON COLUMN db_logs.trace_id IS '链路追踪ID';
COMMENT ON COLUMN db_logs.client_ip IS '客户端IP地址(INET类型)';
COMMENT ON COLUMN db_logs.app_name IS '应用名称';
COMMENT ON COLUMN db_logs.class_name IS '类名';
COMMENT ON COLUMN db_logs.method_name IS '方法名';
COMMENT ON COLUMN db_logs.uri IS '请求URI';
COMMENT ON COLUMN db_logs.table_name IS '操作的表名';
COMMENT ON COLUMN db_logs.detail IS '操作详情（JSONB）';
COMMENT ON COLUMN db_logs.created_at IS '操作时间';

-- 15. subscriptions (订阅表)
DROP TABLE IF EXISTS subscriptions;
CREATE TABLE subscriptions
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    plan_id    BIGINT NOT NULL,
    status     VARCHAR(32) DEFAULT 'active',
    start_date TIMESTAMPTZ,
    end_date   TIMESTAMPTZ,
    auto_renew BOOLEAN     DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE INDEX idx_subscriptions_tenant ON subscriptions (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_subscriptions_plan ON subscriptions (plan_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_subscriptions_status ON subscriptions (status);

COMMENT ON TABLE subscriptions IS '订阅表 - 租户订阅计划';
COMMENT ON COLUMN subscriptions.id IS '订阅ID';
COMMENT ON COLUMN subscriptions.tenant_id IS '租户ID';
COMMENT ON COLUMN subscriptions.plan_id IS '套餐ID';
COMMENT ON COLUMN subscriptions.status IS '订阅状态：active-激活，expired-过期，cancelled-取消';
COMMENT ON COLUMN subscriptions.start_date IS '开始日期';
COMMENT ON COLUMN subscriptions.end_date IS '结束日期';
COMMENT ON COLUMN subscriptions.auto_renew IS '自动续费';

-- 16. plans (套餐表)
DROP TABLE IF EXISTS plans;
CREATE TABLE plans
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code       VARCHAR(32) NOT NULL,
    name       VARCHAR(64) NOT NULL,
    price      DECIMAL(10, 2) DEFAULT 0,
    quota      INT            DEFAULT 100,
    period     VARCHAR(20)    DEFAULT 'month',
    features   JSONB,
    sort       INT         DEFAULT 0,
    status     SMALLINT    DEFAULT 1,
    created_at TIMESTAMPTZ    DEFAULT NOW(),
    updated_at TIMESTAMPTZ    DEFAULT NOW(),
    is_deleted BOOLEAN        DEFAULT FALSE
);
CREATE UNIQUE INDEX uk_plans_code ON plans (code) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX uk_plans_name ON plans (name) WHERE is_deleted = FALSE;

COMMENT ON TABLE plans IS '套餐表 - SaaS订阅套餐';
COMMENT ON COLUMN plans.id IS '套餐ID';
COMMENT ON COLUMN plans.code IS '套餐编码';
COMMENT ON COLUMN plans.name IS '套餐名称';
COMMENT ON COLUMN plans.price IS '价格';
COMMENT ON COLUMN plans.quota IS '配额限制';
COMMENT ON COLUMN plans.period IS '计费周期：day-日，week-周，month-月，year-年';
COMMENT ON COLUMN plans.features IS '套餐功能（JSONB）';
COMMENT ON COLUMN plans.sort IS '排序号';
COMMENT ON COLUMN plans.status IS '状态：0-下架，1-上架';

-- 17. usage_records (使用记录表)
DROP TABLE IF EXISTS usage_records;
CREATE TABLE usage_records
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,
    action     VARCHAR(64) NOT NULL,
    resource   VARCHAR(64),
    count      INT         DEFAULT 1,
    ip         INET,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE INDEX idx_usage_tenant ON usage_records (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_usage_action ON usage_records (action) WHERE is_deleted = FALSE;
CREATE INDEX idx_usage_created ON usage_records (created_at) WHERE is_deleted = FALSE;

COMMENT ON TABLE usage_records IS '使用记录表 - 租户资源使用情况';
COMMENT ON COLUMN usage_records.id IS '记录ID';
COMMENT ON COLUMN usage_records.tenant_id IS '租户ID';
COMMENT ON COLUMN usage_records.action IS '操作类型';
COMMENT ON COLUMN usage_records.resource IS '资源标识';
COMMENT ON COLUMN usage_records.count IS '使用数量';
COMMENT ON COLUMN usage_records.ip IS '请求IP';

-- 18. ai_documents (AI文档表)
DROP TABLE IF EXISTS ai_documents;
CREATE TABLE ai_documents
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    title      VARCHAR(256) NOT NULL,
    content    TEXT,
    tags       VARCHAR(256)[],
    status     SMALLINT    DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE INDEX idx_ai_docs_tenant ON ai_documents (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_ai_docs_status ON ai_documents (status) WHERE is_deleted = FALSE;
CREATE INDEX idx_ai_docs_tags ON ai_documents USING GIN (tags);

COMMENT ON TABLE ai_documents IS 'AI文档表 - AI知识库文档';
COMMENT ON COLUMN ai_documents.id IS '文档ID';
COMMENT ON COLUMN ai_documents.tenant_id IS '租户ID';
COMMENT ON COLUMN ai_documents.title IS '文档标题';
COMMENT ON COLUMN ai_documents.content IS '文档内容';
COMMENT ON COLUMN ai_documents.tags IS '标签数组';
COMMENT ON COLUMN ai_documents.status IS '状态：0-禁用，1-启用';

-- 19. auth_sessions (登录会话表)
DROP TABLE IF EXISTS auth_sessions;
CREATE TABLE auth_sessions
(
    session_id VARCHAR(64) PRIMARY KEY,
    account_id BIGINT      NOT NULL,
    user_id    BIGINT,
    tenant_id  BIGINT      NOT NULL DEFAULT 0,
    role       VARCHAR(64),
    ip         INET,
    user_agent VARCHAR(512),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ          DEFAULT NOW()
);
CREATE INDEX idx_auth_sessions_expires_at ON auth_sessions (expires_at);
CREATE INDEX idx_auth_sessions_account ON auth_sessions (account_id);
CREATE INDEX idx_auth_sessions_user ON auth_sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_session_expires
    ON auth_sessions (session_id, expires_at);

COMMENT ON TABLE auth_sessions IS '登录会话表 - Redis 不可用时的会话持久化兜底';
COMMENT ON COLUMN auth_sessions.session_id IS '会话ID';
COMMENT ON COLUMN auth_sessions.account_id IS '账号ID';
COMMENT ON COLUMN auth_sessions.user_id IS '用户ID（可选）';
COMMENT ON COLUMN auth_sessions.tenant_id IS '租户ID';
COMMENT ON COLUMN auth_sessions.role IS '角色编码';
COMMENT ON COLUMN auth_sessions.ip IS '登录IP';
COMMENT ON COLUMN auth_sessions.user_agent IS 'User Agent';
COMMENT ON COLUMN auth_sessions.expires_at IS '过期时间';
COMMENT ON COLUMN auth_sessions.created_at IS '创建时间';

-- 20. tenant_user_seq (租户用户序号表) - 用于生成自增user_code
DROP TABLE IF EXISTS tenant_user_seq;
CREATE TABLE tenant_user_seq
(
    tenant_id        BIGINT PRIMARY KEY,
    seq              BIGINT       NOT NULL DEFAULT 0,
    user_code_format VARCHAR(32)  NOT NULL DEFAULT 'sequential',
    updated_at       TIMESTAMPTZ           DEFAULT NOW()
);
COMMENT ON TABLE tenant_user_seq IS '租户用户序号表 - 用于生成自增user_code';
COMMENT ON COLUMN tenant_user_seq.tenant_id IS '租户ID';
COMMENT ON COLUMN tenant_user_seq.seq IS '当前序号';
COMMENT ON COLUMN tenant_user_seq.user_code_format IS '用户编码格式: sequential(U00000001), tenant_prefix(U001-00001), tenant_random(U001-AB3F2)';

-- 21. account_roles (平台级账号角色表) - 新增
DROP TABLE IF EXISTS account_roles;
CREATE TABLE account_roles
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    account_id BIGINT NOT NULL,
    role_code  VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE UNIQUE INDEX uk_account_role ON account_roles (account_id, role_code);
CREATE INDEX idx_ar_account ON account_roles (account_id);
CREATE INDEX idx_ar_role ON account_roles (role_code);

COMMENT ON TABLE account_roles IS '平台级账号角色表 - 超级管理员、平台运营等跨租户角色';
COMMENT ON COLUMN account_roles.id IS '记录ID';
COMMENT ON COLUMN account_roles.account_id IS '账号ID';
COMMENT ON COLUMN account_roles.role_code IS '角色编码：super_admin-超级管理员，platform_admin-平台管理员';

-- 22. attachments (附件资源表)
DROP TABLE IF EXISTS attachments;
CREATE TABLE attachments
(
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT      NOT NULL DEFAULT 0,
    user_id     BIGINT,
    file_name   TEXT,
    file_ext    VARCHAR(20),
    mime_type   VARCHAR(100),
    file_size   BIGINT,
    storage     VARCHAR(20),
    object_key  TEXT,
    url         TEXT,
    hash        VARCHAR(64),
    status      SMALLINT    DEFAULT 1,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    is_deleted  BOOLEAN     DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_attachments_tenant_hash ON attachments(tenant_id, hash) WHERE is_deleted = FALSE;
CREATE INDEX idx_attachments_tenant ON attachments(tenant_id) WHERE is_deleted = FALSE;

COMMENT ON TABLE attachments IS '附件资源表 - 多租户文件存储元数据';
COMMENT ON COLUMN attachments.id IS '主键ID，自增序列';
COMMENT ON COLUMN attachments.tenant_id IS '租户ID，用于多租户隔离';
COMMENT ON COLUMN attachments.user_id IS '上传用户ID';
COMMENT ON COLUMN attachments.file_name IS '原始文件名';
COMMENT ON COLUMN attachments.file_ext IS '文件扩展名';
COMMENT ON COLUMN attachments.mime_type IS 'MIME类型';
COMMENT ON COLUMN attachments.file_size IS '文件大小(字节)';
COMMENT ON COLUMN attachments.storage IS '存储方式: local / oss / s3';
COMMENT ON COLUMN attachments.object_key IS '存储路径(关键)';
COMMENT ON COLUMN attachments.url IS '访问URL';
COMMENT ON COLUMN attachments.hash IS '去重用文件哈希值';
COMMENT ON COLUMN attachments.status IS '状态：1-正常，0-隐藏';
COMMENT ON COLUMN attachments.created_at IS '创建时间';
COMMENT ON COLUMN attachments.updated_at IS '更新时间';
COMMENT ON COLUMN attachments.is_deleted IS '软删除标记';


-- ============================================
-- 🔐 多租户 RLS (行级安全) 策略 — 纵深防御层
-- ============================================
--
-- ⚠️ 重要说明：
-- 本 RLS 策略已降级为"纵深防御"，主要依赖应用层 SET app.tenant_id。
-- 本层作用：防止应用层漏 SET 时的数据泄漏，或运维直接操作 DB 时的越权。
--
-- ⚠️ app.mode 配置：
--   single：不约束 tenant_id，所有租户数据均可访问（单租户模式）
--   saas：必须约束 tenant_id，未设置 tenant_id 时拒绝所有操作（SaaS 多租户模式，按 tenant_id 隔离行）
--   schema：每个租户独立 schema，RLS 不约束 tenant_id（由连接层 schema 隔离）
--   database：每个租户独立数据库，RLS 不约束 tenant_id（由连接层 database 隔离）
--
-- ⚠️ 注意：accounts 表为平台级，不在 RLS 覆盖范围内（同一账号可跨租户存在）。
--

-- 1. 对所有租户数据表启用 RLS
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
ALTER TABLE account_auths ENABLE ROW LEVEL SECURITY;
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE role_data_scopes ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE menus ENABLE ROW LEVEL SECURITY;
ALTER TABLE resources ENABLE ROW LEVEL SECURITY;
ALTER TABLE routes ENABLE ROW LEVEL SECURITY;
ALTER TABLE permissions ENABLE ROW LEVEL SECURITY;
ALTER TABLE dicts ENABLE ROW LEVEL SECURITY;
ALTER TABLE dict_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE usage_records ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE attachments ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_user_seq ENABLE ROW LEVEL SECURITY;

-- 2. 创建租户隔离策略 (读取 & 写入)
-- app.mode = single：不约束 tenant_id（放行所有行）
-- app.mode = saas：必须约束 tenant_id（tenant_id 必须匹配）
-- is_deleted = TRUE 的行默认不可见，除非 SET app.show_deleted = true
-- 未设置 app.mode：默认 single（向后兼容）
-- 未设置 app.tenant_id：saas 模式下拒绝所有行（安全默认值）

CREATE POLICY tenant_isolation_policy ON tenants
    USING (is_deleted = FALSE OR COALESCE(current_setting('app.show_deleted', true)::boolean, false));

CREATE POLICY tenant_isolation_policy ON account_auths
    USING (
        (current_setting('app.mode') = 'single' OR (current_setting('app.mode') = 'saas' AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT))
        AND (is_deleted = FALSE OR COALESCE(current_setting('app.show_deleted', true)::boolean, false))
    );

CREATE POLICY tenant_isolation_policy ON organizations
    USING (
        (current_setting('app.mode') = 'single' OR (current_setting('app.mode') = 'saas' AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT))
        AND (is_deleted = FALSE OR COALESCE(current_setting('app.show_deleted', true)::boolean, false))
    );

CREATE POLICY tenant_isolation_policy ON users
    USING (
        (current_setting('app.mode') = 'single' OR (current_setting('app.mode') = 'saas' AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT))
        AND (is_deleted = FALSE OR COALESCE(current_setting('app.show_deleted', true)::boolean, false))
    );

CREATE POLICY tenant_isolation_policy ON roles
    USING (
        (current_setting('app.mode') = 'single' OR (current_setting('app.mode') = 'saas' AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT))
        AND (is_deleted = FALSE OR COALESCE(current_setting('app.show_deleted', true)::boolean, false))
    );

CREATE POLICY tenant_isolation_policy ON role_data_scopes
    USING (
        current_setting('app.mode') = 'single' OR (current_setting('app.mode') = 'saas' AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT)
    );

CREATE POLICY tenant_isolation_policy ON user_roles
    USING (
        (current_setting('app.mode') = 'single' OR (current_setting('app.mode') = 'saas' AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT))
        AND (is_deleted = FALSE OR COALESCE(current_setting('app.show_deleted', true)::boolean, false))
    );

CREATE POLICY tenant_isolation_policy ON menus
    USING (
        (current_setting('app.mode') = 'single' OR (current_setting('app.mode') = 'saas' AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT))
        AND (is_deleted = FALSE OR COALESCE(current_setting('app.show_deleted', true)::boolean, false))
    );

CREATE POLICY tenant_isolation_policy ON resources
    USING (
        (current_setting('app.mode') = 'single' OR (current_setting('app.mode') = 'saas' AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT))
        AND (is_deleted = FALSE OR COALESCE(current_setting('app.show_deleted', true)::boolean, false))
    );

CREATE POLICY tenant_isolation_policy ON routes
    USING (
        (current_setting('app.mode') = 'single' OR (current_setting('app.mode') = 'saas' AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT))
        AND (is_deleted = FALSE OR COALESCE(current_setting('app.show_deleted', true)::boolean, false))
    );

CREATE POLICY tenant_isolation_policy ON permissions
    USING (
        (current_setting('app.mode') = 'single' OR (current_setting('app.mode') = 'saas' AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT))
        AND (is_deleted = FALSE OR COALESCE(current_setting('app.show_deleted', true)::boolean, false))
    );

CREATE POLICY tenant_isolation_policy ON dicts
    USING (
        (current_setting('app.mode') = 'single' OR (current_setting('app.mode') = 'saas' AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT))
        AND (is_deleted = FALSE OR COALESCE(current_setting('app.show_deleted', true)::boolean, false))
    );

CREATE POLICY tenant_isolation_policy ON dict_items
    USING (
        (current_setting('app.mode') = 'single' OR (current_setting('app.mode') = 'saas' AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT))
        AND (is_deleted = FALSE OR COALESCE(current_setting('app.show_deleted', true)::boolean, false))
    );

CREATE POLICY tenant_isolation_policy ON subscriptions
    USING (
        (current_setting('app.mode') = 'single' OR (current_setting('app.mode') = 'saas' AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT))
        AND (is_deleted = FALSE OR COALESCE(current_setting('app.show_deleted', true)::boolean, false))
    );

CREATE POLICY tenant_isolation_policy ON usage_records
    USING (
        (current_setting('app.mode') = 'single' OR (current_setting('app.mode') = 'saas' AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT))
        AND (is_deleted = FALSE OR COALESCE(current_setting('app.show_deleted', true)::boolean, false))
    );

CREATE POLICY tenant_isolation_policy ON ai_documents
    USING (
        (current_setting('app.mode') = 'single' OR (current_setting('app.mode') = 'saas' AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT))
        AND (is_deleted = FALSE OR COALESCE(current_setting('app.show_deleted', true)::boolean, false))
    );

CREATE POLICY tenant_isolation_policy ON attachments
    USING (
        (current_setting('app.mode') = 'single' OR (current_setting('app.mode') = 'saas' AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT))
        AND (is_deleted = FALSE OR COALESCE(current_setting('app.show_deleted', true)::boolean, false))
    );

CREATE POLICY tenant_isolation_policy ON tenant_user_seq
    USING (
        current_setting('app.mode') = 'single' OR (current_setting('app.mode') = 'saas' AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT)
    );


-- ============================================
-- 初始化数据
-- ============================================
INSERT INTO tenants (code, name, status, created_by, updated_by)
VALUES ('default', '默认租户', 1, 0, 0);

INSERT INTO accounts (phone, email, password, username, real_name, status)
VALUES ('13800138000', 'admin@example.com', '$argon2id$v=19$m=19456,t=2,p=1$nMpweyGYDB9dvGMQAkzcHw$Tfc9vn1or7d0KMg0h6aRYFuDMxZbuK2cO8o6VaOyBk4', 'admin', '系统管理员', 1);

INSERT INTO users (tenant_id, account_id, code, real_name, status, created_by, updated_by)
VALUES (1, 1, 'admin', '系统管理员', 1, 0, 0);

INSERT INTO roles (tenant_id, code, name, description, data_scope, is_default, sort, status, created_by, updated_by)
VALUES (1, 'admin', '管理员', '系统管理员', 5, FALSE, 1, 1, 0, 0),
       (1, 'user', '普通用户', '普通用户', 4, TRUE, 2, 1, 0, 0);

INSERT INTO user_roles (tenant_id, user_id, role_id)
VALUES (1, 1, 1);

INSERT INTO tenant_user_seq (tenant_id, seq, user_code_format) VALUES (1, 0, 'sequential');

-- 初始化超级管理员角色
INSERT INTO account_roles (account_id, role_code) VALUES (1, 'super_admin');