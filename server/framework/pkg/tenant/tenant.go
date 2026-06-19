// Package tenant exposes the public TenantRepository contract used
// by framework's AppContext Reader/Writer and by apps/boot/tenant
// (the producer) and apps/rbac/user + apps/reference/weixin (the
// consumers).
//
// Phase 3 cleanup: the historical Register/Get global variables
// are gone. Modules exchange tenant repositories through the
// AppContext, not via this package.
package tenant

import (
	"context"
	"time"
)

// TenantRecord is the minimal tenant row representation consumed by
// the framework's cross-module providers (cms, extapi). apps/boot/tenant
// publishes values of this shape via its Writer.
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
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TenantRepository is the cross-module tenant data access contract.
type TenantRepository interface {
	GetByID(ctx context.Context, id uint) (*TenantRecord, error)
}
