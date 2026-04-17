package boot

import (
	"fmt"

	"github.com/xin-framework/xin/configs"
	"github.com/xin-framework/xin/internal/core/server"
	"github.com/xin-framework/xin/internal/infra/cache"
	"github.com/xin-framework/xin/internal/infra/db"
	"github.com/xin-framework/xin/internal/infra/logger"
)

func Init(cfg *configs.Config) (*server.XinServer, error) {
	logger.Init()
	if err := db.Init(&cfg.Database); err != nil {
		return nil, fmt.Errorf("db init failed: %w", err)
	}
	cache.Init(&cfg.Redis)

	srv := server.New(cfg)
	return srv, nil
}
