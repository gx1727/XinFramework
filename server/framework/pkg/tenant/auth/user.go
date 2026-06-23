// Package auth exposes the public contracts that business modules use
// to interact with the tenant RBAC suite (user / role / permission /
// menu / resource / organization).
//
// weixin module (still in framework/internal) and any future
// framework-internal consumer must depend only on this pkg, not on
// apps/. The concrete implementations (PostgresUserRepository,
// PostgresRoleRepository, ...) live in apps/tenant/<name>/.
//
// globals are gone. Modules exchange repositories through the
package auth

import (
	"context"
	"encoding/json"
	"time"

	"gx1727.com/xin/framework/pkg/identity"
)

// User is the cross-module user representation. apps/tenant/user aliases
// its local User struct to this type so the rest of the system sees
// one canonical definition.
//
// Struct composition:
//   - identity.User (10 fields): the cross-domain base (ID, AccountID,
//     OrgID, Code, RealName, Nickname, Avatar, Status, CreatedAt,
//     UpdatedAt). Locked in place by framework/pkg/identity/identity_test.go.
//   - 4 tenant-domain fields: TenantID, Phone, Email, OrgName.
//
// JSON output is byte-level compatible with the legacy pkgrbac.User
// (pre-0023.3) via the custom MarshalJSON below. apps/tenant/user/model_test.go
// golden JSON test pins this contract.
type User struct {
	identity.User
	TenantID uint   `json:"tenant_id"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	OrgName  string `json:"org_name,omitempty"`
}

// MarshalJSON serializes User in the legacy field order for byte-level
// compatibility with downstream JSON consumers (cms/handler, weixin/service,
// extapi.User, etc.). The structural change (embed + MarshalJSON) is
// transparent to them.
//
// Field order (must stay aligned with the golden JSON in
// apps/tenant/user/model_test.go):
//
//	id, tenant_id, account_id, org_id, org_name, code, nickname,
//	real_name, avatar, phone, email, status, created_at, updated_at
//
// Why not just rely on Go embed order? Go's encoding/json uses the
// host struct's field order: identity.User's 10 fields first, then
// the 4 extension fields, yielding a different order than the legacy
// 14-field struct. Custom MarshalJSON forces the legacy order.
func (u User) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
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
	}{
		ID:        u.ID,
		TenantID:  u.TenantID,
		AccountID: u.AccountID,
		OrgID:     u.OrgID,
		OrgName:   u.OrgName,
		Code:      u.Code,
		Nickname:  u.Nickname,
		RealName:  u.RealName,
		Avatar:    u.Avatar,
		Phone:     u.Phone,
		Email:     u.Email,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	})
}

// UserRepository is the subset of user data access that other
// framework-internal modules (notably weixin) need. The concrete
// implementation in apps/tenant/user/ satisfies this interface
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
