# 登录安全策略

> 本文描述 XinFramework 的登录安全机制：账号锁定、登录尝试流水、
> 登录历史 / IP 审计、异地登录告警。
> 配套迁移：`migrations/login_security.sql`
> 实现位置：`framework/pkg/login_security/`

---

## 1. 为什么需要这套机制

未引入本机制前，登录路径仅有 [apps/boot/auth](file://d:\work\xin\XinFramework\server\apps\boot\auth) 的密码校验
（[password.go](file://d:\work\xin\XinFramework\server\apps\boot\auth\password.go) 的 `argon2id`），
不解决以下问题：

| 风险 | 现状 | 引入本机制后 |
|---|---|---|
| 密码爆破 | 无任何限制 | 滑动窗口失败计数 → 自动锁定 |
| 账号盗用 | 登录即用，无审计 | 登录历史全量记录，IP 审计 |
| 异地登录 | 无感知 | 检测 new_ip / new_device → 通知用户 |
| 跨账号爆破 | 单一账号锁定无法阻止 | 同 IP 失败计数（已实现计数，预留封禁位） |

---

## 2. 三大数据表

> 迁移脚本：[migrations/login_security.sql](file://d:\work\xin\XinFramework\server\migrations\login_security.sql)
> 三张表都不带 RLS——登录路径必须在跨域事务里跑。

### 2.1 `login_attempts` 登录尝试流水

每次登录（无论成败）写一行。`success=false` 时 `failure_reason` 记录原因枚举。

关键索引：
- `(account, created_at DESC)`：滑动窗口判定
- `(ip, created_at DESC)`：按 IP 维度封禁（未来扩展）

### 2.2 `account_locks` 当前生效的锁定

`account` 字段 UNIQUE，一锁一行。`locked_until` 到期后由定时任务清理。

### 2.3 `login_history` 登录成功历史

仅记录成功登录，用于异地判定与安全审计。

关键索引：
- `(account_id, login_at DESC)`：取最近 N 次对比
- `(ip)`：按 IP 维度查"哪些账号从这个 IP 登录过"

---

## 3. 包结构

```
framework/pkg/login_security/
├── types.go       // 类型 / 错误原因枚举
├── lock.go        // LockManager + PGLockManager
├── attempt.go     // AttemptStore + PGAttemptStore
├── history.go     // HistoryRecorder + PGHistoryRecorder
├── notify.go      // Notifier + LogNotifier + MultiNotifier
├── security.go    // SecurityService 主入口（编排上述组件）
└── login_security_test.go  // 14 个单元测试
```

| 类型 | 职责 |
|---|---|
| `LockManager` | 账号锁定的 Get / Lock / Unlock / CleanupExpired |
| `AttemptStore` | 登录尝试记录与滑动窗口计数 |
| `HistoryRecorder` | 成功登录历史 + 最近 N 次查询 |
| `Notifier` | 通知通道抽象（SMS / Email / InSite） |
| `RecipientResolver` | 根据 account_id 查联系方式 |
| `SecurityService` | 业务侧入口：`CheckLock` / `RecordFailure` / `RecordSuccess` |

---

## 4. 调用流程

### 4.1 登录入口（tenant-login / platform-login）

```
[Handler]                              [auth.Service]                        [SecurityService]
TenantLogin(c)
  ├─ svc.Login(ctx, req)
  │   ├─ checkAccountLock(ctx, account)  ──────────────────► CheckLock ──► LockManager.Get
  │   │                                     ◄────────────────── (*AccountLock | nil, nil)
  │   │                                     └─ 如果被锁 → return ErrAccountLocked
  │   ├─ ResolveLoginIdentity            (查账号 + 绑定的 tenant user)
  │   │   └─ 失败时 recordFailure(...)
  │   ├─ verifyPassword
  │   │   └─ 失败时 recordFailure(...)
  │   ├─ generateTokens                  (签 JWT + 写 session)
  │   └─ recordSuccess(ctx, ...)         ──────────────────► RecordSuccess ──► HistoryRecorder.Record
  │                                                                  └─ detectAnomaly ──► 若异地触发 Notifier
  └─ resp.Success(c, token)
```

### 4.2 失败计数 + 自动锁定

`SecurityService.RecordFailure` 行为：

```go
count, triggered, err := s.RecordFailure(ctx, account, ip, ua, reason, scope, tenantID)
// count = 当前窗口 [now - FailureWindow, now] 内的失败次数（含本次）
// triggered = true 表示本次失败触发了账号锁定
```

锁定触发条件：`count >= MaxFailedAttempts`（默认 5 次 / 10 分钟）。
锁定时长：`LockDuration`（默认 30 分钟）。

### 4.3 异地检测

`SecurityService.RecordSuccess` 内部调用 `detectAnomaly`：

```go
sig := detectAnomaly(now, recent, AnomalyDeviceMatch)
// sig.IsAnomaly = true 表示命中异地规则
// sig.Reasons = ["new_ip"] 或 ["new_ip", "new_device"]
// sig.KnownIPs = ["1.1.1.1", "1.1.1.2"]   用于通知里给用户对照
```

判定规则（保守优先）：

1. **没有任何历史** → 不是异地（新账号）
2. **历史上所有 IP 都相同 + 本次 IP 不同** → 异地（new_ip）
3. **AnomalyDeviceMatch=true + device_id 一致才算"未异地"** → 异地（new_device）

### 4.4 通知通道

`Notifier` 接口：

```go
type Notifier interface {
    Notify(ctx context.Context, payload NotificationPayload) error
}
```

默认实现 `LogNotifier`：仅写日志，便于未集成真实短信 / 邮件的过渡期。

业务模块可注入真实通道（如腾讯云 SMS / SendGrid）：

```go
smsNotifier := tencent.NewSMSNotifier(cfg)
emailNotifier := sendgrid.NewEmailNotifier(cfg)

multi := login_security.NewMultiNotifier(LogNotifier{}, smsNotifier, emailNotifier)
svc := login_security.NewSecurityService(cfg, ..., multi, nil)
```

---

## 5. 配置项

完整配置见 [config/config.go 的 LoginSecurityConfig](file://d:\work\xin\XinFramework\server\framework\pkg\config\config.go)：

```yaml
login_security:
  enabled: true              # 总开关
  max_failed_attempts: 5     # 滑动窗口内最大失败次数
  lock_duration_min: 30      # 锁定时长
  failure_window_min: 10     # 滑动窗口长度
  ip_failure_threshold: 20   # 同 IP 失败阈值
  ip_failure_window_min: 5   # IP 维度窗口

  # 异地告警
  anomaly_enabled: true
  anomaly_history_limit: 5
  anomaly_device_match: false
  anomaly_notify_in_site: true
  anomaly_notify_email: true
  anomaly_notify_sms: false

  # 锁定通知
  lock_notify_in_site: true
  lock_notify_email: true
  lock_notify_sms: true
```

环境变量覆盖：`XIN_LOGIN_SECURITY_*`（每条配置都有对应环境变量）。

---

## 6. 错误码

| Code | 错误 | 触发场景 |
|---|---|---|
| 1020 | ErrAccountLocked | 账号处于锁定状态 |
| 1021 | ErrTooManyAttempts | 滑动窗口内失败次数超限（已锁定） |
| 1022 | ErrAnomalyLoginConfirmed | 检测到异地登录（保留扩展位） |
| 1023 | ErrLoginSecurityUnavailable | SecurityService 装配失败 |

---

## 7. 请求元数据采集

### 7.1 `xincontext.Context` 新增字段

```go
type Context struct {
    // ... 已有字段 ...
    IP        string // c.ClientIP()，处理 X-Forwarded-For 后
    UserAgent string // User-Agent header
    DeviceID  string // X-Device-ID header（前端可选设置）
}
```

### 7.2 Auth 中间件注入

[framework/internal/core/middleware/auth.go](file://d:\work\xin\XinFramework\server\framework\internal\core\middleware\auth.go) 在 `Auth` / `AuthLite` / `OptionalAuth` 三个中间件末尾统一调 `injectRequestMeta(c)` 把请求元数据灌到 ctx。
业务代码无需关心——`auth.Service` 内通过 `attemptFromContext(ctx)` 读取。

---

## 8. 接入步骤（已完成）

### 8.1 数据库

```bash
psql -U xin_user -d xin -f migrations/login_security.sql
```

`migrate.Run` 启动期会自动扫 `migrations/` 目录，无需手动。

### 8.2 配置

`config/config.yaml` 已有 `login_security:` 段，按需调整。

### 8.3 集成点

`apps/boot/auth/module.go` 已自动装配：

```go
security := login_security.NewSecurityService(
    buildSecurityConfig(app),
    login_security.NewPGLockManager(pool),
    login_security.NewPGAttemptStore(pool),
    login_security.NewPGHistoryRecorder(pool),
    nil,  // notifier：留 nil → 自动 fallback 到 LogNotifier
    nil,  // recipients：留 nil → 通知发送会降级
)
```

业务模块可后续注入 `RecipientResolver` 实现，让告警真正走 SMS / Email 通道。

---

## 9. 运维 / 监控

### 9.1 关键查询

```sql
-- 当前被锁账号
SELECT account, locked_until, reason, attempts, ip, created_at
FROM account_locks
WHERE locked_until > NOW()
ORDER BY locked_until DESC;

-- 最近 24h 失败次数 Top 10 账号
SELECT account, COUNT(*) AS fails
FROM login_attempts
WHERE success = FALSE AND created_at > NOW() - INTERVAL '1 day'
GROUP BY account
ORDER BY fails DESC LIMIT 10;

-- 最近 24h 异地登录（device 或 IP 变化）
SELECT h.account_id, h.ip, h.device_id, h.login_at
FROM login_history h
WHERE h.login_at > NOW() - INTERVAL '1 day'
ORDER BY h.login_at DESC LIMIT 100;
```

### 9.2 推荐监控指标

| 指标 | 来源 | 告警阈值 |
|---|---|---|
| 当前锁定账号数 | `SELECT COUNT(*) FROM account_locks WHERE locked_until > NOW()` | > 50 异常 |
| 5 分钟失败登录率 | `login_attempts` 聚合 | > 基线 5 倍 |
| 异地登录数 | `login_history` 配合 detectAnomaly | 单账号 1 次/分钟 |
| Notifier 失败率 | 应用日志 | > 5% |

### 9.3 维护任务

建议每日定时清理过期锁定：

```go
// 接到 cron 调度器或 main 启动后 goroutine
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    for range ticker.C {
        n, err := lockMgr.CleanupExpired(ctx, time.Now())
        if err != nil {
            log.Printf("cleanup locks: %v", err)
        }
        if n > 0 {
            log.Printf("cleaned %d expired locks", n)
        }
    }
}()
```

---

## 10. 已知限制 / 后续演进

| 限制 | 描述 | 后续路径 |
|---|---|---|
| RecipientResolver 留 nil | 通知仅走 LogNotifier | 业务模块注入真实 SMS / Email 实现 |
| IP 维度封禁仅计数 | 暂未触发封禁 | 增加 IP 黑名单接口 + CDN 回源鉴权 |
| 无验证码机制 | 锁定后需等 30 分钟 | 接入图形验证码 / 短信验证 |
| 无设备指纹 | device_id 来自前端 header，可被伪造 | 集成专业 SDK（如 fingerprint.com） |
| 无风险评分 | 固定阈值（5 次） | 引入动态阈值（IP 信誉 / 时间段） |

---

## 11. 测试

`framework/pkg/login_security/login_security_test.go` 14 个测试覆盖：

- Notifier 接口合规性
- MultiNotifier 行为（空 / 部分失败 / 全失败）
- AccountLock.IsActive 时间判定
- detectAnomaly 6 种场景
- SecurityService.CheckLock 锁前 / 锁后
- SecurityService Enabled=false 时全 noop

新增功能时务必同步加测试；测试应覆盖：
1. 锁定触发边界（恰好 MaxFailedAttempts 次）
2. 锁过期后自动失效
3. 异地判定各分支
4. 通知降级（recipient resolver nil）