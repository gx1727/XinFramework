-- ============================================
-- 长时任务系统：background_tasks 表
-- 0024.x 配套迁移文件
--
-- 设计目标：
--   - 零新依赖：用 PostgreSQL 当任务存储（事务一致 + 已有 + 易调试）
--   - 多 worker 并发安全：SELECT ... FOR UPDATE SKIP LOCKED
--   - 完整的可观测：状态机 / 心跳 / 重试次数 / 最后错误
--
-- 为什么不放 RLS：worker 任务需要跨 tenant（清理过期锁、聚合统计等），
--   与具体业务 tenant_id 解耦。task 模块自身有 tenant_id 列做业务域隔离。
-- ============================================

CREATE TABLE IF NOT EXISTS background_tasks
(
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    kind          VARCHAR(64)  NOT NULL,                                       -- 任务类型: 'send_email' / 'cleanup_expired_locks'
    payload       JSONB        NOT NULL DEFAULT '{}'::jsonb,                  -- 任务参数（任意 JSON 可序列化结构）
    status        VARCHAR(16)  NOT NULL DEFAULT 'pending',                    -- pending/running/succeeded/failed/cancelled/dead
    priority      INT          NOT NULL DEFAULT 0,                            -- 数字越大越优先；同优先级按 run_at ASC
    run_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),                        -- 计划执行时间（支持延迟任务）
    attempts      INT          NOT NULL DEFAULT 0,                            -- 已执行次数（含失败）
    max_attempts  INT          NOT NULL DEFAULT 3,                            -- 最大重试次数；超限进 dead
    last_error    TEXT,                                                       -- 最后一次错误信息（仅失败时记录）
    worker_id     VARCHAR(64),                                               -- 抢占任务的 worker 标识（hostname:pid 或 UUID）
    started_at    TIMESTAMPTZ,                                               -- 任务被 claim 的时间
    heartbeat_at  TIMESTAMPTZ,                                               -- worker 心跳时间（> heartbeat_timeout_sec 视为僵死）
    finished_at   TIMESTAMPTZ,                                               -- 任务终态时间
    timeout_sec   INT          NOT NULL DEFAULT 300,                          -- 单次执行超时（秒），worker 进程内 context.WithTimeout
    tenant_id     BIGINT,                                                    -- 业务域隔离（可选；清理类任务为 NULL）
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ============================================
-- 索引
-- ============================================

-- worker 轮询索引：worker 在每个 poll 周期都查这条
-- 部分索引：只覆盖 pending 状态，避免 failed/dead 行膨胀索引体积
CREATE INDEX IF NOT EXISTS idx_bg_tasks_pending
    ON background_tasks (priority DESC, run_at ASC)
    WHERE status = 'pending';

-- 监控 / 排查索引：按 kind 聚合失败次数、最近任务时间
CREATE INDEX IF NOT EXISTS idx_bg_tasks_kind_status
    ON background_tasks (kind, status, created_at DESC);

-- 僵死回收索引：worker 定期扫描 running 状态中 heartbeat 过期者
CREATE INDEX IF NOT EXISTS idx_bg_tasks_running_heartbeat
    ON background_tasks (heartbeat_at)
    WHERE status = 'running';

-- 按租户聚合（用于管理 API 分页 + 按 tenant_id 过滤）
CREATE INDEX IF NOT EXISTS idx_bg_tasks_tenant_created
    ON background_tasks (tenant_id, created_at DESC)
    WHERE tenant_id IS NOT NULL;

-- ============================================
-- 注释
-- ============================================
COMMENT ON TABLE background_tasks IS '长时任务队列。worker 通过 SELECT FOR UPDATE SKIP LOCKED 抢占任务。';
COMMENT ON COLUMN background_tasks.kind IS '任务类型标识（注册到 handler registry 时唯一）';
COMMENT ON COLUMN background_tasks.payload IS '任务参数 JSON；handler 反序列化后使用';
COMMENT ON COLUMN background_tasks.status IS 'pending/running/succeeded/failed/cancelled/dead';
COMMENT ON COLUMN background_tasks.priority IS '数字越大越优先；同优先级按 run_at ASC';
COMMENT ON COLUMN background_tasks.attempts IS '已执行次数（含失败）；每次 Claim 自增 1';
COMMENT ON COLUMN background_tasks.max_attempts IS '最大重试次数；attempts >= max_attempts 进 dead';
COMMENT ON COLUMN background_tasks.worker_id IS '抢占该任务的 worker 标识（用于排查僵死任务归属）';
COMMENT ON COLUMN background_tasks.heartbeat_at IS 'worker 上报心跳；定时清理会回收 heartbeat_at < NOW() - heartbeat_timeout 的 running 任务';
COMMENT ON COLUMN background_tasks.timeout_sec IS '单次执行超时；worker 用 context.WithTimeout 包住 handler.Handle';

-- ============================================
-- 默认内置任务（注册到 handler registry）
-- ============================================
-- framework/pkg/task 启动时自动注册两个 cleanup handler：
--   1. cleanup_expired_locks  — 清理 login_security.account_locks 中 locked_until <= NOW() 的记录
--   2. cleanup_old_tasks      — 清理 background_tasks 中 finished_at < NOW() - cleanup_keep_days 的记录
-- 详见 framework/pkg/task/builtin_handlers.go

-- ============================================
-- 与现有 migrations 的兼容性
-- ============================================
-- 本表与现有 migrations/init_schema.sql 的 RLS 白名单冲突解决：
-- background_tasks 不在 RLS 白名单（详见 init_schema.sql §7）。
-- 该表启用全局访问（无 RLS），由 worker 进程和 admin 平台域负责读写。

-- ============================================
-- 索引命名约定
-- ============================================
-- idx_bg_tasks_<列>_<状态> 形式：清晰表达索引用途

COMMENT ON INDEX idx_bg_tasks_pending IS 'worker 轮询热路径索引（部分索引，只覆盖 pending）';
COMMENT ON INDEX idx_bg_tasks_running_heartbeat IS '僵死回收热路径索引（部分索引，只覆盖 running）';
COMMENT ON INDEX idx_bg_tasks_kind_status IS '管理 API / 监控按 kind + status 查询的索引';
COMMENT ON INDEX idx_bg_tasks_tenant_created IS '管理 API 按租户分页查询索引';