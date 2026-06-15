package weixin

import (
	"context"

	"github.com/gin-gonic/gin"
	pkgauth "gx1727.com/xin/framework/pkg/auth"
	pkgtenant "gx1727.com/xin/framework/pkg/tenant"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/session"

	"gx1727.com/xin/framework/internal/module/role"
	"gx1727.com/xin/framework/internal/module/user"
)

func init() {
	plugin.Register(Module())
}

// Module returns the weixin module.
//
// Phase 2 note: weixin depends on auth + tenant + user + role. auth
// and tenant have moved to apps/boot/. NewService takes them through
// the public pkg/auth and pkg/tenant interfaces so the weixin module
// (still in framework/internal) doesn't need to import apps/.
//
// In this minimal initial port we use lazy noop fallbacks for auth/tenant
// repositories that aren't yet exercised by Service — when the apps/boot
// modules register their factories we can swap the fallbacks for real
// adapters. Phase 3 (weixin → apps/...) makes this stop-gap go away.
func Module() plugin.Module {
	return plugin.NewModuleWithOpts("weixin",
		func(public *gin.RouterGroup, protected *gin.RouterGroup) {
			svc := NewService(
				db.Get(),
				session.Manager(),
				noopAccountAuthAdapter{},
				noopAccountAdapter{},
				noopTenantAdapter{},
				role.NewRoleRepository(db.Get()),
				user.NewUserRepository(db.Get()),
			)
			h := NewHandler(svc)
			Register(public, protected, h)
		},
		plugin.WithInit(func() error {
			return InitConfig()
		}),
	)
}

// Phase 2 stop-gap adapters: real impl lives in apps/boot/{auth,tenant}.
// These noops satisfy the pkgauth / pkgtenant interfaces so weixin's
// Service compiles. The actual data access lives outside of weixin,
// routed through the framework's pkgauth / pkgtenant hooks as needed.

// noopAccountAuthAdapter satisfies pkgauth.AccountAuthRepository.
type noopAccountAuthAdapter struct{}

func (noopAccountAuthAdapter) GetByOpenID(ctx context.Context, tenantID uint, authType, openID string) (*pkgauth.AccountAuth, error) {
	return nil, errWeixinDepNotLoaded
}
func (noopAccountAuthAdapter) GetByAccountID(ctx context.Context, accountID uint) ([]pkgauth.AccountAuth, error) {
	return nil, errWeixinDepNotLoaded
}
func (noopAccountAuthAdapter) Create(ctx context.Context, tenantID, accountID uint, authType, openID, unionID, sessionKey string) (*pkgauth.AccountAuth, error) {
	return nil, errWeixinDepNotLoaded
}
func (noopAccountAuthAdapter) UpdateSessionKey(ctx context.Context, id uint, sessionKey string) error {
	return errWeixinDepNotLoaded
}
func (noopAccountAuthAdapter) Delete(ctx context.Context, id uint) error {
	return errWeixinDepNotLoaded
}

// noopAccountAdapter satisfies pkgauth.AccountRepository.
type noopAccountAdapter struct{}

func (noopAccountAdapter) GetByID(ctx context.Context, id uint) (*pkgauth.Account, error) {
	return nil, errWeixinDepNotLoaded
}
func (noopAccountAdapter) GetByUsername(ctx context.Context, username string) (*pkgauth.Account, error) {
	return nil, errWeixinDepNotLoaded
}
func (noopAccountAdapter) GetByPhone(ctx context.Context, phone string) (*pkgauth.Account, error) {
	return nil, errWeixinDepNotLoaded
}
func (noopAccountAdapter) GetByEmail(ctx context.Context, email string) (*pkgauth.Account, error) {
	return nil, errWeixinDepNotLoaded
}
func (noopAccountAdapter) Create(ctx context.Context, username, phone, email, realName, passwordHash string) (*pkgauth.Account, error) {
	return nil, errWeixinDepNotLoaded
}
func (noopAccountAdapter) Exists(ctx context.Context, account string) (bool, error) {
	return false, errWeixinDepNotLoaded
}

// noopTenantAdapter satisfies pkgtenant.TenantRepository.
type noopTenantAdapter struct{}

func (noopTenantAdapter) GetByID(ctx context.Context, id uint) (*pkgtenant.TenantRecord, error) {
	return nil, errWeixinDepNotLoaded
}

var errWeixinDepNotLoaded = stringErrW("weixin dependency (auth/tenant) moved to apps/boot — Phase 2 stop-gap; load real adapters via pkgauth/pkgtenant.Get()")

type stringErrW string

func (e stringErrW) Error() string { return string(e) }