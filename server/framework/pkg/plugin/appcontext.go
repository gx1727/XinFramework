// Package plugin 定义了 AppContext 类型 —— 这是一个共享的依赖容器，
// 用于替代框架原先依赖的 12 个跨模块包级全局变量（Phase B 重构）。
//
// 设计理念
//
//   - 每个模块既需要"我能消费什么"（数据库、配置、其他模块的仓库），
//     也需要"我能贡献什么"（自己的仓库 / 服务）。
//     将角色拆分为 Reader（读端）和 Writer（写端）让这两种意图在类型层面一目了然。
//
//   - Reader 是每个模块在 Init() 和 Register() 阶段获得的读端句柄。
//     所有字段都是强类型（没有 any 类型断言），
//     当读取到 nil 值时，意味着对应的依赖模块未被启用，这是文档化的检测方式。
//
//   - Writer 是写端句柄，仅提供给拥有特定插槽的模块。
//     没有声明"我贡献 X"的模块不会获得该插槽的 Writer，
//     因此无法意外覆盖其他模块的仓库。
//
//   - AppContext 是一个具体结构体，而非接口。
//     该结构体仅在 boot.Init 中构造一次，然后以指针形式传递。
//     Reader / Writer 接口的存在纯粹是为了让模块契约在编译期即可被强制执行。
package plugin

import (
	"errors"

	"github.com/go-redis/redis/v8"

	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/auth"
	"gx1727.com/xin/framework/pkg/authz"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/session"
	"gx1727.com/xin/framework/pkg/task"
	"gx1727.com/xin/framework/pkg/tenant"
	pkgauth "gx1727.com/xin/framework/pkg/tenant/auth"
)

// Reader 是每个模块可见的读端句柄。它对外暴露以下内容：
//
//   - 基础设施：数据库连接池、Redis 客户端、配置、会话管理器。
//   - 跨模块的服务和仓库：由其他模块在 Init() 阶段填充。
//
// 读端方法若返回强类型指针，则在 cfg.Module 中对应模块未启用时返回 nil。
// 模块在使用仓库前必须自行进行 nil 检查。
//
// 为什么要使用接口，而不是直接使用具体的 AppContext 结构体？
//
//   - 在 Init() 阶段，模块不应能够调用 SetX() 方法 ——
//     这种角色拆分能在编译期就阻止此类误用。
//   - 测试中可以构造一个轻量的伪造 Reader，无需构建完整的 AppContext。
type Reader interface {
	// 基础设施（boot.Init 之后始终存在）。
	DB() *appx.Pool
	Cache() *redis.Client // 当 Redis 禁用时可能返回 nil
	Config() *config.Config
	Session() session.SessionManager

	// 跨模块服务（在所有模块的 Init 之后填充）。
	Authz() authz.Authorization

	// 由其他模块贡献的仓库。当提供方模块未在 cfg.Module 中启用时返回 nil。
	AccountRepo() auth.AccountRepository
	AccountAuthRepo() auth.AccountAuthRepository
	TenantRepo() tenant.TenantRepository
	UserRepo() pkgauth.UserRepository
	RoleRepo() pkgauth.RoleRepository
	OrgRepo() pkgauth.OrganizationRepository
	PermRepo() pkgauth.RoleResourceRepository

	// TaskQueue 长期任务队列（由 apps/task 注入）。Nil 表示 task 模块未启用。
	TaskQueue() task.Queue
}

// Writer 是写端句柄，仅提供给拥有特定插槽的模块。每个 SetX 方法设计为仅由匹配的模块调用一次：
//
//   - SetAuthz             ← framework (boot.Init) 或 apps/<authn-svc>
//   - SetAccountRepo       ← apps/boot/auth
//   - SetAccountAuthRepo   ← apps/boot/auth
//   - SetTenantRepo        ← apps/boot/tenant
//   - SetUserRepo          ← apps/tenant/user
//   - SetRoleRepo          ← apps/tenant/role
//   - SetOrgRepo           ← apps/tenant/organization
//   - SetPermRepo          ← apps/tenant/permission
type Writer interface {
	SetAuthz(a authz.Authorization)
	SetAccountRepo(r auth.AccountRepository)
	SetAccountAuthRepo(r auth.AccountAuthRepository)
	SetTenantRepo(r tenant.TenantRepository)
	SetUserRepo(r pkgauth.UserRepository)
	SetRoleRepo(r pkgauth.RoleRepository)
	SetOrgRepo(r pkgauth.OrganizationRepository)
	SetPermRepo(r pkgauth.RoleResourceRepository)
	SetTaskQueue(q task.Queue)
}

// AppContext 是具体的实现。它在 framework/internal/core/boot.Init 中构造，
// 然后以指针形式传递给每个模块的 Init 和 Register。
//
// 注意：零值不可用 —— 必须由 boot.Init 中的 NewAppContext 构造。
type AppContext struct {
	// 基础设施，Init 之前设置一次。
	db      *appx.Pool
	cache   *redis.Client
	cfg     *config.Config
	session session.SessionManager

	// 跨模块贡献，由模块的 Init() 阶段填充。
	authz_       authz.Authorization
	accountRepo  auth.AccountRepository
	accountAuthR auth.AccountAuthRepository
	tenantRepo   tenant.TenantRepository
	userRepo     pkgauth.UserRepository
	roleRepo     pkgauth.RoleRepository
	orgRepo      pkgauth.OrganizationRepository
	permRepo     pkgauth.RoleResourceRepository
	taskQueue    task.Queue
}

