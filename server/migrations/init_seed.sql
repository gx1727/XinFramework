-- ============================================
-- init_seed.sql : 全部种子数据（开发期统一入口）
-- ============================================
-- 依赖：先跑 init_schema.sql。
--
-- 0023.3 关键变化：
--   - account_roles 表 drop → admin 走 sys_users + sys_user_roles
--   - menus → tenant_menus（去掉 scope 字段）
--   - resources → tenant_permissions
--   - users → tenant_users
--   - roles → tenant_roles
--   - user_roles → tenant_user_roles
--   - role_menus / role_resources / role_data_scopes / organizations 全部 tenant_ 前缀
--
-- 内容（按种子概念分组）：
--   1. 平台账号（admin 账号）
--   2. bootstrap 租户
--   3. admin 用户 / 角色 / 用户-角色（bootstrap 租户下）
--   4. 基础菜单（dashboard / system 及其子菜单）
--   5. 权限码（resource / dict / 超级通配）
--   6. admin 角色绑菜单 + 通配权限
--   7. 字典（平台级 gender / user_status / education + bootstrap 副本）
--   8. Config 模块（4 个分组 + 19 个项）
--   9. 平台管理菜单（tenants / config / frames / avatars）
--  10. flag 资源 seed
--  11. 平台域 sys_* 种子（admin 账号的 sys_user + super_admin 角色绑定）
--  12. 序列号兜底
--
-- 全部用 ON CONFLICT DO NOTHING / 条件 UPDATE —— 重复跑不会出错。
-- 业务模块（asset / cms / flag）如果有 seed 在各自 .sql 里；本文件只覆盖核心域。
-- ============================================

-- ============================================
-- 1. 平台账号
-- ============================================
-- 0023.3 终态：admin 账号走 sys_users + sys_user_roles，不再有 account_roles 表。
-- 登录路径：accounts → sys_users → sys_user_roles → sys_roles.code = 'super_admin'
--
-- 1.1 账号 (password: admin123)
-- 此 hash 必须与 framework/pkg/auth.HashPassword 输出格式一致（m=65536,t=3,p=4）。
INSERT INTO accounts (phone, email, password, username, real_name, status)
VALUES ('13800138000', 'admin@example.com',
        '$argon2id$v=19$m=65536,t=3,p=4$l9OpXE4q2opC5q1SZSSFMg$sKlfP3vLGM+/UJPa51OLGonHhYmsACGYjV9f8AveDes', 'admin',
        '系统管理员', 1)
ON CONFLICT (phone) WHERE is_deleted = FALSE DO NOTHING;

-- ============================================
-- 2. bootstrap 租户
-- ============================================
-- 单一系统租户：是 admin 用户的"初始租户管理员"居住地（路径 B）。
-- 同时是"新租户克隆源"——first_install.go 从这里复制 menus / permissions / dicts / config_*。
INSERT INTO tenants (code, name, status, created_by, updated_by)
VALUES ('bootstrap', '[系统] Bootstrap 租户', 1, 0, 0)
ON CONFLICT (code) WHERE is_deleted = FALSE DO NOTHING;

-- ============================================
-- 3. admin 用户 / 角色 / 用户-角色（bootstrap 租户下）
-- ============================================

-- 3.1 admin 用户（tenant_users）
INSERT INTO tenant_users (tenant_id, account_id, code, org_id, real_name, status, created_by, updated_by)
VALUES (1, 1, 'admin', NULL, '系统管理员', 1, 0, 0)
ON CONFLICT (account_id, tenant_id) WHERE is_deleted = FALSE DO NOTHING;

-- 3.2 角色：admin + user（tenant_roles）
INSERT INTO tenant_roles (tenant_id, code, name, description, data_scope, is_default, sort, status, created_by, updated_by)
VALUES (1, 'admin', '管理员', '系统管理员', 1, FALSE, 1, 1, 0, 0),
       (1, 'user', '普通用户', '普通用户', 4, TRUE, 2, 1, 0, 0)
ON CONFLICT (tenant_id, code) WHERE is_deleted = FALSE DO NOTHING;

-- 3.3 用户-角色：admin → admin 角色（tenant_user_roles）
INSERT INTO tenant_user_roles (tenant_id, user_id, role_id)
VALUES (1, 1, 1)
ON CONFLICT (user_id, role_id) WHERE is_deleted = FALSE DO NOTHING;

-- ============================================
-- 4. 基础菜单（框架 + system 子菜单，tenant_menus）
-- ============================================
INSERT INTO tenant_menus (id, tenant_id, code, name, path, icon, sort, parent_id, ancestors, visible, enabled)
    OVERRIDING SYSTEM VALUE
