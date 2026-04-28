package service

import (
	"context"
	"time"

	"gx1727.com/xin/framework/pkg/model"
)

type Service struct {
	userRepo   model.UserRepository
	tenantRepo model.TenantRepository
	postRepo   model.CmsPostRepository
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) SetRepos(userRepo model.UserRepository, tenantRepo model.TenantRepository, postRepo model.CmsPostRepository) {
	s.userRepo = userRepo
	s.tenantRepo = tenantRepo
	s.postRepo = postRepo
}

func (s *Service) GetUser(ctx context.Context, userID uint) (*model.User, error) {
	if s.userRepo == nil {
		return &model.User{ID: userID, Code: "demo", RealName: "Demo User"}, nil
	}
	return s.userRepo.GetByID(ctx, userID)
}

func (s *Service) ListUsers(ctx context.Context, tenantID uint, keyword string, page, size int) ([]model.User, int64, error) {
	if s.userRepo == nil {
		return []model.User{}, 0, nil
	}
	return s.userRepo.List(ctx, tenantID, keyword, page, size)
}

func (s *Service) GetTenant(ctx context.Context, tenantID uint) (*model.Tenant, error) {
	if s.tenantRepo == nil {
		return &model.Tenant{ID: tenantID, Code: "demo", Name: "Demo Tenant"}, nil
	}
	return s.tenantRepo.GetByID(ctx, tenantID)
}

func (s *Service) ListPosts(ctx context.Context, tenantID uint, keyword string, status *int16, page, size int) ([]model.CmsPost, int64, error) {
	if s.postRepo == nil {
		return []model.CmsPost{}, 0, nil
	}
	return s.postRepo.List(ctx, tenantID, keyword, status, page, size)
}

func (s *Service) GetPost(ctx context.Context, id uint) (*model.CmsPost, error) {
	if s.postRepo == nil {
		return &model.CmsPost{ID: id, Title: "Demo Post", Content: "Demo Content", Status: 1, CreatedAt: time.Now()}, nil
	}
	return s.postRepo.GetByID(ctx, id)
}

func (s *Service) CreatePost(ctx context.Context, tenantID uint, title, content string, status int16) (*model.CmsPost, error) {
	if s.postRepo == nil {
		return &model.CmsPost{
			ID:        1,
			TenantID:  tenantID,
			Title:     title,
			Content:   content,
			Status:    status,
			CreatedAt: time.Now(),
		}, nil
	}
	return s.postRepo.Create(ctx, tenantID, title, content, status)
}

func (s *Service) UpdatePost(ctx context.Context, id uint, title, content string, status int16) error {
	if s.postRepo == nil {
		return nil
	}
	return s.postRepo.Update(ctx, id, title, content, status)
}

func (s *Service) DeletePost(ctx context.Context, id uint) error {
	if s.postRepo == nil {
		return nil
	}
	return s.postRepo.Delete(ctx, id)
}
