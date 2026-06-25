-- ============================================
-- 登录安全策略：锁定 / 尝试 / 历史 / IP 审计
-- 0024.x 配套迁移文件
--
-- 三张表的设计目的：
--   login_attempts  滑动窗口内失败计数 → 触发锁定
--   account_locks   当前生效的账号锁定记录（TTL 到了自动作废）
--   login_history   每次成功登录的 IP/UA/位置  → 异地告警基础数据
--
-- 三张表都不带 RLS：登录路径必须在跨域事务里跑（用户登录时 tenant 上下文尚未建立）。
-- ============================================

-- ============================================
-- 1. login_attempts 登录尝试记录
-- ============================================
CREATE TABLE IF NOT EXISTS login_attempts
(
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    account        VARCHAR(255) NOT NULL,                -- 登录尝试的账号（username/phone/email 原值）
    ip             VARCHAR(64)  NOT NULL,
    user_agent     VARCHAR(512),
    success        BOOLEAN      NOT NULL,
    failure_reason VARCHAR(32),                          -- 'invalid_password' / 'account_not_found' / 'user_disabled' / 'locked'
    scope          VARCHAR(16),                          -- 'tenant' / 'platform' / 'precheck'
    tenant_id      BIGINT,                                -- 登录尝试的 tenant（precheck 时为 NULL）
    created_at     TIMESTAMPTZ DEFAULT NOW()
);

-- 按账号 + 时间倒序索引：用于"最近 N 次失败"窗口查询
CREATE INDEX IF NOT EXISTS idx_login_attempts_account_time
    ON login_attempts (account, created_at DESC);

-- 按 IP + 时间倒序索引：用于按 IP 维度封禁（防止同一 IP 跨账号爆破）
CREATE INDEX IF NOT EXISTS idx_login_attempts_ip_time
    ON login_attempts (ip, created_at DESC);

COMMENT ON TABLE login_attempts IS '登录尝试流水（含成功与失败），用于锁定判定与安全审计';
COMMENT ON COLUMN login_attempts.failure_reason IS '失败原因枚举：invalid_password / account_not_found / user_disabled / locked';

-- ============================================
-- 2. account_locks 账号锁定状态
-- ============================================
CREATE TABLE IF NOT EXISTS account_locks
(
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    account       VARCHAR(255) NOT NULL UNIQUE,          -- 一锁一行，账号维度唯一
    locked_until  TIMESTAMPTZ NOT NULL,                 -- 超过此时间视为自动解锁
    reason        VARCHAR(32) NOT NULL,                 -- 'too_many_failures' / 'manual' / 'security_alert'
    attempts      INT NOT NULL DEFAULT 0,               -- 触发锁定时的累计失败次数
    ip            VARCHAR(64),                          -- 触发锁定时的 IP（取证用）
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

-- 按锁定到期时间索引：用于定期清理过期记录 + 扫描即将到期
CREATE INDEX IF NOT EXISTS idx_account_locks_until
    ON account_locks (locked_until);

COMMENT ON TABLE account_locks IS '账号锁定记录，locked_until 到期自动作废';
COMMENT ON COLUMN account_locks.reason IS '锁定原因：too_many_failures / manual / security_alert';

-- ============================================
-- 3. login_history 登录成功历史（异地告警基础）
-- ============================================
CREATE TABLE IF NOT EXISTS login_history
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    account_id  BIGINT NOT NULL,                          -- accounts.id
    user_id     BIGINT,                                  -- tenant_users.id 或 sys_users.id（platform 登录时为 sys_users.id）
    tenant_id   BIGINT,                                  -- 登录所属 tenant（platform 登录时为 0）
    scope       VARCHAR(16) NOT NULL,                    -- 'tenant' / 'platform'
    ip          VARCHAR(64) NOT NULL,
    user_agent  VARCHAR(512),
    device_id   VARCHAR(128),                            -- 前端设备指纹（可选）
    location    VARCHAR(128),                            -- IP 反查的省市（GeoIP，留空由 Notifier 异步解析）
    session_id  VARCHAR(64),                            -- 关联 auth_sessions.id
    login_at    TIMESTAMPTZ DEFAULT NOW()
);

-- 按账号 + 时间倒序：取"最近 N 次成功登录"对比 IP/UA
CREATE INDEX IF NOT EXISTS idx_login_history_account_time
    ON login_history (account_id, login_at DESC);

-- 按 IP 索引：按 IP 维度查"哪些账号从这个 IP 登录过"
CREATE INDEX IF NOT EXISTS idx_login_history_ip
    ON login_history (ip);

COMMENT ON TABLE login_history IS '登录成功历史，用于异地登录检测与安全审计';
COMMENT ON COLUMN login_history.device_id IS '前端设备指纹（可选），用于识别同设备切换';
COMMENT ON COLUMN login_history.location IS 'IP 反查的省市，由 GeoIP 服务异步填充';