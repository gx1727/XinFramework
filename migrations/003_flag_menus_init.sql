-- ============================================
-- Flag App 菜单初始化
-- 版本: PostgreSQL 14+
-- ============================================

SET client_encoding = 'UTF8';
SET app.mode = 'single';

-- 一级菜单
INSERT INTO menus (id, tenant_id, code, name, path, icon, sort, parent_id, ancestors, visible, enabled)
OVERRIDING SYSTEM VALUE VALUES
(1,  0, 'dashboard', '仪表盘',   '/dashboard',       'LayoutDashboardIcon', 1, 0, '1',  TRUE, TRUE),
(2,  0, 'analytics', '数据分析', '/analytics',        'ChartBarIcon',        2, 0, '2',  TRUE, TRUE),
(3,  0, 'projects',  '项目管理', '/projects',         'FolderIcon',          3, 0, '3',  TRUE, TRUE),
(4,  0, 'team',      '团队管理', '/team',             'UsersIcon',           4, 0, '4',  TRUE, TRUE),
(6,  0, 'frames',    '相框管理', '/frames',           'FrameIcon',           6, 0, '6',  TRUE, TRUE),
(7,  0, 'avatars',   '头像管理', '/avatars',          'ImageIcon',           7, 0, '7',  TRUE, TRUE),
(5,  0, 'system',    '系统管理', '/system',           'SettingsIcon',        5, 0, '5',  TRUE, TRUE);

-- 二级菜单 - 相框管理
INSERT INTO menus (id, tenant_id, code, name, path, icon, sort, parent_id, ancestors, visible, enabled)
OVERRIDING SYSTEM VALUE VALUES
(61, 0, 'frame-list',       '相框列表', '/frames',           'FileIcon', 1, 6, '6.61',  TRUE, TRUE),
(62, 0, 'frame-categories', '相框分类', '/frame-categories', 'ListIcon', 2, 6, '6.62',  TRUE, TRUE);

-- 二级菜单 - 头像管理
INSERT INTO menus (id, tenant_id, code, name, path, icon, sort, parent_id, ancestors, visible, enabled)
OVERRIDING SYSTEM VALUE VALUES
(71, 0, 'avatar-list',       '头像列表', '/avatars',           'FileIcon', 1, 7, '7.71',  TRUE, TRUE),
(72, 0, 'avatar-categories', '头像分类', '/avatar-categories', 'ListIcon', 2, 7, '7.72',  TRUE, TRUE);

-- 二级菜单 - 系统管理
INSERT INTO menus (id, tenant_id, code, name, path, icon, sort, parent_id, ancestors, visible, enabled)
OVERRIDING SYSTEM VALUE VALUES
(51, 0, 'users', '用户管理', '/users', 'FileIcon',    1, 5, '5.51',  TRUE, TRUE),
(52, 0, 'roles', '角色管理', '/roles', 'ShieldIcon',  2, 5, '5.52',  TRUE, TRUE),
(53, 0, 'menus', '菜单管理', '/menus', 'MenuIcon',    3, 5, '5.53',  TRUE, TRUE);

-- 重置序列，确保后续自增ID不冲突
SELECT setval('menus_id_seq', 200, true);
