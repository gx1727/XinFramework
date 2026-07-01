// Package db 数据库连接与事务工具。
//
// 本包不再持有进程级全局状态。Init(ctx, cfg) 返回 *pgxpool.Pool，
// 由调用方（通常是 framework/internal/core/boot.App）持有。
//
// Phase 4 重构要点：
//   - 删除 var Pool / Get() / Close()（包级全局）
//   - Init(ctx, cfg) (*pgxpool.Pool, error) 返回池
//   - GetQuerier(ctx, pool) 需要显式传入 pool
//   - RunInTx / RunInTenantTx / RunInSysTx 一直就是 pool 参数化
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

// Init 解析配置并创建连接池。返回的 pool 由调用方持有与关闭。
func Init(ctx context.Context, cfg *config.DatabaseConfig) (*pgxpool.Pool, error) {
	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
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

	pool, err := pgxpool.NewWithConfig(initCtx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(initCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return pool, nil
}

// Querier 是 Pool 与 Tx 共同实现的最小查询接口。
type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, optionsAndArgs ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, optionsAndArgs ...any) pgx.Row
}

// txKey 是事务注入到 context 的私有键。
type txKey struct{}

// WithTx 注入一个 pgx.Tx 到 context，使下游 Repository 能自动加入同一事务。
func WithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// GetQuerier 从 ctx 提取事务（如果有），否则回退到 pool。
func GetQuerier(ctx context.Context, pool *pgxpool.Pool) (Querier, error) {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx, nil
	}
	if pool == nil {
		return nil, fmt.Errorf("db pool is not initialized")
	}
	return pool, nil
}

// RunInTx 在事务中执行 fn；如 ctx 已携带事务则直接复用（嵌套安全）。
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

// RunInTenantTx 在事务中设置 app.tenant_id = tenantID，并执行 fn。
func RunInTenantTx(ctx context.Context, pool *pgxpool.Pool, tenantID uint, fn func(ctx context.Context) error) error {
	return RunInTx(ctx, pool, func(ctx context.Context) error {
		tx := ctx.Value(txKey{}).(pgx.Tx)

		if tenantID > 0 {
			if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", strconv.Itoa(int(tenantID))); err != nil {
				return err
			}
		} else {
			if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', '0', true)"); err != nil {
				return err
			}
		}
		return fn(ctx)
	})
}

// RunInSysTx 在事务中开启 app.bypass_rls，用于 sys 级跨租户操作。
func RunInSysTx(ctx context.Context, pool *pgxpool.Pool, fn func(ctx context.Context) error) error {
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
