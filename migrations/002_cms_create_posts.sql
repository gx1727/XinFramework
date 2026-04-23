-- 删除已存在的CMS文章表（如果存在）
DROP TABLE IF EXISTS cms_posts;

-- 创建CMS文章表
CREATE TABLE IF NOT EXISTS cms_posts (
    id BIGSERIAL PRIMARY KEY,                    -- 主键ID，自增序列
    tenant_id BIGINT NOT NULL DEFAULT 0,         -- 租户ID，用于多租户隔离
    title VARCHAR(255) NOT NULL,                 -- 文章标题
    content TEXT,                                -- 文章内容
    status SMALLINT DEFAULT 1,                   -- 文章状态：1-发布，0-草稿，-1-下架
    created_at TIMESTAMPTZ DEFAULT NOW(),        -- 创建时间
    updated_at TIMESTAMPTZ DEFAULT NOW(),        -- 更新时间
    is_deleted BOOLEAN DEFAULT FALSE              -- 软删除标记
);

-- 创建租户ID索引，仅针对未删除的记录
CREATE INDEX IF NOT EXISTS idx_cms_posts_tenant ON cms_posts (tenant_id) WHERE is_deleted = FALSE;

-- 添加表注释
COMMENT ON TABLE cms_posts IS 'CMS文章表';

-- 添加字段注释
COMMENT ON COLUMN cms_posts.id IS '主键ID，自增序列';
COMMENT ON COLUMN cms_posts.tenant_id IS '租户ID，用于多租户隔离';
COMMENT ON COLUMN cms_posts.title IS '文章标题';
COMMENT ON COLUMN cms_posts.content IS '文章内容';
COMMENT ON COLUMN cms_posts.status IS '文章状态：1-发布，0-草稿，-1-下架';
COMMENT ON COLUMN cms_posts.created_at IS '创建时间';
COMMENT ON COLUMN cms_posts.updated_at IS '更新时间';
COMMENT ON COLUMN cms_posts.is_deleted IS '软删除标记';

-- ============================================
-- 🔐 多租户 RLS (行级安全) 策略 — 纵深防御层
-- ============================================
-- 注意：cms_posts 表的多租户隔离以应用层 SET app.tenant_id 为主要机制，RLS 作为纵深防御。
-- app.mode 配置：
--   single：不约束 tenant_id（放行所有行）
--   saas：必须约束 tenant_id（tenant_id 必须匹配）
-- is_deleted = TRUE 的行默认不可见，除非 SET app.show_deleted = true
ALTER TABLE cms_posts ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON cms_posts
    USING (
    (
        current_setting('app.mode') = 'single'
        OR (
            current_setting('app.mode') = 'saas'
            AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
        )
    )
    AND (
        is_deleted = FALSE
        OR COALESCE(current_setting('app.show_deleted', true)::boolean, false)
    )
);

-- ============================================
CREATE POLICY tenant_rw_policy ON users
FOR SELECT, UPDATE, DELETE
    USING (...);

CREATE POLICY tenant_insert_policy ON users
FOR INSERT
WITH CHECK (
    current_setting('app.mode') = 'single'
    OR (
        current_setting('app.mode') = 'saas'
        AND tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    )
);