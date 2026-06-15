package boot

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pkgauth "gx1727.com/xin/framework/pkg/auth"
	"gx1727.com/xin/framework/pkg/db"
)

// BootstrapConfig 启动期引导配置（仅在初始化阶段使用一次）
type BootstrapConfig struct {
	Enabled    bool   // 是否启用（要求 XIN_BOOTSTRAP_TOKEN 非空）
	Token      string // 期望匹配的令牌（仅在 XIN_BOOTSTRAP_TOKEN 设置后才生效）
	Account    string // 账号（username/phone/email）
	Password   string // 明文密码
	RealName   string // 真实姓名
	Role       string // 平台角色，默认 super_admin
	TenantCode string // 要绑定的租户 code，默认 "default"
}

// LoadBootstrapConfig 从环境变量读取引导配置
// 任何字段缺失（特别是 Token）即视为禁用
func LoadBootstrapConfig() BootstrapConfig {
	cfg := BootstrapConfig{
		Token:      os.Getenv("XIN_BOOTSTRAP_TOKEN"),
		Account:    os.Getenv("XIN_BOOTSTRAP_ACCOUNT"),
		Password:   os.Getenv("XIN_BOOTSTRAP_PASSWORD"),
		RealName:   os.Getenv("XIN_BOOTSTRAP_REAL_NAME"),
		Role:       os.Getenv("XIN_BOOTSTRAP_ROLE"),
		TenantCode: os.Getenv("XIN_BOOTSTRAP_TENANT_CODE"),
	}
	if cfg.Role == "" {
		cfg.Role = "super_admin"
	}
	if cfg.TenantCode == "" {
		cfg.TenantCode = "default"
	}
	cfg.Enabled = cfg.Token != "" && cfg.Account != "" && cfg.Password != ""
	return cfg
}

// RunBootstrap 在启动时确保存在一个 super_admin 账号
// 调用条件：cfg.Enabled == true
// 该函数只读不写普通业务表，仅操作 accounts/account_roles/users/user_roles。
func RunBootstrap(ctx context.Context, pool *pgxpool.Pool, cfg BootstrapConfig) error {
	if !cfg.Enabled {
		return nil
	}

	// 简单防御：至少 16 位 token
	if len(cfg.Token) < 16 {
		return errors.New("XIN_BOOTSTRAP_TOKEN 至少 16 位")
	}

	// 1. accounts：若已存在则取 id；否则创建
	accountID, created, err := upsertBootstrapAccount(ctx, pool, cfg)
	if err != nil {
		return err
	}
	if created {
		log.Printf("[bootstrap] created bootstrap account %q (id=%d)", cfg.Account, accountID)
	} else {
		log.Printf("[bootstrap] bootstrap account %q already exists (id=%d)", cfg.Account, accountID)
	}

	// 2. account_roles：授予平台角色（幂等）
	// 注意：migrations 里用的是 CREATE UNIQUE INDEX，PG 的 ON CONFLICT ON CONSTRAINT
	// 只认 UNIQUE / PRIMARY KEY / EXCLUSION 约束，纯唯一索引必须用列名形式。
	if _, err := pool.Exec(ctx, `
		INSERT INTO account_roles (account_id, role)
		VALUES ($1, $2)
		ON CONFLICT (account_id, role) DO NOTHING
	`, accountID, cfg.Role); err != nil {
		return err
	}
	log.Printf("[bootstrap] granted platform role %q to account %d", cfg.Role, accountID)

	// 3. users：把账号绑定到目标租户（登录时 LoginIdentity 需要 users 行）
	if err := upsertBootstrapUser(ctx, pool, accountID, cfg); err != nil {
		return err
	}

	return nil
}

