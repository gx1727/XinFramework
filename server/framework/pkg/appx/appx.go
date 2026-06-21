// Package appx 持有进程级、模块真正需要的基础设施（DB、Config）。
//
// 设计原因：
//   - framework/internal/core/boot 启动期构造这些资源；
//   - boot 是 internal 包，apps/* 不能直接 import；
//   - appx 作为公开包，把构造好的 *App 暴露给应用模块。
//
// App 故意只保留业务模块直接消费的字段；framework 内部使用的 HTTP server
// 与 AppContext 走 framework.Runtime 通道，不在 App 上暴露，原因有两个：
//   1. Server / AppContext 的具体类型来自 framework/internal/*，在公开包
//      struct 上保留字段会绕过 internal 包的"禁止外部引用"设计意图；
//   2. App 历史上是一个"超集容器"——14 个模块里只有 DB / Config 是真用到的，
//      其余字段（Server / PermService / Authz / SessionMgr / AppContext）
//      仅 framework 自己消费，留在 App 上只会鼓励"反正都拿到了"的写法。
//
// App 中 Session / Authz 等跨模块共享资源请走 plugin.AppContext（在 framework.Runtime
// 里持有）；模块通过 Module(app *appx.App) 注入参数读取。
package appx

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/framework/pkg/config"
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
}