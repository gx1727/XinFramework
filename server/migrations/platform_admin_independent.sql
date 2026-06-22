-- ============================================
-- 平台账号独立迁移
-- ============================================
-- 目的：让平台账号（super_admin）独立于租户，不在 users 表里借用 bootstrap 租户。
--
-- 路径 B 终态约定：
--   - accounts            全局账号表（无 tenant_id）
--   - account_roles       账号级平台角色表（白名单 + CHECK 约束）
--   - users               租户内的用户身份表（一个 account 可在多个 tenant 各有一条）
--
-- 本次改动：
--   1. account_roles.role 加 CHECK 约束（白名单：当前仅 super_admin）
--   2. seed 默认 admin 账号的 super_admin 平台角色
--   3. users 表 unique 索引从 (account_id) 改为 (account_id, tenant_id)
--      —— 支持规则 4（一个账号在多个租户各有一条 users 记录）
--   4. 清掉 admin 账号借用 bootstrap 租户的 users 记录 + 关联的 user_roles
--
-- 兼容性：
--   - 现有登录流程已经是路径 B（PlatformLogin 不走 users 表），无 Go 代码改动
--   - 迁移机制 _schema_migrations 跳过已应用版本
--   - 本文件按字母序排在 framework.sql 之后、tenant.sql 之前
-- ============================================

-- ============================================
-- 1. account_roles.role 加 CHECK 约束
-- ============================================
-- 防止拼写错误（如 superadmin / SuperAdmin）和脏数据写入。
-- 当前已知唯一平台角色：super_admin（路径 B 终态）。
-- 未来新增角色时 ALTER 约束扩展白名单。

ALTER TABLE account_roles DROP CONSTRAINT IF EXISTS chk_account_role;
ALTER TABLE account_roles ADD CONSTRAINT chk_account_role
    CHECK (role IN ('super_admin'));

-- ============================================
-- 2. seed 默认 admin 账号的 super_admin 平台角色
-- ============================================
-- framework.sql L494-497 seed 了 accounts 表的 admin 账号（phone='13800138000'），
-- 但 account_roles 表是空的。这里补 seed，使平台登录能直接用 admin 账号。

INSERT INTO account_roles (account_id, role)
SELECT id, 'super_admin'
FROM accounts
WHERE phone = '13800138000'
ON CONFLICT (account_id, role) DO NOTHING;

-- ============================================
-- 3. users 表 unique 索引改复合
-- ============================================
-- 原索引 uk_users_account (account_id) 限制一账号一记录，阻塞规则 4。
-- 改为 uk_users_account_tenant (account_id, tenant_id)，允许一个账号在多个
-- 租户各有一条 users 记录（每个租户一个身份）。

DROP INDEX IF EXISTS uk_users_account;
CREATE UNIQUE INDEX IF NOT EXISTS uk_users_account_tenant
    ON users (account_id, tenant_id) WHERE is_deleted = FALSE;

-- ============================================
-- 4. 清掉 admin 账号借用 bootstrap 租户的 users 记录
-- ============================================
-- framework.sql L500-501 seed 时把 admin 账号放进了 tenant_id=1（bootstrap）
-- 的 users 表借用为"系统管理员"。路径 B 下不再需要这个借用。
--
-- 安全检查：如果 admin 账号有 > 1 条 users 记录，说明不是简单借用，是合法
-- 多租户身份，停止迁移避免误删。

DO $$
DECLARE
    cnt INT;
    admin_account_id BIGINT;
BEGIN
    SELECT id INTO admin_account_id FROM accounts WHERE phone = '13800138000';
    IF admin_account_id IS NULL THEN
        RAISE NOTICE 'admin 账号 (phone=13800138000) 不存在，跳过清理步骤';
        RETURN;
    END IF;
    SELECT COUNT(*) INTO cnt FROM users WHERE account_id = admin_account_id;
    IF cnt > 1 THEN
        RAISE EXCEPTION 'admin 账号有 % 条 users 记录，超出预期（1 条 = bootstrap 借用）。需人工确认后再清。', cnt;
    END IF;

    -- 先清 user_roles（避免外键悬挂）
    DELETE FROM user_roles
    WHERE user_id IN (
        SELECT id FROM users WHERE account_id = admin_account_id
    );

    -- 再清 users
    DELETE FROM users WHERE account_id = admin_account_id;

    RAISE NOTICE '已清理 admin 账号 (account_id=%) 的 users 借用记录', admin_account_id;
END $$;

-- ============================================
-- 5. 验证
-- ============================================
-- 迁移完成后预期状态：
--   - accounts: 至少 1 条（admin 账号）
--   - account_roles: 至少 1 条（admin → super_admin）
--   - users: 0 条 admin 借用的记录（可能还有其他用户的 records）
--   - 索引: uk_users_account_tenant 已建，uk_users_account 已删

DO $$
DECLARE
    admin_id BIGINT;
    ar_count INT;
    u_count INT;
    idx_exists BOOLEAN;
BEGIN
    SELECT id INTO admin_id FROM accounts WHERE phone = '13800138000';
    IF admin_id IS NULL THEN
        RAISE NOTICE 'admin 账号不存在，跳过验证';
        RETURN;
    END IF;

    SELECT COUNT(*) INTO ar_count FROM account_roles WHERE account_id = admin_id;
    SELECT COUNT(*) INTO u_count FROM users WHERE account_id = admin_id;
    SELECT EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE schemaname = 'public' AND indexname = 'uk_users_account_tenant'
    ) INTO idx_exists;

    RAISE NOTICE '验证结果：';
    RAISE NOTICE '  - admin 账号 account_id = %', admin_id;
    RAISE NOTICE '  - account_roles 数 = %（预期 >= 1）', ar_count;
    RAISE NOTICE '  - users 借用记录数 = %（预期 0）', u_count;
    RAISE NOTICE '  - 复合索引 uk_users_account_tenant 存在 = %', idx_exists;

    IF ar_count = 0 THEN
        RAISE EXCEPTION '验证失败：admin 账号没有 platform role 记录';
    END IF;
    IF u_count > 0 THEN
        RAISE EXCEPTION '验证失败：admin 账号仍有 users 借用记录';
    END IF;
    IF NOT idx_exists THEN
        RAISE EXCEPTION '验证失败：复合索引未创建';
    END IF;
END $$;