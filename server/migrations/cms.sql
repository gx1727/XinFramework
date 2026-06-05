-- ============================================
-- CMS App 数据库表
-- ============================================

SET client_encoding = 'UTF8';

-- CMS文章表
DROP TABLE IF EXISTS cms_posts;
CREATE TABLE cms_posts (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL DEFAULT 0,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    status SMALLINT DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_cms_posts_tenant ON cms_posts (tenant_id) WHERE is_deleted = FALSE;
COMMENT ON TABLE cms_posts IS 'CMS文章表';
COMMENT ON COLUMN cms_posts.id IS '主键ID，自增序列';
COMMENT ON COLUMN cms_posts.tenant_id IS '租户ID，用于多租户隔离';
COMMENT ON COLUMN cms_posts.title IS '文章标题';
COMMENT ON COLUMN cms_posts.content IS '文章内容';
COMMENT ON COLUMN cms_posts.status IS '文章状态：1-发布，0-草稿，-1-下架';

-- RLS
ALTER TABLE cms_posts ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_policy ON cms_posts USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT);