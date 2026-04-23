package db

import (
	"context"
	"fmt"
	"time"

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

func SetTenantID(ctx context.Context, tenantID uint) error {
	if Pool != nil && tenantID > 0 {
		_, err := Pool.Exec(ctx, "SET app.tenant_id = $1", tenantID)
		return err
	}
	return nil
}

func ClearTenantID(ctx context.Context) error {
	if Pool != nil {
		_, err := Pool.Exec(ctx, "RESET app.tenant_id")
		return err
	}
	return nil
}

func Close() {
	if Pool != nil {
		Pool.Close()
	}
}
