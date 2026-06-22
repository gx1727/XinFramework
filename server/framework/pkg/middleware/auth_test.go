package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	pkgcontext "gx1727.com/xin/framework/pkg/context"
	jwtpkg "gx1727.com/xin/framework/pkg/jwt"
	"gx1727.com/xin/framework/pkg/permission"
	"gx1727.com/xin/framework/pkg/resp"
)

// buildCtx constructs a *gin.Context pre-loaded with the supplied
// UserContext. This mirrors what the Auth middleware does at runtime:
// it injects both the request's UserContext (consumed by Require* /
// RequireAny / RequireAll via MustNewUserContext) and the embedded
// XinContext (consumed by RequirePlatformRole via xinContext.New).
func buildCtx(uc *pkgcontext.UserContext) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if uc != nil {
		// Order matters: XinContext last so it wins if both collide.
		// UserContext must be set first so the chained WithValue
		// keeps the UserContext slot populated.
		c.Request = c.Request.WithContext(
			pkgcontext.WithUserContext(c.Request.Context(), uc),
		)
		c.Request = c.Request.WithContext(
			pkgcontext.WithXinContext(c.Request.Context(), uc.XinContext),
		)
	}
	return c, rec
}

// makeUC is a small constructor that defaults Permissions to the
// provided map and leaves DataScope at zero value.
func makeUC(userID uint, platformRoles []string, perms map[string]bool) *pkgcontext.UserContext {
	return &pkgcontext.UserContext{
		XinContext:  &pkgcontext.XinContext{UserID: userID, PlatformRoles: platformRoles},
		Permissions: perms,
	}
}

// runOnce dispatches a single request through the handler chain and
// returns (was-aborted, response-status). We use Abort() as the signal
// for "middleware denied the request", since resp.Forbidden/Error
// always call c.Abort().
func runOnce(c *gin.Context, handler gin.HandlerFunc) (aborted bool, status int) {
	w := c.Writer
	handler(c)
	return c.IsAborted(), w.Status()
}

// ============================================================================
// Require
// ============================================================================

func TestRequire_PlatformSuperAdminBypassesAllChecks(t *testing.T) {
	// Platform super_admin with NO permissions at all — Require must still pass.
	uc := makeUC(1, []string{jwtpkg.PlatformRoleSuperAdmin}, nil)
	c, _ := buildCtx(uc)

	mw := Require(permission.P("user", "delete"))
	aborted, _ := runOnce(c, mw)
	if aborted {
		t.Error("platform super admin must bypass Require even with empty Permissions")
	}
}

func TestRequire_ExactMatchAllowed(t *testing.T) {
	uc := makeUC(1, nil, map[string]bool{"user:list": true})
	c, _ := buildCtx(uc)
	mw := Require(permission.P("user", "list"))
	if aborted, _ := runOnce(c, mw); aborted {
		t.Error("exact match should pass")
	}
}

func TestRequire_ResourceWildcardAllowed(t *testing.T) {
	uc := makeUC(1, nil, map[string]bool{"user:*": true})
	c, _ := buildCtx(uc)
	mw := Require(permission.P("user", "delete"))
	if aborted, _ := runOnce(c, mw); aborted {
		t.Error(`"user:*" should grant "user:delete"`)
	}
}

func TestRequire_DeniedWritesForbiddenAndAborts(t *testing.T) {
	uc := makeUC(1, nil, map[string]bool{"user:list": true})
	c, rec := buildCtx(uc)
	mw := Require(permission.P("user", "delete"))
	aborted, status := runOnce(c, mw)
	if !aborted {
		t.Error("denied request must abort the chain")
	}
	if status != http.StatusForbidden {
		t.Errorf("denied request should write 403, got %d (body=%q)", status, rec.Body.String())
	}
}

// ============================================================================
// RequireAuthenticated
// ============================================================================

func TestRequireAuthenticated_LoggedInPasses(t *testing.T) {
	uc := makeUC(1, nil, nil)
	c, _ := buildCtx(uc)
	mw := RequireAuthenticated()
	if aborted, _ := runOnce(c, mw); aborted {
		t.Error("authenticated user must pass RequireAuthenticated with any permission map")
	}
}

// ============================================================================
// RequireAny
// ============================================================================

func TestRequireAny_OneMatchPasses(t *testing.T) {
	uc := makeUC(1, nil, map[string]bool{"user:list": true})
	c, _ := buildCtx(uc)
	mw := RequireAny(
		permission.P("user", "delete"),
		permission.P("user", "list"),
	)
	if aborted, _ := runOnce(c, mw); aborted {
		t.Error("RequireAny should pass when at least one spec matches")
	}
}

