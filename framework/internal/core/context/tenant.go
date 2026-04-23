package context

import (
	"context"
)

type tenantKey struct{}

func WithTenantID(parent context.Context, tenantID uint) context.Context {
	return context.WithValue(parent, tenantKey{}, tenantID)
}

func TenantIDFrom(parent context.Context) (uint, bool) {
	v, ok := parent.Value(tenantKey{}).(uint)
	return v, ok
}
