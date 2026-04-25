package auth

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserCodeFormat string

const (
	UserCodeFormatSequential   UserCodeFormat = "sequential"
	UserCodeFormatTenantPrefix UserCodeFormat = "tenant_prefix"
	UserCodeFormatTenantRandom UserCodeFormat = "tenant_random"
)

func generateUserCode(ctx context.Context, db *pgxpool.Pool, tenantID uint, format UserCodeFormat) (string, error) {
	seq, format, err := getNextSeqWithFormat(ctx, db, tenantID, format)
	if err != nil {
		return "", err
	}

	switch format {
	case UserCodeFormatTenantPrefix:
		return fmt.Sprintf("U%03d-%05d", tenantID%1000, seq%100000), nil
	case UserCodeFormatTenantRandom:
		return generateRandomCode(tenantID, seq), nil
	default:
		return fmt.Sprintf("U%08d", seq), nil
	}
}

type seqResult struct {
	seq    int64
	format UserCodeFormat
}

func getNextSeqWithFormat(ctx context.Context, db *pgxpool.Pool, tenantID uint, defaultFormat UserCodeFormat) (int64, UserCodeFormat, error) {
	var result seqResult
	err := db.QueryRow(ctx, `
		INSERT INTO tenant_user_seq (tenant_id, seq, user_code_format)
		VALUES ($1, 1, $2)
		ON CONFLICT (tenant_id)
		DO UPDATE SET seq = tenant_user_seq.seq + 1, updated_at = NOW()
		RETURNING tenant_user_seq.seq, tenant_user_seq.user_code_format
	`, tenantID, string(defaultFormat)).Scan(&result.seq, &result.format)
	if err != nil {
		return 0, "", fmt.Errorf("get next user seq: %w", err)
	}
	return result.seq, result.format, nil
}

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateRandomCode(tenantID uint, seq int64) string {
	r := rand.New(rand.NewSource(seq + int64(tenantID)*1000))
	b := make([]byte, 5)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return fmt.Sprintf("U%03d-%s", tenantID%1000, string(b))
}

// SetUserCodeFormat 更新租户的用户编码格式
func SetUserCodeFormat(ctx context.Context, db *pgxpool.Pool, tenantID uint, format UserCodeFormat) error {
	_, err := db.Exec(ctx, `
		INSERT INTO tenant_user_seq (tenant_id, seq, user_code_format)
		VALUES ($1, 0, $2)
		ON CONFLICT (tenant_id)
		DO UPDATE SET user_code_format = $2, updated_at = NOW()
	`, tenantID, string(format))
	return err
}

// GetUserCodeFormat 获取租户的用户编码格式
func GetUserCodeFormat(ctx context.Context, db *pgxpool.Pool, tenantID uint) (UserCodeFormat, error) {
	var format string
	err := db.QueryRow(ctx, `
		SELECT user_code_format FROM tenant_user_seq WHERE tenant_id = $1
	`, tenantID).Scan(&format)
	if err != nil {
		return UserCodeFormatSequential, err
	}
	return UserCodeFormat(format), nil
}
