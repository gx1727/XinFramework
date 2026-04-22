package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gx1727.com/xin/framework/pkg/config"
	jwtpkg "gx1727.com/xin/framework/pkg/jwt"
)

type Service struct {
	db      *gorm.DB
	config  *config.Config
	session SessionManager
}

func NewService(deps Dependencies) *Service {
	return &Service{
		db:      deps.DB,
		config:  deps.Config,
		session: deps.Session,
	}
}

func (s *Service) Login(req loginRequest) (*loginResult, error) {
	identity, err := ResolveLoginIdentity(s.db, req.Account, req.TenantID)
	if err != nil {
		switch {
		case errors.Is(err, ErrBackendUnavailable):
			return nil, ErrBackendUnavailable
		case errors.Is(err, ErrAccountNotFound):
			return nil, ErrInvalidAccountOrPassword
		case errors.Is(err, ErrTenantBindingNotFound):
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

	if s.config == nil || s.session == nil {
		return nil, ErrBackendUnavailable
	}

	sessionID := uuid.NewString()
	if err := s.session.Create(sessionID, identity.UserID, identity.TenantID, identity.RoleCode, time.Duration(s.config.JWT.Expire)*time.Second); err != nil {
		return nil, ErrSessionCreateFailed
	}

	token, err := jwtpkg.Generate(&s.config.JWT, identity.UserID, identity.TenantID, identity.RoleCode, sessionID)
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
	if s.config == nil || s.session == nil {
		return ErrBackendUnavailable
	}
	if sessionID == "" {
		return ErrInvalidToken
	}
	if err := s.session.Revoke(sessionID); err != nil {
		return ErrSessionRevokeFailed
	}
	return nil
}

func (s *Service) Register(req registerRequest) (*registerResult, error) {
	if s.db == nil || s.config == nil || s.session == nil {
		return nil, ErrBackendUnavailable
	}
	d := s.db

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
		Where("id = ?", req.TenantID).
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
		acc := struct {
			ID       uint `gorm:"primaryKey"`
			Phone    string
			Email    string
			Username string
			Password string
			RealName string
		}{
			Phone:    req.Account,
			Email:    req.Account,
			Username: req.Account,
			Password: passwordHash,
			RealName: req.RealName,
		}
		if err := tx.Table("accounts").Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).Create(&acc).Error; err != nil {
			return ErrRegisterFailed
		}
		if acc.ID == 0 {
			return ErrRegisterFailed
		}
		newAccountID = acc.ID

		userCode := uuid.NewString()[:8]
		usr := struct {
			ID        uint `gorm:"primaryKey"`
			TenantID  uint
			AccountID uint
			Code      string
			Status    int
		}{
			TenantID:  req.TenantID,
			AccountID: newAccountID,
			Code:      userCode,
			Status:    1,
		}
		if err := tx.Table("users").Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).Create(&usr).Error; err != nil {
			return ErrRegisterFailed
		}
		if usr.ID == 0 {
			return ErrRegisterFailed
		}
		newUserID = usr.ID
		newUserCode = userCode

		var role struct {
			ID uint
		}
		if err := tx.Table("roles").
			Select("id").
			Where("is_deleted = FALSE").
			Where("tenant_id = ?", req.TenantID).
			Where("is_default = TRUE").
			First(&role).Error; err != nil {
			return ErrDefaultRoleNotFound
		}

		if err := tx.Table("user_roles").Create(&struct {
			TenantID uint
			UserID   uint
			RoleID   uint
		}{
			TenantID: req.TenantID,
			UserID:   newUserID,
			RoleID:   role.ID,
		}).Error; err != nil {
			return ErrRegisterFailed
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	sessionID := uuid.NewString()
	if err := s.session.Create(sessionID, newUserID, req.TenantID, "user", time.Duration(s.config.JWT.Expire)*time.Second); err != nil {
		return nil, ErrSessionCreateFailed
	}

	token, err := jwtpkg.Generate(&s.config.JWT, newUserID, req.TenantID, "user", sessionID)
	if err != nil {
		return nil, ErrGenerateTokenFailed
	}

	res := &registerResult{Token: token}
	res.User.ID = newUserID
	res.User.TenantID = req.TenantID
	res.User.Code = newUserCode
	res.User.Role = "user"
	return res, nil
}
