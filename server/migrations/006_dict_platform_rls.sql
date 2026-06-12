-- dicts 平台级共享：dicts/dict_items RLS 放行 tenant_id=0；现有 seed 升为平台
-- 解决：每个租户都要 seed 基础字典 / 字典维护 N 份
-- 行为：tenant_id=0 的字典对所有租户可见；租户可创建同名 code 字典做覆盖（项级合并）

-- 1. dicts RLS policy 升级
DROP POLICY IF EXISTS tenant_isolation_policy ON dicts;
CREATE POLICY tenant_isolation_policy ON dicts
    USING (tenant_id = 0 OR tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT);

-- 2. dict_items RLS policy 升级
DROP POLICY IF EXISTS tenant_isolation_policy ON dict_items;
CREATE POLICY tenant_isolation_policy ON dict_items
    USING (tenant_id = 0 OR tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT);

-- 3. 现有 seed 数据（tenant_id=1 的 3 个字典 + 7 个字典项）升为平台（tenant_id=0）
-- 只在还没升过的时候改（用 NOT EXISTS 判断）
UPDATE dicts SET tenant_id = 0
WHERE tenant_id != 0
  AND code IN ('gender', 'user_status', 'education');

UPDATE dict_items SET tenant_id = 0
WHERE tenant_id != 0
  AND dict_id IN (SELECT id FROM dicts WHERE code IN ('gender', 'user_status', 'education'));
