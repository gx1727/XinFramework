package user

import (
	"context"
	"errors"
	"fmt"

	"gx1727.com/xin/framework/pkg/model"
)

type Service struct {
	userRepo model.UserRepository
	roleRepo model.RoleRepository
}

func NewService(userRepo model.UserRepository, roleRepo model.RoleRepository) *Service {
	return &Service{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

func (s *Service) List(ctx context.Context, tenantID uint, req listRequest) ([]UserInfo, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 {
		req.Size = 20
	}

	users, total, err := s.userRepo.List(ctx, tenantID, req.Keyword, req.Page, req.Size)
	if err != nil {
		return nil, 0, err
	}

	result := make([]UserInfo, len(users))
	for i, u := range users {
		result[i] = UserInfo{
			ID:       u.ID,
			TenantID: u.TenantID,
			Code:     u.Code,
			Status:   u.Status,
			RealName: u.RealName,
			Avatar:   u.Avatar,
			Phone:    u.Phone,
			Email:    u.Email,
		}

		roles, err := s.roleRepo.GetUserRoles(ctx, u.ID)
		if err == nil && len(roles) > 0 {
			result[i].Role = roles[0].Code
		}
	}

	return result, total, nil
}

func (s *Service) Get(ctx context.Context, tenantID, userID uint) (*UserInfo, error) {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if u.TenantID != tenantID {
		return nil, ErrUserNotFound
	}

	info := &UserInfo{
		ID:       u.ID,
		TenantID: u.TenantID,
		Code:     u.Code,
		Status:   u.Status,
		RealName: u.RealName,
		Avatar:   u.Avatar,
		Phone:    u.Phone,
		Email:    u.Email,
	}

	roles, err := s.roleRepo.GetUserRoles(ctx, u.ID)
	if err == nil && len(roles) > 0 {
		info.Role = roles[0].Code
	}

	return info, nil
}

func (s *Service) UpdateStatus(ctx context.Context, tenantID, userID uint, status int8) error {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if u.TenantID != tenantID {
		return ErrUserNotFound
	}

	if err := s.userRepo.UpdateStatus(ctx, userID, status); err != nil {
		return fmt.Errorf("update user status: %w", err)
	}
	return nil
}
