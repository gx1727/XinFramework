// Package plugin defines the AppContext type, which is the single
// shared dependency container that replaces the 12 cross-module package-level
// globals the framework used to rely on (Phase B refactor).
//
// Design rationale
//
//   - Each module needs both "what can I consume" (DB, Config, other
//     modules' repositories) and "what do I contribute" (my own
//     repositories / services). Splitting the role into a Reader and a
//     Writer makes both intents explicit at the type level.
//
//   - Reader is the read-side handle every module receives during
//     Init() and Register(). Fields are typed (no `interface{}` casts),
//     and reading a nil value is the documented way to detect that
//     a dependency module was not enabled.
//
//   - Writer is the write-side handle ONLY given to a module that owns
//     a particular slot. A module that did not declare
//     "I contribute X" cannot accidentally overwrite another module's
//     repository because it never sees the Writer for that slot.
//
//   - AppContext is a concrete struct, not an interface. The struct is
//     constructed exactly once in boot.Init and is then passed by
//     pointer. The Reader/Writer interfaces exist purely to make the
//     module contract compiler-enforceable.
package plugin

import (
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/framework/pkg/auth"
	"gx1727.com/xin/framework/pkg/authz"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/rbac"
	"gx1727.com/xin/framework/pkg/session"
	"gx1727.com/xin/framework/pkg/tenant"
)

// Reader is the read-side handle every Module sees. It exposes:
//
//   - Infrastructure: DB pool, Redis client, Config, Session manager.
//   - Cross-module services and repositories contributed by other
//     modules (filled during Init() of the contributor).
//
// Read-side methods that return a typed pointer return nil if the
// corresponding module was not enabled in cfg.Module. Modules must
// nil-check the repositories they actually need.
//
// Why interfaces, not just the concrete AppContext struct?
//
//   - During Init() a module must not be able to call SetX() — the
//     role split catches that misuse at compile time.
//   - Tests can construct a tiny fake Reader without building a full
//     AppContext.
type Reader interface {
	// Infrastructure (always present after boot.Init).
	DB() *pgxpool.Pool
	Cache() *redis.Client // may return nil if Redis is disabled
	Config() *config.Config
	Session() session.SessionManager

	// Cross-module services (filled after every module's Init).
	Authz() authz.Authorization

	// Repositories contributed by other modules. Each returns nil
	// when the producing module was not enabled in cfg.Module.
	AccountRepo() auth.AccountRepository
	AccountAuthRepo() auth.AccountAuthRepository
	TenantRepo() tenant.TenantRepository
	UserRepo() rbac.UserRepository
	RoleRepo() rbac.RoleRepository
	OrgRepo() rbac.OrganizationRepository
	PermRepo() rbac.RoleResourceRepository
}

// Writer is the write-side handle given only to a Module that owns
// a particular slot. Each SetX method is intended to be called at
// most once, by the module whose name matches the slot:
//
//   - SetAuthz             ← framework (boot.Init) or apps/<authn-svc>
//   - SetAccountRepo       ← apps/boot/auth
//   - SetAccountAuthRepo   ← apps/boot/auth
//   - SetTenantRepo        ← apps/boot/tenant
//   - SetUserRepo          ← apps/rbac/user
//   - SetRoleRepo          ← apps/rbac/role
//   - SetOrgRepo           ← apps/rbac/organization
//   - SetPermRepo          ← apps/rbac/permission
type Writer interface {
	SetAuthz(a authz.Authorization)
	SetAccountRepo(r auth.AccountRepository)
	SetAccountAuthRepo(r auth.AccountAuthRepository)
	SetTenantRepo(r tenant.TenantRepository)
	SetUserRepo(r rbac.UserRepository)
	SetRoleRepo(r rbac.RoleRepository)
	SetOrgRepo(r rbac.OrganizationRepository)
	SetPermRepo(r rbac.RoleResourceRepository)
}

// AppContext is the concrete implementation. It is constructed in
// framework/internal/core/boot.Init and passed by pointer to every
// module's Init and Register.
//
// The zero value is NOT useful — call NewAppContext in boot.Init.
type AppContext struct {
	// Infrastructure, set once before module Init.
	db      *pgxpool.Pool
	cache   *redis.Client
	cfg     *config.Config
	session session.SessionManager

	// Cross-module contributions, set by module Init().
	authz_         authz.Authorization
	accountRepo    auth.AccountRepository
	accountAuthR   auth.AccountAuthRepository
	tenantRepo     tenant.TenantRepository
	userRepo       rbac.UserRepository
	roleRepo       rbac.RoleRepository
	orgRepo        rbac.OrganizationRepository
	permRepo       rbac.RoleResourceRepository
}

// NewAppContext constructs the AppContext with the infrastructure slots
// pre-filled. Modules will fill the rest during Init().
//
// Both the db pool and the config must be non-nil. cache and session
// may be nil only if their respective subsystems are intentionally
// disabled at boot (e.g. Redis off + DB session manager).
func NewAppContext(
	db *pgxpool.Pool,
	cache *redis.Client,
	cfg *config.Config,
	session session.SessionManager,
) *AppContext {
	if db == nil {
		panic("NewAppContext: db pool must not be nil")
	}
	if cfg == nil {
		panic("NewAppContext: config must not be nil")
	}
	return &AppContext{
		db:      db,
		cache:   cache,
		cfg:     cfg,
		session: session,
	}
}

// Compile-time assertion that *AppContext satisfies both interfaces.
var (
	_ Reader = (*AppContext)(nil)
	_ Writer = (*AppContext)(nil)
)

// --- Reader ---

func (a *AppContext) DB() *pgxpool.Pool             { return a.db }
func (a *AppContext) Cache() *redis.Client             { return a.cache }
func (a *AppContext) Config() *config.Config        { return a.cfg }
func (a *AppContext) Session() session.SessionManager { return a.session }

func (a *AppContext) Authz() authz.Authorization { return a.authz_ }

func (a *AppContext) AccountRepo() auth.AccountRepository {
	return a.accountRepo
}
func (a *AppContext) AccountAuthRepo() auth.AccountAuthRepository {
	return a.accountAuthR
}
func (a *AppContext) TenantRepo() tenant.TenantRepository {
	return a.tenantRepo
}
func (a *AppContext) UserRepo() rbac.UserRepository     { return a.userRepo }
func (a *AppContext) RoleRepo() rbac.RoleRepository     { return a.roleRepo }
func (a *AppContext) OrgRepo() rbac.OrganizationRepository { return a.orgRepo }
func (a *AppContext) PermRepo() rbac.RoleResourceRepository { return a.permRepo }

// --- Writer ---

func (a *AppContext) SetAuthz(v authz.Authorization)            { a.authz_ = v }
func (a *AppContext) SetAccountRepo(v auth.AccountRepository)  { a.accountRepo = v }
func (a *AppContext) SetAccountAuthRepo(v auth.AccountAuthRepository) { a.accountAuthR = v }
func (a *AppContext) SetTenantRepo(v tenant.TenantRepository)   { a.tenantRepo = v }
func (a *AppContext) SetUserRepo(v rbac.UserRepository)         { a.userRepo = v }
func (a *AppContext) SetRoleRepo(v rbac.RoleRepository)         { a.roleRepo = v }
func (a *AppContext) SetOrgRepo(v rbac.OrganizationRepository)  { a.orgRepo = v }
func (a *AppContext) SetPermRepo(v rbac.RoleResourceRepository) { a.permRepo = v }
