-- ============================================
-- 角色菜单和资源权限分离重构
-- 新增 role_menus（角色-菜单关联）
-- 新增 role_resources（角色-资源关联）
-- 保留 permissions 表用于向后兼容（可选逐步迁移）
-- ============================================

-- 1. role_menus（角色-菜单关联表）
DROP TABLE IF EXISTS role_menus;
CREATE TABLE role_menus
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,
    role_id    BIGINT      NOT NULL,
    menu_id    BIGINT      NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    is_deleted  BOOLEAN     DEFAULT FALSE
);

-- 唯一索引：同一角色同一菜单不能重复
CREATE UNIQUE INDEX uk_role_menu ON role_menus (role_id, menu_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_role_menus_tenant ON role_menus (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_role_menus_role ON role_menus (role_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_role_menus_menu ON role_menus (menu_id) WHERE is_deleted = FALSE;

COMMENT ON TABLE role_menus IS '角色菜单关联表 - 角色能访问的菜单';
COMMENT ON COLUMN role_menus.role_id IS '角色ID';
COMMENT ON COLUMN role_menus.menu_id IS '菜单ID';

-- 2. role_resources（角色-资源关联表）
DROP TABLE IF EXISTS role_resources;
CREATE TABLE role_resources
(
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id    BIGINT      NOT NULL,
    role_id      BIGINT      NOT NULL,
    resource_id  BIGINT      NOT NULL,
    effect       SMALLINT    DEFAULT 1,  -- 1: grant, 2: deny
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW(),
    is_deleted   BOOLEAN     DEFAULT FALSE
);

-- 唯一索引：同一角色同一资源不能重复
CREATE UNIQUE INDEX uk_role_resource ON role_resources (role_id, resource_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_role_resources_tenant ON role_resources (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_role_resources_role ON role_resources (role_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_role_resources_resource ON role_resources (resource_id) WHERE is_deleted = FALSE;

COMMENT ON TABLE role_resources IS '角色资源关联表 - 角色的资源权限';
COMMENT ON COLUMN role_resources.effect IS '权限效果: 1=授予(grant), 2=拒绝(deny)';

-- 3. 为关键表启用 RLS（纵深防御）
ALTER TABLE role_menus ENABLE ROW LEVEL SECURITY;
ALTER TABLE role_resources ENABLE ROW LEVEL SECURITY;

-- 4. 创建 RLS 策略
DROP POLICY IF EXISTS tenant_isolation_policy ON role_menus;
CREATE POLICY tenant_isolation_policy ON role_menus
    USING (
        tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    );

DROP POLICY IF EXISTS tenant_isolation_policy ON role_resources;
CREATE POLICY tenant_isolation_policy ON role_resources
    USING (
        tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    );

-- 5. 可选：为 permissions 表添加 resource_type 注释，说明新设计
-- 注意：保留 permissions 表用于向后兼容，新设计使用 role_menus 和 role_resources