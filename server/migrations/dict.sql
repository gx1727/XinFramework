-- ============================================
-- Dict 模块数据表（由 apps/reference/dict 拥有）
-- 版本: PostgreSQL 14+
-- ============================================

-- dicts（字典主表）
CREATE TABLE IF NOT EXISTS dicts
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,
    code       VARCHAR(32) NOT NULL,
    name       VARCHAR(64) NOT NULL,
    status     SMALLINT    DEFAULT 1,
    sort       INT         DEFAULT 0,
    extend     JSONB       DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_dict_code ON dicts (tenant_id, code) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_dicts_tenant ON dicts (tenant_id);

-- dict_items（字典项表）
CREATE TABLE IF NOT EXISTS dict_items
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,
    dict_id    BIGINT      NOT NULL,
    code       VARCHAR(64) NOT NULL,
    name       VARCHAR(128) NOT NULL,
    sort       INT         DEFAULT 0,
    status     SMALLINT    DEFAULT 1,
    extend     JSONB       DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_dict_items_dict ON dict_items (dict_id) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_dict_items_code ON dict_items (dict_id, code) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_dict_items_tenant ON dict_items (tenant_id);