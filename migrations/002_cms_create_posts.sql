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
    is_deleted BOOLEAN DEFAULT FALSE             -- 软删除标记
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
