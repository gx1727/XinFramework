// Package appx 持有进程级应用资源（DB、Config、Session、Authz、AppContext）。
//
// 设计原因：
//   - framework/internal/core/boot 启动期构造这些资源；
//   - boot 是 internal 包，apps/* 不能直接 import；
//   - appx 作为公开包，把构造好的 *App 暴露给应用模块。
//
// Phase 5：appx 是显式 Build 阶段的产物，main.go 调用 boot.Init 拿到 *App，
// 然后显式把它传给每个模块的 Module(app) 构造函数，不再走 bootx.Pool()
// 这种过渡期访问器。
//
// App 中的 PermService / Authz / Server 字段类型来自 framework/internal/*，
// apps 不应该直接访问这些字段；如果确实需要，appx 应该再补一层薄包装。
package appx

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/framework/internal/core/server"
	"gx1727.com/xin/framework/internal/service"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/session"
)

// App 是显式 Build 阶段构造的进程级资源容器。
//
// 一个进程只有一个 App，模块通过 Module(app *appx.App) 接收它的引用，
// 不再依赖任何包级全局。
type App struct {
	// Config 启动时加载的配置
	Config *config.Config

	// DB PostgreSQL 连接池，模块用它做所有数据库操作
	DB *pgxpool.Pool

	// SessionMgr session 管理器（Redis 或 DB 实现）
	SessionMgr session.SessionManager

	// Server gin server 引用，由 framework.Serve 用来 Start / Shutdown
	Server *server.XinServer

	// PermService 权限服务，提供 LoadUserSecurityContext 等
	PermService *service.PermissionService

	// Authz 授权服务（authz.Wrap 后的 Authz 适配器），用于中间件与业务校验
	Authz *service.AuthorizationService

	// AppContext 跨模块共享的 Reader/Writer 容器。
	// 写：模块在 Init 阶段把自己的 repository 写进对应的 slot；
	// 读：模块在 RegFn 里通过 ctx plugin.Reader 读到别的模块写进来的依赖。
	AppContext *plugin.AppContext
}
