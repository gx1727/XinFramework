// Package appx 持有进程级、模块真正需要的基础设施（DB、Config）。
//
// 设计原因：
//   - framework/internal/core/boot 启动期构造这些资源；
//   - boot 是 internal 包，apps/* 不能直接 import；
//   - appx 作为公开包，把构造好的 *App 暴露给应用模块。
//
// App 故意只保留业务模块直接消费的字段；framework 内部使用的 HTTP server
// 与 AppContext 走 framework.Runtime 通道，不在 App 上暴露，原因有两个：
//  1. Server / AppContext 的具体类型来自 framework/internal/*，在公开包
//     struct 上保留字段会绕过 internal 包的"禁止外部引用"设计意图；
//  2. App 历史上是一个"超集容器"——14 个模块里只有 DB / Config 是真用到的，
//     其余字段（Server / PermService / Authz / SessionMgr / AppContext）
//     仅 framework 自己消费，留在 App 上只会鼓励"反正都拿到了"的写法。
//
// App 中 Session / Authz 等跨模块共享资源请走 plugin.AppContext（在 framework.Runtime
// 里持有）；模块通过 Module(app *appx.App) 注入参数读取。
//
// # Phase 0024 改造
//
// 将 DB 字段从 *pgxpool.Pool 升级为 Pool 强类型包装：
//   - NewPool / MustNewPool 在构造期 fail-fast（保证非空）
//   - 业务模块用 app.Pool().Raw() 拿原生 pool，所有下游调用方零感知
//   - 消除散落在每个 module 里的 `if p := ctx.DB(); p != nil { ... }` 模式
package appx

import (
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/framework/pkg/config"
)

// Pool 是 *pgxpool.Pool 的强类型包装，构造期必须非空。
//
// 设计目的：
//   - 编译期 + 运行期双重保证池非空，消灭"app.DB == nil"防御性检查
//   - 业务模块用 Pool() 方法直接拿，零 nil-check
//   - Raw() 暴露原生 *pgxpool.Pool 给所有 pgx 兼容的 API（透明）
type Pool struct {
	raw *pgxpool.Pool
}

// NewPool 从 *pgxpool.Pool 构造强类型 Pool。pool 为 nil 时返回 error。
//
// 启动期应使用 MustNewPool（构造期 fail-fast）；此函数保留用于运行时构造场景
// （如多租户多 DB 路由，V2 阶段再考虑）。
func NewPool(pool *pgxpool.Pool) (*Pool, error) {
	if pool == nil {
		return nil, errors.New("appx: pgxpool cannot be nil")
	}
	return &Pool{raw: pool}, nil
}

// MustNewPool 是 NewPool 的 panic 版本。pool 为 nil 直接 panic（构造期错误）。
//
// 启动期使用：DB 不可用 = 进程不应该启动，立即退出比让后续 panic 更友好。
func MustNewPool(pool *pgxpool.Pool) *Pool {
	if pool == nil {
		panic("appx: pgxpool is nil at boot — this is a wiring bug, fix boot.Init")
	}
	return &Pool{raw: pool}
}

// Raw 返回原生 *pgxpool.Pool。绝大多数业务代码（pgx 调用、tx 工具、RLS）需要
// 原生类型，本方法作为透明桥接，不引入任何性能开销。
func (p *Pool) Raw() *pgxpool.Pool { return p.raw }

// Close 关闭底层连接池。幂等。
func (p *Pool) Close() {
	if p == nil || p.raw == nil {
		return
	}
	p.raw.Close()
}

// App 是显式 Build 阶段构造的进程级资源容器。
//
// 一个进程只有一个 App，模块通过 Module(app *appx.App) 接收它的引用，
// 不再依赖任何包级全局。
type App struct {
	// Config 启动时加载的配置
	Config *config.Config

	// DB PostgreSQL 连接池（强类型包装，构造期必非空）。
	//
	// 业务模块用 app.DB.Raw() 拿原生 *pgxpool.Pool 传给 Repository 构造函数。
	// 不需要 nil-check（构造期已保证非空）。
	DB *Pool
}

// NewApp 构造 App。pool / config 都不能为 nil；pool 应是 MustNewPool 的产物。
func NewApp(cfg *config.Config, pool *Pool) (*App, error) {
	if cfg == nil {
		return nil, errors.New("appx: config is required")
	}
	if pool == nil {
		return nil, errors.New("appx: pool is required (use MustNewPool)")
	}
	return &App{Config: cfg, DB: pool}, nil
}

// MustNewApp 是 NewApp 的 panic 版本。cfg / pool 为 nil 直接 panic。
func MustNewApp(cfg *config.Config, pool *Pool) *App {
	app, err := NewApp(cfg, pool)
	if err != nil {
		panic("appx.MustNewApp: " + err.Error())
	}
	return app
}
