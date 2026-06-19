// Package bootx 暴露 boot 内部 App 的少量过渡期访问器。
//
// Phase 4 设计：app 模块不应再使用 db.Get() / config.Get() 这类包级全局，
// 但完全切到显式注入需要把所有 14 个模块重写为 New(deps Deps) 形态，
// 工作量大且与本阶段"先消除全局变量"的目标正交。
//
// 因此：
//   - 真正的资源（pool、config）仍只在 framework/internal/core/boot 内部持有；
//   - 这里只暴露 Pool() / Config() 两个 accessor 给 app 模块使用；
//   - 这两个 accessor 在未来 main.go 显式 Build 阶段后整体删除。
//
// app 模块代码风格建议：模块 init() 中调用 bootx.Pool() 取 pool，
// 然后构造自己的 service / repository。
package bootx

import (
	"gx1727.com/xin/framework/internal/core/boot"
	"gx1727.com/xin/framework/pkg/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool 返回当前进程的 *pgxpool.Pool。仅供过渡期使用。
func Pool() *pgxpool.Pool {
	return boot.Pool()
}

// Config 返回当前进程的 *config.Config。仅供过渡期使用。
func Config() *config.Config {
	return boot.Config()
}
