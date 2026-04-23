package user

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"strconv"
)

type LoginIdentity struct {
	UserID       uint
	TenantID     uint
	UserCode     string
	UserStatus   int16
	RoleCode     string
	PasswordHash string
}

func ResolveLoginIdentity(ctx context.Context, d *pgxpool.Pool, account string, tenantID uint) (*LoginIdentity, error) {
	if d == nil {
		return nil, ErrBackendUnavailable
	}

	tx, err := d.Begin(ctx)
	if err != nil {
		return nil, ErrRegisterFailed
	}
	defer tx.Rollback(ctx)

	if tenantID > 0 {
		_, err = tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", strconv.Itoa(int(tenantID)))
		if err != nil {
			return nil, fmt.Errorf("set tenant_id: %w", err)
		}
	} else {
		_, err = tx.Exec(ctx, "SELECT set_config('app.mode', $1, true)", "single")
		if err != nil {
			return nil, fmt.Errorf("set mode: %w", err)
		}
	}

	var accID uint
	var password string
	err = tx.QueryRow(ctx, `
		SELECT id, password 
		FROM accounts 
		WHERE username = $1 OR phone = $1 OR email = $1
		LIMIT 1`, account).Scan(&accID, &password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errAccountNotFound
		}
		return nil, err
	}

	query := `
		SELECT id, tenant_id, code, status 
		FROM users 
		WHERE  account_id = $1`
	args := []interface{}{accID}

	query += " ORDER BY id ASC LIMIT 1"

	var uID uint
	var uTenantID uint
	var uCode string
	var uStatus int16
	err = tx.QueryRow(ctx, query, args...).Scan(&uID, &uTenantID, &uCode, &uStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errTenantBindingNotFound
		}
		return nil, err
	}

	if tenantID == 0 {
		_, err = tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", strconv.Itoa(int(uTenantID)))
		if err != nil {
			return nil, fmt.Errorf("set tenant_id: %w", err)
		}

		_, err = tx.Exec(ctx, "SELECT set_config('app.mode', $1, true)", "saas")
		if err != nil {
			return nil, fmt.Errorf("set mode: %w", err)
		}
	}

	roleCode := "user"
	err = tx.QueryRow(ctx, `
		SELECT r.code 
		FROM user_roles ur 
		JOIN roles r ON r.id = ur.role_id 
		WHERE ur.user_id = $1 
		ORDER BY ur.id ASC LIMIT 1`, uID).Scan(&roleCode)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		// Ignore role error, fallback to "user"
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, ErrRegisterFailed
	}

	return &LoginIdentity{
		UserID:       uID,
		TenantID:     uTenantID,
		UserCode:     uCode,
		UserStatus:   uStatus,
		RoleCode:     roleCode,
		PasswordHash: password,
	}, nil
}