// NewAppContext 构造 AppContext 并预填充基础设施插槽。
// 其他插槽由模块的 Init() 阶段填充。
//
// pool / config 必须非 nil（pool 应是 appx.MustNewPool 的产物）；
// cache / session 可能为 nil，仅当对应的子系统在启动时被显式禁用。
func NewAppContext(
	db *appx.Pool,
	cache *redis.Client,
	cfg *config.Config,
	session session.SessionManager,
) (*AppContext, error) {
	if db == nil {
		return nil, errors.New("NewAppContext: db pool cannot be nil")
	}
	if cfg == nil {
		return nil, errors.New("NewAppContext: config cannot be nil")
	}
	return &AppContext{
		db:      db,
		cache:   cache,
		cfg:     cfg,
		session: session,
	}, nil
}

// Compile-time assertions: ensure *AppContext satisfies both Reader and Writer.
var (
	_ Reader = (*AppContext)(nil)
	_ Writer = (*AppContext)(nil)
)

// --- Reader ---

// DB returns the underlying database connection pool (强类型包装，构造期必非空)。
//
// 业务模块用 ctx.DB().Raw() 拿原生 *pgxpool.Pool 传给 Repository 构造函数。
// 不需要 nil-check（构造期已保证非空）。
func (a *AppContext) DB() *appx.Pool { return a.db }

// Cache returns the Redis client (may be nil if Redis is disabled).
func (a *AppContext) Cache() *redis.Client { return a.cache }

// Config returns the global config object.
func (a *AppContext) Config() *config.Config { return a.cfg }

// Session returns the session manager.
func (a *AppContext) Session() session.SessionManager { return a.session }

// Authz returns the authorization service.
func (a *AppContext) Authz() authz.Authorization { return a.authz_ }

// AccountRepo returns the account repository (injected by apps/boot/auth).
func (a *AppContext) AccountRepo() auth.AccountRepository {
	return a.accountRepo
}

// AccountAuthRepo returns the account auth repository (injected by apps/boot/auth).
func (a *AppContext) AccountAuthRepo() auth.AccountAuthRepository {
	return a.accountAuthR
}

// TenantRepo returns the tenant repository (injected by apps/boot/tenant).
func (a *AppContext) TenantRepo() tenant.TenantRepository {
	return a.tenantRepo
}

// UserRepo returns the user repository (injected by apps/tenant/user).
func (a *AppContext) UserRepo() pkgauth.UserRepository { return a.userRepo }

// RoleRepo returns the role repository (injected by apps/tenant/role).
func (a *AppContext) RoleRepo() pkgauth.RoleRepository { return a.roleRepo }

// OrgRepo returns the organization repository (injected by apps/tenant/organization).
func (a *AppContext) OrgRepo() pkgauth.OrganizationRepository { return a.orgRepo }

// PermRepo returns the role-resource (permission) repository (injected by apps/tenant/permission).
func (a *AppContext) PermRepo() pkgauth.RoleResourceRepository { return a.permRepo }

// TaskQueue 实现 Reader.TaskQueue。
//
// nil 表示未注入 task 模块，调用方需自行 nil-check 跳过。
func (a *AppContext) TaskQueue() task.Queue { return a.taskQueue }

// --- Writer ---

// SetAuthz sets the authorization service.
func (a *AppContext) SetAuthz(v authz.Authorization) { a.authz_ = v }

// SetAccountRepo sets the account repository.
func (a *AppContext) SetAccountRepo(v auth.AccountRepository) { a.accountRepo = v }

// SetAccountAuthRepo sets the account auth repository.
func (a *AppContext) SetAccountAuthRepo(v auth.AccountAuthRepository) { a.accountAuthR = v }

// SetTenantRepo sets the tenant repository.
func (a *AppContext) SetTenantRepo(v tenant.TenantRepository) { a.tenantRepo = v }

// SetUserRepo sets the user repository.
func (a *AppContext) SetUserRepo(v pkgauth.UserRepository) { a.userRepo = v }

// SetRoleRepo sets the role repository.
func (a *AppContext) SetRoleRepo(v pkgauth.RoleRepository) { a.roleRepo = v }

// SetOrgRepo sets the organization repository.
func (a *AppContext) SetOrgRepo(v pkgauth.OrganizationRepository) { a.orgRepo = v }

// SetPermRepo sets the role-resource (permission) repository.
func (a *AppContext) SetPermRepo(v pkgauth.RoleResourceRepository) { a.permRepo = v }

// SetTaskQueue 注入任务队列。供 apps/task 模块在 Init 阶段调用。
func (a *AppContext) SetTaskQueue(v task.Queue) { a.taskQueue = v }
