package context

import (
	"context"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/permission"
)

type XinContext struct {
	TenantID  uint
	UserID    uint
	SessionID string
	Role      string
	// PlatformRoles 平台级角色（如 super_admin），不绑定具体租户
	PlatformRoles []string
}

// IsSuperAdmin 判断当前上下文是否携带 super_admin 平台角色
func (x *XinContext) IsSuperAdmin() bool {
	if x == nil {
		return false
	}
	for _, r := range x.PlatformRoles {
		if r == "super_admin" {
			return true
		}
	}
	return false
}

// Clone returns a copy of XinContext
func (x *XinContext) Clone() *XinContext {
	clone := &XinContext{
		TenantID:      x.TenantID,
		UserID:        x.UserID,
		SessionID:     x.SessionID,
		Role:          x.Role,
		PlatformRoles: append([]string(nil), x.PlatformRoles...),
	}
	return clone
}

// UserContext extends XinContext with RBAC + DataScope
type UserContext struct {
	*XinContext
	OrgID       int64
	Roles       []string
	Permissions map[string]bool
	DataScope   permission.DataScope
}

type xinContextKey struct{}
type userContextKey struct{}
type userContextLoaderKey struct{}

// userContextWrapper 用于实现懒加载且只执行一次
type userContextWrapper struct {
	once   sync.Once
	uc     *UserContext
	loader func() *UserContext
}

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

func WithUserContext(parent context.Context, uc *UserContext) context.Context {
	return context.WithValue(parent, userContextKey{}, uc)
}

func WithUserContextLoader(parent context.Context, loader func() *UserContext) context.Context {
	wrapper := &userContextWrapper{loader: loader}
	return context.WithValue(parent, userContextLoaderKey{}, wrapper)
}

func UserContextFrom(parent context.Context) (*UserContext, bool) {
	// 先看上下文里是不是已经有了生成好的实体
	if v, ok := parent.Value(userContextKey{}).(*UserContext); ok {
		return v, true
	}

	// 如果没有实体，看看有没有注册懒加载生成器
	if wrapper, ok := parent.Value(userContextLoaderKey{}).(*userContextWrapper); ok {
		// 使用 sync.Once 确保 loader 只执行一次
		wrapper.once.Do(func() {
			wrapper.uc = wrapper.loader()
		})
		return wrapper.uc, true
	}

	return nil, false
}

func NewUserContext(c *gin.Context) *UserContext {
	if uc, ok := UserContextFrom(c.Request.Context()); ok {
		return uc
	}
	return &UserContext{XinContext: &XinContext{}}
}

// MustNewUserContext returns the UserContext or panics if it's not present or invalid.
// This is useful for catching missing middleware configuration.
func MustNewUserContext(c *gin.Context) *UserContext {
	uc, ok := UserContextFrom(c.Request.Context())
	if !ok || uc.UserID == 0 {
		panic("UserContext not found or UserID is 0. Did you forget to add the Auth middleware?")
	}
	return uc
}

func UserContextFromRequest(req *http.Request) *UserContext {
	if uc, ok := UserContextFrom(req.Context()); ok {
		return uc
	}
	return &UserContext{XinContext: &XinContext{}}
}

// HasPermission checks if user has the specified permission
func (u *UserContext) HasPermission(resource, action string) bool {
	return permission.HasPermission(u.Permissions, resource, action)
}

// GetDataScopeFilter returns SQL WHERE clause and args for data filtering
func (u *UserContext) GetDataScopeFilter() (string, []any, error) {
	filter, err := u.GetDataScopeFilterFor(permission.DefaultScopeColumns)
	if err != nil {
		return "", nil, err
	}
	return filter.SQL, filter.Args, nil
}

// GetDataScopeFilterFor returns a data-scope filter using explicit column mapping.
func (u *UserContext) GetDataScopeFilterFor(columns permission.ScopeColumns) (permission.ScopeFilter, error) {
	return permission.BuildDataScopeFilter(u.DataScope, u.UserID, u.OrgID, columns)
}

// XinContext getters

func (x *XinContext) GetTenantID() uint {
	if x == nil {
		return 0
	}
	return x.TenantID
}

func (x *XinContext) GetUserID() uint {
	if x == nil {
		return 0
	}
	return x.UserID
}

func (x *XinContext) GetSessionID() string {
	if x == nil {
		return ""
	}
	return x.SessionID
}

func (x *XinContext) GetRole() string {
	if x == nil {
		return ""
	}
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