VALUES (1, 1, 'dashboard', '仪表盘', '/dashboard', 'LayoutDashboardIcon', 1, 0, '1', TRUE, TRUE),
       (2, 1, 'analytics', '数据分析', '/analytics', 'ChartBarIcon', 2, 0, '2', TRUE, TRUE),
       (3, 1, 'projects', '项目管理', '/projects', 'FolderIcon', 3, 0, '3', TRUE, TRUE),
       (4, 1, 'team', '团队管理', '/team', 'UsersIcon', 4, 0, '4', TRUE, TRUE),
       (5, 1, 'system', '系统管理', '/system', 'SettingsIcon', 5, 0, '5', TRUE, TRUE);

INSERT INTO tenant_menus (id, tenant_id, code, name, path, icon, sort, parent_id, ancestors, visible, enabled)
    OVERRIDING SYSTEM VALUE
VALUES (51, 1, 'users', '用户管理', '/users', 'FileIcon', 1, 5, '5.51', TRUE, TRUE),
       (52, 1, 'roles', '角色管理', '/roles', 'ShieldIcon', 2, 5, '5.52', TRUE, TRUE),
       (53, 1, 'menus', '菜单管理', '/menus', 'MenuIcon', 3, 5, '5.53', TRUE, TRUE),
       (54, 1, 'resources', '资源管理', '/resources', 'ResourceIcon', 4, 5, '5.54', TRUE, TRUE),
       (55, 1, 'organizations', '组织管理', '/organizations', 'ResourceIcon', 5, 5, '5.55', TRUE, TRUE),
       (56, 1, 'dicts', '数据字典', '/dicts', 'BookIcon', 6, 5, '5.56', TRUE, TRUE);

SELECT setval('tenant_menus_id_seq', 300, true);

-- ============================================
-- 5. 权限码（tenant_permissions，原 resources）
-- ============================================

-- 资源管理下的权限码
INSERT INTO tenant_permissions (tenant_id, menu_id, code, name, action, description, sort, status)
VALUES (1, 54, 'resource:list', '查询资源', 'list', '查询资源列表', 1, 1),
       (1, 54, 'resource:get', '查看资源', 'get', '查看单个资源详情', 2, 1),
       (1, 54, 'resource:create', '创建资源', 'create', '新建资源', 3, 1),
       (1, 54, 'resource:update', '更新资源', 'update', '更新资源信息', 4, 1),
       (1, 54, 'resource:delete', '删除资源', 'delete', '删除资源', 5, 1);

-- 字典管理下的权限码
INSERT INTO tenant_permissions (tenant_id, menu_id, code, name, action, description, sort, status)
VALUES (1, 56, 'dict:list', '查询字典', 'list', '查询字典列表', 1, 1),
       (1, 56, 'dict:get', '查看字典', 'get', '查看单个字典及字典项', 2, 1),
       (1, 56, 'dict:create', '创建字典', 'create', '新建字典', 3, 1),
       (1, 56, 'dict:update', '更新字典', 'update', '更新字典及字典项', 4, 1),
       (1, 56, 'dict:delete', '删除字典', 'delete', '删除字典及字典项', 5, 1);

-- 超级通配权限码（0024+ 约定：code 用完整串 "*:*"；运行时按 *:* 走全局通配）
INSERT INTO tenant_permissions (tenant_id, code, name, action, description, status)
VALUES (1, '*:*', '超级管理员通配权限', '*', '拥有系统所有权限', 1);

SELECT setval('tenant_permissions_id_seq', 100, true);

-- ============================================
-- admin 角色绑菜单 + 绑通配权限码
INSERT INTO tenant_role_menus (tenant_id, role_id, menu_id)
SELECT 1, 1, id FROM tenant_menus WHERE is_deleted = FALSE;

INSERT INTO tenant_role_resources (tenant_id, role_id, permission_id, effect)
VALUES (1, 1, (SELECT id FROM tenant_permissions WHERE code = '*:*'), 1);

-- ============================================
-- 7. 字典（平台级 + bootstrap 副本）
-- ============================================
-- 7.1 平台级字典（tenant_id=0）
INSERT INTO dicts (tenant_id, code, name, sort, status)
VALUES (0, 'gender', '性别', 1, 1),
       (0, 'user_status', '用户状态', 2, 1),
       (0, 'education', '学历', 3, 1);

INSERT INTO dict_items (tenant_id, dict_id, code, name, sort, status)
SELECT 0, d.id, x.code, x.name, x.sort, 1
FROM dicts d
JOIN (VALUES
  ('gender', 'male', '男', 1),
  ('gender', 'female', '女', 2),
  ('user_status', 'active', '启用', 1),
  ('user_status', 'disabled', '停用', 2),
  ('education', 'bachelor', '本科', 1),
  ('education', 'master', '硕士', 2),
  ('education', 'doctor', '博士', 3)
) AS x(dict_code, code, name, sort) ON x.dict_code = d.code;

-- 7.2 bootstrap 租户的字典副本（从 platform 复制，新租户 first_install 时再从这里克隆）
INSERT INTO dicts (tenant_id, code, name, sort, status, extend)
SELECT (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE),
       code, name, sort, status, extend
