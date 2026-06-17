-- ============================================
-- Asset 模块数据表（由 apps/reference/asset 拥有）
-- 版本: PostgreSQL 14+
-- ============================================

-- attachments（附件表）
CREATE TABLE IF NOT EXISTS attachments
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    user_id    BIGINT,
    file_name  VARCHAR(255),
    file_ext   VARCHAR(32),
    mime_type  VARCHAR(64),
    file_size  BIGINT,
    storage    VARCHAR(32),
    object_key VARCHAR(255),
    url        VARCHAR(512),
    hash       VARCHAR(64),
    status     SMALLINT    DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_attachments_tenant ON attachments (tenant_id) WHERE is_deleted = FALSE;