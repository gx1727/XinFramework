package auth

import (
	"time"

	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/login_security"
	"gx1727.com/xin/framework/pkg/permission"
	"gx1727.com/xin/framework/pkg/plugin"
	pkgtenant "gx1727.com/xin/framework/pkg/tenant"
)

// Module returns the auth module as a BaseModule.
func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "auth",
		InitFn: func(_ plugin.Reader, w plugin.Writer) error {
			pool := app.DB.Raw()
			w.SetAccountRepo(NewAccountRepository(pool))
			w.SetAccountAuthRepo(NewAccountAuthRepository(pool))
			return nil
		},
		RegFn: func(ctx plugin.Reader, slots plugin.RouterSlots) {
			public := slots.MustGet(plugin.SlotPublic).Group
			protected := slots.MustGet(plugin.SlotProtected).Group
			pool := app.DB.Raw()
			if ctx != nil {
				if p := ctx.DB().Raw(); p != nil {
					pool = p
				}
			}

			// 跨模块依赖（tenant repo）从 AppContext 拿，遵循 DI 设计意图。
			// auth 与 tenants 都是 alwaysOn 模块，理论 ctx.TenantRepo() 必非 nil；
			// 若为 nil（启动顺序异常），拒绝注册路由，避免后续 panic。
			tenantRepo := pkgtenant.TenantRepository(nil)
			if ctx != nil {
				tenantRepo = ctx.TenantRepo()
			}
			if tenantRepo == nil {
				return
			}

			repos := Repositories{
				Account:  NewAccountRepository(pool),
				Tenant:   tenantRepo,
				Platform: permission.NewPlatformRoleRepository(pool),
				// PermissionLoader 复用中间件同款 PostgresPermissionRepository，
				// 保证登录响应里的 Permissions 与运行时段 Require(P(Res, Act)) 判定完全一致。
				// 不走 PermissionService 是因为 PermissionService 在 framework/internal/service
				// （internal 限制 apps/boot/auth 不能直接 import），这里只用其底层 repo
				// 走一次性 GetUserPermissions，结果与中间件懒加载路径等价。
				PermLoader: permission.NewPermissionRepository(pool),
			}

			// 装配登录安全服务（账号锁定 + 异地告警）。
			// 不传 RecipientResolver 时所有告警走 LogNotifier 仅写日志，
			// 业务模块可后续注入带邮件/短信通道的 Resolver 实现。
			securityCfg := buildSecurityConfig(app)
			security := login_security.NewSecurityService(
				securityCfg,
				login_security.NewPGLockManager(pool),
				login_security.NewPGAttemptStore(pool),
				login_security.NewPGHistoryRecorder(pool),
				nil, // notifier 留 nil → 自动 fallback 到 LogNotifier
				nil, // recipients resolver 留 nil → 告警发送会降级为不发送
			)

			deps := DefaultDependencies(app.Config, pool, repos, security)
			h := NewHandler(NewService(deps))
			Register(public, protected, h)
		},
	}
}

// buildSecurityConfig 从全局 config 装配 SecurityConfig（懒转换）。
func buildSecurityConfig(app *appx.App) login_security.SecurityConfig {
	if app == nil || app.Config == nil {
		return login_security.SecurityConfig{Enabled: false}
	}
	ls := app.Config.LoginSecurity
	if !ls.Enabled {
		return login_security.SecurityConfig{Enabled: false}
	}
	c := login_security.DefaultSecurityConfig()
	if ls.MaxFailedAttempts > 0 {
		c.MaxFailedAttempts = ls.MaxFailedAttempts
	}
	if ls.LockDurationMin > 0 {
		c.LockDuration = time.Duration(ls.LockDurationMin) * time.Minute
	}
	if ls.FailureWindowMin > 0 {
		c.FailureWindow = time.Duration(ls.FailureWindowMin) * time.Minute
	}
	if ls.IPFailureThreshold > 0 {
		c.IPFailureThreshold = ls.IPFailureThreshold
	}
	if ls.IPFailureWindowMin > 0 {
		c.IPFailureWindow = time.Duration(ls.IPFailureWindowMin) * time.Minute
	}
	if ls.AnomalyHistoryLimit > 0 {
		c.AnomalyHistoryLimit = ls.AnomalyHistoryLimit
	}
	c.AnomalyDeviceMatch = ls.AnomalyDeviceMatch
	c.AnomalyNotifyInSite = ls.AnomalyNotifyInSite
	c.AnomalyNotifyEmail = ls.AnomalyNotifyEmail
	c.AnomalyNotifySMS = ls.AnomalyNotifySMS
	c.LockNotifyInSite = ls.LockNotifyInSite
	c.LockNotifyEmail = ls.LockNotifyEmail
	c.LockNotifySMS = ls.LockNotifySMS
	return c
}
