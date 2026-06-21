package authz

import (
	"context"

	"gx1727.com/xin/framework/pkg/permission"
)

// adapter wraps the concrete *permission.AuthorizationService so that
// its *permission.DataScope return type matches the public
// authz.Authorization interface (which returns interface{}).
//
// Why interface{}? apps/ can't import *permission.DataScope directly
// (it lives in framework/pkg/permission which apps CAN import, but the
// concrete authorization service is in framework/internal/service which
// apps CANNOT import). Returning interface{} keeps apps from depending
// on the internal concrete type while still allowing them to do
// `if ds, ok := result.(*permission.DataScope); ok { ... }`.
//
// LoadUserSecurityContext returns *permission.DataScope directly because
// the auth middleware needs the concrete type to wire it into gin.Context.
// framework is allowed to depend on framework/pkg/permission.
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

func (a *adapter) LoadDataScope(ctx context.Context, userID uint) (interface{}, error) {
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