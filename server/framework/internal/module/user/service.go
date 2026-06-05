package user

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"

	"gx1727.com/xin/framework/internal/module/asset"
	"gx1727.com/xin/framework/internal/module/auth"
	"gx1727.com/xin/framework/internal/module/role"
	"gx1727.com/xin/framework/pkg/db"
)

type Service struct {
	userRepo    UserRepository
	roleRepo    role.RoleRepository
	assetSvc    *asset.FileService
	accountRepo auth.AccountRepository
}

func NewService(userRepo UserRepository, roleRepo role.RoleRepository, assetSvc *asset.FileService, accountRepo auth.AccountRepository) *Service {
	return &Service{
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		assetSvc:    assetSvc,
		accountRepo: accountRepo,
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
				Status:   u.Status,
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
			Status:   u.Status,
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

// Update 全量更新；会校验租户归属
func (s *Service) Update(ctx context.Context, tenantID, userID uint, req updateUserRequest) (*UserInfo, error) {
	var info *UserInfo
	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		u, err := s.userRepo.Update(ctx, userID, UpdateUserRepoReq{
			Nickname: req.Nickname,
			RealName: req.RealName,
			Avatar:   req.Avatar,
			Status:   req.Status,
		})
		if err != nil {
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
			Status:    u.Status,
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

// Patch 局部更新；nil 字段保持原值。空 body 等价于 GET
func (s *Service) Patch(ctx context.Context, tenantID, userID uint, req patchUserRequest) (*UserInfo, error) {
	var info *UserInfo
	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		u, err := s.userRepo.Patch(ctx, userID, PatchUserRepoReq{
			Nickname: req.Nickname,
			RealName: req.RealName,
			Avatar:   req.Avatar,
			Status:   req.Status,
		})
		if err != nil {
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
			Status:    u.Status,
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
			Status:    u.Status,
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

func (s *Service) Create(ctx context.Context, tenantID, creatorID uint, req createRequest) (*createResponse, error) {
	var (
		newUserID   uint
		newUserCode string
		newAccount  *auth.Account
	)

	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		if s.accountRepo != nil {
			exists, err := s.accountRepo.Exists(ctx, req.Username)
			if err != nil {
				return fmt.Errorf("check account exists: %w", err)
			}
			if exists {
				return ErrUserAlreadyExists
			}
		}

		status := req.Status
		if status == 0 {
			status = 1
		}

		passwordHash, err := auth.HashPassword(req.Password)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}

		newAccount, err = s.accountRepo.Create(ctx, req.Username, req.Phone, req.Email, req.RealName, passwordHash)
		if err != nil {
			return fmt.Errorf("create account: %w", err)
		}

		querier, err := db.GetQuerier(ctx)
		if err != nil {
			return fmt.Errorf("get querier: %w", err)
		}

		err = querier.QueryRow(ctx, `
			INSERT INTO users (tenant_id, account_id, code, status, created_by)
			VALUES ($1, $2, '', $3, $4)
			RETURNING id`, tenantID, newAccount.ID, status, creatorID).Scan(&newUserID)
		if err != nil {
			return fmt.Errorf("create user: %w", err)
		}

		var roleID uint
		err = querier.QueryRow(ctx, `
			SELECT id FROM roles
			WHERE is_deleted = FALSE AND tenant_id = $1 AND is_default = TRUE
			LIMIT 1`, tenantID).Scan(&roleID)
		if err != nil {
			return ErrDefaultRoleNotFound
		}

		_, err = querier.Exec(ctx, `
			INSERT INTO user_roles (tenant_id, user_id, role_id)
			VALUES ($1, $2, $3)`, tenantID, newUserID, roleID)
		if err != nil {
			return fmt.Errorf("assign default role: %w", err)
		}

		newUserCode = fmt.Sprintf("U%07d", newUserID)

		_, err = querier.Exec(ctx, `UPDATE users SET code = $1 WHERE id = $2`, newUserCode, newUserID)
		if err != nil {
			return fmt.Errorf("update user code: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &createResponse{
		ID:       newUserID,
		TenantID: tenantID,
		Code:     newUserCode,
		Username: newAccount.Username,
		RealName: newAccount.RealName,
		Phone:    newAccount.Phone,
		Status:   newAccount.Status,
	}, nil
}
