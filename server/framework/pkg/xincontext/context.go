// Package xincontext 提供框架自定义的请求上下文类型（XinContext），
// 封装从 JWT claims 中解析出的 UserID、TenantID、SessionID 等身份信息。
//
// 注意：包名刻意使用 xincontext 而非 context，以避免遮蔽标准库 context 包，
// 这样调用方不需要每次都写 import alias。
package xincontext

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

// HasPlatformRole 判断当前上下文是否携带指定的平台级角色。
//
// 平台角色独立于租户内 RBAC（permission.HasGlobalPermission），
// 用于跨租户 / 平台级特权校验，如租户管理、平台字典、平台配置等。
// 典型调用：HasPlatformRole(jwt.PlatformRoleSuperAdmin)。
//
// 与 RBAC 通配符判定的区别：
//   - HasPlatformRole：检查 PlatformRoles 切片中的角色字符串
//   - permission.HasGlobalPermission：检查 perms map 中的 "*:*" 通配符
func (x *XinContext) HasPlatformRole(role string) bool {
	if x == nil || role == "" {
		return false
	}
	for _, r := range x.PlatformRoles {
		if r == role {
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

// DataScopeFilter 返回 SQL WHERE 子句片段和绑定参数。
//
// Deprecated: 请改用 xincontext.ScopeFilterFrom(ctx, columns)，无需手动构造 UserContext。
func (u *UserContext) DataScopeFilter() (string, []any, error) {
	filter, err := u.DataScopeFilterFor(permission.DefaultScopeColumns)
	if err != nil {
		return "", nil, err
	}
	return filter.SQL, filter.Args, nil
}

// DataScopeFilterFor 返回基于指定列映射的 DataScope SQL 过滤片段。
//
// Deprecated: 请改用 xincontext.ScopeFilterFrom(ctx, columns)，无需手动取 UserContext。
func (u *UserContext) DataScopeFilterFor(columns permission.ScopeColumns) (permission.ScopeFilter, error) {
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