FROM dicts
WHERE tenant_id = 0 AND is_deleted = FALSE
ON CONFLICT (tenant_id, code) WHERE tenant_id <> 0 AND is_deleted = FALSE DO NOTHING;

INSERT INTO dict_items (tenant_id, dict_id, code, name, sort, status, extend)
SELECT (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE),
       new_d.id, di.code, di.name, di.sort, di.status, di.extend
FROM dict_items di
JOIN dicts old_d ON old_d.id = di.dict_id AND old_d.tenant_id = 0 AND old_d.is_deleted = FALSE
JOIN dicts new_d ON new_d.code = old_d.code
                AND new_d.tenant_id = (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE)
                AND new_d.is_deleted = FALSE
WHERE di.tenant_id = 0 AND di.is_deleted = FALSE
ON CONFLICT (tenant_id, dict_id, code) WHERE tenant_id <> 0 AND is_deleted = FALSE DO NOTHING;

-- ============================================
-- 8. Config 模块（4 个分组 + 19 个项）
-- ============================================

-- 8.1 config_categories（4 个预置分组）
INSERT INTO config_categories (tenant_id, code, name, description, icon, sort, is_system, is_public)
SELECT (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE),
       'site', '站点信息', '站点名称、Logo、版权等公开信息', 'GlobeIcon', 1, TRUE, TRUE
ON CONFLICT (tenant_id, code) WHERE tenant_id <> 0 AND is_deleted = FALSE DO NOTHING;

INSERT INTO config_categories (tenant_id, code, name, description, icon, sort, is_system, is_public)
SELECT (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE),
       'security', '安全策略', '密码强度、会话超时等安全相关配置', 'ShieldIcon', 2, TRUE, FALSE
ON CONFLICT (tenant_id, code) WHERE tenant_id <> 0 AND is_deleted = FALSE DO NOTHING;

INSERT INTO config_categories (tenant_id, code, name, description, icon, sort, is_system, is_public)
SELECT (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE),
       'email', '邮件服务', 'SMTP 邮件服务配置', 'MailIcon', 3, TRUE, FALSE
ON CONFLICT (tenant_id, code) WHERE tenant_id <> 0 AND is_deleted = FALSE DO NOTHING;

INSERT INTO config_categories (tenant_id, code, name, description, icon, sort, is_system, is_public)
SELECT (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE),
       'feature_flag', '功能开关', '系统级功能启用/禁用开关', 'ToggleLeftIcon', 4, TRUE, FALSE
ON CONFLICT (tenant_id, code) WHERE tenant_id <> 0 AND is_deleted = FALSE DO NOTHING;

-- 8.2 site items
INSERT INTO config_items (tenant_id, category_id, key, value, default_value, type, label, description, sort, is_public, is_system)
SELECT
    (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE),
    (SELECT id FROM config_categories WHERE code = 'site' AND tenant_id = (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE) AND is_deleted = FALSE),
    s.key, s.value, s.default_value, s.type, s.label, s.description, s.sort, s.is_public, TRUE
FROM (VALUES
    ('site_name',         '"XinFramework"'::jsonb, '"XinFramework"'::jsonb, 'string', '站点名称', '显示在页面标题、登录页等位置', 1, TRUE),
    ('site_logo',         '""'::jsonb,              '""'::jsonb,             'image',  '站点 Logo', '建议 PNG/SVG，背景透明', 2, TRUE),
    ('site_favicon',      '""'::jsonb,              '""'::jsonb,             'image',  'Favicon',  '浏览器标签图标', 3, TRUE),
    ('site_copyright',    '""'::jsonb,              '""'::jsonb,             'string', '版权信息', '页面底部显示', 4, TRUE),
    ('site_icp',          '""'::jsonb,              '""'::jsonb,             'string', 'ICP 备案号', '中国大陆站点必填', 5, TRUE),
    ('site_locale_default', '"zh-CN"'::jsonb,       '"zh-CN"'::jsonb,        'select', '默认语言', 'zh-CN / en-US', 6, TRUE),
    ('login_background',  '""'::jsonb,              '""'::jsonb,             'image',  '登录页背景', '登录页右侧大图', 7, TRUE)
) AS s(key, value, default_value, type, label, description, sort, is_public)
ON CONFLICT (tenant_id, category_id, key) WHERE tenant_id <> 0 AND is_deleted = FALSE DO NOTHING;

-- 8.3 security items
INSERT INTO config_items (tenant_id, category_id, key, value, default_value, type, label, description, validation, sort, is_public, is_system)
SELECT
    (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE),
    (SELECT id FROM config_categories WHERE code = 'security' AND tenant_id = (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE) AND is_deleted = FALSE),
    s.key, s.value, s.default_value, s.type, s.label, s.description, s.validation, s.sort, FALSE, TRUE
