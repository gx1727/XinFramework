package service

import (
	"context"

	"gx1727.com/xin/framework/pkg/model"
)

type Service struct {
	userRepo    model.UserRepository
	tenantRepo  model.TenantRepository
	cmsPostRepo model.CmsPostRepository
}

func NewService(userRepo model.UserRepository, tenantRepo model.TenantRepository, cmsPostRepo model.CmsPostRepository) *Service {
	return &Service{
		userRepo:    userRepo,
		tenantRepo:  tenantRepo,
		cmsPostRepo: cmsPostRepo,
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

func (s *Service) ListPosts(ctx context.Context, tenantID uint, keyword string, status *int16, page, size int) ([]model.CmsPost, int64, error) {
	return s.cmsPostRepo.List(ctx, tenantID, keyword, status, page, size)
}

func (s *Service) GetPost(ctx context.Context, id uint) (*model.CmsPost, error) {
	return s.cmsPostRepo.GetByID(ctx, id)
}

func (s *Service) CreatePost(ctx context.Context, tenantID uint, title, content string, status int16) (*model.CmsPost, error) {
	return s.cmsPostRepo.Create(ctx, tenantID, title, content, status)
}

func (s *Service) UpdatePost(ctx context.Context, id uint, title, content string, status int16) error {
	return s.cmsPostRepo.Update(ctx, id, title, content, status)
}

func (s *Service) DeletePost(ctx context.Context, id uint) error {
	return s.cmsPostRepo.Delete(ctx, id)
}
