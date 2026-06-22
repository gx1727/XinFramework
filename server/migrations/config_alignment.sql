-- ============================================
-- Config 模块对齐 Dict 的结构改造（Phase 0022）
-- 与 dict.sql 的设计对齐：
--   - config_categories 加 scope / visibility 字段
--   - config_items 加 platform_item_id / is_override 字段
--   - 新建 config_visibility 表（与 dict_visibility 同构）
--   - 索引按 platform/tenant 分层（部分唯一索引）
--
-- 运行时机：迁移器发现此文件 → 顺序执行（migration.Run 已支持按文件名排序）
-- 向后兼容：
--   - 所有新增字段都有 DEFAULT，旧数据自动填充
--   - 原 uk_config_group_code 索引被替换为 platform/tenant 双索引
--   - 不删除任何旧列、不破坏已部署数据
-- ============================================

-- 1. config_categories 加 scope / visibility
-- scope:      'platform' | 'tenant'
--   platform = 平台级，super_admin 维护，所有租户继承
--   tenant   = 租户级，租户自建
-- visibility: 'all' | 'whitelist' | 'blacklist'
ALTER TABLE config_categories
    ADD COLUMN IF NOT EXISTS scope VARCHAR(16) NOT NULL DEFAULT 'tenant';
ALTER TABLE config_categories
    ADD COLUMN IF NOT EXISTS visibility VARCHAR(16) NOT NULL DEFAULT 'all';

-- 2. config_items 加 platform_item_id / is_override
-- platform_item_id: 指向被覆盖的 platform config_item.id（仅 override 行非空）
-- is_override:      TRUE 表示这是租户对某 platform_item 的覆盖
ALTER TABLE config_items
    ADD COLUMN IF NOT EXISTS platform_item_id BIGINT;
ALTER TABLE config_items
    ADD COLUMN IF NOT EXISTS is_override BOOLEAN NOT NULL DEFAULT FALSE;

-- 3. config_visibility 表
-- 与 dict_visibility 同构：平台 group 对各租户的访问级别
-- access: 'invisible' | 'readonly' | 'editable'
CREATE TABLE IF NOT EXISTS config_visibility
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    category_id   BIGINT      NOT NULL,
    tenant_id  BIGINT      NOT NULL,
    access     VARCHAR(16) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT uk_config_visibility UNIQUE (category_id, tenant_id)
);
CREATE INDEX IF NOT EXISTS idx_config_visibility_tenant ON config_visibility (tenant_id);
CREATE INDEX IF NOT EXISTS idx_config_visibility_category  ON config_visibility (category_id);

-- 4. 替换 config_categories 唯一索引为 platform/tenant 分层
-- 旧索引：uk_config_group_code ON (tenant_id, code)
-- 新索引：
--   platform code 全局唯一（tenant_id=0）
--   tenant   code 按 (tenant_id, code) 唯一
DROP INDEX IF EXISTS uk_config_group_code;
CREATE UNIQUE INDEX IF NOT EXISTS uk_config_group_code_platform
    ON config_categories (code) WHERE scope = 'platform' AND is_deleted = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_config_group_code_tenant
    ON config_categories (tenant_id, code) WHERE scope = 'tenant' AND is_deleted = FALSE;

-- 5. config_items 覆盖索引
-- 同一租户对同一 platform_item 仅一条覆盖
CREATE UNIQUE INDEX IF NOT EXISTS uk_config_item_override
    ON config_items (tenant_id, platform_item_id)
    WHERE is_override = TRUE AND is_deleted = FALSE;

-- 6. config_items 按 platform/tenant 分层的 (category_id, key) 唯一
-- 旧 uk_config_item_key 索引可能只按 (category_id, key)
-- 需要替换为双索引（与 dict_items 同款）
DROP INDEX IF EXISTS uk_config_item_key;
CREATE UNIQUE INDEX IF NOT EXISTS uk_config_item_key_platform
    ON config_items (category_id, key) WHERE tenant_id = 0 AND is_deleted = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_config_item_key_tenant
    ON config_items (tenant_id, category_id, key)
    WHERE tenant_id <> 0 AND is_deleted = FALSE;

-- 7. config_items 平台项引用加速
CREATE INDEX IF NOT EXISTS idx_config_items_platform_ref
    ON config_items (category_id, id) WHERE tenant_id = 0 AND is_deleted = FALSE;
