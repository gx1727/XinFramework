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

-- 注意：flag 模块的菜单/资源 seed 已搬到 migrations/framework.sql 末尾的 bootstrap 段。
-- 原因：flag.sql 按字母序在 framework.sql 之前跑，但 menu/resource seed 依赖 tenants/menus 表。
-- 拆分原则：flag.sql 只建表，seed 跟 bootstrap 一起进 framework.sql（与 config 模块一致）。