// upsertBootstrapUser 把 account 绑定到 cfg.TenantCode 租户。
// - tenants 表无 RLS，可直接查
// - users / user_roles 受 RLS 限制，必须在 RunInTenantTx 内执行
func upsertBootstrapUser(ctx context.Context, pool *pgxpool.Pool, accountID uint, cfg BootstrapConfig) error {
	// 3.1 查目标租户（tenants 不受 RLS 限制）
	var tenantID uint
	var tenantStatus int16
	err := pool.QueryRow(ctx, `
		SELECT id, status FROM tenants
		WHERE is_deleted = FALSE AND code = $1
		LIMIT 1
	`, cfg.TenantCode).Scan(&tenantID, &tenantStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("XIN_BOOTSTRAP_TENANT_CODE=" + cfg.TenantCode + " 的租户不存在")
		}
		return err
	}
	if tenantStatus != 1 {
		return errors.New("租户 " + cfg.TenantCode + " 已禁用")
	}

	// 3.2 在租户事务内 upsert users 行 + 绑定 admin 角色
	var userID uint
	var userCreated bool
	err = db.RunInTenantTx(ctx, pool, tenantID, func(ctx context.Context) error {
		querier, err := db.GetQuerier(ctx)
		if err != nil {
			return err
		}

		// 该 account 是否已经有任意 users 绑定（任一租户）
		err = querier.QueryRow(ctx, `
			SELECT id FROM users
			WHERE is_deleted = FALSE AND account_id = $1
			ORDER BY id ASC LIMIT 1
		`, accountID).Scan(&userID)
		if err == nil {
			// 已存在：更新真实姓名，避免令牌泄露后无法回收
			if _, err := querier.Exec(ctx, `
				UPDATE users SET real_name = COALESCE(NULLIF($1, ''), real_name), updated_at = NOW()
				WHERE id = $2
			`, cfg.RealName, userID); err != nil {
				return err
			}
		} else {
			if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
			// 创建 users 行
			if err := querier.QueryRow(ctx, `
				INSERT INTO users (tenant_id, account_id, code, real_name, status)
				VALUES ($1, $2, $3, $4, 1)
				RETURNING id
			`, tenantID, accountID, cfg.Account, cfg.RealName).Scan(&userID); err != nil {
				return err
			}
			userCreated = true
		}

		// 3.3 绑定 tenant 内的 admin 角色（如存在则给，便于带 admin 权限登录）
		var roleID uint
		err = querier.QueryRow(ctx, `
			SELECT id FROM roles
			WHERE is_deleted = FALSE AND tenant_id = $1 AND code = 'admin'
			LIMIT 1
		`, tenantID).Scan(&roleID)
		if err == nil {
			// uk_ur_unique 是部分唯一索引（WHERE is_deleted = FALSE），
			// ON CONFLICT ON CONSTRAINT 不认 index 形式的 arbiter，必须显式列名 + WHERE。
			if _, err := querier.Exec(ctx, `
				INSERT INTO user_roles (tenant_id, user_id, role_id)
				VALUES ($1, $2, $3)
				ON CONFLICT (user_id, role_id) WHERE is_deleted = FALSE DO NOTHING
			`, tenantID, userID, roleID); err != nil {
				return err
			}
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	if userCreated {
		log.Printf("[bootstrap] created user binding account %d -> tenant %q (user_id=%d)", accountID, cfg.TenantCode, userID)
	} else {
		log.Printf("[bootstrap] user binding for account %d already exists (user_id=%d)", accountID, userID)
	}

	return nil
}

func upsertBootstrapAccount(ctx context.Context, pool *pgxpool.Pool, cfg BootstrapConfig) (uint, bool, error) {
	passwordHash, err := pkgauth.HashPassword(cfg.Password)
	if err != nil {
		return 0, false, err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, false, err
	}
	defer tx.Rollback(ctx)

	ctxTx := db.WithTx(ctx, tx)

	var accountID uint
	err = tx.QueryRow(ctxTx, `
		SELECT id FROM accounts
		WHERE is_deleted = FALSE AND (username = $1 OR phone = $1 OR email = $1)
		LIMIT 1
	`, cfg.Account).Scan(&accountID)
	if err == nil {
		// 已存在：更新密码 + 真实姓名，避免令牌泄露后无法回收
		if _, err := tx.Exec(ctxTx, `
			UPDATE accounts SET password = $1, real_name = COALESCE(NULLIF($2, ''), real_name), updated_at = NOW()
			WHERE id = $3
		`, passwordHash, cfg.RealName, accountID); err != nil {
			return 0, false, err
		}
		if err := tx.Commit(ctxTx); err != nil {
			return 0, false, err
		}
		return accountID, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, false, err
	}

	// 创建账号
	err = tx.QueryRow(ctxTx, `
		INSERT INTO accounts (username, phone, email, real_name, password, status)
		VALUES ($1, $1, $1, $2, $3, 1)
		RETURNING id
	`, cfg.Account, cfg.RealName, passwordHash).Scan(&accountID)
	if err != nil {
		return 0, false, err
	}

	if err := tx.Commit(ctxTx); err != nil {
		return 0, false, err
	}
	return accountID, true, nil
}