FROM (VALUES
    ('password_min_length',  '8'::jsonb,    '8'::jsonb,    'number', '密码最小长度', '新建/修改密码时校验',           '{"min":6,"max":32,"required":true}'::jsonb, 1),
    ('password_complexity',  '"standard"'::jsonb, '"standard"'::jsonb, 'select', '密码复杂度', 'low/standard/strong', '[{"label":"低(纯字母数字)","value":"low"},{"label":"标准(字母+数字)","value":"standard"},{"label":"强(字母+数字+符号)","value":"strong"}]'::jsonb, 2),
    ('session_timeout_min',  '30'::jsonb,   '30'::jsonb,   'number', '会话超时(分钟)', '空闲超过此时间强制下线',       '{"min":5,"max":1440,"required":true}'::jsonb, 3),
    ('max_login_attempts',   '5'::jsonb,    '5'::jsonb,    'number', '最大登录失败次数', '超过后锁定账户',                '{"min":1,"max":20,"required":true}'::jsonb, 4),
    ('lock_duration_min',    '5'::jsonb,    '5'::jsonb,    'number', '锁定时长(分钟)',   '失败次数超限后的锁定时长',       '{"min":1,"max":1440,"required":true}'::jsonb, 5)
) AS s(key, value, default_value, type, label, description, validation, sort)
ON CONFLICT (tenant_id, category_id, key) WHERE tenant_id <> 0 AND is_deleted = FALSE DO NOTHING;

-- 8.4 email items
INSERT INTO config_items (tenant_id, category_id, key, value, default_value, type, label, description, sort, is_public, is_readonly, is_system)
SELECT
    (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE),
    (SELECT id FROM config_categories WHERE code = 'email' AND tenant_id = (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE) AND is_deleted = FALSE),
    s.key, s.value, s.default_value, s.type, s.label, s.description, s.sort, FALSE, s.is_readonly, TRUE
FROM (VALUES
    ('smtp_host',     '""'::jsonb,         '""'::jsonb,         'string',  'SMTP 主机',   '如 smtp.example.com',  1, FALSE),
    ('smtp_port',     '465'::jsonb,        '465'::jsonb,        'number',  'SMTP 端口',   '常用 25/465/587',     2, FALSE),
    ('smtp_user',     '""'::jsonb,         '""'::jsonb,         'string',  'SMTP 用户',   '通常为邮箱地址',       3, FALSE),
    ('smtp_password', '""'::jsonb,         '""'::jsonb,         'password','SMTP 密码',   '授权码或登录密码',     4, TRUE),
    ('smtp_from',     '""'::jsonb,         '""'::jsonb,         'string',  '发件人邮箱',  '邮件 From 头',         5, FALSE),
    ('smtp_use_tls',  'true'::jsonb,       'true'::jsonb,       'boolean', '启用 TLS',    '465 通常 TLS，587 STARTTLS', 6, FALSE)
) AS s(key, value, default_value, type, label, description, sort, is_readonly)
ON CONFLICT (tenant_id, category_id, key) WHERE tenant_id <> 0 AND is_deleted = FALSE DO NOTHING;

-- 8.5 feature_flag items
INSERT INTO config_items (tenant_id, category_id, key, value, default_value, type, label, description, sort, is_public, is_system)
SELECT
    (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE),
    (SELECT id FROM config_categories WHERE code = 'feature_flag' AND tenant_id = (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE) AND is_deleted = FALSE),
    s.key, s.value, s.default_value, s.type, s.label, s.description, s.sort, FALSE, TRUE
FROM (VALUES
    ('enable_registration', 'true'::jsonb, 'true'::jsonb, 'boolean', '开放注册', '允许外部用户自助注册', 1),
    ('enable_audit_log',    'true'::jsonb, 'true'::jsonb, 'boolean', '审计日志', '记录关键操作审计日志', 2)
) AS s(key, value, default_value, type, label, description, sort)
ON CONFLICT (tenant_id, category_id, key) WHERE tenant_id <> 0 AND is_deleted = FALSE DO NOTHING;

-- ============================================
-- 9. 平台管理菜单（tenants / config / frames / avatars，tenant_menus）
-- ============================================
-- 这些菜单在 bootstrap 租户里，是 admin 登录后能看到的"平台管理"入口。
-- first_install.go 会全量复制到新租户。

