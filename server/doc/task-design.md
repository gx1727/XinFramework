# 长时任务系统设计

> 本文描述 XinFramework 的长时任务（后台异步任务）系统：任务队列、worker 池、
> 重试 / DLQ、僵死回收、管理 API。
> 配套迁移：[migrations/background_tasks.sql](../../migrations/background_tasks.sql)
> 实现位置：[framework/pkg/task/](../../framework/pkg/task/)、[apps/task/](../../apps/task/)

---

## 1. 设计目标

未引入本系统前，框架所有写入操作都是同步的：

| 模块 | 同步点 | 问题 |
|---|---|---|
| [audit.Log](../../framework/pkg/audit/audit.go#L34) | INSERT db_logs | OK（事务内） |
| [message.Send](../../apps/tenant/message/service.go#L104) | INSERT messages | OK |
| login_security notify | 仅写日志 | 不发真实通知 |
| flag/generate | 合成头像框 | 长事务阻塞接口 |

引入长时任务系统后：

- 异步：业务路径入队即返回，第三方 API 失败不阻塞登录
- 可靠：DB 持久化任务，worker 崩溃后另一实例可接管
- 可观测：状态机 / 重试次数 / 最后错误 / 心跳全量记录
- 可水平扩展：`SELECT FOR UPDATE SKIP LOCKED` 支持多 worker 并发
- 可中断：长任务带 ctx + heartbeat；僵死任务自动回收

---

## 2. 核心组件

```
framework/pkg/task/
├── types.go          Task / Status / 枚举 / EnqueueOption
├── handler.go        Handler interface + Registry + HandlerFunc 适配器
├── retry.go          BackoffConfig（指数 / 线性 / 固定退避）
├── queue.go          Queue interface + PGQueue 实现 + SKIP LOCKED 抢占
└── worker.go         Worker pool：轮询 / 心跳 / 僵死回收 / 优雅停机

apps/task/
├── module.go         Module 工厂：Init 注册 Queue；Register 启动 Worker；内置 cleanup handler
├── service.go        管理 API 的 service 层（list / get / cancel / requeue / stats / cleanup）
├── handler.go        sys 域管理 API（仅 super_admin 可访问）
├── routes.go         路由注册（sys 域）
├── types.go          DTO + 别名（让业务代码读起来更短）
└── errors.go         错误码段（1700-1799）
```

---

## 3. 数据模型

```sql
CREATE TABLE background_tasks (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    kind          VARCHAR(64)  NOT NULL,                  -- 任务类型
    payload       JSONB        NOT NULL DEFAULT '{}'::jsonb,
    status        VARCHAR(16)  NOT NULL DEFAULT 'pending',-- pending/running/succeeded/failed/cancelled/dead
    priority      INT          NOT NULL DEFAULT 0,        -- 越大越优先
    run_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),    -- 计划执行时间
    attempts      INT          NOT NULL DEFAULT 0,
    max_attempts  INT          NOT NULL DEFAULT 3,
    last_error    TEXT,
    worker_id     VARCHAR(64),
    started_at    TIMESTAMPTZ,
    heartbeat_at  TIMESTAMPTZ,                           -- > heartbeat_timeout_sec 视为僵死
    finished_at   TIMESTAMPTZ,
    timeout_sec   INT          NOT NULL DEFAULT 300,      -- 单次执行超时
    tenant_id     BIGINT,                                -- 业务域隔离（可选）
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
```

**关键索引**（部分索引，避免状态切换后索引膨胀）：

```sql
-- worker 轮询热路径
CREATE INDEX idx_bg_tasks_pending
    ON background_tasks (priority DESC, run_at ASC)
    WHERE status = 'pending';

-- 僵死回收热路径
CREATE INDEX idx_bg_tasks_running_heartbeat
    ON background_tasks (heartbeat_at)
    WHERE status = 'running';
```

---

## 4. 状态机

```
                  失败（attempts < max_attempts）
        ┌─────────────────────────────────┐
        ▼                                 │
   [pending] ──Claim──► [running] ──成功──► [succeeded]
       │                  │
       │                  ├──心跳超时──回收──► [pending]
       │                  │
       │                  └──失败──┬─► [pending]（重试）
       │                            └─► [dead]（attempts 超限）
       │
       ├─手动取消──► [cancelled]
       └─执行成功──► [succeeded]
```

`Status.IsTerminal()` 判定 succeeded/failed/cancelled/dead 为终态。

---

## 5. 核心接口

### 5.1 Handler 业务侧实现

```go
type Handler interface {
    Kind() string
    Handle(ctx context.Context, t *Task) error
    Timeout() int
}

// 注册：apps/task/module.go 启动期
taskpkg.RegisterHandler(HandlerFunc{
    KindStr:  "send_notification",
    TimeoutV: 60,
    HandleFn: func(ctx context.Context, t *Task) error {
        var p login_security.NotificationPayload
        json.Unmarshal(t.Payload, &p)
        return sendSMS(ctx, p.Recipient, p.Body)
    },
})
```

### 5.2 Queue 入队 + 抢占

```go
type Queue interface {
    Enqueue(ctx, kind, payload, opts...) (uint64, error)
    Claim(ctx, workerID, kinds) (*Task, error)  // SELECT FOR UPDATE SKIP LOCKED
    Complete(ctx, taskID) error
    Fail(ctx, taskID, errMsg, backoff) (willRetry bool, err error)
    Heartbeat(ctx, taskID) error
    Cancel(ctx, taskID) error
    Requeue(ctx, taskID) error
    ReclaimDead(ctx, olderThan) (int, error)
    Get(ctx, taskID) (*Task, error)
    List(ctx, filter) ([]*Task, int, error)
    Stats(ctx) (Stats, error)
    Cleanup(ctx, cutoff, statuses) (int, error)
}
```

**Claim 关键 SQL**（CTE + FOR UPDATE SKIP LOCKED）：

```sql
WITH cte AS (
    SELECT id FROM background_tasks
    WHERE status = 'pending' AND run_at <= NOW()
      AND ($1::text[] IS NULL OR kind = ANY($1))
    ORDER BY priority DESC, run_at ASC
    LIMIT 1 FOR UPDATE SKIP LOCKED
)
UPDATE background_tasks t
SET status = 'running', worker_id = $2, started_at = NOW(),
    heartbeat_at = NOW(), attempts = attempts + 1
FROM cte WHERE t.id = cte.id
RETURNING t.*;
```

### 5.3 Worker

```go
type WorkerConfig struct {
    WorkerID          string
    WorkerCount       int
    PollInterval      time.Duration
    HeartbeatInterval time.Duration
    ReclaimInterval   time.Duration
    ReclaimOlderThan  time.Duration
    Backoff           BackoffConfig
    WatchedKinds      []string
    Queue             Queue
    Registry          Registry
}

worker, _ := NewWorker(cfg)
worker.Start(ctx)         // 启动 N 个 worker + 1 个 reclaim goroutine
defer worker.Stop(30*time.Second)  // 优雅停机
```

Worker 行为保证：
- handler panic → recover 转 error，不杀 worker goroutine
- worker 停机 → 当前任务跑完才退出（grace period）
- 心跳超时 → reclaimLoop 自动回收回 pending

---

## 6. 与 plugin.AppContext 集成

```go
// framework/pkg/plugin/appcontext.go
type Reader interface {
    // ... 现有 8 个 slot ...
    TaskQueue() task.Queue  // 新增
}

type Writer interface {
    // ... 现有 8 个 setter ...
    SetTaskQueue(task.Queue)  // 新增
}
```

业务模块在 `Init` 阶段注入：

```go
func (m *Module) Init(_ plugin.Reader, w plugin.Writer) error {
    w.SetTaskQueue(taskpkg.NewPGQueue(m.DB))
    return nil
}
```

其他模块通过 `ctx.TaskQueue()` 拿：

```go
func (s *SomeService) DoAsyncWork(ctx context.Context) error {
    id, err := ctx.TaskQueue().Enqueue(ctx, "send_email",
        task.MarshalPayload(myPayload),
        task.WithPriority(10),
    )
    return err
}
```

---

## 7. 内置 Cleanup Handler

apps/task 模块启动期注册两个内置 handler：

| Kind | 行为 | 触发方式 |
|---|---|---|
| `cleanup_expired_locks` | `DELETE FROM account_locks WHERE locked_until <= NOW()` | 外部 cron / 业务侧入队 |
| `cleanup_old_tasks` | 删除 background_tasks 中 finished_at < NOW() - 7d 的 succeeded/dead | 外部 cron |

清理任务自身也是任务——所以可以"递归清理"：
```bash
# 每天凌晨 3 点清理过期锁定 + 历史任务
0 3 * * * curl -X POST http://localhost:8087/api/v1/sys/tasks/cleanup?keep_days=7
```

---

## 8. 管理 API

| 方法 | 路径 | 鉴权 |
|---|---|---|
| `GET /api/v1/sys/tasks?kind=&status=&tenant_id=&page=&size=` | 列任务 | super_admin |
| `GET /api/v1/sys/tasks/:id` | 任务详情 | super_admin |
| `POST /api/v1/sys/tasks/:id/cancel` | 取消（pending/running） | super_admin |
| `POST /api/v1/sys/tasks/:id/requeue` | 重新入队（failed/dead） | super_admin |
| `GET /api/v1/sys/tasks/stats` | 队列统计 | super_admin |
| `POST /api/v1/sys/tasks/cleanup?keep_days=7&statuses=succeeded,dead` | 清理历史 | super_admin |

---

## 9. 配置项

```yaml
task:
  worker_count: 4               # 进程内 worker goroutine 数（建议 = CPU/2）
  poll_interval_ms: 1000        # 轮询间隔
  heartbeat_interval_sec: 30    # 心跳上报周期
  heartbeat_timeout_sec: 90     # 心跳超时视为僵死，自动回收
  reclaim_interval_sec: 60      # 僵死回收周期
  default_max_attempts: 3       # 默认重试次数
  default_timeout_sec: 300      # 默认单任务超时
  retry_strategy: "exponential" # exponential / linear / fixed
  cleanup:
    succeeded_keep_days: 7
    failed_keep_days: 30
    dead_keep_days: 90
```

环境变量：`XIN_TASK_*`（每条配置都有对应环境变量）。

---

## 10. 与 login_security 的串联

[login_security.QueueNotifier](../../framework/pkg/login_security/notify.go) 把"发短信/邮件"改造为异步入队：

```go
// login_security 内部
type QueueNotifier struct {
    queue task.Queue
    KindName string  // 默认 "send_notification"
}

// Notify 把 payload 序列化后入队
func (q *QueueNotifier) Notify(ctx context.Context, p NotificationPayload) error {
    payload := task.MarshalPayload(p)
    _, err := q.queue.Enqueue(ctx, q.KindName, payload,
        task.WithPriority(10),
        task.WithMaxAttempts(5),
        task.WithTimeout(60),
    )
    return err
}

// apps/task/module.go 注册 handler
taskpkg.RegisterHandler(taskpkg.HandlerFunc{
    KindStr: "send_notification",
    HandleFn: func(ctx context.Context, t *taskpkg.Task) error {
        var p login_security.NotificationPayload
        json.Unmarshal(t.Payload, &p)
        return sendSMSOrEmail(ctx, p)
    },
})
```

业务侧零感知：login_security.SecurityService 仍按原来的 Notify 调用，
但 Notifier 实现已经是 QueueNotifier（而非 LogNotifier），
所以调用 Notify 就是入队，不阻塞业务路径。

---

## 11. 可观测指标

### 11.1 关键查询

```sql
-- 当前正在执行的任务
SELECT id, kind, worker_id, started_at, heartbeat_at
FROM background_tasks
WHERE status = 'running';

-- 僵死任务
SELECT id, kind, worker_id, heartbeat_at, NOW() - heartbeat_at AS stalled_for
FROM background_tasks
WHERE status = 'running'
  AND heartbeat_at < NOW() - INTERVAL '90 seconds';

-- 24h 失败率（按 kind 聚合）
SELECT kind, status, COUNT(*)
FROM background_tasks
WHERE created_at > NOW() - INTERVAL '1 day'
GROUP BY kind, status
ORDER BY kind, status;

-- DLQ 内容
SELECT id, kind, attempts, last_error, created_at
FROM background_tasks
WHERE status = 'dead'
ORDER BY created_at DESC
LIMIT 50;
```

### 11.2 推荐 Prometheus 指标（V2 引入）

| 指标 | 类型 | 含义 |
|---|---|---|
| `task_queue_pending{kind}` | Gauge | 队列积压 |
| `task_running_count{kind}` | Gauge | 正在执行 |
| `task_dlq_count{kind}` | Gauge | 进 DLQ |
| `task_duration_seconds{kind,result}` | Histogram | 执行耗时 |
| `task_attempts_total{kind,result}` | Counter | 失败次数 |

---

## 12. 演进路径

### V1（MVP，当前）
- 存储：PostgreSQL
- 任务量：< 10k/天
- 任务时长：< 1h

### V2（按需升级）
- 任务量 > 10k/天 → 引入 Redis Stream 或 asynq
- 需要精确调度 → 引入 robfig/cron + cron_jobs 表
- 长任务（> 1h）→ 拆分 + ack 心跳

### V3（任务编排）
- 任务依赖图：pending_for 字段
- 任务链：A 完成后触发 B
- 工作流引擎：DAG 定义

---

## 13. 风险与缓解

| 风险 | 缓解 |
|---|---|
| PG 表膨胀 | cleanup_old_tasks handler 周期清理 |
| worker 僵死后任务永远 running | ReclaimDead 自动回收 |
| handler panic 杀死 worker | worker.execute 包裹 defer recover |
| 任务占用事务过长 | handler 禁止长事务；如需可拆 saga |
| 入队失败影响业务 | QueueNotifier 入队失败仅 warn，业务路径不感知 |
| 误删除重要任务 | 管理 API 需 super_admin；统计 + 审计日志 |

**回滚方案**：停止 worker（`worker.Stop(30s)`），所有 `Enqueue` 调用降级为 fire-and-forget（写表但不消费），不影响主流程。

---

## 14. 测试

[framework/pkg/task/task_test.go](../../framework/pkg/task/task_test.go) 12 个测试覆盖：

- Status.IsTerminal 6 种状态
- BackoffConfig.NextDelay 4 种策略
- Registry 注册 / 重名 / Names
- HandlerFunc 适配器
- EnqueueOption 默认值
- MarshalPayload JSON 序列化

集成测试（需 PG）：
- Claim SKIP LOCKED 并发安全
- Heartbeat + ReclaimDead 回收
- Fail 重试与进 DLQ 边界
- Cancel / Requeue 状态转换

---

## 15. 落地清单（已完成）

| # | 文件 | 状态 |
|---|---|---|
| 1 | [migrations/background_tasks.sql](../../migrations/background_tasks.sql) | ✅ |
| 2 | [framework/pkg/task/types.go](../../framework/pkg/task/types.go) | ✅ |
| 3 | [framework/pkg/task/handler.go](../../framework/pkg/task/handler.go) | ✅ |
| 4 | [framework/pkg/task/retry.go](../../framework/pkg/task/retry.go) | ✅ |
| 5 | [framework/pkg/task/queue.go](../../framework/pkg/task/queue.go) | ✅ |
| 6 | [framework/pkg/task/worker.go](../../framework/pkg/task/worker.go) | ✅ |
| 7 | [framework/pkg/task/task_test.go](../../framework/pkg/task/task_test.go) | ✅ 12 tests |
| 8 | [framework/pkg/plugin/appcontext.go](../../framework/pkg/plugin/appcontext.go) TaskQueue slot | ✅ |
| 9 | [framework/pkg/config/config.go](../../framework/pkg/config/config.go) TaskConfig | ✅ |
| 10 | [apps/task/](../../apps/task/) 业务模块（6 文件） | ✅ |
| 11 | [cmd/xin/main.go](../../cmd/xin/main.go) 注册 task 模块 | ✅ |
| 12 | [framework/pkg/login_security/notify.go](../../framework/pkg/login_security/notify.go) QueueNotifier | ✅ |

**编译**：✅ 全量通过
**测试**：✅ 框架层全部通过（task 包 12 个 + 其他包无回归）