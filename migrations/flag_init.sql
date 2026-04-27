-- ============================================
-- 🚩 Flag App - 头像框生成器 / 活动头像工具
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
COMMENT ON COLUMN flag_frame_categories.id IS '分类ID';
COMMENT ON COLUMN flag_frame_categories.tenant_id IS '租户ID(0表示系统公共分类)';
COMMENT ON COLUMN flag_frame_categories.code IS '分类编码';
COMMENT ON COLUMN flag_frame_categories.name IS '分类名称';
COMMENT ON COLUMN flag_frame_categories.type IS '类型: public-公开, emotion-情绪, tag-标签, hot-热点, custom-自定义';
COMMENT ON COLUMN flag_frame_categories.sort IS '排序号';
COMMENT ON COLUMN flag_frame_categories.status IS '状态: 0-禁用, 1-启用';

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
COMMENT ON COLUMN flag_frames.id IS '模板ID';
COMMENT ON COLUMN flag_frames.tenant_id IS '租户ID(0表示系统公共模板)';
COMMENT ON COLUMN flag_frames.category_id IS '所属分类ID';
COMMENT ON COLUMN flag_frames.name IS '模板名称';
COMMENT ON COLUMN flag_frames.description IS '模板描述';
COMMENT ON COLUMN flag_frames.preview_url IS '预览图URL';
COMMENT ON COLUMN flag_frames.template_url IS '模板底图URL';
COMMENT ON COLUMN flag_frames.template_config IS '模板配置(JSON): 头像区域位置、大小、装饰元素位置等';
COMMENT ON COLUMN flag_frames.type IS '类型: public-公开, private-私有, space-活动专属';

-- 3. flag_spaces (活动空间表 - Space)
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

COMMENT ON TABLE flag_spaces IS '活动空间表 - Space';
COMMENT ON COLUMN flag_spaces.id IS '空间ID';
COMMENT ON COLUMN flag_spaces.tenant_id IS '租户ID(组织/学校/企业)';
COMMENT ON COLUMN flag_spaces.name IS '活动名称';
COMMENT ON COLUMN flag_spaces.description IS '活动描述';
COMMENT ON COLUMN flag_spaces.frame_id IS '绑定的头像框模板ID';
COMMENT ON COLUMN flag_spaces.space_config IS '空间配置(JSON): 动态字段、显示设置等';
COMMENT ON COLUMN flag_spaces.access_type IS '访问类型: public-公开, invite-邀请码, limit-限制次数';
COMMENT ON COLUMN flag_spaces.invite_code IS '邀请码';
COMMENT ON COLUMN flag_spaces.max_usage IS '最大使用次数限制';
COMMENT ON COLUMN flag_spaces.usage_count IS '已使用次数';
COMMENT ON COLUMN flag_spaces.start_at IS '活动开始时间';
COMMENT ON COLUMN flag_spaces.end_at IS '活动结束时间';

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
COMMENT ON COLUMN flag_user_generated.id IS '记录ID';
COMMENT ON COLUMN flag_user_generated.tenant_id IS '租户ID';
COMMENT ON COLUMN flag_user_generated.user_id IS '用户ID(可选,游客可为空)';
COMMENT ON COLUMN flag_user_generated.space_id IS '关联的Space ID';
COMMENT ON COLUMN flag_user_generated.frame_id IS '使用的头像框模板ID';
COMMENT ON COLUMN flag_user_generated.source_image IS '用户上传的原图URL';
COMMENT ON COLUMN flag_user_generated.result_url IS '生成结果URL';
COMMENT ON COLUMN flag_user_generated.result_key IS '存储路径key';
COMMENT ON COLUMN flag_user_generated.field_values IS '动态字段值(JSON): 如 {grade: "2024届", college: "计算机学院"}';
COMMENT ON COLUMN flag_user_generated.share_text IS '分享文案';

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
COMMENT ON COLUMN flag_avatar_categories.id IS '分类ID';
COMMENT ON COLUMN flag_avatar_categories.tenant_id IS '租户ID(0表示系统公共分类)';
COMMENT ON COLUMN flag_avatar_categories.code IS '分类编码';
COMMENT ON COLUMN flag_avatar_categories.name IS '分类名称';
COMMENT ON COLUMN flag_avatar_categories.icon IS '分类图标URL';
COMMENT ON COLUMN flag_avatar_categories.type IS '类型: public-公开, custom-自定义';
COMMENT ON COLUMN flag_avatar_categories.sort IS '排序号';
COMMENT ON COLUMN flag_avatar_categories.status IS '状态: 0-禁用, 1-启用';

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
COMMENT ON COLUMN flag_avatars.id IS '头像ID';
COMMENT ON COLUMN flag_avatars.tenant_id IS '租户ID';
COMMENT ON COLUMN flag_avatars.user_id IS '用户ID';
COMMENT ON COLUMN flag_avatars.category_id IS '所属分类ID';
COMMENT ON COLUMN flag_avatars.name IS '头像名称(可选)';
COMMENT ON COLUMN flag_avatars.source_url IS '原图URL';
COMMENT ON COLUMN flag_avatars.thumbnail_url IS '缩略图URL';
COMMENT ON COLUMN flag_avatars.file_size IS '文件大小(字节)';
COMMENT ON COLUMN flag_avatars.width IS '图片宽度';
COMMENT ON COLUMN flag_avatars.height IS '图片高度';
COMMENT ON COLUMN flag_avatars.type IS '类型: custom-自定义, system-系统';
COMMENT ON COLUMN flag_avatars.is_public IS '是否公开(公开后其他用户可使用)';
COMMENT ON COLUMN flag_avatars.like_count IS '点赞数';
COMMENT ON COLUMN flag_avatars.view_count IS '浏览数';
COMMENT ON COLUMN flag_avatars.status IS '状态: 0-禁用, 1-启用';


