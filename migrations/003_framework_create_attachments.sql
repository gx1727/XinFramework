-- 删除已存在的attachments表（如果存在）
DROP TABLE IF EXISTS attachments;

-- Create attachments table for Asset Service
CREATE TABLE IF NOT EXISTS attachments ( 
    id              BIGSERIAL PRIMARY KEY,                       -- 主键ID，自增序列
    tenant_id       BIGINT NOT NULL DEFAULT 0,                   -- 租户ID，用于多租户隔离
    user_id         BIGINT,                                      -- 上传用户ID
    
    file_name       TEXT,                                        -- 原始文件名
    file_ext        VARCHAR(20),                                 -- 文件扩展名
    mime_type       VARCHAR(100),                                -- MIME类型
    
    file_size       BIGINT,                                      -- 文件大小(字节)
    storage         VARCHAR(20),                                 -- 存储方式: local / oss / s3 
    
    object_key      TEXT,                                        -- 存储路径(关键)
    url             TEXT,                                        -- 访问URL
    
    hash            VARCHAR(64),                                 -- 去重用（md5/sha256） 
    
    status          SMALLINT DEFAULT 1,                          -- 1正常 0删除 
    
    created_at      TIMESTAMPTZ DEFAULT NOW(),                   -- 创建时间
    updated_at      TIMESTAMPTZ DEFAULT NOW(),                   -- 更新时间
    is_deleted      BOOLEAN DEFAULT FALSE                        -- 软删除标记
);

-- Index for tenant deduplication
CREATE INDEX IF NOT EXISTS idx_attachments_tenant_hash ON attachments(tenant_id, hash) WHERE is_deleted = FALSE;

-- 添加表注释
COMMENT ON TABLE attachments IS '附件资源表';

-- 添加字段注释
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
-- 注意：attachments 表的多租户隔离以应用层 SET app.tenant_id 为主要机制，RLS 作为纵深防御。
-- app.mode 配置：
--   single：不约束 tenant_id（放行所有行）
--   saas：必须约束 tenant_id（tenant_id 必须匹配）
-- is_deleted = TRUE 的行默认不可见，除非 SET app.show_deleted = true
ALTER TABLE attachments ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON attachments
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
