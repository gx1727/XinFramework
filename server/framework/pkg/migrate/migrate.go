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
func ensureTable(ctx context.Context) {
	pool := db.Get()
	if pool == nil {
		return
	}
	// 创建迁移版本记录表
	_, _ = pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS _schema_migrations (
    version VARCHAR(255) PRIMARY KEY,     -- 迁移版本号（文件名）
    applied_at TIMESTAMPTZ DEFAULT NOW()  -- 应用时间
)`)
}

// isApplied 检查指定版本的迁移是否已应用
func isApplied(ctx context.Context, version string) bool {
	pool := db.Get()
	if pool == nil {
		return false
	}
	var exists bool
	_ = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM _schema_migrations WHERE version = $1)", version).Scan(&exists)
	return exists
}

// markApplied 标记指定版本的迁移为已应用
func markApplied(ctx context.Context, version string) error {
	pool := db.Get()
	if pool == nil {
		return ErrDBNotInitialized
	}
	_, err := pool.Exec(ctx, "INSERT INTO _schema_migrations (version) VALUES ($1)", version)
	return err
}

// Migration 迁移结构，表示单个SQL迁移文件
type Migration struct {
	Version string // 迁移版本号（文件名）
	SQL     string // SQL内容
}

// loadFromDir 从指定目录加载所有SQL迁移文件
func loadFromDir(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		// 如果目录不存在，返回空列表（不是错误）
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var migrations []Migration
	for _, entry := range entries {
		// 跳过子目录和非SQL文件
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

	// 按版本号排序（文件名排序）
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// Run 执行指定目录下的所有未应用的数据库迁移
func Run(dir string) error {
	pool := db.Get()
	if pool == nil {
		return ErrDBNotInitialized
	}

	ctx := context.Background()

	// 确保迁移记录表存在
	ensureTable(ctx)

	// 加载所有迁移文件
	migrations, err := loadFromDir(dir)
	if err != nil {
		return err
	}
	// 如果没有迁移文件，直接返回
	if len(migrations) == 0 {
		return nil
	}

	// 逐个应用未执行的迁移
	for _, m := range migrations {
		// 跳过已应用的迁移
		if isApplied(ctx, m.Version) {
			continue
		}

		fmt.Printf("[migrate] applying %s ...\n", m.Version)
		// 在事务内执行迁移 SQL，开头 SET LOCAL row_security = off 关闭 RLS。
		// 原因：migrations 可能跨租户/系统级 INSERT（如 framework_002_template_tenant.sql
		// 的 __template__ 租户 + menus / dicts 复制），RLS policy 会拦截非 owner 连接。
		// 关闭 RLS 是迁移期的合理行为——业务层仍由 app.bypass_rls（policy 需识别）+ tenant_id 控制。
		if err := runMigration(ctx, pool, m.SQL); err != nil {
			return fmt.Errorf("migration %s failed: %w", m.Version, err)
		}

		// 标记为已应用（事务已提交，再走普通连接）
		if err := markApplied(ctx, m.Version); err != nil {
			return fmt.Errorf("mark %s applied failed: %w", m.Version, err)
		}
		fmt.Printf("[migrate] %s done\n", m.Version)
	}

	return nil
}

// runMigration 在 RunInTx 内执行单条迁移，开头关闭 RLS。
// 失败自动回滚（RunInTx 内部 defer Rollback），保证不会留下半成品。
func runMigration(ctx context.Context, pool *pgxpool.Pool, sql string) error {
	return db.RunInTx(ctx, pool, func(ctx context.Context) error {
		q, err := db.GetQuerier(ctx)
		if err != nil {
			return err
		}
		// SET LOCAL 仅在本事务内生效，事务结束自动失效，不污染连接池。
		// row_security = off 等价于该 session 内 RLS 完全不参与检查。
		if _, err := q.Exec(ctx, "SET LOCAL row_security = off"); err != nil {
			return fmt.Errorf("set row_security off: %w", err)
		}
		if _, err := q.Exec(ctx, sql); err != nil {
			return err
		}
		return nil
	})
}
