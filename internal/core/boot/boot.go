package boot

import (
	"fmt"
	"sync"

	"gx1727.com/xin/internal/core/server"
	"gx1727.com/xin/internal/infra/cache"
	"gx1727.com/xin/internal/infra/db"
	"gx1727.com/xin/internal/infra/logger"
	"gx1727.com/xin/pkg/config"
)

var (
	globalSrv *server.XinServer
	once      sync.Once
)

func Init(cfg *config.Config) (*server.XinServer, error) {
	logger.Init(cfg.Log.Dir, cfg.Log.Level)
	if err := db.Init(&cfg.Database); err != nil {
		return nil, fmt.Errorf("db init failed: %w", err)
	}
	cache.Init(&cfg.Redis)

	srv := server.New(cfg)
	once.Do(func() {
		globalSrv = srv
	})
	return srv, nil
}

func GetServer() *server.XinServer {
	return globalSrv
}
