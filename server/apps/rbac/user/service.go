package user

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"

	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/apps/rbac/organization"
	"gx1727.com/xin/apps/rbac/role"
	"gx1727.com/xin/apps/reference/asset"
	pkgauth "gx1727.com/xin/framework/pkg/auth"
	"gx1727.com/xin/framework/pkg/audit"
	"gx1727.com/xin/framework/pkg/db"
)

type Service struct {
	pool       *pgxpool.Pool
	userRepo   UserRepository
	roleRepo   role.RoleRepository
	orgRepo    organization.OrganizationRepository
	assetSvc   *asset.FileService
	accountRepo pkgauth.AccountRepository
}

// toUserInfo 把 User 转换成返回给前端的 UserInfo。
//
// 不设置 Role 字段——Role 由调用方通过 roleRepo.GetUserRoles 单独查询后赋值，
// 这里保持单一职责。
//
// 接收指针以便直接消费 repo.GetByXxx 返回的 *User；nil 时返回零值，
// 避免在 UpdateOrg / Patch 等"读后写"路径上漏 nil check 导致 panic。
//
// Phase 6 抽出：之前 Get / List / Update / Patch / Profile / UpdateOrg 6 处
// 重复构造 UserInfo 字面量，且 UpdateOrg 漏写 AccountID 字段（前端可见 bug）。
// 统一通过此 helper 保证字段一致。
func toUserInfo(u *User) UserInfo {
	if u == nil {
		return UserInfo{}
	}
	return UserInfo{
		ID:        u.ID,
		TenantID:  u.TenantID,
		AccountID: u.AccountID,
		OrgID:     u.OrgID,
		OrgName:   u.OrgName,
		Code:      u.Code,
		Nickname:  u.Nickname,
		RealName:  u.RealName,
		Avatar:    u.Avatar,
		Phone:     u.Phone,
		Email:     u.Email,
		Status:    u.Status,
	}
}

// validateOrg 校验主组织 ID 是否存在且与租户一致。nil 表示允许"未指定"。
// 与 0 等价的请求也允许（表示移出主组织），但 0 不需要查数据库。
func (s *Service) validateOrg(ctx context.Context, tenantID uint, orgID *uint) error {
	if orgID == nil || *orgID == 0 {
		return nil
	}
	if s.orgRepo == nil {
		return errors.New("organization repository not wired")
	}
	org, err := s.orgRepo.GetByIDScoped(ctx, *orgID)
	if err != nil {
		return ErrOrgNotFound
	}
	if org.TenantID != tenantID {
		return ErrOrgNotFound
	}
	if org.Status != 1 {
		return ErrOrgNotFound
	}
	return nil
}

