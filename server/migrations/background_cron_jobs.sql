-- ============================================
-- 周期性任务定义（cron jobs）
-- 0024.x 配套迁移文件
--
-- 设计目标：
--   - 让 scheduler 每分钟扫描 next_run_at 到期的任务
--   - 多实例并发安全（SELECT FOR UPDATE SKIP LOCKED）
--   - 错过窗口策略可配置（skip / run_immediately）
--   - 每个 cron_job 独立时区（业务跨时区场景）
--
-- 与 background_tasks 的关系：
--   cron job 定义"什么时候触发"，触发后入队到 background_tasks
--   由现有 worker 异步执行；本表不存任务历史，只存定义 + 调度状态
-- ============================================

CREATE TABLE IF NOT EXISTS background_cron_jobs (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        VARCHAR(64)  UNIQUE NOT NULL,                -- 唯一业务名（如 'cleanup_daily'）
    cron_expr   VARCHAR(128) NOT NULL,                       -- 标准 5/6 字段 cron 表达式
    timezone    VARCHAR(64)  NOT NULL DEFAULT 'UTC',         -- IANA 时区名
    kind        VARCHAR(64)  NOT NULL,                       -- 入队的 task kind
    payload     JSONB        NOT NULL DEFAULT '{}'::jsonb,   -- 入队时的 payload
    enabled     BOOLEAN      NOT NULL DEFAULT TRUE,

    -- 错过窗口策略
    --   skip         : 过期不补跑（下次 cron 再触发）
    --   run_immediately: scanner 扫到时立刻跑一次（仅 1 次）
    miss_policy VARCHAR(16)  NOT NULL DEFAULT 'skip',

    -- 触发统计
    last_run_at       TIMESTAMPTZ,
    next_run_at       TIMESTAMPTZ NOT NULL,                  -- 下次计划触发时间
    last_run_status   VARCHAR(16),                           -- 'success' / 'failed' / 'skipped' / 'missed'
    last_run_error    TEXT,
    total_runs        BIGINT     NOT NULL DEFAULT 0,
    total_failures    BIGINT     NOT NULL DEFAULT 0,

    -- 元数据
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 扫描热路径：scheduler 每分钟扫这条
CREATE INDEX IF NOT EXISTS idx_bg_cron_due
    ON background_cron_jobs (next_run_at)
    WHERE enabled = TRUE;

-- 按名字查找（管理 API / 删除）
CREATE INDEX IF NOT EXISTS idx_bg_cron_name
    ON background_cron_jobs (name);

COMMENT ON TABLE background_cron_jobs IS '周期性任务定义。scheduler 每分钟扫描 enabled=TRUE AND next_run_at<=NOW() 的记录，命中后入队到 background_tasks。';
COMMENT ON COLUMN background_cron_jobs.cron_expr IS '标准 cron 表达式：分 时 日 月 周（5 字段）或 秒 分 时 日 月 周（6 字段）';
COMMENT ON COLUMN background_cron_jobs.timezone IS 'IANA 时区名（Asia/Shanghai / UTC / America/New_York 等）；影响 next_run_at 计算';
COMMENT ON COLUMN background_cron_jobs.kind IS '触发时入队的 task kind；handler 注册时必须用同样的字符串';
COMMENT ON COLUMN background_cron_jobs.miss_policy IS '错过窗口的处理策略：skip（默认）/ run_immediately';
COMMENT ON COLUMN background_cron_jobs.last_run_status IS '上一次执行的状态：success / failed / skipped / missed';
COMMENT ON COLUMN background_cron_jobs.next_run_at IS '下次计划触发时间。scheduler 扫描时找 next_run_at <= NOW() AND enabled=TRUE 的记录';