// Package migrate 数据库迁移工具（按目录扫 SQL 文件，按序应用）。
//
// Phase 4 重构：所有 helper 与 Run 都接收 pool 作为参数，删除对
// 包级 db.Pool / db.Get() 的依赖。
package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/db"
)

// ensureTable 确保数据库迁移记录表存在
func ensureTable(ctx context.Context, pool *pgxpool.Pool) {
	if pool == nil {
		return
	}
	_, _ = pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS _schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMPTZ DEFAULT NOW()
)`)
}

// isApplied 检查指定版本的迁移是否已应用
func isApplied(ctx context.Context, pool *pgxpool.Pool, version string) bool {
	if pool == nil {
		return false
	}
	var exists bool
	_ = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM _schema_migrations WHERE version = $1)", version).Scan(&exists)
	return exists
}

// markApplied 标记指定版本的迁移为已应用
func markApplied(ctx context.Context, pool *pgxpool.Pool, version string) error {
	if pool == nil {
		return ErrDBNotInitialized
	}
	_, err := pool.Exec(ctx, "INSERT INTO _schema_migrations (version) VALUES ($1)", version)
	return err
}

// Migration 迁移结构
type Migration struct {
	Version string
	SQL     string
}

// loadFromDir 从指定目录加载所有SQL迁移文件
func loadFromDir(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		migrations = append(migrations, Migration{
			Version: entry.Name(),
			SQL:     string(data),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// Run 执行指定目录下的所有未应用的数据库迁移。pool 由调用方持有。
func Run(pool *pgxpool.Pool, dir string) error {
	if pool == nil {
		return ErrDBNotInitialized
	}

	ctx := context.Background()

	ensureTable(ctx, pool)

	migrations, err := loadFromDir(dir)
	if err != nil {
		return err
	}
	if len(migrations) == 0 {
		return nil
	}

	for _, m := range migrations {
		if isApplied(ctx, pool, m.Version) {
			continue
		}

		fmt.Printf("[migrate] applying %s ...\n", m.Version)
		if err := runMigration(ctx, pool, m.SQL); err != nil {
			return fmt.Errorf("migration %s failed: %w", m.Version, err)
		}

		if err := markApplied(ctx, pool, m.Version); err != nil {
			return fmt.Errorf("mark %s applied failed: %w", m.Version, err)
		}
		fmt.Printf("[migrate] %s done\n", m.Version)
	}

	return nil
}

// runMigration 在 RunInTx 内执行单条迁移，开头关闭 RLS。
func runMigration(ctx context.Context, pool *pgxpool.Pool, sql string) error {
	return db.RunInTx(ctx, pool, func(ctx context.Context) error {
		q, err := db.GetQuerier(ctx, pool)
		if err != nil {
			return err
		}
		if _, err := q.Exec(ctx, "SET LOCAL row_security = off"); err != nil {
			return fmt.Errorf("set row_security off: %w", err)
		}
		if _, err := q.Exec(ctx, sql); err != nil {
			return err
		}
		return nil
	})
}
