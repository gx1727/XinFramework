-- CMS posts table
CREATE TABLE IF NOT EXISTS cms_posts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL DEFAULT 0,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    status SMALLINT DEFAULT 1,
    author_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_cms_posts_tenant ON cms_posts (tenant_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_cms_posts_author ON cms_posts (author_id) WHERE is_deleted = FALSE;
