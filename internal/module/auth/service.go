package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
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
