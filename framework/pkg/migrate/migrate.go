package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gx1727.com/xin/framework/pkg/db"
)

func ensureTable() {
	d := db.Get()
	if d == nil {
		return
	}
	d.Exec(`
CREATE TABLE IF NOT EXISTS _schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMPTZ DEFAULT NOW()
)`)
}

func isApplied(version string) bool {
	d := db.Get()
	if d == nil {
		return false
	}
	var count int64
	d.Table("_schema_migrations").Where("version = ?", version).Count(&count)
	return count > 0
}

func markApplied(version string) error {
	d := db.Get()
	if d == nil {
		return fmt.Errorf("db not initialized")
	}
	return d.Table("_schema_migrations").Create(map[string]interface{}{"version": version}).Error
}

type Migration struct {
	Version string
	SQL     string
}

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

func Run(dir string) error {
	d := db.Get()
	if d == nil {
		return fmt.Errorf("db not initialized")
	}

	ensureTable()

	migrations, err := loadFromDir(dir)
	if err != nil {
		return err
	}
	if len(migrations) == 0 {
		return nil
	}

	for _, m := range migrations {
		if isApplied(m.Version) {
			continue
		}

		fmt.Printf("[migrate] applying %s ...\n", m.Version)
		if err := d.Exec(m.SQL).Error; err != nil {
			return fmt.Errorf("migration %s failed: %w", m.Version, err)
		}

		if err := markApplied(m.Version); err != nil {
			return fmt.Errorf("mark %s applied failed: %w", m.Version, err)
		}
		fmt.Printf("[migrate] %s done\n", m.Version)
	}

	return nil
}
