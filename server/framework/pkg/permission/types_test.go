package permission

import "testing"

// TestHasPermission_NilMap covers the defensive nil-map branch.
func TestHasPermission_NilMap(t *testing.T) {
	if HasPermission(nil, "user", "list") {
		t.Fatal("nil permission map should return false")
	}
	if HasPermission(map[string]bool{}, "user", "list") {
		t.Fatal("empty permission map should return false")
	}
}

// TestHasPermission_ExactMatch exercises the primary code path:
// the perm map contains exactly the "resource:action" key the
// caller asked for.
func TestHasPermission_ExactMatch(t *testing.T) {
	perms := map[string]bool{"user:list": true, "user:create": true}
	if !HasPermission(perms, "user", "list") {
		t.Error(`perms={"user:list":true} → "user:list" should be allowed`)
	}
	if !HasPermission(perms, "user", "create") {
		t.Error(`perms={"user:create":true} → "user:create" should be allowed`)
	}
	if HasPermission(perms, "user", "delete") {
		t.Error(`perms={"user:list","user:create"} → "user:delete" should be denied`)
	}
	if HasPermission(perms, "order", "list") {
		t.Error("permission for `user:*` must not leak into a different resource")
	}
}

// TestHasPermission_ResourceWildcard covers the secondary code path:
// the map carries "user:*" which should grant every action on
// the user resource.
func TestHasPermission_ResourceWildcard(t *testing.T) {
	perms := map[string]bool{"user:*": true}
	for _, action := range []string{"list", "create", "update", "delete", "audit"} {
		if !HasPermission(perms, "user", action) {
			t.Errorf(`perms={"user:*":true} should grant "user:%s"`, action)
		}
	}
	// Resource wildcard MUST NOT cross resource boundaries.
	if HasPermission(perms, "order", "list") {
		t.Error(`"user:*" must not grant "order:list"`)
	}
}

// TestHasPermission_GlobalWildcard covers the super-admin escape hatch.
// "*:*" should grant everything regardless of resource/action.
func TestHasPermission_GlobalWildcard(t *testing.T) {
	perms := map[string]bool{"*:*": true}
	if !HasPermission(perms, "user", "list") {
		t.Error(`"*:*" should grant user:list`)
	}
	if !HasPermission(perms, "billing", "export") {
		t.Error(`"*:*" should grant billing:export`)
	}
	if !HasPermission(perms, "anything", "goes") {
		t.Error(`"*:*" should grant arbitrary keys`)
	}
}

// TestHasPermission_Precedence documents the lookup order:
//   1. exact match
//   2. resource wildcard
//   3. global wildcard
// If multiple are present, exact match wins (no surprise).
func TestHasPermission_Precedence(t *testing.T) {
	// Exact + resource wildcard present → exact match still returns true
	// (function never even reaches the wildcard branch in this case,
	// but we verify behavior stays correct when both are populated).
	perms := map[string]bool{
		"user:list": true,
		"user:*":    true,
		"*:*":       true,
	}
	if !HasPermission(perms, "user", "list") {
		t.Error("exact match must succeed when populated alongside wildcards")
	}
	// A non-exact action falls through to resource wildcard, not global.
	if !HasPermission(perms, "user", "delete") {
		t.Error("resource wildcard should grant user:delete")
	}
	// An unrelated resource falls through to global wildcard.
	if !HasPermission(perms, "order", "list") {
		t.Error("global wildcard should grant order:list")
	}
}

// TestHasGlobalPermission pairs with TestHasPermission_GlobalWildcard: a user
// whose perm map carries "*:*" is granted all permissions.
func TestHasGlobalPermission(t *testing.T) {
	if !HasGlobalPermission(map[string]bool{"*:*": true}) {
		t.Error(`{"*:*":true} must report HasGlobalPermission=true`)
	}
	if HasGlobalPermission(map[string]bool{"user:*": true}) {
		t.Error(`{"user:*":true} must NOT be global (resource-scoped wildcard)`)
	}
	if HasGlobalPermission(nil) {
		t.Error("nil perms must not be global")
	}
}

// BuildDataScopeSQL was removed (was a deprecated thin wrapper that
// delegated to BuildDataScopeFilter with DefaultScopeColumns). The two
// old tests that exercised it are also removed: no API left to cover.