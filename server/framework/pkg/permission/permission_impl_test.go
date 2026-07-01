package permission

import (
	"context"
	"sort"
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

// TestExpandPermissionCode 覆盖 0024 通配展开逻辑。
// 规则：
//   - "x"    → {"x", "x:*"}                                                  菜单无关资源
//   - "x:y"  → {"x:y"}                                                        菜单相关具体 action
//   - "x:*"  → allActions 展开                                                菜单相关所有 action
//   - "*:*"  → {"*:*"}                                                        全局通配
func TestExpandPermissionCode(t *testing.T) {
	tests := []struct {
		name string
		code string
		want []string
	}{
		{"normal", "user:list", []string{"user:list"}},
		{"normal-create", "platform-permissions:create", []string{"platform-permissions:create"}},
		{"global-wildcard", "*:*", []string{"*:*"}},
		{
			"resource-wildcard",
			"platform-permissions:*",
			[]string{
				"platform-permissions:list",
				"platform-permissions:get",
				"platform-permissions:create",
				"platform-permissions:update",
				"platform-permissions:delete",
				"platform-permissions:tree",
			},
		},
		{
			"menu-less-resource",
			"changepwd",
			[]string{"changepwd", "changepwd:*"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandPermissionCode(tt.code)
			sort.Strings(got)
			want := append([]string(nil), tt.want...)
			sort.Strings(want)
			if len(got) != len(want) {
				t.Fatalf("expandPermissionCode(%q) len=%d want=%d (got=%v)", tt.code, len(got), len(want), got)
			}
			for i := range got {
				if got[i] != want[i] {
					t.Errorf("expandPermissionCode(%q)[%d]=%q want=%q", tt.code, i, got[i], want[i])
				}
			}
		})
	}
}
