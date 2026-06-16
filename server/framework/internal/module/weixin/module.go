package weixin

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	pkgauth "gx1727.com/xin/framework/pkg/auth"
	pkgrbac "gx1727.com/xin/framework/pkg/rbac"
	pkgtenant "gx1727.com/xin/framework/pkg/tenant"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/session"
)

func init() {
	plugin.Register(Module())
}

// Module returns the weixin module.
//
// Phase 3 status: weixin still lives in framework/internal/module/.
// Its dependencies — auth + tenant + user + role — have all moved to
// apps/boot/ and apps/rbac/. weixin accesses them through the public
// pkg hooks (pkgauth, pkgtenant, pkgrbac), so no direct apps/ imports.
//
// Phase 3b will move weixin itself to apps/reference/weixin/ and
// remove the noop fallbacks below.
func Module() plugin.Module {
	return plugin.NewModuleWithOpts("weixin",
		func(public *gin.RouterGroup, protected *gin.RouterGroup) {
			svc := NewService(
				db.Get(),
				session.Manager(),
				noopAccountAuthAdapter{},
				noopAccountAdapter{},
				noopTenantAdapter{},
				resolveRoleRepository(),
				resolveUserRepository(),
			)
			h := NewHandler(svc)
			Register(public, protected, h)
		},
		plugin.WithInit(func() error {
			return InitConfig()
		}),
	)
}

// resolveRoleRepository returns the apps/rbac/role RoleRepository
// factory if registered, otherwise a noop. Used at module init time.
func resolveRoleRepository() pkgrbac.RoleRepository {
	if f := pkgrbac.GetRoleRepository(); f != nil {
		return f()
	}
	return nil
}

// resolveUserRepository returns the apps/rbac/user UserRepository
// factory if registered, otherwise a noop.
func resolveUserRepository() pkgrbac.UserRepository {
	if f := pkgrbac.GetUserRepository(); f != nil {
		return f()
	}
	return nil
}

// Phase 2 stop-gap adapters: real impl lives in apps/boot/{auth,tenant}.
// These noops satisfy the pkgauth / pkgtenant interfaces so weixin's
// Service compiles. The actual data access lives outside of weixin,
// routed through the framework's pkgauth / pkgtenant hooks as needed.

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

type noopTenantAdapter struct{}

func (noopTenantAdapter) GetByID(ctx context.Context, id uint) (*pkgtenant.TenantRecord, error) {
	return nil, errWeixinDepNotLoaded
}

var errWeixinDepNotLoaded = errors.New("weixin dependency not loaded — register apps/boot/{auth,tenant} and apps/rbac/{user,role} via main.go side-effect import")