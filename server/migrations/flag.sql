-- ============================================
-- Flag App - 头像框生成器 / 活动头像工具
-- 版本: PostgreSQL 14+
-- ============================================

SET client_encoding = 'UTF8';

-- 1. flag_frame_categories (头像框分类表)
DROP TABLE IF EXISTS flag_frame_categories;
CREATE TABLE flag_frame_categories
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id   BIGINT      NOT NULL DEFAULT 0,
    code        VARCHAR(32) NOT NULL,
    name        VARCHAR(64) NOT NULL,
    type        VARCHAR(20) DEFAULT 'public',
    sort        INT         DEFAULT 0,
    status      SMALLINT    DEFAULT 1,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    is_deleted  BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX uk_flag_category_code ON flag_frame_categories(tenant_id, code) WHERE is_deleted = FALSE;
CREATE INDEX idx_flag_category_tenant ON flag_frame_categories(tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_flag_category_type ON flag_frame_categories(type) WHERE is_deleted = FALSE;
COMMENT ON TABLE flag_frame_categories IS '头像框分类表';

-- 2. flag_frames (头像框模板表)
DROP TABLE IF EXISTS flag_frames;
CREATE TABLE flag_frames
(
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id       BIGINT      NOT NULL DEFAULT 0,
    category_id     BIGINT,
    name            VARCHAR(64) NOT NULL,
    description     VARCHAR(256),
    preview_url     VARCHAR(512),
    template_url    VARCHAR(512),
    template_config JSONB,
    type            VARCHAR(20) DEFAULT 'public',
    sort            INT         DEFAULT 0,
    status          SMALLINT    DEFAULT 1,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    is_deleted      BOOLEAN     DEFAULT FALSE
);
CREATE INDEX idx_flag_frame_tenant ON flag_frames(tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_flag_frame_category ON flag_frames(category_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_flag_frame_type ON flag_frames(type) WHERE is_deleted = FALSE;
CREATE INDEX idx_flag_frame_status ON flag_frames(status) WHERE is_deleted = FALSE;
COMMENT ON TABLE flag_frames IS '头像框模板表';

-- 3. flag_spaces (活动空间表)
DROP TABLE IF EXISTS flag_spaces;
CREATE TABLE flag_spaces
(
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id       BIGINT      NOT NULL,
    name            VARCHAR(128) NOT NULL,
    description     VARCHAR(512),
    frame_id        BIGINT,
    space_config    JSONB,
    access_type     VARCHAR(20) DEFAULT 'public',
    invite_code     VARCHAR(32),
    max_usage       INT,
    usage_count     INT         DEFAULT 0,
    status          SMALLINT    DEFAULT 1,
    start_at        TIMESTAMPTZ,
    end_at          TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    is_deleted      BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX uk_flag_space_invite ON flag_spaces(tenant_id, invite_code) WHERE is_deleted = FALSE AND invite_code IS NOT NULL;
CREATE INDEX idx_flag_space_tenant ON flag_spaces(tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_flag_space_status ON flag_spaces(status) WHERE is_deleted = FALSE;
CREATE INDEX idx_flag_space_frame ON flag_spaces(frame_id) WHERE is_deleted = FALSE;
COMMENT ON TABLE flag_spaces IS '活动空间表';

-- 4. flag_user_generated (用户生成的头像记录)
DROP TABLE IF EXISTS flag_user_generated;
CREATE TABLE flag_user_generated
(
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id       BIGINT      NOT NULL DEFAULT 0,
    user_id         BIGINT,
    space_id        BIGINT,
    frame_id        BIGINT,
    source_image    VARCHAR(512),
    result_url      VARCHAR(512),
    result_key      VARCHAR(256),
    field_values    JSONB,
    share_text      VARCHAR(256),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    is_deleted      BOOLEAN     DEFAULT FALSE
);
CREATE INDEX idx_flag_gen_tenant ON flag_user_generated(tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_flag_gen_user ON flag_user_generated(user_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_flag_gen_space ON flag_user_generated(space_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_flag_gen_frame ON flag_user_generated(frame_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_flag_gen_created ON flag_user_generated(created_at) WHERE is_deleted = FALSE;
COMMENT ON TABLE flag_user_generated IS '用户生成的头像记录';

-- 5. flag_avatar_categories (头像分类表)
DROP TABLE IF EXISTS flag_avatar_categories;
CREATE TABLE flag_avatar_categories
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id   BIGINT      NOT NULL DEFAULT 0,
    code        VARCHAR(32) NOT NULL,
    name        VARCHAR(64) NOT NULL,
    icon        VARCHAR(128),
    type        VARCHAR(20) DEFAULT 'public',
    sort        INT         DEFAULT 0,
    status      SMALLINT    DEFAULT 1,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    is_deleted  BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX uk_flag_avatar_cat_code ON flag_avatar_categories(tenant_id, code) WHERE is_deleted = FALSE;
CREATE INDEX idx_flag_avatar_cat_tenant ON flag_avatar_categories(tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_flag_avatar_cat_type ON flag_avatar_categories(type) WHERE is_deleted = FALSE;
COMMENT ON TABLE flag_avatar_categories IS '头像分类表';

-- 6. flag_avatars (头像表)
DROP TABLE IF EXISTS flag_avatars;
CREATE TABLE flag_avatars
(
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id       BIGINT      NOT NULL DEFAULT 0,
    user_id         BIGINT      NOT NULL,
    category_id     BIGINT,
    name            VARCHAR(64),
    source_url      VARCHAR(512),
    thumbnail_url   VARCHAR(512),
    file_size       BIGINT,
    width           INT,
    height          INT,
    type            VARCHAR(20) DEFAULT 'custom',
    is_public       BOOLEAN     DEFAULT FALSE,
    like_count      INT         DEFAULT 0,
    view_count      INT         DEFAULT 0,
    status          SMALLINT    DEFAULT 1,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    is_deleted      BOOLEAN     DEFAULT FALSE
);
CREATE INDEX idx_flag_avatar_tenant ON flag_avatars(tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_flag_avatar_user ON flag_avatars(user_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_flag_avatar_category ON flag_avatars(category_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_flag_avatar_type ON flag_avatars(type) WHERE is_deleted = FALSE;
CREATE INDEX idx_flag_avatar_status ON flag_avatars(status) WHERE is_deleted = FALSE;
CREATE INDEX idx_flag_avatar_created ON flag_avatars(created_at) WHERE is_deleted = FALSE;
COMMENT ON TABLE flag_avatars IS '用户头像表';

-- ============================================
-- 🔐 RLS 策略
-- ============================================
ALTER TABLE flag_frame_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE flag_frames ENABLE ROW LEVEL SECURITY;
ALTER TABLE flag_spaces ENABLE ROW LEVEL SECURITY;
ALTER TABLE flag_user_generated ENABLE ROW LEVEL SECURITY;
ALTER TABLE flag_avatar_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE flag_avatars ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON flag_frame_categories USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT);
CREATE POLICY tenant_isolation_policy ON flag_frames USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT);
CREATE POLICY tenant_isolation_policy ON flag_spaces USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT);
CREATE POLICY tenant_isolation_policy ON flag_user_generated USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT);
CREATE POLICY tenant_isolation_policy ON flag_avatar_categories USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT);
CREATE POLICY tenant_isolation_policy ON flag_avatars USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT);

-- ============================================
-- 📋 菜单 seed（写入 __template__ 租户，新租户首装时由 first_install.go 复制）
-- 注意：用子查询拿 __template__ 里 system 菜单的实际 id（避免硬编码）
-- 注意：ancestors 留空，下面 UPDATE 段统一重建为 parent_id::text（与 framework.sql 2c 一致）
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
    ('frame-list',        '相框列表', '', '/frames',           'FileIcon',  1),
    ('frame-categories',  '相框分类', '', '/frame-categories', 'ListIcon',  2)
) AS s(code, name, subtitle, url, path, icon, sort)
ON CONFLICT (tenant_id, code) WHERE is_deleted = FALSE DO NOTHING;

-- 子菜单：头像列表、头像分类（parent = avatars）
INSERT INTO menus (tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
SELECT (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE),
       s.code, s.name, s.subtitle, s.url, s.path, s.icon, s.sort,
       (SELECT id FROM menus WHERE code = 'avatars' AND tenant_id = (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE) AND is_deleted = FALSE),
       '', TRUE, TRUE
FROM (VALUES
    ('avatar-list',        '头像列表', '', '/avatars',           'FileIcon',  1),
    ('avatar-categories',  '头像分类', '', '/avatar-categories', 'ListIcon',  2)
) AS s(code, name, subtitle, url, path, icon, sort)
ON CONFLICT (tenant_id, code) WHERE is_deleted = FALSE DO NOTHING;

-- 重建 frames/avatars 子菜单的 ancestors
UPDATE menus SET ancestors = parent_id::text
WHERE tenant_id = (SELECT id FROM tenants WHERE code = '__template__' AND is_deleted = FALSE)
  AND code IN ('frame-list', 'frame-categories', 'avatar-list', 'avatar-categories')
  AND parent_id > 0 AND is_deleted = FALSE;

-- 序列号兜底
SELECT setval('menus_id_seq', GREATEST(
    (SELECT COALESCE(MAX(id), 0) FROM menus),
    1000
), true);

-- ============================================
-- 🔑 资源 seed（__template__ 租户；first_install.go 会全量复制）
-- 让 flag 模块的菜单可被角色授权 / RBAC 校验
-- ============================================
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