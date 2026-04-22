package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gx1727.com/xin/framework/pkg/db"
)

// ensureTable 确保数据库迁移记录表存在
func ensureTable() {
	d := db.Get()
	if d == nil {
		return
	}
	// 创建迁移版本记录表
	d.Exec(`
CREATE TABLE IF NOT EXISTS _schema_migrations (
    version VARCHAR(255) PRIMARY KEY,     -- 迁移版本号（文件名）
    applied_at TIMESTAMPTZ DEFAULT NOW()  -- 应用时间
)`)
}

// isApplied 检查指定版本的迁移是否已应用
func isApplied(version string) bool {
	d := db.Get()
	if d == nil {
		return false
	}
	var count int64
	d.Table("_schema_migrations").Where("version = ?", version).Count(&count)
	return count > 0
}

// markApplied 标记指定版本的迁移为已应用
func markApplied(version string) error {
	d := db.Get()
	if d == nil {
		return fmt.Errorf("db not initialized")
	}
	return d.Table("_schema_migrations").Create(map[string]interface{}{"version": version}).Error
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
	d := db.Get()
	if d == nil {
		return fmt.Errorf("db not initialized")
	}

	// 确保迁移记录表存在
	ensureTable()

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
		if isApplied(m.Version) {
			continue
		}

		fmt.Printf("[migrate] applying %s ...\n", m.Version)
		// 执行SQL迁移
		if err := d.Exec(m.SQL).Error; err != nil {
			return fmt.Errorf("migration %s failed: %w", m.Version, err)
		}

		// 标记为已应用
		if err := markApplied(m.Version); err != nil {
			return fmt.Errorf("mark %s applied failed: %w", m.Version, err)
		}
		fmt.Printf("[migrate] %s done\n", m.Version)
	}

	return nil
}
