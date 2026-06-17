// Package rbac exposes the public contracts that business modules use
// to interact with the RBAC suite (user / role / permission / menu /
// resource / organization).
//
// Phase 3 rationale: the 6 RBAC modules have moved from
// framework/internal/module/* to apps/rbac/*. framework's own
// weixin module (still in framework/internal) and any future
// framework-internal consumer must depend only on this pkg, not on
// apps/. The concrete implementations (PostgresUserRepository,
// PostgresRoleRepository, …) live in apps/rbac/<name>/.
//
// Phase 4 cleanup: the historical RegisterUserRepository / GetUserRepository
// globals are gone. Modules exchange repositories through the
// AppContext (plugin.Reader / plugin.Writer).
package rbac

import (
	"context"
	"time"
)

// User is the cross-module user representation. apps/rbac/user aliases
// its local User struct to this type so the rest of the system sees
// one canonical definition.
type User struct {
	ID         uint      `json:"id"`
	TenantID   uint      `json:"tenant_id"`
	AccountID  uint      `json:"account_id"`
	OrgID      *uint     `json:"org_id"`
	OrgName    string    `json:"org_name,omitempty"`
	Code       string    `json:"code"`
	Nickname   string    `json:"nickname"`
	RealName   string    `json:"real_name"`
	Avatar     string    `json:"avatar"`
	Phone      string    `json:"phone"`
	Email      string    `json:"email"`
	Status     int8      `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// UserRepository is the subset of user data access that other
// framework-internal modules (notably weixin) need. The concrete
// implementation in apps/rbac/user/ satisfies this interface
// implicitly because field types are identical.
type UserRepository interface {
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByAccount(ctx context.Context, tenantID, accountID uint) (*User, error)
	GetByCode(ctx context.Context, tenantID uint, code string) (*User, error)
	List(ctx context.Context, tenantID uint, keyword string, page, size int) ([]User, int64, error)
	UpdateStatus(ctx context.Context, id uint, status int8) error
	UpdatePhone(ctx context.Context, id uint, phone string) error
}

// UserService is the optional business-level abstraction. Currently
// unused cross-module; reserved for Phase 4+ use cases where
// non-RBAC apps need user operations beyond raw CRUD.
type UserService interface {
	GetByID(ctx context.Context, tenantID, id uint) (*User, error)
}