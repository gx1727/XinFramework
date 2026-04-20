package user

import (
	"errors"

	"gx1727.com/xin/internal/infra/db"
)

var (
	ErrBackendUnavailable     = errors.New("backend unavailable")
	ErrAccountNotFound        = errors.New("account not found")
	ErrTenantBindingNotFound  = errors.New("tenant binding not found")
)

// LoginIdentity is the minimum user data required by auth login flow.
type LoginIdentity struct {
	UserID       uint
	TenantID     uint
	UserCode     string
	UserStatus   int16
	RoleCode     string
	PasswordHash string
}

// ResolveLoginIdentity queries account, tenant binding, and role for login.
func ResolveLoginIdentity(account string, tenantID uint) (*LoginIdentity, error) {
	d := db.Get()
	if d == nil {
		return nil, ErrBackendUnavailable
	}

	var acc struct {
		ID       uint
		Password string
	}
	if err := d.Table("accounts").
		Select("id, password").
		Where("is_deleted = FALSE").
		Where("username = ? OR phone = ? OR email = ?", account, account, account).
		First(&acc).Error; err != nil {
		return nil, ErrAccountNotFound
	}

	q := d.Table("users").
		Select("id, tenant_id, code, status").
		Where("is_deleted = FALSE").
		Where("account_id = ?", acc.ID)
	if tenantID > 0 {
		q = q.Where("tenant_id = ?", tenantID)
	}

	var u struct {
		ID       uint
		TenantID uint
		Code     string
		Status   int16
	}
	if err := q.Order("id ASC").First(&u).Error; err != nil {
		return nil, ErrTenantBindingNotFound
	}

	roleCode := "user"
	var role struct {
		Code string
	}
	_ = d.Table("user_roles ur").
		Select("r.code").
		Joins("JOIN roles r ON r.id = ur.role_id").
		Where("ur.is_deleted = FALSE").
		Where("r.is_deleted = FALSE").
		Where("ur.user_id = ?", u.ID).
		Order("ur.id ASC").
		First(&role).Error
	if role.Code != "" {
		roleCode = role.Code
	}

	return &LoginIdentity{
		UserID:       u.ID,
		TenantID:     u.TenantID,
		UserCode:     u.Code,
		UserStatus:   u.Status,
		RoleCode:     roleCode,
		PasswordHash: acc.Password,
	}, nil
}
