package boot

import (
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/internal/core/server"
	"gx1727.com/xin/framework/internal/repository"
	"gx1727.com/xin/framework/internal/service"
	"gx1727.com/xin/framework/pkg/cache"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/dict"
	"gx1727.com/xin/framework/pkg/logger"
	"gx1727.com/xin/framework/pkg/permission"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/session"
)

type App struct {
	Config      *config.Config
	DB          *pgxpool.Pool
	Repository  *repository.Provider
	SessionMgr  session.SessionManager
	Server      *server.XinServer
	PermService *service.PermissionService
}

func Init(cfg *config.Config) (*App, error) {
	logger.Init(cfg.Log.Dir, cfg.Log.Level)
	if err := db.Init(&cfg.Database, cfg.Saas.Mode); err != nil {
		return nil, fmt.Errorf("db init failed: %w", err)
	}

	// 初始化 dict 缓存
	dict.Init(db.Get())

	// 初始化 repository
	repoProvider := repository.NewProvider(db.Get())
	repository.Init(repoProvider)

	if err := cache.Init(&cfg.Redis); err != nil {
		return nil, fmt.Errorf("cache init failed: %w", err)
	}

	// 初始化 session manager
	var sm session.SessionManager
	if cache.Get() != nil {
		sm = session.NewRedisSessionManager()
	} else {
		sm = session.NewDBSessionManager(db.Get())
	}
	session.Init(sm)

	// 初始化 permission service
	var permCache permission.PermissionCache
	if cache.Get() != nil {
		permCache = repository.NewRedisPermissionCache()
	}

	permService := service.NewPermissionService(
		repoProvider.Permission(),
		repoProvider.DataScope(),
		permCache,
	)

	return &App{
		Config:      cfg,
		DB:          db.Get(),
		Repository:  repoProvider,
		SessionMgr:  sm,
		Server:      server.New(cfg),
		PermService: permService,
	}, nil
}

func Shutdown(app *App) {
	for _, m := range plugin.All() {
		if err := m.Shutdown(); err != nil {
			log.Printf("module %s shutdown failed: %v", m.Name(), err)
		}
	}
	if err := cache.Close(); err != nil {
		log.Printf("cache close failed: %v", err)
	}
	db.Close()
	logger.Close()
}
