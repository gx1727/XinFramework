-- ============================================
-- Flag App 资源初始化
-- 版本: PostgreSQL 14+
-- ============================================

SET client_encoding = 'UTF8';
SET app.tenant_id = '1';

-- ----------------------------------------
-- 用户管理 /resources -> 54 -> menu_id=54
-- ----------------------------------------
INSERT INTO resources (tenant_id, menu_id, code, name, action, description, sort, status)
OVERRIDING SYSTEM VALUE VALUES
(1, 54, 'resource:list',   '查询资源',    'GET',    '查询资源列表',    1, 1),
(1, 54, 'resource:get',    '查看资源',    'GET',    '查看单个资源详情', 2, 1),
(1, 54, 'resource:create', '创建资源',    'POST',   '新建资源',        3, 1),
(1, 54, 'resource:update', '更新资源',   'PUT',    '更新资源信息',    4, 1),
(1, 54, 'resource:delete', '删除资源',   'DELETE', '删除资源',        5, 1);

-- 重置序列，确保后续自增ID不冲突
SELECT setval('resources_id_seq', 100, true);