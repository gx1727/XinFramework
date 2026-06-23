// 流程：
//  1. config.Load("config/config.yaml")
//  2. boot.Init(cfg) → *appx.App
//  3. 显式构造 []plugin.Module 列表，调用各模块的 Module(app) 工厂
//  4. framework.Serve(cfg, app, modules) 启动服务
//
// 没有任何 side-effect import 或全局注册表，每个模块的依赖都在这里一目了然。
//
// Phase 0023 全阶段已完成（2026-06-23）：
//   - 0023.0：新增 sys_user / sys_role / sys_menu / sys_permission 四个 platform 域模块
//   - 0023.1：数据迁移 + account_roles drop；admin 走 sys_users + sys_user_roles
//   - 0023.2：登录路径切到 sys_user_roles + sys_roles.code = 'super_admin'
//   - 0023.3：Go 包重命名（apps/rbac → apps/tenant、framework/pkg/rbac → framework/pkg/tenant/auth）、
//     SQL 表 7 张 rename（users → tenant_users 等）+ resources → tenant_permissions +
//     menus 物理拆 tenant_menus / sys_menus
//   - 0023.4：apps/platform/menu 模块删除（broken：旧 tenant_menus WHERE tenant_id=0 已失效），
//     由 apps/platform/sys_menu 接管 /platform/sys-menus 路由
//   - 0023.5：文档同步（architecture.md / modules.md / AGENTS.md / migrations/README.md）
//
// 终态分层：
//   - 平台域 sys_*（无 tenant_id、不启用 RLS）
//   - 租户域 tenant_*（带 tenant_id + RLS）
//   - 共享层 accounts / tenants / auth_sessions
package main

import (
	"log"

	"gx1727.com/xin/apps/boot/auth"
	"gx1727.com/xin/apps/cms"
	"gx1727.com/xin/apps/flag"
	sysmenu "gx1727.com/xin/apps/platform/sys_menu"
	syspermission "gx1727.com/xin/apps/platform/sys_permission"
	sysrole "gx1727.com/xin/apps/platform/sys_role"
	sysuser "gx1727.com/xin/apps/platform/sys_user"
	"gx1727.com/xin/apps/platform/tenants"
	"gx1727.com/xin/apps/tenant/menu"
	"gx1727.com/xin/apps/tenant/organization"
	"gx1727.com/xin/apps/tenant/permission"
	"gx1727.com/xin/apps/tenant/resource"
	"gx1727.com/xin/apps/tenant/role"
	"gx1727.com/xin/apps/tenant/user"
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
		tenants.Module(app),         // 管 tenants 表
		sysuser.Module(app),        // Phase 0023.0：sys_users 表
		sysrole.Module(app),        // Phase 0023.0：sys_roles 表
		sysmenu.Module(app),        // Phase 0023.0：sys_menus 表
		syspermission.Module(app),  // Phase 0023.0：sys_permissions 表

		// rbac 套件（租户域）
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
