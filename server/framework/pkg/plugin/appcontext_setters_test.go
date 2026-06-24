package plugin

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/framework/pkg/auth"
	"gx1727.com/xin/framework/pkg/config"
	pkgauth "gx1727.com/xin/framework/pkg/tenant/auth"
	"gx1727.com/xin/framework/pkg/tenant"
)

// -----------------------------------------------------------------------------
// Fakes
//
// Each fake embeds the interface it is meant to satisfy. Calling any
// method on the embedded interface panics with a nil-receiver — that's
// intentional, because these tests only verify that the slot stored
// what we put in it. They never invoke a repository method.
// -----------------------------------------------------------------------------

type fakeAccountRepo struct {
	auth.AccountRepository
}

type fakeAccountAuthRepo struct {
	auth.AccountAuthRepository
}

type fakeTenantRepo struct {
	tenant.TenantRepository
}

type fakeUserRepo struct {
	pkgauth.UserRepository
}

type fakeRoleRepo struct {
	pkgauth.RoleRepository
}

type fakeOrgRepo struct {
	pkgauth.OrganizationRepository
}

type fakePermRepo struct {
	pkgauth.RoleResourceRepository
}

// -----------------------------------------------------------------------------
// Setter round-trip tests
//
// One test per setter, each verifying that the Reader getter returns
// the exact same instance after SetX. We use Go interface equality
// (==) which compares the underlying concrete pointer.
// -----------------------------------------------------------------------------

func TestAppContext_SetAccountRepo_RoundTrip(t *testing.T) {
	ctx := newTestContext()
	v := &fakeAccountRepo{}
	ctx.SetAccountRepo(v)
	if got := ctx.AccountRepo(); got != v {
		t.Error("AccountRepo() did not return the value set by SetAccountRepo")
	}
}

func TestAppContext_SetAccountAuthRepo_RoundTrip(t *testing.T) {
	ctx := newTestContext()
	v := &fakeAccountAuthRepo{}
	ctx.SetAccountAuthRepo(v)
	if got := ctx.AccountAuthRepo(); got != v {
		t.Error("AccountAuthRepo() did not return the value set by SetAccountAuthRepo")
	}
}

func TestAppContext_SetTenantRepo_RoundTrip(t *testing.T) {
	ctx := newTestContext()
	v := &fakeTenantRepo{}
	ctx.SetTenantRepo(v)
	if got := ctx.TenantRepo(); got != v {
		t.Error("TenantRepo() did not return the value set by SetTenantRepo")
	}
}

func TestAppContext_SetUserRepo_RoundTrip(t *testing.T) {
	ctx := newTestContext()
	v := &fakeUserRepo{}
	ctx.SetUserRepo(v)
	if got := ctx.UserRepo(); got != v {
		t.Error("UserRepo() did not return the value set by SetUserRepo")
	}
}

func TestAppContext_SetRoleRepo_RoundTrip(t *testing.T) {
	ctx := newTestContext()
	v := &fakeRoleRepo{}
	ctx.SetRoleRepo(v)
	if got := ctx.RoleRepo(); got != v {
		t.Error("RoleRepo() did not return the value set by SetRoleRepo")
	}
}

func TestAppContext_SetOrgRepo_RoundTrip(t *testing.T) {
	ctx := newTestContext()
	v := &fakeOrgRepo{}
	ctx.SetOrgRepo(v)
	if got := ctx.OrgRepo(); got != v {
		t.Error("OrgRepo() did not return the value set by SetOrgRepo")
	}
}

func TestAppContext_SetPermRepo_RoundTrip(t *testing.T) {
	ctx := newTestContext()
	v := &fakePermRepo{}
	ctx.SetPermRepo(v)
	if got := ctx.PermRepo(); got != v {
		t.Error("PermRepo() did not return the value set by SetPermRepo")
	}
}

// -----------------------------------------------------------------------------
// Overwrite semantics: a second SetX(nil) on the same slot must clear
// the value. This is the documented behavior — Writer is last-wins.
// -----------------------------------------------------------------------------

func TestAppContext_AllSetters_CanBeClearedWithNil(t *testing.T) {
	ctx := newTestContext()

	// Fill every slot, then clear it.
	slots := []struct {
		name string
		set  func()
		get  func() any
	}{
		{"AccountRepo", func() { ctx.SetAccountRepo(&fakeAccountRepo{}) }, func() any { return ctx.AccountRepo() }},
		{"AccountAuthRepo", func() { ctx.SetAccountAuthRepo(&fakeAccountAuthRepo{}) }, func() any { return ctx.AccountAuthRepo() }},
		{"TenantRepo", func() { ctx.SetTenantRepo(&fakeTenantRepo{}) }, func() any { return ctx.TenantRepo() }},
		{"UserRepo", func() { ctx.SetUserRepo(&fakeUserRepo{}) }, func() any { return ctx.UserRepo() }},
		{"RoleRepo", func() { ctx.SetRoleRepo(&fakeRoleRepo{}) }, func() any { return ctx.RoleRepo() }},
		{"OrgRepo", func() { ctx.SetOrgRepo(&fakeOrgRepo{}) }, func() any { return ctx.OrgRepo() }},
		{"PermRepo", func() { ctx.SetPermRepo(&fakePermRepo{}) }, func() any { return ctx.PermRepo() }},
	}

	for _, s := range slots {
		s.set()
		if s.get() == nil {
			t.Errorf("%s: get() returned nil after set", s.name)
		}
	}

	// Clear all by setting nil.
	ctx.SetAccountRepo(nil)
	ctx.SetAccountAuthRepo(nil)
	ctx.SetTenantRepo(nil)
	ctx.SetUserRepo(nil)
	ctx.SetRoleRepo(nil)
	ctx.SetOrgRepo(nil)
	ctx.SetPermRepo(nil)

	for _, s := range slots {
		if got := s.get(); got != nil {
			t.Errorf("%s: expected nil after SetX(nil), got %#v", s.name, got)
		}
	}
}

// -----------------------------------------------------------------------------
// Helper: the bare-minimum context for setter tests. We only need db
// and cfg non-nil; everything else stays nil/zero.
// -----------------------------------------------------------------------------

func newTestContext() *AppContext {
	ctx, err := NewAppContext(&pgxpool.Pool{}, nil, &config.Config{}, nil)
	if err != nil {
		panic(err) // 测试辅助函数,boot 期 panic 可接受
	}
	return ctx
}