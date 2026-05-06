package db

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/config"
)

var Pool *pgxpool.Pool

func Init(cfg *config.DatabaseConfig, saasMode string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return fmt.Errorf("parse dsn: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		poolConfig.MinConns = int32(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetimeSec > 0 {
		poolConfig.MaxConnLifetime = time.Duration(cfg.ConnMaxLifetimeSec) * time.Second
	}
	if cfg.ConnMaxIdleTimeSec > 0 {
		poolConfig.MaxConnIdleTime = time.Duration(cfg.ConnMaxIdleTimeSec) * time.Second
	}

	if saasMode != "" {
		mode := saasMode
		poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
			_, err := conn.Exec(ctx, "SELECT set_config('app.mode', $1, false)", mode)
			if err != nil {
				return fmt.Errorf("set app.mode: %w", err) // ✅ 包装错误，提供上下文
			}
			return nil // ✅ 明确返回 nil
		}
	}

	Pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("create pool: %w", err)
	}

	if err := Pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	return nil
}

func Get() *pgxpool.Pool {
	return Pool
}

func Close() {
	if Pool != nil {
		Pool.Close()
	}
}

type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, optionsAndArgs ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, optionsAndArgs ...any) pgx.Row
}

type txKey struct{}

// WithTx injects a pgx.Tx into the context so that underlying repositories
// can join the same transaction automatically.
func WithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func BeginTenantTx(ctx context.Context, pool *pgxpool.Pool, tenantID uint) (pgx.Tx, error) {
	if pool == nil {
		return nil, fmt.Errorf("db pool is not initialized")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	if tenantID > 0 {
		if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", strconv.Itoa(int(tenantID))); err != nil {
			_ = tx.Rollback(ctx)
			return nil, err
		}
	}

	return tx, nil
}

func GetTenantQuerier(ctx context.Context, pool *pgxpool.Pool, tenantID uint) (context.Context, Querier, pgx.Tx, error) {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return ctx, tx, nil, nil
	}

	if tenantID > 0 {
		tx, err := BeginTenantTx(ctx, pool, tenantID)
		if err != nil {
			return ctx, nil, nil, err
		}
		ctx = WithTx(ctx, tx)
		return ctx, tx, tx, nil
	}

	if pool == nil {
		return ctx, nil, nil, fmt.Errorf("db pool is not initialized")
	}

	return ctx, pool, nil, nil
}

func FinishTx(ctx context.Context, tx pgx.Tx, opErr error) error {
	if tx == nil {
		return opErr
	}
	if opErr != nil {
		_ = tx.Rollback(ctx)
		return opErr
	}
	return tx.Commit(ctx)
}

func GetQuerier(ctx context.Context) (Querier, func(), error) {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx, func() {}, nil
	}

	if Pool == nil {
		return nil, nil, fmt.Errorf("db pool is not initialized")
	}

	return Pool, func() {}, nil
}
