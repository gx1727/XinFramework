-- ============================================
-- Config 模块数据表（由 apps/reference/config 拥有）
-- 通用配置系统：分组 (config_categories) + 项 (config_items)
-- 版本: PostgreSQL 14+
--
-- 注意：本文件只做建表 + 索引 + RLS，不写 seed。
-- 4 个预置分组 (site/security/email/feature_flag) + 19 个预置项 +
-- 1 个菜单 (config) + 5 个资源 (config:list/get/create/update/delete)
-- 在 migrations/framework.sql 的 `bootstrap` 段写入。
-- 这样字母序执行时 config.sql 跑在 framework.sql 前后都不依赖。
-- ============================================

-- config_categories（配置分组）
CREATE TABLE IF NOT EXISTS config_categories
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id   BIGINT       NOT NULL,
    code        VARCHAR(64)  NOT NULL,
    name        VARCHAR(64)  NOT NULL,
    description VARCHAR(255),
    icon        VARCHAR(64),
    sort        INT          DEFAULT 0,
    is_system   BOOLEAN      DEFAULT FALSE,
    is_public   BOOLEAN      DEFAULT FALSE,
    status      SMALLINT     DEFAULT 1,
    created_at  TIMESTAMPTZ  DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  DEFAULT NOW(),
    is_deleted  BOOLEAN      DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_config_groups_code ON config_categories (tenant_id, code) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_config_groups_tenant ON config_categories (tenant_id) WHERE is_deleted = FALSE;

-- config_items（配置项 = 定义 + 值）
CREATE TABLE IF NOT EXISTS config_items
(
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id     BIGINT       NOT NULL,
    category_id      BIGINT       NOT NULL,
    key           VARCHAR(128) NOT NULL,
    value         JSONB        DEFAULT NULL,
    default_value JSONB        DEFAULT NULL,
    type          VARCHAR(32)  NOT NULL,
    label         VARCHAR(128),
    description   VARCHAR(512),
    options       JSONB        DEFAULT NULL,
    validation    JSONB        DEFAULT NULL,
    sort          INT          DEFAULT 0,
    is_public     BOOLEAN      DEFAULT FALSE,
    is_readonly   BOOLEAN      DEFAULT FALSE,
    is_system     BOOLEAN      DEFAULT FALSE,
    status        SMALLINT     DEFAULT 1,
    created_at    TIMESTAMPTZ  DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  DEFAULT NOW(),
    is_deleted    BOOLEAN      DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_config_items_key ON config_items (tenant_id, category_id, key) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_config_items_category ON config_items (category_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_config_items_tenant ON config_items (tenant_id) WHERE is_deleted = FALSE;

-- RLS（与 dict 一致）
ALTER TABLE config_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE config_items  ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON config_categories;
CREATE POLICY tenant_isolation_policy ON config_categories
USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);

DROP POLICY IF EXISTS tenant_isolation_policy ON config_items;
CREATE POLICY tenant_isolation_policy ON config_items
USING (
    tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
    OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
);
