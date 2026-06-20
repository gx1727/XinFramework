-- ============================================
-- Dict 模块数据表（由 apps/reference/dict 拥有）
-- 版本: PostgreSQL 14+
-- 包含：platform/tenant 分层、可见性矩阵、租户覆盖
-- ============================================

-- dicts（字典主表）
-- scope:      'platform' | 'tenant'   — 平台级由 super_admin 维护；租户级由租户自建
-- visibility: 'all' | 'whitelist' | 'blacklist' — 平台字典对租户的可见性策略
CREATE TABLE IF NOT EXISTS dicts
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL,
    code       VARCHAR(32) NOT NULL,
    name       VARCHAR(64) NOT NULL,
    scope      VARCHAR(16) NOT NULL DEFAULT 'tenant',
    visibility VARCHAR(16) NOT NULL DEFAULT 'all',
    status     SMALLINT    DEFAULT 1,
    sort       INT         DEFAULT 0,
    extend     JSONB       DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN     DEFAULT FALSE
);
-- 平台字典按 code 全局唯一；租户字典按 (tenant_id, code) 唯一
CREATE UNIQUE INDEX IF NOT EXISTS uk_dict_code_platform
    ON dicts (code) WHERE scope = 'platform' AND is_deleted = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_dict_code_tenant
    ON dicts (tenant_id, code) WHERE scope = 'tenant' AND is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_dicts_tenant ON dicts (tenant_id);

-- dict_items（字典项表）
-- platform_item_id: 指向被覆盖的 dict_items.id（仅 override 行非空）
-- is_override:      TRUE 表示这是租户对某 platform_item 的覆盖
CREATE TABLE IF NOT EXISTS dict_items
(
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id        BIGINT       NOT NULL,
    dict_id          BIGINT       NOT NULL,
    code             VARCHAR(64)  NOT NULL,
    name             VARCHAR(128) NOT NULL,
    platform_item_id BIGINT,
    is_override      BOOLEAN      NOT NULL DEFAULT FALSE,
    sort             INT          DEFAULT 0,
    status           SMALLINT     DEFAULT 1,
    extend           JSONB        DEFAULT '{}'::jsonb,
    created_at       TIMESTAMPTZ  DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  DEFAULT NOW(),
    is_deleted       BOOLEAN      DEFAULT FALSE
);
-- 平台字典项按 (dict_id, code) 唯一
CREATE UNIQUE INDEX IF NOT EXISTS uk_dict_item_platform
    ON dict_items (dict_id, code) WHERE tenant_id = 0 AND is_deleted = FALSE;
-- 租户字典项按 (tenant_id, dict_id, code) 唯一
CREATE UNIQUE INDEX IF NOT EXISTS uk_dict_item_tenant
    ON dict_items (tenant_id, dict_id, code)
    WHERE tenant_id <> 0 AND is_deleted = FALSE;
-- 同一租户对同一 platform_item 仅一条覆盖
CREATE UNIQUE INDEX IF NOT EXISTS uk_dict_item_override
    ON dict_items (tenant_id, platform_item_id)
    WHERE is_override = TRUE AND is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_dict_items_dict ON dict_items (dict_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_dict_items_tenant ON dict_items (tenant_id);
-- 覆盖查询加速
CREATE INDEX IF NOT EXISTS idx_dict_items_platform_ref
    ON dict_items (dict_id, id) WHERE tenant_id = 0 AND is_deleted = FALSE;

-- dict_visibility：平台字典对各租户的访问级别
-- access: 'invisible' | 'readonly' | 'editable'
CREATE TABLE IF NOT EXISTS dict_visibility
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    dict_id    BIGINT      NOT NULL,
    tenant_id  BIGINT      NOT NULL,
    access     VARCHAR(16) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT uk_dict_visibility UNIQUE (dict_id, tenant_id)
);
CREATE INDEX IF NOT EXISTS idx_dict_visibility_tenant ON dict_visibility (tenant_id);
CREATE INDEX IF NOT EXISTS idx_dict_visibility_dict   ON dict_visibility (dict_id);