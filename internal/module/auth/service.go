package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gx1727.com/xin/internal/infra/db"
	"gx1727.com/xin/internal/infra/session"
	"gx1727.com/xin/pkg/config"
	jwtpkg "gx1727.com/xin/pkg/jwt"
)

var (
	ErrInvalidAccountOrPassword = errors.New("invalid account or password")
	ErrTenantBindingNotFound    = errors.New("user is not bound to any tenant")
	ErrUserDisabled             = errors.New("user is disabled")
	ErrInvalidToken             = errors.New("invalid token")
	ErrBackendUnavailable       = errors.New("backend unavailable")
	ErrSessionCreateFailed      = errors.New("create session failed")
	ErrSessionRevokeFailed      = errors.New("revoke session failed")
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Login(req loginRequest) (*loginResult, error) {
	d := db.Get()
	if d == nil {
		return nil, ErrBackendUnavailable
	}

	var acc accountRow
	if err := d.Table("accounts").
		Select("id, username, phone, email, password").
		Where("is_deleted = FALSE").
		Where("username = ? OR phone = ? OR email = ?", req.Account, req.Account, req.Account).
		First(&acc).Error; err != nil {
		return nil, ErrInvalidAccountOrPassword
	}

	ok, err := verifyPassword(acc.Password, req.Password)
	if err != nil || !ok {
		return nil, ErrInvalidAccountOrPassword
	}

	q := d.Table("users").
		Select("id, tenant_id, code, status").
		Where("is_deleted = FALSE").
		Where("account_id = ?", acc.ID)
	if req.TenantID > 0 {
		q = q.Where("tenant_id = ?", req.TenantID)
	}

	var u userRow
	if err := q.Order("id ASC").First(&u).Error; err != nil {
		return nil, ErrTenantBindingNotFound
	}
	if u.Status != 1 {
		return nil, ErrUserDisabled
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

	cfg := config.Get()
	if cfg == nil {
		return nil, ErrBackendUnavailable
	}

	sessionID := uuid.NewString()
	if err := session.Create(sessionID, u.ID, u.TenantID, roleCode, time.Duration(cfg.JWT.Expire)*time.Second); err != nil {
		return nil, ErrSessionCreateFailed
	}

	token, err := jwtpkg.Generate(&cfg.JWT, u.ID, u.TenantID, roleCode, sessionID)
	if err != nil {
		return nil, err
	}

	res := &loginResult{Token: token}
	res.User.ID = u.ID
	res.User.TenantID = u.TenantID
	res.User.Code = u.Code
	res.User.Role = roleCode
	return res, nil
}

func (s *Service) Logout(authHeader string) error {
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	cfg := config.Get()
	if cfg == nil {
		return ErrBackendUnavailable
	}

	claims := &jwtpkg.Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWT.Secret), nil
	})
	if err != nil || !token.Valid {
		return ErrInvalidToken
	}

	if err := session.Revoke(claims.SessionID); err != nil {
		return ErrSessionRevokeFailed
	}
	return nil
}
