package boot

import (
	"fmt"

	"gx1727.com/xin/internal/core/server"
	"gx1727.com/xin/internal/infra/cache"
	"gx1727.com/xin/internal/infra/db"
	"gx1727.com/xin/internal/infra/logger"
	"gx1727.com/xin/pkg/config"
)

func Init(cfg *config.Config) (*server.XinServer, error) {
	logger.Init()
	if err := db.Init(&cfg.Database); err != nil {
		return nil, fmt.Errorf("db init failed: %w", err)
	}
	cache.Init(&cfg.Redis)

	srv := server.New(cfg)
	return srv, nil
}
