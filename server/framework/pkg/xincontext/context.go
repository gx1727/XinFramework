// Package xincontext 提供框架自定义的请求上下文类型（Context），
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

type Context struct {
	// Phase 0024 决定：TenantID / UserID 保持 uint，避免全量迁移 50+ 个调用点。
	// 强类型 ID（TenantID / UserID / AccountID / OrgID / RoleID）已在 types.go 定义，
	// 暂以 alias 形式提供（v2 阶段全面替换）。当前业务通过 uint 兼容层 + 文档约定避免误用。
	TenantID  uint
	UserID    uint
	SessionID SessionID
	Role      string
	// SysRoles sys 级角色（如 super_admin），不绑定具体租户
	SysRoles []string

	// 客户端请求元数据（由 Auth / OptionalAuth 中间件统一注入，
	// 供 login_security、audit、notify 等模块使用）。
	IP        string // 客户端 IP（处理代理头后）
	UserAgent string // User-Agent 原始值
	DeviceID  string // 前端设备指纹（可选，从 header X-Device-ID 取）
}

// HasSysRole 判断当前上下文是否携带指定的 sys 级角色。
//
// sys 角色独立于租户内 RBAC（permission.HasGlobalPermission），
// 用于跨租户 / sys 级特权校验，如租户管理、平台字典、平台配置等。
// 典型调用：HasSysRole(jwt.SysRoleSuperAdmin)。
//
// 与 RBAC 通配符判定的区别：
//   - HasSysRole：检查 SysRoles 切片中的角色字符串
//   - permission.HasGlobalPermission：检查 perms map 中的 "*:*" 通配符
func (x *Context) HasSysRole(role string) bool {
	if x == nil || role == "" {
		return false
	}
	for _, r := range x.SysRoles {
		if r == role {
			return true
		}
	}
	return false
}

// Clone returns a copy of Context
func (x *Context) Clone() *Context {
	clone := &Context{
		TenantID:  x.TenantID,
		UserID:    x.UserID,
		SessionID: x.SessionID,
		Role:      x.Role,
		SysRoles:  append([]string(nil), x.SysRoles...),
		IP:        x.IP,
		UserAgent: x.UserAgent,
		DeviceID:  x.DeviceID,
	}
	return clone
}

// UserContext extends Context with RBAC + DataScope
type UserContext struct {
	*Context
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

func WithXinContext(parent context.Context, xc *Context) context.Context {
	return context.WithValue(parent, xinContextKey{}, xc)
}

func XinContextFrom(parent context.Context) (*Context, bool) {
	v, ok := parent.Value(xinContextKey{}).(*Context)
	return v, ok
}

func New(c *gin.Context) *Context {
	if xc, ok := XinContextFrom(c.Request.Context()); ok {
		return xc
	}
	return &Context{}
}

func FromRequest(req *http.Request) *Context {
	if xc, ok := XinContextFrom(req.Context()); ok {
		return xc
	}
	return &Context{}
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
	return &UserContext{Context: &Context{}}
}

func UserContextFromRequest(req *http.Request) *UserContext {
	if uc, ok := UserContextFrom(req.Context()); ok {
		return uc
	}
	return &UserContext{Context: &Context{}}
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

// Context getters

// GetTenantID 返回租户 ID（0 表示平台域 / 未指定）。
func (x *Context) GetTenantID() uint {
	if x == nil {
		return 0
	}
	return x.TenantID
}

// GetUserID 返回用户 ID。
func (x *Context) GetUserID() uint {
	if x == nil {
		return 0
	}
	return x.UserID
}

// GetSessionID 返回强类型会话 ID。
func (x *Context) GetSessionID() SessionID {
	if x == nil {
		return ""
	}
	return x.SessionID
}

func (x *Context) GetRole() string {
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
