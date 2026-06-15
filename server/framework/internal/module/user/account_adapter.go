package user

import (
	"context"

	pkgauth "gx1727.com/xin/framework/pkg/auth"
)

// localAccountAdapter bridges from user (which lives in
// framework/internal) to the apps/boot/auth-side AccountRepository.
//
// Phase 2 setup:
//   - apps/boot/auth's init() calls pkgauth.Register(NewAccountRepository)
//   - user/module.go constructs a localAccountAdapter and asks
//     pkgauth.Get() for the factory
//   - Once user also moves to apps/rbac/user/ (Phase 3), this file
//     goes away entirely.
//
// Until Phase 3 lands, if auth is not loaded (e.g. user-only build)
// the adapter falls back to noopAccountAdapter so user still loads.
type localAccountAdapter struct {
	factory func() pkgauth.AccountRepository
}

func newLocalAccountAdapter(_ interface{}) pkgauth.AccountRepository {
	if f := pkgauth.Get(); f != nil {
		return f()
	}
	return noopAccountAdapter{}
}

// noopAccountAdapter is a safe fallback when auth is not loaded —
// account operations return errAccountNotLoaded so the caller sees
// a clean error instead of a nil-deref panic.
type noopAccountAdapter struct{}

func (noopAccountAdapter) GetByID(ctx context.Context, id uint) (*pkgauth.Account, error) {
	return nil, errAccountNotLoaded
}
func (noopAccountAdapter) GetByUsername(ctx context.Context, username string) (*pkgauth.Account, error) {
	return nil, errAccountNotLoaded
}
func (noopAccountAdapter) GetByPhone(ctx context.Context, phone string) (*pkgauth.Account, error) {
	return nil, errAccountNotLoaded
}
func (noopAccountAdapter) GetByEmail(ctx context.Context, email string) (*pkgauth.Account, error) {
	return nil, errAccountNotLoaded
}
func (noopAccountAdapter) Create(ctx context.Context, username, phone, email, realName, passwordHash string) (*pkgauth.Account, error) {
	return nil, errAccountNotLoaded
}
func (noopAccountAdapter) Exists(ctx context.Context, account string) (bool, error) {
	return false, errAccountNotLoaded
}

var errAccountNotLoaded = stringError("auth module not loaded — register apps/boot/auth in main.go")

type stringError string

func (e stringError) Error() string { return string(e) }