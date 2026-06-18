package permission

import (
	"context"
	"testing"
)

// Tests in this file assume db.GetQuerier(ctx) returns an error
// because db.Pool was never initialised. The test process never
// calls db.Init(), so the global pool stays nil. This lets us
// exercise the "db not initialised → return error" branch of every
// PostgresPermissionRepository / DataScopeRepository method.
//
// Each test verifies that the error is propagated (not swallowed),
// and that the method does not panic on a nil Querier.

// TestPostgresPermissionRepository_NilDB_AllMethodsPropagateError
// covers GetUserPermissions, GetUserRoles, GetUserIDsByRole,
// GetUserIDsByResource.
func TestPostgresPermissionRepository_NilDB_AllMethodsPropagateError(t *testing.T) {
	ctx := context.Background()
	r := NewPermissionRepository(nil)

	// All four read paths should return a non-nil error.
	if perms, err := r.GetUserPermissions(ctx, 1); err == nil {
		t.Errorf("GetUserPermissions(nil-db) returned perms=%v with nil error", perms)
	}
	if roles, err := r.GetUserRoles(ctx, 1); err == nil {
		t.Errorf("GetUserRoles(nil-db) returned roles=%v with nil error", roles)
	}
	if ids, err := r.GetUserIDsByRole(ctx, 1); err == nil {
		t.Errorf("GetUserIDsByRole(nil-db) returned ids=%v with nil error", ids)
	}
	if ids, err := r.GetUserIDsByResource(ctx, 1); err == nil {
		t.Errorf("GetUserIDsByResource(nil-db) returned ids=%v with nil error", ids)
	}
}

// TestPostgresDataScopeRepository_NilDB_AllMethodsPropagateError
// covers GetDataScope, GetUserOrgID, GetByRoleID, SetForRole.
func TestPostgresDataScopeRepository_NilDB_AllMethodsPropagateError(t *testing.T) {
	ctx := context.Background()
	r := NewDataScopeRepository(nil)

	if ds, err := r.GetDataScope(ctx, 1); err == nil {
		t.Errorf("GetDataScope(nil-db) returned ds=%v with nil error", ds)
	}
	if org, err := r.GetUserOrgID(ctx, 1); err == nil {
		t.Errorf("GetUserOrgID(nil-db) returned org=%v with nil error", org)
	}
	if ids, err := r.GetByRoleID(ctx, 1); err == nil {
		t.Errorf("GetByRoleID(nil-db) returned ids=%v with nil error", ids)
	}
	if err := r.SetForRole(ctx, 1, []uint{2, 3}); err == nil {
		t.Errorf("SetForRole(nil-db) returned nil error")
	}
}