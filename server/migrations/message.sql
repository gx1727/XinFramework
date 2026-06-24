-- ============================================
-- message.sql : 站内信（tenant_messages）
-- ============================================
-- 目的：
--   站内信是租户域基础能力。表放在 tenant_* 命名空间，启用 RLS，
--   平台广播走 tenant_id = 0 短路口子（参考 dicts / config_items）。
--
-- 数据域：tenant（带 RLS）
-- 错误码段：16001-16999
-- 资源码：message
--
-- 接入：
--   1. 在 cmd/xin/main.go 显式 import 并加进 []plugin.Module
--   2. 在 framework/pkg/config/config.go optOutModules 加 "message"
--   3. 在 framework/pkg/permission/constants.go 加 ResMessage = "message"
--   4. 在 framework/pkg/resp/errors.go 加 CodeMessage = 16000
--   5. 在 migrations/init_seed.sql 加 message:* 资源码 seed（bootstrap 租户 tenant_id=1）
--      新租户会通过 apps/platform/tenants/first_install.go 自动复制过去
-- ============================================

SET client_encoding = 'UTF8';

-- ============================================
-- 1. tenant_messages 站内信主表
-- ============================================
CREATE TABLE IF NOT EXISTS tenant_messages
(
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id     BIGINT       NOT NULL,                   -- 0 = 平台广播
    sender_id     BIGINT       NOT NULL DEFAULT 0,         -- 0 = 系统消息
    recipient_id  BIGINT       NOT NULL,                   -- 收件人 tenant_users.id
    subject       VARCHAR(255) NOT NULL,
    body          TEXT         NOT NULL DEFAULT '',
    msg_type      SMALLINT     NOT NULL DEFAULT 1,         -- 1=私信 2=通知 3=系统公告
    priority      SMALLINT     NOT NULL DEFAULT 0,         -- 0=普通 1=重要 2=紧急
    is_read       BOOLEAN      NOT NULL DEFAULT FALSE,
    read_at       TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    is_deleted    BOOLEAN      NOT NULL DEFAULT FALSE
);

COMMENT ON TABLE  tenant_messages             IS '站内信主表（租户域，tenant_id=0 表示平台广播）';
COMMENT ON COLUMN tenant_messages.tenant_id    IS '租户 ID；0 表示平台广播，所有租户可见';
COMMENT ON COLUMN tenant_messages.sender_id    IS '发件人 tenant_users.id；0 表示系统消息';
COMMENT ON COLUMN tenant_messages.recipient_id IS '收件人 tenant_users.id';
COMMENT ON COLUMN tenant_messages.msg_type     IS '1=私信 2=通知 3=系统公告';
COMMENT ON COLUMN tenant_messages.priority     IS '0=普通 1=重要 2=紧急';

-- ============================================
-- 2. 索引
-- ============================================
-- 收件箱按 recipient + is_read 过滤（最热路径：未读列表）
CREATE INDEX IF NOT EXISTS idx_tenant_messages_recipient_unread
    ON tenant_messages (recipient_id, is_read)
    WHERE is_deleted = FALSE;

-- 发件箱按 sender 过滤
CREATE INDEX IF NOT EXISTS idx_tenant_messages_sender
    ON tenant_messages (sender_id)
    WHERE is_deleted = FALSE;

-- 租户内按时间倒序分页
CREATE INDEX IF NOT EXISTS idx_tenant_messages_tenant_created
    ON tenant_messages (tenant_id, created_at DESC)
    WHERE is_deleted = FALSE;

-- ============================================
-- 3. RLS 策略（短路口子：tenant_id=0 表示平台广播，所有租户可见）
-- ============================================
ALTER TABLE tenant_messages ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_messages FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_messages_isolation_policy ON tenant_messages;
CREATE POLICY tenant_messages_isolation_policy ON tenant_messages
    USING (
        tenant_id = 0
        OR tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::BIGINT
        OR NULLIF(current_setting('app.bypass_rls', true), 'off') = 'on'
    );

-- ============================================
-- 4. 资源码 seed 不在本文件，参考 cms.sql / flag.sql 约定
-- ============================================
-- tenant_permissions 表的 unique key 是 (tenant_id, code)，且 code 字段约定
-- 为 `<resource>:<action>` 合并格式（如 'message:list'），permission.HasPermission
-- 也是按 'message:list' 这种 key 查找的。
--
-- 正确做法：在 migrations/init_seed.sql 末尾追加（tenant_id=1 是 bootstrap 租户）：
--
--   INSERT INTO tenant_permissions (tenant_id, code, action, name, description, sort, status)
--   VALUES
--       (1, 'message:list',   'list',   '查看站内信', '查看收件箱/发件箱', 1, 1),
--       (1, 'message:create', 'create', '发送站内信', '发送站内信给指定用户', 2, 1),
--       (1, 'message:update', 'update', '更新站内信', '标记已读 / 改自己发出去的信', 3, 1),
--       (1, 'message:delete', 'delete', '删除站内信', '软删除收件或发件', 4, 1);
--
-- 新租户创建时会通过 apps/platform/tenants/first_install.go 的 INSERT ... SELECT
-- 从 bootstrap 租户复制所有 tenant_permissions 行，无需在每个租户里手插。