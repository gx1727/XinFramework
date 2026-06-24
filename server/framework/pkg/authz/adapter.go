package authz

import (
	"context"

	"gx1727.com/xin/framework/pkg/permission"
)

// adapter 把 framework/internal/service.AuthorizationService 适配成公开的
// Authorization 接口。两端签名现在完全一致,adapter 只是 1:1 转发,留着
// 是为了让 boot 包不直接 import 内部 service(由 Go internal 规则隔离)。
type adapter struct {
	inner interface {
		LoadPermissions(ctx context.Context, userID uint) (map[string]bool, error)
		LoadRoles(ctx context.Context, userID uint) ([]string, error)
		LoadDataScope(ctx context.Context, userID uint) (*permission.DataScope, error)
		LoadUserSecurityContext(ctx context.Context, userID uint) (map[string]bool, []string, *permission.DataScope, int64, error)
		InvalidateUser(ctx context.Context, userID uint) error
		InvalidateRole(ctx context.Context, roleID uint) error
		InvalidateResource(ctx context.Context, resourceID uint) error
	}
}

func (a *adapter) LoadPermissions(ctx context.Context, userID uint) (map[string]bool, error) {
	return a.inner.LoadPermissions(ctx, userID)
}

func (a *adapter) LoadRoles(ctx context.Context, userID uint) ([]string, error) {
	return a.inner.LoadRoles(ctx, userID)
}

func (a *adapter) LoadDataScope(ctx context.Context, userID uint) (*permission.DataScope, error) {
	return a.inner.LoadDataScope(ctx, userID)
}

func (a *adapter) LoadUserSecurityContext(ctx context.Context, userID uint) (map[string]bool, []string, *permission.DataScope, int64, error) {
	return a.inner.LoadUserSecurityContext(ctx, userID)
}

func (a *adapter) InvalidateUser(ctx context.Context, userID uint) error {
	return a.inner.InvalidateUser(ctx, userID)
}

func (a *adapter) InvalidateRole(ctx context.Context, roleID uint) error {
	return a.inner.InvalidateRole(ctx, roleID)
}

func (a *adapter) InvalidateResource(ctx context.Context, resourceID uint) error {
	return a.inner.InvalidateResource(ctx, resourceID)
}

// Wrap takes a concrete authorization service that matches the
// signature shape above and returns an Authorization that apps can
// consume.
func Wrap(inner interface {
	LoadPermissions(ctx context.Context, userID uint) (map[string]bool, error)
	LoadRoles(ctx context.Context, userID uint) ([]string, error)
	LoadDataScope(ctx context.Context, userID uint) (*permission.DataScope, error)
	LoadUserSecurityContext(ctx context.Context, userID uint) (map[string]bool, []string, *permission.DataScope, int64, error)
	InvalidateUser(ctx context.Context, userID uint) error
	InvalidateRole(ctx context.Context, roleID uint) error
	InvalidateResource(ctx context.Context, resourceID uint) error
}) Authorization {
	return &adapter{inner: inner}
}