-- ============================================
-- System 模块兼容补丁
-- 目的:把 init_seed.sql 11.4 段后续追加的 sys_menus 行同步给已走过的库
--     （init_seed 已被 _schema_migrations 记录,不会重跑）
-- 幂等:ON CONFLICT DO NOTHING,新库走 init_seed、老库走本文件
-- ============================================

-- platform-cache 菜单（Cache.tsx → /platform/cache）
-- 2026-06 新增:Redis cache 运维移到 platform 域（super_admin 专属）
INSERT INTO sys_menus (id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
    OVERRIDING SYSTEM VALUE
VALUES (105, 'platform-cache', '缓存管理', 'Redis cache 运维 (Cache.tsx)', '', '/platform/cache', 'DatabaseIcon', 5, 100, '100', TRUE, TRUE)
ON CONFLICT (code) WHERE is_deleted = FALSE DO NOTHING;
