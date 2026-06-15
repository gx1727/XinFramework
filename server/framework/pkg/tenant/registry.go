// Package tenant exposes the public contract for the tenant module.
//
// Phase 2: apps/boot/tenant registers its TenantRepository factory
// here at process start. framework's internal extapi provider looks
// up the factory through this hook. apps/ imports are not allowed
// from framework/internal/, but this public pkg/ is fair game.
//
// Phase 3 will retire this indirection once tenant's only consumer
// (extapi.Provider) moves out of framework/internal as well.
package tenant

import "context"

// TenantRecord is the minimal tenant row representation consumed by
// extapi.TenantFacade. apps/boot/tenant provides values of this
// shape through its factory registration.
type TenantRecord struct {
	ID        uint
	Code      string
	Name      string
	Status    int16
	Contact   string
	Phone     string
	Email     string
	Province  string
	City      string
	Area      string
	Address   string
	Config    string
	Dashboard string
	CreatedAt interface{ /* time.Time via duck typing */ }
	UpdatedAt interface{}
}

// TenantRepository is the minimal interface extapi needs.
type TenantRepository interface {
	GetByID(ctx context.Context, id uint) (*TenantRecord, error)
}

var globalFactory func() TenantRepository

// Register is called by apps/boot/tenant's init().
func Register(f func() TenantRepository) {
	globalFactory = f
}

// Get returns the registered factory, or nil if tenant is not loaded.
func Get() func() TenantRepository {
	return globalFactory
}