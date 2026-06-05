package audit

import "context"

// ctxKey 是 audit 包内部用的 context key 类型，避免与其它包冲突。
type ctxKey struct{}

// WithIP 把客户端 IP 放入 ctx，供 audit.Log 写 db_logs.ip 列。
// 由 ClientIP 中间件在 setupRouter 里统一注入。
func WithIP(parent context.Context, ip string) context.Context {
	if ip == "" {
		return parent
	}
	return context.WithValue(parent, ctxKey{}, ip)
}

// IPFrom 读出 ctx 里携带的 IP；不存在时返回空串。
func IPFrom(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKey{}).(string); ok {
		return v
	}
	return ""
}
