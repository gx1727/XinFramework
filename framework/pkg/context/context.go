package context

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type XinContext struct {
	TenantID  uint
	UserID    uint
	SessionID string
	Role      string
}

type xinContextKey struct{}

func WithXinContext(parent context.Context, xc *XinContext) context.Context {
	return context.WithValue(parent, xinContextKey{}, xc)
}

func XinContextFrom(parent context.Context) (*XinContext, bool) {
	v, ok := parent.Value(xinContextKey{}).(*XinContext)
	return v, ok
}

func New(c *gin.Context) *XinContext {
	if xc, ok := XinContextFrom(c.Request.Context()); ok {
		return xc
	}
	return &XinContext{}
}

func FromRequest(req *http.Request) *XinContext {
	if xc, ok := XinContextFrom(req.Context()); ok {
		return xc
	}
	return &XinContext{}
}

func (x *XinContext) SetTenantID(id uint) {
	x.TenantID = id
}

func (x *XinContext) SetUserID(id uint) {
	x.UserID = id
}

func (x *XinContext) SetSessionID(id string) {
	x.SessionID = id
}

func (x *XinContext) SetRole(role string) {
	x.Role = role
}

func (x *XinContext) GetTenantID() uint {
	return x.TenantID
}

func (x *XinContext) GetUserID() uint {
	return x.UserID
}

func (x *XinContext) GetSessionID() string {
	return x.SessionID
}

func (x *XinContext) GetRole() string {
	return x.Role
}

type tenantKey struct{}

func WithTenantID(parent context.Context, tenantID uint) context.Context {
	return context.WithValue(parent, tenantKey{}, tenantID)
}

func TenantIDFrom(parent context.Context) (uint, bool) {
	v, ok := parent.Value(tenantKey{}).(uint)
	return v, ok
}
