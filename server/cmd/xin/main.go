// 流程：
//  1. config.Load("config/config.yaml")
//  2. boot.Init(cfg) → *appx.App
//  3. 显式构造 []plugin.Module 列表，调用各模块的 Module(app) 工厂
//  4. framework.Serve(cfg, app, modules) 启动服务
//
// 没有任何 side-effect import 或全局注册表，每个模块的依赖都在这里一目了然。
package main

import (
	"log"

	"gx1727.com/xin/apps/boot/auth"
	"gx1727.com/xin/apps/cms"
	"gx1727.com/xin/apps/flag"
	platformmenu "gx1727.com/xin/apps/platform/menu"
	"gx1727.com/xin/apps/platform/tenant"
	"gx1727.com/xin/apps/rbac/menu"
	"gx1727.com/xin/apps/rbac/organization"
	"gx1727.com/xin/apps/rbac/permission"
	"gx1727.com/xin/apps/rbac/resource"
	"gx1727.com/xin/apps/rbac/role"
	"gx1727.com/xin/apps/rbac/user"
	"gx1727.com/xin/apps/reference/asset"
	refconfig "gx1727.com/xin/apps/reference/config"
	"gx1727.com/xin/apps/reference/dict"
	"gx1727.com/xin/apps/reference/weixin"
	"gx1727.com/xin/apps/system"
	"gx1727.com/xin/framework"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/plugin"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	app, rt, err := framework.Boot(cfg)
	if err != nil {
		log.Fatalf("boot init failed: %v", err)
	}

	// 显式 Build：列出会用到的所有模块。
	// 顺序无关，framework.Serve 会按 cfg.Module 中的 enable 列表过滤。
	modules := []plugin.Module{
		// boot 阶段
		auth.Module(app),

		// 平台管理域（必须 super_admin 才能访问）
		platformmenu.Module(app),
		tenant.Module(app),

		// rbac 套件
		menu.Module(app),
		organization.Module(app),
		permission.Module(app),
		resource.Module(app),
		role.Module(app),
		user.Module(app),

		// reference 套件
		asset.Module(app),
		refconfig.Module(app),
		dict.Module(app),
		weixin.Module(app),

		// system
		system.Module(app),

		// external
		cms.Module(app),
		flag.Module(app),
	}

	framework.Serve(cfg, app, rt, modules)
}
