package service

import (
	"context"

	"gx1727.com/xin/framework/pkg/model"
)

type Service struct {
	userRepo   model.UserRepository
	tenantRepo model.TenantRepository
}

func NewService(userRepo model.UserRepository, tenantRepo model.TenantRepository) *Service {
	return &Service{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
	}
}

func (s *Service) GetUser(ctx context.Context, userID uint) (*model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *Service) ListUsers(ctx context.Context, tenantID uint, keyword string, page, pageSize int) ([]model.User, int64, error) {
	return s.userRepo.List(ctx, tenantID, keyword, page, pageSize)
}

func (s *Service) GetTenant(ctx context.Context, tenantID uint) (*model.Tenant, error) {
	return s.tenantRepo.GetByID(ctx, tenantID)
}
