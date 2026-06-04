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

func Init(cfg *config.DatabaseConfig) error {
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

func RunInTx(ctx context.Context, pool *pgxpool.Pool, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return fn(ctx)
	}

	if pool == nil {
		return fmt.Errorf("db pool is not initialized")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	ctx = WithTx(ctx, tx)
	if err := fn(ctx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func RunInTenantTx(ctx context.Context, pool *pgxpool.Pool, tenantID uint, fn func(ctx context.Context) error) error {
	return RunInTx(ctx, pool, func(ctx context.Context) error {
		tx := ctx.Value(txKey{}).(pgx.Tx)

		if tenantID > 0 {
			if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", strconv.Itoa(int(tenantID))); err != nil {
				return err
			}
		} else {
			// For public/system tenant access
			if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', '0', true)"); err != nil {
				return err
			}
		}

		return fn(ctx)
	})
}

// RunInPlatformTx 平台级事务：设置 app.tenant_id = '0'，并打开 app.bypass_rls = 'on'，
// 用于超级管理员跨租户/无租户上下文访问。
// 使用时 RLS policy 应识别 current_setting('app.bypass_rls', true) = 'on' 放行。
func RunInPlatformTx(ctx context.Context, pool *pgxpool.Pool, fn func(ctx context.Context) error) error {
	return RunInTx(ctx, pool, func(ctx context.Context) error {
		tx := ctx.Value(txKey{}).(pgx.Tx)
		if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', '0', true)"); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "SELECT set_config('app.bypass_rls', 'on', true)"); err != nil {
			return err
		}
		return fn(ctx)
	})
}

func GetQuerier(ctx context.Context) (Querier, error) {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx, nil
	}

	if Pool == nil {
		return nil, fmt.Errorf("db pool is not initialized")
	}

	return Pool, nil
}