// UpdateOrg 调整用户的主组织（orgID=nil 或 0 表示移出组织）。
func (s *Service) UpdateOrg(ctx context.Context, tenantID, userID uint, orgID *uint) (*UserInfo, error) {
	var info *UserInfo
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		// 校验目标存在 + 同租户
		// 先读出旧 org_id，写审计要用
		oldUser, err := s.userRepo.GetByIDScoped(ctx, userID)
		if err != nil {
			if errors.Is(err, ErrUserNotFoundDB) {
				return ErrUserNotFound
			}
			return err
		}
		oldOrgID := uint(0)
		if oldUser.OrgID != nil {
			oldOrgID = *oldUser.OrgID
		}
		oldOrgName := oldUser.OrgName

		if err := s.validateOrg(ctx, tenantID, orgID); err != nil {
			return err
		}
		u, err := s.userRepo.UpdateOrg(ctx, userID, orgID)
		if err != nil {
			if errors.Is(err, ErrUserNotFoundDB) {
				return ErrUserNotFound
			}
			return err
		}
		if u.TenantID != tenantID {
			return ErrUserNotFound
		}
		info = &UserInfo{}
		*info = toUserInfo(u)
		roles, err := s.roleRepo.GetUserRoles(ctx, u.ID)
		if err == nil && len(roles) > 0 {
			info.Role = roles[0].Code
		}

		// 审计：组织变更（从 oldOrgID 改到新 orgID）
		newOrgID := uint(0)
		if u.OrgID != nil {
			newOrgID = *u.OrgID
		}
		if oldOrgID != newOrgID {
			audit.Log(ctx, s.pool, audit.Entry{
				TenantID:  u.TenantID,
				Action:    "user:org_change",
				TableName: "users",
				RecordID:  u.ID,
				OldData: map[string]any{
					"org_id":   oldOrgID,
					"org_name": oldOrgName,
				},
				NewData: map[string]any{
					"org_id":   newOrgID,
					"org_name": u.OrgName,
				},
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return info, nil
}

func NewService(
	pool *pgxpool.Pool,
	userRepo UserRepository,
	roleRepo role.RoleRepository,
	orgRepo organization.OrganizationRepository,
	assetSvc *asset.FileService,
	accountRepo pkgauth.AccountRepository,
) *Service {
	return &Service{
		pool:        pool,
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		orgRepo:     orgRepo,
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

	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		users, t, err := s.userRepo.ListScoped(ctx, tenantID, req.Keyword, req.OrgID, req.Page, req.Size)
		if err != nil {
			return err
		}
		total = t

		result = make([]UserInfo, len(users))
		for i, u := range users {
			result[i] = toUserInfo(&u)

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
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
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

		info = &UserInfo{}
		*info = toUserInfo(u)

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
	return db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
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
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		if err := s.validateOrg(ctx, tenantID, req.OrgID); err != nil {
			return err
		}
		u, err := s.userRepo.Update(ctx, userID, UpdateUserRepoReq{
			Nickname: req.Nickname,
			RealName: req.RealName,
			Avatar:   req.Avatar,
			Status:   req.Status,
			OrgID:    req.OrgID,
		})
		if err != nil {
			return err
		}
		if u.TenantID != tenantID {
			return ErrUserNotFound
		}

		info = &UserInfo{}
		*info = toUserInfo(u)
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
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		if err := s.validateOrg(ctx, tenantID, req.OrgID); err != nil {
			return err
		}
		u, err := s.userRepo.Patch(ctx, userID, PatchUserRepoReq{
			Nickname: req.Nickname,
			RealName: req.RealName,
			Avatar:   req.Avatar,
			Status:   req.Status,
			OrgID:    req.OrgID,
		})
		if err != nil {
			return err
		}
		if u.TenantID != tenantID {
			return ErrUserNotFound
		}

		info = &UserInfo{}
		*info = toUserInfo(u)
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
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
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

		info = &UserInfo{}
		*info = toUserInfo(u)

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
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
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
	return db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
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
		newAccount  *pkgauth.Account
	)

	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
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

		passwordHash, err := pkgauth.HashPassword(req.Password)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}

		newAccount, err = s.accountRepo.Create(ctx, req.Username, req.Phone, req.Email, req.RealName, passwordHash)
		if err != nil {
			return fmt.Errorf("create account: %w", err)
		}

		querier, err := db.GetQuerier(ctx, s.pool)
		if err != nil {
			return fmt.Errorf("get querier: %w", err)
		}

		if err := s.validateOrg(ctx, tenantID, req.OrgID); err != nil {
			return err
		}
		var orgIDArg interface{}
		if req.OrgID != nil {
			orgIDArg = *req.OrgID
		}
		err = querier.QueryRow(ctx, `
			INSERT INTO users (tenant_id, account_id, code, status, org_id, created_by)
			VALUES ($1, $2, '', $3, $4, $5)
			RETURNING id`, tenantID, newAccount.ID, status, orgIDArg, creatorID).Scan(&newUserID)
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

	// 补一次 org_name（不入主路径，写在事务外）
	var orgName string
	if req.OrgID != nil && *req.OrgID != 0 {
		if s.orgRepo != nil {
			if org, err := s.orgRepo.GetByIDScoped(ctx, *req.OrgID); err == nil && org.TenantID == tenantID {
				orgName = org.Name
			}
		}
	}
	return &createResponse{
		ID:       newUserID,
		TenantID: tenantID,
		Code:     newUserCode,
		Username: newAccount.Username,
		RealName: newAccount.RealName,
		Phone:    newAccount.Phone,
		OrgID:    req.OrgID,
		OrgName:  orgName,
		Status:   newAccount.Status,
	}, nil
}