-- 9.1 租户管理（id=100）
INSERT INTO tenant_menus (id, tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
    OVERRIDING SYSTEM VALUE
VALUES (100, 1, 'tenants', '租户管理', '跨租户平台管理', '', '/tenants', 'Building2Icon', 0, 0, '', TRUE, TRUE)
ON CONFLICT (tenant_id, code) WHERE tenant_id <> 0 AND is_deleted = FALSE DO NOTHING;

SELECT setval('tenant_menus_id_seq', GREATEST((SELECT COALESCE(MAX(id), 0) FROM tenant_menus), 1000), true);

-- 9.2 平台配置管理（id=101）
INSERT INTO tenant_menus (id, tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
    OVERRIDING SYSTEM VALUE
SELECT 101,
       (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE),
       'config', '配置管理', '系统配置项管理', '', '/settings', 'SlidersHorizontalIcon', 1, 0, '', TRUE, TRUE
ON CONFLICT (tenant_id, code) WHERE tenant_id <> 0 AND is_deleted = FALSE DO NOTHING;

-- 9.3 相框管理（顶级）
INSERT INTO tenant_menus (tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
SELECT (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE),
       'frames', '相框管理', '头像框与活动空间', '', '/frames', 'FrameIcon', 6, 0, '', TRUE, TRUE
ON CONFLICT (tenant_id, code) WHERE tenant_id <> 0 AND is_deleted = FALSE DO NOTHING;

-- 9.4 头像管理（顶级）
INSERT INTO tenant_menus (tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
SELECT (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE),
       'avatars', '头像管理', '用户头像与分类', '', '/avatars', 'ImageIcon', 7, 0, '', TRUE, TRUE
ON CONFLICT (tenant_id, code) WHERE tenant_id <> 0 AND is_deleted = FALSE DO NOTHING;

-- 9.5 相框/头像子菜单
INSERT INTO tenant_menus (tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
SELECT (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE),
       s.code, s.name, s.subtitle, s.url, s.path, s.icon, s.sort,
       (SELECT id FROM tenant_menus WHERE code = 'frames' AND tenant_id = (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE) AND is_deleted = FALSE),
       '', TRUE, TRUE
FROM (VALUES
    ('frame-list',        '相框列表', '', '', '/frames',           'FileIcon',  1),
    ('frame-categories',  '相框分类', '', '', '/frame-categories', 'ListIcon',  2)
) AS s(code, name, subtitle, url, path, icon, sort)
ON CONFLICT (tenant_id, code) WHERE tenant_id <> 0 AND is_deleted = FALSE DO NOTHING;

INSERT INTO tenant_menus (tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
SELECT (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE),
       s.code, s.name, s.subtitle, s.url, s.path, s.icon, s.sort,
       (SELECT id FROM tenant_menus WHERE code = 'avatars' AND tenant_id = (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE) AND is_deleted = FALSE),
       '', TRUE, TRUE
FROM (VALUES
    ('avatar-list',        '头像列表', '', '', '/avatars',           'FileIcon',  1),
    ('avatar-categories',  '头像分类', '', '', '/avatar-categories', 'ListIcon',  2)
) AS s(code, name, subtitle, url, path, icon, sort)
ON CONFLICT (tenant_id, code) WHERE tenant_id <> 0 AND is_deleted = FALSE DO NOTHING;

-- 9.6 重建 frames/avatars 子菜单的 ancestors
UPDATE tenant_menus SET ancestors = parent_id::text
WHERE tenant_id = (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE)
  AND code IN ('frame-list', 'frame-categories', 'avatar-list', 'avatar-categories')
  AND parent_id > 0 AND is_deleted = FALSE;

-- ============================================
-- 10. flag 资源 seed
-- ============================================
INSERT INTO tenant_permissions (tenant_id, menu_id, code, name, action, description, sort, status)
SELECT (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE),
       (SELECT id FROM tenant_menus WHERE code = s.menu_code AND tenant_id = (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE) AND is_deleted = FALSE),
       s.code, s.name, s.action, s.description, s.sort, 1
FROM (VALUES
    ('flag:list',   '查询相框/头像', 'list',   '查询相框、头像、活动空间',  1, 'frames'),
    ('flag:get',    '查看详情',      'get',    '查看相框/头像/空间详情',   2, 'frames'),
    ('flag:create', '创建相框/头像', 'create', '创建相框/头像/活动空间',   3, 'frames'),
    ('flag:update', '更新相框/头像', 'update', '更新相框/头像/活动空间',   4, 'frames'),
    ('flag:delete', '删除相框/头像', 'delete', '软删相框/头像/活动空间',   5, 'frames'),
    ('flag:list',   '查询相框/头像', 'list',   '查询相框、头像、活动空间',  1, 'avatars'),
    ('flag:get',    '查看详情',      'get',    '查看相框/头像/空间详情',   2, 'avatars'),
    ('flag:create', '创建相框/头像', 'create', '创建相框/头像/活动空间',   3, 'avatars'),
    ('flag:update', '更新相框/头像', 'update', '更新相框/头像/活动空间',   4, 'avatars'),
    ('flag:delete', '删除相框/头像', 'delete', '软删相框/头像/活动空间',   5, 'avatars')
) AS s(code, name, action, description, sort, menu_code)
ON CONFLICT (tenant_id, code) WHERE tenant_id <> 0 AND is_deleted = FALSE DO NOTHING;

-- 10.1 平台域资源权限码占位（已迁移到 sys_permissions，见 11.3b2）
-- 历史：曾将 tenant:list/get/create/update/delete 放在 tenant_permissions（模板租户 id=1），
-- 但 apps/platform/tenants/routes.go 的 ResTenant="tenant" 拼出的是 platform 域的 key。
-- 0024+ 终态：tenant 域不放 tenant:*（单一事实源 = platform 域 sys_permissions）。

INSERT INTO tenant_permissions (tenant_id, menu_id, code, name, action, description, sort, status)
SELECT (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE),
       (SELECT id FROM tenant_menus WHERE code = 'config' AND tenant_id = (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE) AND is_deleted = FALSE),
       r.code, r.name, r.action, r.description, r.sort, 1
FROM (VALUES
    ('config:list',   '查询配置', 'list',   '查询配置分组与项',  1),
    ('config:get',    '查看配置', 'get',    '查看分组/项详情',   2),
    ('config:create', '创建配置', 'create', '新建分组或项',       3),
    ('config:update', '更新配置', 'update', '更新分组或项',       4),
    ('config:delete', '删除配置', 'delete', '删除分组或项',       5)
) AS r(code, name, action, description, sort)
ON CONFLICT (tenant_id, code) WHERE tenant_id <> 0 AND is_deleted = FALSE DO NOTHING;

-- ============================================
-- 11. 平台域 sys_* 种子
-- ============================================
-- 0023.3 终态：admin 账号对应 sys_users 里的 admin 平台身份 + super_admin 角色。
-- init_schema.sql 已建好 sys_users / sys_user_roles / sys_roles，
-- 这里 seed admin 的 sys_user 记录和 super_admin 角色绑定。

-- 11.1 平台超管角色
INSERT INTO sys_roles (id, code, name, description, data_scope, sort, status, created_by, updated_by)
OVERRIDING SYSTEM VALUE
VALUES (1, 'super_admin', '超级管理员', '平台级超级管理员', 1, 1, 1, 0, 0)
ON CONFLICT (code) WHERE is_deleted = FALSE DO NOTHING;

-- 11.2 admin 账号的 sys_user 平台身份
INSERT INTO sys_users (id, account_id, code, real_name, status, created_by, updated_by)
OVERRIDING SYSTEM VALUE
SELECT 1, a.id, 'admin', '系统管理员', 1, 0, 0
FROM accounts a
WHERE a.phone = '13800138000' AND a.is_deleted = FALSE
ON CONFLICT (id) DO NOTHING;

-- 11.3 admin 绑定 super_admin 平台角色（替代旧 account_roles 路径）
INSERT INTO sys_user_roles (user_id, role_id)
VALUES (1, 1)
ON CONFLICT (user_id, role_id) WHERE is_deleted = FALSE DO NOTHING;

-- 11.3b 平台域通配权限码
-- 0024+ 终态：super_admin 不再走中间件硬编码短路，改由 `*:*` 通配权限授权。
-- HasPermission(perms, res, act) 在 framework/pkg/permission/types.go 已支持 `*:*` 全局通配。
-- 注意：不显式指定 id，让 sys_permissions.id 的 GENERATED IDENTITY 自然分配，
-- 避免与后续 11.5 段隐式 INSERT 撞主键（id 列默认从 1 开始）。
INSERT INTO sys_permissions (code, name, action, description, sort, status, created_by, updated_by)
VALUES ('*:*', '超级通配权限', '*', '平台域所有资源所有操作（替代旧 super_admin 中间件短路）', 1, 1, 0, 0)
ON CONFLICT (code) WHERE is_deleted = FALSE DO NOTHING;

-- 11.3b2 平台域 tenant 资源权限码（已迁移到 11.5 段）
-- 历史：曾在这里直接 INSERT tenant:list/get/create/update/delete 五行。
-- 0024+ 调整为「所有平台域资源权限码统一放 11.5，并挂对应 menu_id」，
-- 故本段不重复插入。已 dropdb 重建过的 DB 可重跑 init_seed.sql，11.5 段会补齐 menu_id。

-- 11.3c super_admin 绑定通配权限
INSERT INTO sys_role_permissions (role_id, permission_id)
SELECT 1, p.id FROM sys_permissions p
WHERE p.code = '*:*' AND p.is_deleted = FALSE
ON CONFLICT (role_id, permission_id) WHERE is_deleted = FALSE DO NOTHING;

-- 11.4 平台域菜单树（sys_menus）
-- 这些菜单是平台管理分组的入口，与前端 App.tsx 中 /platform/* 路由对齐。
-- 注意：sys_menus 无 tenant_id 字段（平台域单租户概念）。
-- 哪些 platform 角色能看到哪些菜单，由 sys_role_menus 显式分配（11.4b）。
--
-- 顶级菜单：
--   id=100 平台管理（容器）        sort=999
--   id=101 租户管理                sort=1   （从原平台管理下提升为一级）
--   id=106 平台用户                sort=2   （从原平台管理下提升为一级）
-- 其余 id=102/103/104/105/107/108 仍挂在 id=100 之下。

-- 顶级：平台管理（id=100）
INSERT INTO sys_menus (id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
    OVERRIDING SYSTEM VALUE
VALUES (100, 'platform-admin', '平台管理', '平台域管理入口', '', '/platform', 'ShieldIcon', 999, 0, '0', TRUE, TRUE)
ON CONFLICT (code) WHERE is_deleted = FALSE DO NOTHING;

-- 顶级：租户管理（id=101）— 已从平台管理下提升为一级菜单
INSERT INTO sys_menus (id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
    OVERRIDING SYSTEM VALUE
VALUES (101, 'platform-tenants', '租户管理', '跨租户平台管理', '', '/platform/tenants', 'Building2Icon', 1, 0, '0', TRUE, TRUE)
ON CONFLICT (code) WHERE is_deleted = FALSE DO NOTHING;

-- 顶级：平台用户（id=106）— 已从平台管理下提升为一级菜单
INSERT INTO sys_menus (id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
    OVERRIDING SYSTEM VALUE
VALUES (106, 'platform-users', '平台用户', 'sys_users CRUD + 分配角色', '', '/platform/users', 'UsersIcon', 2, 0, '0', TRUE, TRUE)
ON CONFLICT (code) WHERE is_deleted = FALSE DO NOTHING;

-- 平台管理的子菜单（id=102/103/104/105/107/108）
INSERT INTO sys_menus (id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
    OVERRIDING SYSTEM VALUE
VALUES (102, 'platform-menus',       '平台菜单', 'sys_menus CRUD',                   '', '/platform/menus',       'MenuIcon',       2, 100, '100', TRUE, TRUE),
       (103, 'platform-configs',     '平台配置', 'config_categories / config_items 维护', '', '/platform/configs', 'SettingsIcon',   3, 100, '100', TRUE, TRUE),
       (104, 'platform-dicts',       '平台字典', 'dicts / dict_items 维护',          '', '/platform/dicts',       'BookIcon',       4, 100, '100', TRUE, TRUE),
       (105, 'platform-cache',       '缓存管理', 'Redis cache 运维 (Cache.tsx)',     '', '/platform/cache',       'DatabaseIcon',   5, 100, '100', TRUE, TRUE),
       (107, 'platform-roles',       '平台角色', 'sys_roles CRUD + 菜单/权限绑定',     '', '/platform/roles',       'ShieldIcon',     7, 100, '100', TRUE, TRUE),
       (108, 'platform-permissions', '平台权限', 'sys_permissions CRUD',             '', '/platform/permissions', 'KeyIcon',        8, 100, '100', TRUE, TRUE)
ON CONFLICT (code) WHERE is_deleted = FALSE DO NOTHING;

-- 11.4b super_admin 绑定所有平台菜单
-- 替代旧 sys_menu.Tree service 里的 isSuperAdmin 分支：super_admin 看全菜单改为
-- "绑定所有菜单"这一数据事实。
INSERT INTO sys_role_menus (role_id, menu_id)
SELECT 1, m.id FROM sys_menus m
WHERE m.is_deleted = FALSE
ON CONFLICT (role_id, menu_id) WHERE is_deleted = FALSE DO NOTHING;

-- 11.5 平台域资源权限码（5 组 = 25 行）
-- apps/platform/*/routes.go 用 ResXxx 常量拼出 "resource:action"，中间件查 sys_permissions.code。
-- 挂 menu_id 用于 UI 展示"哪个菜单下有哪些资源"。
-- 必须在 11.4 段 sys_menus INSERT 之后执行——子查询依赖 platform-xxx menu 的 id。
--
-- ON CONFLICT DO UPDATE：首跑等价 DO NOTHING；已 dropdb 重建过、又跑过旧版 seed
-- 留下 5 行无 menu_id 的记录，重跑时会补齐 menu_id。
--
-- super_admin 靠 11.3c 绑的 *:* 通配自动拥有；其他 platform 角色（如 platform_admin）
-- 需通过 UI 显式绑定具体 code。
INSERT INTO sys_permissions (code, name, action, description, sort, status, created_by, updated_by, menu_id)
SELECT s.code, s.name, s.action, s.description, s.sort, 1, 0, 0,
       (SELECT id FROM sys_menus WHERE code = 'platform-tenants' AND is_deleted = FALSE)
FROM (VALUES
    ('tenant:list',   '查询租户', 'list',   '查询所有租户列表',  1),
    ('tenant:get',    '查看租户', 'get',    '查看单个租户详情',  2),
    ('tenant:create', '创建租户', 'create', '新建租户并触发首装', 3),
    ('tenant:update', '更新租户', 'update', '更新租户档案信息',  4),
    ('tenant:delete', '删除租户', 'delete', '软删/硬删租户',     5)
) AS s(code, name, action, description, sort)
ON CONFLICT (code) WHERE is_deleted = FALSE
DO UPDATE SET menu_id = EXCLUDED.menu_id;

INSERT INTO sys_permissions (code, name, action, description, sort, status, created_by, updated_by, menu_id)
SELECT s.code, s.name, s.action, s.description, s.sort, 1, 0, 0,
       (SELECT id FROM sys_menus WHERE code = 'platform-permissions' AND is_deleted = FALSE)
FROM (VALUES
    ('permission:list',   '查询权限', 'list',   '查询平台权限码', 1),
    ('permission:get',    '查看权限', 'get',    '查看平台权限码', 2),
    ('permission:create', '创建权限', 'create', '新建平台权限码', 3),
    ('permission:update', '更新权限', 'update', '更新平台权限码', 4),
    ('permission:delete', '删除权限', 'delete', '软删平台权限码', 5)
) AS s(code, name, action, description, sort)
ON CONFLICT (code) WHERE is_deleted = FALSE
DO UPDATE SET menu_id = EXCLUDED.menu_id;

INSERT INTO sys_permissions (code, name, action, description, sort, status, created_by, updated_by, menu_id)
SELECT s.code, s.name, s.action, s.description, s.sort, 1, 0, 0,
       (SELECT id FROM sys_menus WHERE code = 'platform-roles' AND is_deleted = FALSE)
FROM (VALUES
    ('role:list',   '查询角色', 'list',   '查询平台角色', 1),
    ('role:get',    '查看角色', 'get',    '查看平台角色', 2),
    ('role:create', '创建角色', 'create', '新建平台角色', 3),
    ('role:update', '更新角色', 'update', '更新平台角色', 4),
    ('role:delete', '删除角色', 'delete', '软删平台角色', 5)
) AS s(code, name, action, description, sort)
ON CONFLICT (code) WHERE is_deleted = FALSE
DO UPDATE SET menu_id = EXCLUDED.menu_id;

INSERT INTO sys_permissions (code, name, action, description, sort, status, created_by, updated_by, menu_id)
SELECT s.code, s.name, s.action, s.description, s.sort, 1, 0, 0,
       (SELECT id FROM sys_menus WHERE code = 'platform-users' AND is_deleted = FALSE)
FROM (VALUES
    ('user:list',   '查询用户', 'list',   '查询平台用户', 1),
    ('user:get',    '查看用户', 'get',    '查看平台用户', 2),
    ('user:create', '创建用户', 'create', '新建平台用户', 3),
    ('user:update', '更新用户', 'update', '更新平台用户', 4),
    ('user:delete', '删除用户', 'delete', '软删平台用户', 5)
) AS s(code, name, action, description, sort)
ON CONFLICT (code) WHERE is_deleted = FALSE
DO UPDATE SET menu_id = EXCLUDED.menu_id;

INSERT INTO sys_permissions (code, name, action, description, sort, status, created_by, updated_by, menu_id)
SELECT s.code, s.name, s.action, s.description, s.sort, 1, 0, 0,
       (SELECT id FROM sys_menus WHERE code = 'platform-menus' AND is_deleted = FALSE)
FROM (VALUES
    ('menu:list',   '查询菜单', 'list',   '查询平台菜单', 1),
    ('menu:get',    '查看菜单', 'get',    '查看平台菜单', 2),
    ('menu:create', '创建菜单', 'create', '新建平台菜单', 3),
    ('menu:update', '更新菜单', 'update', '更新平台菜单', 4),
    ('menu:delete', '删除菜单', 'delete', '软删平台菜单', 5)
) AS s(code, name, action, description, sort)
ON CONFLICT (code) WHERE is_deleted = FALSE
DO UPDATE SET menu_id = EXCLUDED.menu_id;

-- ============================================
-- 12. 序列号兜底（防止 first_install 复制时 id 冲突）
-- ============================================
SELECT setval('config_categories_id_seq', GREATEST(
    (SELECT COALESCE(MAX(id), 0) FROM config_categories),
    (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE) * 1000
), true);

SELECT setval('config_items_id_seq', GREATEST(
    (SELECT COALESCE(MAX(id), 0) FROM config_items),
    (SELECT id FROM tenants WHERE code = 'bootstrap' AND is_deleted = FALSE) * 1000
), true);

SELECT setval('sys_users_id_seq', GREATEST((SELECT COALESCE(MAX(id), 0) FROM sys_users), 1), true);
SELECT setval('sys_roles_id_seq', GREATEST((SELECT COALESCE(MAX(id), 0) FROM sys_roles), 1), true);
SELECT setval('sys_menus_id_seq', GREATEST((SELECT COALESCE(MAX(id), 0) FROM sys_menus), 1000), true);
