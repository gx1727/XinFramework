package framework

import (
	"gx1727.com/xin/framework/internal/core/server"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Runtime 持有 framework 启动后内部使用的运行时资源。
//
// 这些资源只供 framework 包内代码消费：
//   - Server：HTTP server 控制（Start / Shutdown）
//   - AppCtx：驱动模块 Init / Register 的 plugin.Reader/Writer 容器；
//     Session / Authz 等共享资源由它持有，framework 通过
//     rt.AppCtx.Session() / rt.AppCtx.Authz() 读取，不再单独持有字段。
//
// Runtime 故意不传给业务模块——业务模块拿到的只是 *appx.App
// （Config + DB），跨模块共享通过 plugin.AppContext。
type Runtime struct {
	Server *server.Server
	AppCtx *plugin.AppContext
}