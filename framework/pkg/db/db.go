package db

import (
	"context"
	"fmt"
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

type Conn struct {
	pool   *pgxpool.Pool
	conn   *pgxpool.Conn
	tenant uint
}

func Acquire(ctx context.Context) (*Conn, error) {
	conn, err := Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	return &Conn{pool: Pool, conn: conn}, nil
}

func (c *Conn) SetTenant(ctx context.Context, tenantID uint) error {
	c.tenant = tenantID
	_, err := c.conn.Exec(ctx, fmt.Sprintf("SET app.tenant_id = '%d'", tenantID))
	return err
}

func (c *Conn) ShowDeleted(ctx context.Context) error {
	_, err := c.conn.Exec(ctx, "SET app.show_deleted = true")
	return err
}

func (c *Conn) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return c.conn.Exec(ctx, sql, args...)
}

func (c *Conn) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return c.conn.Query(ctx, sql, args...)
}

func (c *Conn) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return c.conn.QueryRow(ctx, sql, args...)
}

func (c *Conn) Release() {
	if c.tenant > 0 {
		_, _ = c.conn.Exec(context.Background(), "RESET app.tenant_id")
	}
	c.conn.Release()
}

func Get() *pgxpool.Pool {
	return Pool
}

func Close() {
	if Pool != nil {
		Pool.Close()
	}
}

// Querier interface allows the repository layer to execute SQL seamlessly
// whether using a single connection or participating in a transaction.
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

// GetQuerier intelligently returns a Querier (either an existing Tx from context or a new Conn).
// It also returns a release function that the caller MUST defer.
func GetQuerier(ctx context.Context) (Querier, func(), error) {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		// Existing transaction found in context. Release is a no-op because
		// the transaction is managed by the outer caller (e.g., Service layer).
		return tx, func() {}, nil
	}

	conn, err := Acquire(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Single connection. Must be released when done.
	return conn, func() { conn.Release() }, nil
}
