package user

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"

	"gx1727.com/xin/framework/internal/module/asset"
	"gx1727.com/xin/framework/internal/module/role"
	"gx1727.com/xin/framework/pkg/db"
)

type Service struct {
	userRepo UserRepository
	roleRepo role.RoleRepository
	assetSvc *asset.FileService
}

func NewService(userRepo UserRepository, roleRepo role.RoleRepository, assetSvc *asset.FileService) *Service {
	return &Service{
		userRepo: userRepo,
		roleRepo: roleRepo,
		assetSvc: assetSvc,
	}
}

func (s *Service) List(ctx context.Context, tenantID uint, req listRequest) ([]UserInfo, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 {
		req.Size = 20
	}

	var result []UserInfo
	var total int64

	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		users, t, err := s.userRepo.ListScoped(ctx, tenantID, req.Keyword, req.Page, req.Size)
		if err != nil {
			return err
		}
		total = t

		result = make([]UserInfo, len(users))
		for i, u := range users {
			result[i] = UserInfo{
				ID:       u.ID,
				TenantID: u.TenantID,
				Code:     u.Code,
				Nickname: u.Nickname,
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
		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

func (s *Service) Get(ctx context.Context, tenantID, userID uint) (*UserInfo, error) {
	var info *UserInfo
	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		u, err := s.userRepo.GetByIDScoped(ctx, userID)
		if err != nil {
			if errors.Is(err, ErrUserNotFoundDB) {
				return ErrUserNotFound
			}
			return err
		}

		if u.TenantID != tenantID {
			return ErrUserNotFound
		}

		info = &UserInfo{
			ID:       u.ID,
			TenantID: u.TenantID,
			Code:     u.Code,
			Nickname: u.Nickname,
			RealName: u.RealName,
			Avatar:   u.Avatar,
			Phone:    u.Phone,
			Email:    u.Email,
		}

		roles, err := s.roleRepo.GetUserRoles(ctx, u.ID)
		if err == nil && len(roles) > 0 {
			info.Role = roles[0].Code
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return info, nil
}

func (s *Service) UpdateStatus(ctx context.Context, tenantID, userID uint, status int8) error {
	return db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		u, err := s.userRepo.GetByIDScoped(ctx, userID)
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
	})
}

func (s *Service) Profile(ctx context.Context, tenantID, userID uint) (*UserInfo, error) {
	var info *UserInfo
	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		u, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			if errors.Is(err, ErrUserNotFoundDB) {
				return ErrUserNotFound
			}
			return err
		}

		if u.TenantID != tenantID {
			return ErrUserNotFound
		}

		info = &UserInfo{
			ID:        u.ID,
			TenantID:  u.TenantID,
			AccountID: u.AccountID,
			Code:      u.Code,
			Nickname:  u.Nickname,
			RealName:  u.RealName,
			Avatar:    u.Avatar,
			Phone:     u.Phone,
			Email:     u.Email,
		}

		roles, err := s.roleRepo.GetUserRoles(ctx, u.ID)
		if err == nil && len(roles) > 0 {
			info.Role = roles[0].Code
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return info, nil
}

func (s *Service) UploadAvatar(ctx context.Context, tenantID, userID uint, file *multipart.FileHeader) (string, error) {
	var url string
	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		uploadResp, err := s.assetSvc.Upload(ctx, tenantID, userID, file)
		if err != nil {
			return err
		}
		url = uploadResp.URL

		if err := s.userRepo.UpdateAvatar(ctx, userID, url); err != nil {
			return fmt.Errorf("update avatar: %w", err)
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	return url, nil
}

func (s *Service) UpdateProfile(ctx context.Context, tenantID, userID uint, nickname, avatar string) error {
	return db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		u, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			return err
		}
		if u.TenantID != tenantID {
			return ErrUserNotFound
		}

		if err := s.userRepo.UpdateProfile(ctx, userID, nickname, avatar); err != nil {
			return fmt.Errorf("update user profile: %w", err)
		}
		return nil
	})
}