func TestRequireAny_AllMissingFails(t *testing.T) {
	uc := makeUC(1, nil, map[string]bool{"order:list": true}) // unrelated
	c, rec := buildCtx(uc)
	mw := RequireAny(
		permission.P("user", "delete"),
		permission.P("user", "create"),
	)
	aborted, status := runOnce(c, mw)
	if !aborted {
		t.Error("RequireAny must deny when no spec matches")
	}
	if status != http.StatusForbidden {
		t.Errorf("expected 403, got %d (body=%q)", status, rec.Body.String())
	}
}

// ============================================================================
// RequireAll
// ============================================================================

func TestRequireAll_AllMatchPasses(t *testing.T) {
	uc := makeUC(1, nil, map[string]bool{
		"user:list":   true,
		"user:create": true,
	})
	c, _ := buildCtx(uc)
	mw := RequireAll(
		permission.P("user", "list"),
		permission.P("user", "create"),
	)
	if aborted, _ := runOnce(c, mw); aborted {
		t.Error("RequireAll must pass when every spec matches")
	}
}

func TestRequireAll_OneMissingFails(t *testing.T) {
	uc := makeUC(1, nil, map[string]bool{"user:list": true}) // missing "user:create"
	c, rec := buildCtx(uc)
	mw := RequireAll(
		permission.P("user", "list"),
		permission.P("user", "create"),
	)
	aborted, status := runOnce(c, mw)
	if !aborted {
		t.Error("RequireAll must deny when at least one spec is missing")
	}
	if status != http.StatusForbidden {
		t.Errorf("expected 403, got %d (body=%q)", status, rec.Body.String())
	}
}

// ============================================================================
// RequirePlatformRole
// ============================================================================

func TestRequirePlatformRole_MatchPasses(t *testing.T) {
	uc := makeUC(1, []string{jwtpkg.PlatformRoleSuperAdmin}, nil)
	c, _ := buildCtx(uc)
	mw := RequirePlatformRole(jwtpkg.PlatformRoleSuperAdmin)
	if aborted, _ := runOnce(c, mw); aborted {
		t.Error("matching platform role should pass")
	}
}

func TestRequirePlatformRole_NoMatchFails(t *testing.T) {
	uc := makeUC(1, []string{"viewer"}, nil)
	c, rec := buildCtx(uc)
	mw := RequirePlatformRole(jwtpkg.PlatformRoleSuperAdmin)
	aborted, status := runOnce(c, mw)
	if !aborted {
		t.Error("non-matching role should be denied")
	}
	if status != http.StatusForbidden {
		t.Errorf("expected 403, got %d (body=%q)", status, rec.Body.String())
	}
}

func TestRequirePlatformRole_EmptyRolesListPasses(t *testing.T) {
	// Empty roles list is a documented "no restriction" sentinel.
	uc := makeUC(1, nil, nil)
	c, _ := buildCtx(uc)
	mw := RequirePlatformRole() // no roles required
	if aborted, _ := runOnce(c, mw); aborted {
		t.Error("empty roles list must be a no-op (caller didn't request any role)")
	}
}

// ============================================================================
// Spec helper coverage (Spec types tested alongside middleware)
// ============================================================================

func TestSpec_ConstructorsAndPredicates(t *testing.T) {
	// P() builds a permission spec that requires both Resource and Action.
	s := permission.P("user", "list")
	if !s.IsPermission() {
		t.Error("P() result should report IsPermission=true")
	}
	if !s.IsValid() {
		t.Error("P(\"user\",\"list\") should be valid")
	}
	if s.IsAuthOnly() {
		t.Error("P() result must not be AuthOnly")
	}

	// AuthOnly() builds a spec that requires login but no specific RBAC.
	a := permission.AuthOnly()
	if !a.IsAuthOnly() {
		t.Error("AuthOnly() must report IsAuthOnly=true")
	}
	if !a.IsValid() {
		t.Error("AuthOnly() must be valid")
	}
	if a.IsPermission() {
		t.Error("AuthOnly() must not be a permission spec")
	}

	// An invalid spec (no Authenticated flag) must fail IsValid().
	invalid := permission.Spec{Resource: "user", Action: "list"} // Authenticated=false
	if invalid.IsValid() {
		t.Error("Spec without Authenticated=true must fail IsValid")
	}
}

// Compile-time sanity: resp package is used inside the middleware.
// This catches accidental package removal that would silently break us.
var _ = resp.Forbidden