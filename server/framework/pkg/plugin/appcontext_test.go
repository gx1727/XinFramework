package plugin

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/framework/pkg/config"
)

// -----------------------------------------------------------------------------
// Construction: nil-handling panics
// -----------------------------------------------------------------------------

// TestNewAppContext_NilDB_Panics verifies that NewAppContext refuses to
// construct an AppContext without a database pool — a nil db would let
// every repository downstream crash with a confusing nil-pointer
// dereference far from the root cause.
func TestNewAppContext_NilDB_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewAppContext(nil, nil, &cfg, nil) must panic on nil db")
		}
	}()
	_ = NewAppContext(nil, nil, &config.Config{}, nil)
}

// TestNewAppContext_NilConfig_Panics: same rationale as the db panic.
func TestNewAppContext_NilConfig_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewAppContext(pool, nil, nil, nil) must panic on nil cfg")
		}
	}()
	_ = NewAppContext(&pgxpool.Pool{}, nil, nil, nil)
}

// TestNewAppContext_HappyPath constructs a context with zero-value infra
// and reads it back via the Reader interface.
func TestNewAppContext_HappyPath(t *testing.T) {
	var (
		pool = &pgxpool.Pool{}
		cfg  = &config.Config{}
	)

	ctx := NewAppContext(pool, nil, cfg, nil)
	if ctx.DB() != pool {
		t.Error("Reader.DB() should return the pool passed to NewAppContext")
	}
	if ctx.Cache() != nil {
		t.Error("Reader.Cache() should be nil when cache was nil at construction")
	}
	if ctx.Config() != cfg {
		t.Error("Reader.Config() should return the config passed to NewAppContext")
	}
	if ctx.Session() != nil {
		t.Error("Reader.Session() should be nil when session was nil at construction")
	}
}

// -----------------------------------------------------------------------------
// Reader/Writer separation: every slot defaults to nil
// -----------------------------------------------------------------------------

// TestAppContext_DefaultsAreNil asserts that an uninitialised AppContext
// has every repository / service slot reading back as nil. This is the
// single source of truth that "module not enabled" is observable as a
// nil repository rather than a panic.
func TestAppContext_DefaultsAreNil(t *testing.T) {
	ctx := NewAppContext(&pgxpool.Pool{}, nil, &config.Config{}, nil)

	if ctx.Authz() != nil {
		t.Error("Authz() should default to nil")
	}
	if ctx.AccountRepo() != nil {
		t.Error("AccountRepo() should default to nil")
	}
	if ctx.AccountAuthRepo() != nil {
		t.Error("AccountAuthRepo() should default to nil")
	}
	if ctx.TenantRepo() != nil {
		t.Error("TenantRepo() should default to nil")
	}
	if ctx.UserRepo() != nil {
		t.Error("UserRepo() should default to nil")
	}
	if ctx.RoleRepo() != nil {
		t.Error("RoleRepo() should default to nil")
	}
	if ctx.OrgRepo() != nil {
		t.Error("OrgRepo() should default to nil")
	}
	if ctx.PermRepo() != nil {
		t.Error("PermRepo() should default to nil")
	}
}

// -----------------------------------------------------------------------------
// Writer side: SetX must populate the matching Reader getter
// -----------------------------------------------------------------------------

// fakeAuthz is a stand-in Authorization implementation. We can't use
// nil because some code paths may dereference; a struct with zero-value
// methods satisfies the interface trivially.
type fakeAuthz struct{}

func (fakeAuthz) LoadPermissions(context.Context, uint) (map[string]bool, error) {
	return nil, nil
}
func (fakeAuthz) LoadRoles(context.Context, uint) ([]string, error)          { return nil, nil }
func (fakeAuthz) LoadDataScope(context.Context, uint) (interface{}, error)    { return nil, nil }
func (fakeAuthz) InvalidateUser(context.Context, uint) error                 { return nil }
func (fakeAuthz) InvalidateRole(context.Context, uint) error                 { return nil }
func (fakeAuthz) InvalidateResource(context.Context, uint) error             { return nil }

// TestAppContext_SetAuthz_RoundTrip is the most important Writer/Reader
// property: after SetAuthz, Reader.Authz() returns the same value.
func TestAppContext_SetAuthz_RoundTrip(t *testing.T) {
	ctx := NewAppContext(&pgxpool.Pool{}, nil, &config.Config{}, nil)
	v := fakeAuthz{}
	ctx.SetAuthz(v)
	if got := ctx.Authz(); got == nil {
		t.Error("after SetAuthz(fakeAuthz{}), Reader.Authz() must return non-nil")
	} else if got != v {
		t.Error("Reader.Authz() returned a different value than what was set")
	}
}

// TestAppContext_SetAuthz_Overwrite documents that SetX overwrites
// (last writer wins). This is intentional: boot.Init is the canonical
// writer and only runs once, but tests of third-party wiring may set
// then override.
func TestAppContext_SetAuthz_Overwrite(t *testing.T) {
	ctx := NewAppContext(&pgxpool.Pool{}, nil, &config.Config{}, nil)
	ctx.SetAuthz(fakeAuthz{})
	ctx.SetAuthz(nil) // explicit clear
	if ctx.Authz() != nil {
		t.Error("second SetAuthz(nil) should clear the slot")
	}
}

// -----------------------------------------------------------------------------
// Reader/Writer interface assertions (compile-time already enforced, but
// a runtime check documents the contract for anyone reading the test).
// -----------------------------------------------------------------------------

// TestAppContext_SatisfiesReaderAndWriter exists to make the intent
// explicit in the test output. The compile-time assertion in
// appcontext.go (var _ Reader = ...) is the real guarantee.
func TestAppContext_SatisfiesReaderAndWriter(t *testing.T) {
	var _ Reader = (*AppContext)(nil)
	var _ Writer = (*AppContext)(nil)
}

// -----------------------------------------------------------------------------
// Helper placeholders above keep the imports balanced without pulling in
// redis / session concrete types (the tests pass nil intentionally).
// -----------------------------------------------------------------------------

type redisCache = struct {
	// Shape placeholder so we can keep imports stable while still
	// explicitly typing nil values in callers if needed.
}

// memorySession is a placeholder type for tests that want to pass a
// typed nil SessionManager (not used directly; declared for symmetry).
type memorySession = struct{}