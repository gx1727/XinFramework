package permission

import "testing"

// TestSpec_P builds a permission spec via P() and verifies each
// predicate behaves as documented:
//
//   - Resource and Action are populated verbatim.
//   - Authenticated defaults to true (P() is for protected routes).
//   - IsPermission() and IsValid() return true.
//   - IsAuthOnly() returns false (it has a Resource/Action pair).
//   - String() returns "resource:action".
func TestSpec_P(t *testing.T) {
	s := P("user", "list")
	if s.Resource != "user" || s.Action != "list" {
		t.Errorf("P() did not set Resource/Action correctly: %+v", s)
	}
	if !s.Authenticated {
		t.Error("P() result must default Authenticated=true")
	}
	if !s.IsPermission() {
		t.Error(`P("user","list") must report IsPermission=true`)
	}
	if !s.IsValid() {
		t.Error(`P("user","list") must be valid`)
	}
	if s.IsAuthOnly() {
		t.Error(`P() result must not be AuthOnly`)
	}
	if got, want := s.String(), "user:list"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

// TestSpec_AuthOnly verifies the AuthOnly() constructor produces
// a spec that requires login but no specific RBAC permission.
func TestSpec_AuthOnly(t *testing.T) {
	s := AuthOnly()
	if s.Resource != "" || s.Action != "" {
		t.Errorf("AuthOnly() must not populate Resource/Action, got %+v", s)
	}
	if !s.Authenticated {
		t.Error("AuthOnly() must default Authenticated=true")
	}
	if !s.IsAuthOnly() {
		t.Error("AuthOnly() must report IsAuthOnly=true")
	}
	if s.IsPermission() {
		t.Error("AuthOnly() must NOT report IsPermission=true (no Resource/Action)")
	}
	if !s.IsValid() {
		t.Error("AuthOnly() must be valid (Authenticated=true + no Resource)")
	}
	if got := s.String(); got != "auth" {
		t.Errorf("AuthOnly().String() = %q, want \"auth\"", got)
	}
}

// TestSpec_IsValid_RejectedWithoutAuthenticated documents that a spec
// without Authenticated=true must fail validation, even if it carries
// a Resource/Action pair. This protects against accidentally dropping
// the Authenticated flag and creating a permanently-permissive spec.
func TestSpec_IsValid_RejectedWithoutAuthenticated(t *testing.T) {
	invalid := Spec{Resource: "user", Action: "list"} // Authenticated=false
	if invalid.IsValid() {
		t.Error("Spec without Authenticated=true must fail IsValid()")
	}
	// And the empty spec should also fail.
	if (Spec{}).IsValid() {
		t.Error("zero-value Spec must fail IsValid()")
	}
}

// TestSpec_IsPermission covers the "any of Resource/Action populated"
// predicate. AuthOnly() returns false here; P() returns true; an
// arbitrary half-populated spec also returns true.
func TestSpec_IsPermission(t *testing.T) {
	if !P("user", "list").IsPermission() {
		t.Error("P() result must report IsPermission=true")
	}
	if AuthOnly().IsPermission() {
		t.Error("AuthOnly() must NOT report IsPermission=true")
	}
	// A spec with only Resource set is technically a permission spec,
	// even if incomplete. The predicate is intentionally loose; the
	// strict gate is IsValid().
	half := Spec{Resource: "user", Authenticated: true}
	if !half.IsPermission() {
		t.Error(`{Resource:"user"} must report IsPermission=true`)
	}
}