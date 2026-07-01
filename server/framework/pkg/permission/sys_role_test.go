package permission

import (
	"context"
	"testing"
)

// TestPostgresSysRoleRepository_NilDB_AllMethodsPropagateError
// covers GetRolesByAccountID, GetRolesByUserID, Grant, Revoke.
// Same setup as permission_impl_test.go: db.Pool is nil in the test
// process, so db.GetQuerier returns an error and every method should
// propagate it without panic.
func TestPostgresSysRoleRepository_NilDB_AllMethodsPropagateError(t *testing.T) {
	ctx := context.Background()
	r := NewSysRoleRepository(nil)

	if roles, err := r.GetRolesByAccountID(ctx, 1); err == nil {
		t.Errorf("GetRolesByAccountID(nil-db) returned roles=%v with nil error", roles)
	}
	if roles, err := r.GetRolesByUserID(ctx, 1); err == nil {
		t.Errorf("GetRolesByUserID(nil-db) returned roles=%v with nil error", roles)
	}
	if err := r.Grant(ctx, 1, "super_admin"); err == nil {
		t.Error("Grant(nil-db) returned nil error")
	}
	if err := r.Revoke(ctx, 1, "super_admin"); err == nil {
		t.Error("Revoke(nil-db) returned nil error")
	}
}
