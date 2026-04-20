package boot

import (
	"fmt"
	"log"

	"gx1727.com/xin/internal/core/server"
	"gx1727.com/xin/internal/infra/cache"
	"gx1727.com/xin/internal/infra/db"
	"gx1727.com/xin/internal/infra/logger"
	"gx1727.com/xin/pkg/config"
)

func Init(cfg *config.Config) (*server.XinServer, error) {
	logger.Init(cfg.Log.Dir, cfg.Log.Level)
	if err := db.Init(&cfg.Database); err != nil {
		return nil, fmt.Errorf("db init failed: %w", err)
	}
	if err := cache.Init(&cfg.Redis); err != nil {
		return nil, fmt.Errorf("cache init failed: %w", err)
	}

	srv := server.New(cfg)
	return srv, nil
}

func Shutdown() {
	if err := cache.Close(); err != nil {
		log.Printf("cache close failed: %v", err)
	}
	if err := db.Close(); err != nil {
		log.Printf("db close failed: %v", err)
	}
	logger.Close()
}
