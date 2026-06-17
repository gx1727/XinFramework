package boot

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/internal/core/ext_impl"
	"gx1727.com/xin/framework/internal/core/server"
	"gx1727.com/xin/framework/internal/service"
	"gx1727.com/xin/framework/pkg/authz"
	"gx1727.com/xin/framework/pkg/cache"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/dict"
	"gx1727.com/xin/framework/pkg/logger"
	"gx1727.com/xin/framework/pkg/permission"
	"gx1727.com/xin/framework/pkg/session"
)

type App struct {
	Config      *config.Config
	DB          *pgxpool.Pool
	SessionMgr  session.SessionManager
	Server      *server.XinServer
	PermService *service.PermissionService
	Authz       *service.AuthorizationService
}

func Init(cfg *config.Config) (*App, error) {
	logger.Init(cfg.Log.Dir, cfg.Log.Level)
	if err := db.Init(&cfg.Database); err != nil {
		return nil, fmt.Errorf("db init failed: %w", err)
	}

	dict.Init(db.Get())

	if err := cache.Init(&cfg.Redis); err != nil {
		return nil, fmt.Errorf("cache init failed: %w", err)
	}

	var sm session.SessionManager
	if cache.Get() != nil {
		sm = session.NewRedisSessionManager()
	} else {
		sm = session.NewDBSessionManager(db.Get())
	}
	session.Init(sm)

	var permCache permission.PermissionCache
	if cache.Get() != nil {
		permCache = permission.NewRedisPermissionCache()
	}

	// Construct one AppContext and share it. Phase 5 will populate
	// more slots on this instance as each module's Init runs.
	appCtx := plugin.NewAppContext(db.Get(), cache.Get(), cfg, sm)
	ext_impl.InitExtApi(appCtx)

	permService := service.NewPermissionService(
		permission.NewPermissionRepository(db.Get()),
		permission.NewDataScopeRepository(db.Get()),
		permCache,
		permission.NewPlatformRoleRepository(db.Get()),
	)
	authzService := service.NewAuthorizationService(permService)
	// Publish the Authorization onto AppContext so apps can consume it
	// via ctx.Authz() in their module's Register phase. The concrete
	// *service.AuthorizationService lives in framework/internal/, so
	// we wrap it through authz.Wrap() to expose the public interface.
	authzSvc := authz.Wrap(authzService)
	appCtx.SetAuthz(authzSvc)

	app := &App{
		Config:      cfg,
		DB:          db.Get(),
		SessionMgr:  sm,
		Server:      server.New(cfg),
		PermService: permService,
		Authz:       authzService,
	}

	// 启动期引导：在普通业务表就绪前确保存在一个 super_admin
	if bcfg := LoadBootstrapConfig(); bcfg.Enabled {
		if err := RunBootstrap(context.Background(), db.Get(), bcfg); err != nil {
			log.Printf("[bootstrap] failed: %v", err)
		}
	}

	return app, nil
}

func Shutdown(app *App) {
	// Phase 3 will pass app.ContextReader here. For now Shutdown
	// only does connection close, so a nil Reader is fine.
	var reader plugin.Reader
	for _, m := range plugin.Apps() {
		if err := m.Shutdown(reader); err != nil {
			log.Printf("module %s shutdown failed: %v", m.Name(), err)
		}
	}
	if err := cache.Close(); err != nil {
		log.Printf("cache close failed: %v", err)
	}
	db.Close()
	logger.Close()
}
