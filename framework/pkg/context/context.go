package context

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/permission"
)

type XinContext struct {
	TenantID  uint
	UserID    uint
	SessionID string
	Role      string
}

// UserContext extends XinContext with RBAC + DataScope
type UserContext struct {
	TenantID    uint
	UserID      uint
	OrgID       int64
	SessionID   string
	Roles       []string
	Permissions map[string]bool
	DataScope   permission.DataScope
}

type xinContextKey struct{}
type userContextKey struct{}

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

// UserContext methods

func WithUserContext(parent context.Context, uc *UserContext) context.Context {
	return context.WithValue(parent, userContextKey{}, uc)
}

func UserContextFrom(parent context.Context) (*UserContext, bool) {
	v, ok := parent.Value(userContextKey{}).(*UserContext)
	return v, ok
}

func NewUserContext(c *gin.Context) *UserContext {
	if uc, ok := UserContextFrom(c.Request.Context()); ok {
		return uc
	}
	return &UserContext{}
}

func UserContextFromRequest(req *http.Request) *UserContext {
	if uc, ok := UserContextFrom(req.Context()); ok {
		return uc
	}
	return &UserContext{}
}

// HasPermission checks if user has the specified permission
func (u *UserContext) HasPermission(resource, action string) bool {
	return permission.HasPermission(u.Permissions, resource, action)
}

// GetDataScopeFilter returns SQL WHERE clause and args for data filtering
func (u *UserContext) GetDataScopeFilter() (string, []any, error) {
	switch u.DataScope.Type {
	case permission.DataScopeAll:
		return "", nil, nil
	case permission.DataScopeSelf:
		return "creator_id = $1", []any{u.UserID}, nil
	case permission.DataScopeCustom:
		if len(u.DataScope.OrgIDs) == 0 {
			return "creator_id = $1", []any{u.UserID}, nil
		}
		return "org_id = ANY($1)", []any{u.DataScope.OrgIDs}, nil
	case permission.DataScopeDept:
		if u.OrgID == 0 {
			return "creator_id = $1", []any{u.UserID}, nil
		}
		return "org_id = $1", []any{u.OrgID}, nil
	case permission.DataScopeDeptAndBelow:
		if u.OrgID == 0 {
			return "creator_id = $1", []any{u.UserID}, nil
		}
		// Use CTE to find all descendant org IDs
		return `
			org_id = $1
			OR org_id IN (
				WITH RECURSIVE org_tree AS (
					SELECT id FROM organizations WHERE id = $1
					UNION ALL
					SELECT o.id FROM organizations o
					JOIN org_tree ot ON o.parent_id = ot.id
				)
				SELECT id FROM org_tree
			)
		`, []any{u.OrgID}, nil
	default:
		return "creator_id = $1", []any{u.UserID}, nil
	}
}

// XinContext setters/getters (unchanged)

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