-- ============================================
-- 🔐 多租户 RLS (行级安全) 策略
-- ============================================
ALTER TABLE flag_frame_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE flag_frames ENABLE ROW LEVEL SECURITY;
ALTER TABLE flag_spaces ENABLE ROW LEVEL SECURITY;
ALTER TABLE flag_user_generated ENABLE ROW LEVEL SECURITY;
ALTER TABLE flag_avatar_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE flag_avatars ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON flag_frame_categories
    USING (
        (current_setting('app.mode') = 'single' OR (current_setting('app.mode') = 'saas' AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT))
        AND (is_deleted = FALSE OR COALESCE(current_setting('app.show_deleted', true)::boolean, false))
    );

CREATE POLICY tenant_isolation_policy ON flag_frames
    USING (
        (current_setting('app.mode') = 'single' OR (current_setting('app.mode') = 'saas' AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT))
        AND (is_deleted = FALSE OR COALESCE(current_setting('app.show_deleted', true)::boolean, false))
    );

CREATE POLICY tenant_isolation_policy ON flag_spaces
    USING (
        (current_setting('app.mode') = 'single' OR (current_setting('app.mode') = 'saas' AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT))
        AND (is_deleted = FALSE OR COALESCE(current_setting('app.show_deleted', true)::boolean, false))
    );

CREATE POLICY tenant_isolation_policy ON flag_user_generated
    USING (
        (current_setting('app.mode') = 'single' OR (current_setting('app.mode') = 'saas' AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT))
        AND (is_deleted = FALSE OR COALESCE(current_setting('app.show_deleted', true)::boolean, false))
    );

CREATE POLICY tenant_isolation_policy ON flag_avatar_categories
    USING (
        (current_setting('app.mode') = 'single' OR (current_setting('app.mode') = 'saas' AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT))
        AND (is_deleted = FALSE OR COALESCE(current_setting('app.show_deleted', true)::boolean, false))
    );

CREATE POLICY tenant_isolation_policy ON flag_avatars
    USING (
        (current_setting('app.mode') = 'single' OR (current_setting('app.mode') = 'saas' AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT))
        AND (is_deleted = FALSE OR COALESCE(current_setting('app.show_deleted', true)::boolean, false))
    );
