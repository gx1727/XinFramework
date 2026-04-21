package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gx1727.com/xin/internal/infra/db"
	"gx1727.com/xin/internal/infra/session"
	"gx1727.com/xin/internal/module/user"
	"gx1727.com/xin/pkg/config"
	jwtpkg "gx1727.com/xin/pkg/jwt"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Login(req loginRequest) (*loginResult, error) {
	identity, err := user.ResolveLoginIdentity(req.Account, req.TenantID)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrBackendUnavailable):
			return nil, ErrBackendUnavailable
		case errors.Is(err, user.ErrAccountNotFound):
			return nil, ErrInvalidAccountOrPassword
		case errors.Is(err, user.ErrTenantBindingNotFound):
			return nil, ErrTenantBindingNotFound
		default:
			return nil, ErrInvalidAccountOrPassword
		}
	}

	ok, err := verifyPassword(identity.PasswordHash, req.Password)
	if err != nil || !ok {
		return nil, ErrInvalidAccountOrPassword
	}
	if identity.UserStatus != 1 {
		return nil, ErrUserDisabled
	}

	cfg := config.Get()
	if cfg == nil {
		return nil, ErrBackendUnavailable
	}

	sessionID := uuid.NewString()
	if err := session.Create(sessionID, identity.UserID, identity.TenantID, identity.RoleCode, time.Duration(cfg.JWT.Expire)*time.Second); err != nil {
		return nil, ErrSessionCreateFailed
	}

	token, err := jwtpkg.Generate(&cfg.JWT, identity.UserID, identity.TenantID, identity.RoleCode, sessionID)
	if err != nil {
		return nil, ErrGenerateTokenFailed
	}

	res := &loginResult{Token: token}
	res.User.ID = identity.UserID
	res.User.TenantID = identity.TenantID
	res.User.Code = identity.UserCode
	res.User.Role = identity.RoleCode
	return res, nil
}

func (s *Service) Logout(sessionID string) error {
	cfg := config.Get()
	if cfg == nil {
		return ErrBackendUnavailable
	}
	if sessionID == "" {
		return ErrInvalidToken
	}
	if err := session.Revoke(sessionID); err != nil {
		return ErrSessionRevokeFailed
	}
	return nil
}

func (s *Service) Register(req registerRequest) (*registerResult, error) {
	d := db.Get()
	if d == nil {
		return nil, ErrBackendUnavailable
	}

	cfg := config.Get()
	if cfg == nil {
		return nil, ErrBackendUnavailable
	}

	var count int64
	if err := d.Table("accounts").
		Where("is_deleted = FALSE").
		Where("phone = ? OR email = ?", req.Account, req.Account).
		Count(&count).Error; err != nil {
		return nil, ErrRegisterFailed
	}
	if count > 0 {
		return nil, ErrAccountAlreadyExists
	}

	var tenant struct {
		ID     uint
		Status int16
	}
	if err := d.Table("tenants").
		Select("id, status").
		Where("is_deleted = FALSE").
		Where("id = ?", req.TatID).
		First(&tenant).Error; err != nil {
		return nil, ErrTenantNotFound
	}
	if tenant.Status != 1 {
		return nil, ErrTenantNotFound
	}

	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, ErrRegisterFailed
	}

	var newAccountID uint
	var newUserID uint
	var newUserCode string

	err = d.Transaction(func(tx *gorm.DB) error {
		account := map[string]interface{}{
			"phone":      req.Account,
			"email":      req.Account,
			"password":   passwordHash,
			"real_name":  req.RealName,
			"is_deleted": false,
		}
		if err := tx.Table("accounts").Create(account).Error; err != nil {
			return ErrRegisterFailed
		}
		newAccountID = account["id"].(uint)

		userCode := uuid.NewString()[:8]
		user := map[string]interface{}{
			"tenant_id":  req.TatID,
			"account_id": newAccountID,
			"code":       userCode,
			"real_name":  req.RealName,
			"status":     1,
			"is_deleted": false,
		}
		if err := tx.Table("users").Create(user).Error; err != nil {
			return ErrRegisterFailed
		}
		newUserID = user["id"].(uint)
		newUserCode = userCode

		var role struct {
			ID uint
		}
		if err := tx.Table("roles").
			Select("id").
			Where("is_deleted = FALSE").
			Where("tenant_id = ?", req.TatID).
			Where("is_default = TRUE").
			First(&role).Error; err != nil {
			return ErrDefaultRoleNotFound
		}

		userRole := map[string]interface{}{
			"tenant_id":  req.TatID,
			"user_id":    newUserID,
			"role_id":    role.ID,
			"is_deleted": false,
		}
		if err := tx.Table("user_roles").Create(userRole).Error; err != nil {
			return ErrRegisterFailed
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	sessionID := uuid.NewString()
	if err := session.Create(sessionID, newUserID, req.TatID, "user", time.Duration(cfg.JWT.Expire)*time.Second); err != nil {
		return nil, ErrSessionCreateFailed
	}

	token, err := jwtpkg.Generate(&cfg.JWT, newUserID, req.TatID, "user", sessionID)
	if err != nil {
		return nil, ErrGenerateTokenFailed
	}

	res := &registerResult{Token: token}
	res.User.ID = newUserID
	res.User.TenantID = req.TatID
	res.User.Code = newUserCode
	res.User.Role = "user"
	return res, nil
}
