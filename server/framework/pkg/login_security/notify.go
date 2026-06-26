package login_security

import (
	"context"

	"gx1727.com/xin/framework/pkg/logger"
	"gx1727.com/xin/framework/pkg/task"
)

// Channel 通知通道枚举。
type Channel string

const (
	ChannelSMS    Channel = "sms"
	ChannelEmail  Channel = "email"
	ChannelInSite Channel = "in_site" // 站内消息（框架 apps/tenant/message）
)

// NotificationPayload 通知内容。
type NotificationPayload struct {
	Channel   Channel
	Recipient string            // 收件人（手机号 / 邮箱 / 站内消息 user_id）
	Subject   string            // 邮件主题 / 短信签名（站内消息忽略）
	Body      string            // 正文
	AccountID uint              // 关联账号
	Reason    string            // 触发原因："异地登录" / "账号锁定" / ...
	Extra     map[string]string // 通道特定的扩展字段
}

// Notifier 通知通道抽象。
//
// 框架默认只提供 LogNotifier（仅写日志，不实际发短信/邮件）。
// 业务模块可注入自己的实现（如集成腾讯云 SMS / SendGrid）。
type Notifier interface {
	// Notify 发送一条通知。失败不应阻断业务路径，由实现自行处理重试。
	Notify(ctx context.Context, payload NotificationPayload) error
}

// MultiNotifier 组合多个 Notifier（按顺序投递，任一成功即视为整体成功）。
type MultiNotifier struct {
	notifiers []Notifier
}

// NewMultiNotifier 构造组合通知器。
func NewMultiNotifier(notifiers ...Notifier) *MultiNotifier {
	return &MultiNotifier{notifiers: notifiers}
}

// Notify 依次调用每个 Notifier；任一成功即视为成功，全部失败才返回错误。
//
// 实际场景：成功发送的渠道继续跑（用于埋点），失败的只记日志。
func (m *MultiNotifier) Notify(ctx context.Context, p NotificationPayload) error {
	if len(m.notifiers) == 0 {
		return nil
	}
	var lastErr error
	sent := 0
	for _, n := range m.notifiers {
		if n == nil {
			continue
		}
		if err := n.Notify(ctx, p); err != nil {
			lastErr = err
			continue
		}
		sent++
	}
	if sent == 0 && lastErr != nil {
		return lastErr
	}
	return nil
}

// LogNotifier 是默认的 Notifier 实现：仅写日志，不实际发短信/邮件。
//
// 适用于：
//   - dev/test 环境
//   - 尚未集成真实短信/邮件服务的过渡期
//   - 业务路径需要"至少有个通知器在跑"，但暂时不发外部消息
type LogNotifier struct {
	module string // 日志模块名，默认 "login_security"
}

// NewLogNotifier 构造默认日志通知器。
func NewLogNotifier() *LogNotifier {
	return &LogNotifier{module: "login_security"}
}

// Notify 把 payload 写到 zap 日志（INFO 级别）。
//
// 日志结构：channel / recipient / subject / reason / accountID / extra。
// 运维可通过 grep "[login_security]" 拿到所有通知流水。
func (l *LogNotifier) Notify(_ context.Context, p NotificationPayload) error {
	log := logger.Module(l.module)
	extra := ""
	if len(p.Extra) > 0 {
		for k, v := range p.Extra {
			extra += " " + k + "=" + v
		}
	}
	log.Infof(
		"notify channel=%s recipient=%s subject=%q reason=%s accountID=%d body=%q%s",
		p.Channel, p.Recipient, p.Subject, p.Reason, p.AccountID, p.Body, extra,
	)
	return nil
}

// QueueNotifier 把通知入队到后台任务系统，由 worker 异步执行。
//
// 适用场景：
//   - 需要真实短信/邮件通道，但调用方不希望被第三方 API 拖慢
//   - 失败可重试（5 次指数退避）
//   - 不丢失：即使 worker 进程崩溃，DB 里的任务仍可被下一个 worker 消费
//
// 使用方式（在 apps/task/module.go 注册 send_notification handler）：
//
//	taskpkg.RegisterHandler(taskpkg.HandlerFunc{
//	    KindStr: "send_notification",
//	    HandleFn: func(ctx context.Context, t *taskpkg.Task) error {
//	        var p login_security.NotificationPayload
//	        if err := json.Unmarshal(t.Payload, &p); err != nil {
//	            return err
//	        }
//	        return sendSMSOrEmail(ctx, p)
//	    },
//	})
//
// SecurityService 在 Notifier != nil 时自动优先用 QueueNotifier（详见 SecurityService 配置）。
type QueueNotifier struct {
	queue task.Queue
	// KindName 是入队的任务 kind（默认 "send_notification"）。
	// handler 注册时必须用同样的字符串。
	KindName string
}

// NewQueueNotifier 构造异步通知器。
//
// queue 为 nil 时退化为 panic-free noop（Notify 直接返回 nil）。
func NewQueueNotifier(queue task.Queue) *QueueNotifier {
	return &QueueNotifier{queue: queue, KindName: "send_notification"}
}

// Notify 把 payload 序列化为 JSON 后入队。失败仅记日志（不阻塞业务路径）。
func (q *QueueNotifier) Notify(ctx context.Context, p NotificationPayload) error {
	if q.queue == nil {
		return nil
	}
	payload := task.MarshalPayload(p)
	kind := q.KindName
	if kind == "" {
		kind = "send_notification"
	}
	_, err := q.queue.Enqueue(ctx, kind, payload,
		task.WithPriority(10),  // 通知类任务高优先级
		task.WithMaxAttempts(5),
		task.WithTimeout(60),
	)
	if err != nil {
		logger.Module("login_security").Warnf("enqueue notification failed: %v", err)
		// 入队失败不回传给调用方——通知是 best-effort
	}
	return nil
}

// Compile-time guarantee.
var _ Notifier = (*LogNotifier)(nil)
var _ Notifier = (*MultiNotifier)(nil)
var _ Notifier = (*QueueNotifier)(nil)