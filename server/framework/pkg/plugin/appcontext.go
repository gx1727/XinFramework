// Package plugin 定义了 AppContext 类型 —— 这是一个共享的依赖容器，
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
//   - Writer 是写端句柄，只提供给拥有特定插槽的模块。
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
	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/framework/pkg/auth"
	"gx1727.com/xin/framework/pkg/authz"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/session"
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
	DB() *pgxpool.Pool
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
}

// Writer 是写端句柄，仅提供给拥有特定插槽的模块。
// 每个 SetX 方法设计为最多调用一次，由与插槽名称匹配的模块调用：
//
//   - SetAuthz             ← framework（boot.Init）或 apps/<authn-svc>
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
}

// AppContext 是具体的实现。它在 framework/internal/core/boot.Init 中构造，
// 并以指针形式传递给每个模块的 Init 和 Register。
//
// 注意：零值不可用 —— 必须在 boot.Init 中调用 NewAppContext 进行构造。
type AppContext struct {
	// 基础设施，在模块 Init 之前设置一次。
	db      *pgxpool.Pool
	cache   *redis.Client
	cfg     *config.Config
	session session.SessionManager

	// 跨模块贡献，在各模块 Init() 阶段设置。
	authz_       authz.Authorization
	accountRepo  auth.AccountRepository
	accountAuthR auth.AccountAuthRepository
	tenantRepo   tenant.TenantRepository
	userRepo     pkgauth.UserRepository
	roleRepo     pkgauth.RoleRepository
	orgRepo      pkgauth.OrganizationRepository
	permRepo     pkgauth.RoleResourceRepository
}

// NewAppContext 构造 AppContext 并预先填充基础设施插槽。
// 其余插槽将在各模块 Init() 阶段填充。
//
// db 连接池和 config 必须非空。cache 和 session 仅在引导时
// 明确禁用相应子系统时才允许为 nil（例如关闭 Redis 并使用 DB 会话管理器）。
func NewAppContext(
	db *pgxpool.Pool,
	cache *redis.Client,
	cfg *config.Config,
	session session.SessionManager,
) (*AppContext, error) {
	if db == nil {
		return nil, errors.New("NewAppContext: db 连接池不能为空")
	}
	if cfg == nil {
		return nil, errors.New("NewAppContext: config 不能为空")
	}
	return &AppContext{
		db:      db,
		cache:   cache,
		cfg:     cfg,
		session: session,
	}, nil
}

// 编译期断言：确保 *AppContext 同时满足 Reader 和 Writer 两个接口。
var (
	_ Reader = (*AppContext)(nil)
	_ Writer = (*AppContext)(nil)
)

// --- Reader ---

// DB 返回底层数据库连接池。
func (a *AppContext) DB() *pgxpool.Pool { return a.db }

// Cache 返回 Redis 客户端（若禁用 Redis 则可能为 nil）。
func (a *AppContext) Cache() *redis.Client { return a.cache }

// Config 返回全局配置对象。
func (a *AppContext) Config() *config.Config { return a.cfg }

// Session 返回会话管理器。
func (a *AppContext) Session() session.SessionManager { return a.session }

// Authz 返回鉴权（授权）服务。
func (a *AppContext) Authz() authz.Authorization { return a.authz_ }

// AccountRepo 返回账户仓库（由 apps/boot/auth 注入）。
func (a *AppContext) AccountRepo() auth.AccountRepository {
	return a.accountRepo
}

// AccountAuthRepo 返回账户鉴权仓库（由 apps/boot/auth 注入）。
func (a *AppContext) AccountAuthRepo() auth.AccountAuthRepository {
	return a.accountAuthR
}

// TenantRepo 返回租户仓库（由 apps/boot/tenant 注入）。
func (a *AppContext) TenantRepo() tenant.TenantRepository {
	return a.tenantRepo
}

// UserRepo 返回用户仓库（由 apps/tenant/user 注入）。
func (a *AppContext) UserRepo() pkgauth.UserRepository { return a.userRepo }

// RoleRepo 返回角色仓库（由 apps/tenant/role 注入）。
func (a *AppContext) RoleRepo() pkgauth.RoleRepository { return a.roleRepo }

// OrgRepo 返回组织仓库（由 apps/tenant/organization 注入）。
func (a *AppContext) OrgRepo() pkgauth.OrganizationRepository { return a.orgRepo }

// PermRepo 返回角色-资源（权限）仓库（由 apps/tenant/permission 注入）。
func (a *AppContext) PermRepo() pkgauth.RoleResourceRepository { return a.permRepo }

// --- Writer ---

// SetAuthz 设置鉴权服务。
func (a *AppContext) SetAuthz(v authz.Authorization) { a.authz_ = v }

// SetAccountRepo 设置账户仓库。
func (a *AppContext) SetAccountRepo(v auth.AccountRepository) { a.accountRepo = v }

// SetAccountAuthRepo 设置账户鉴权仓库。
func (a *AppContext) SetAccountAuthRepo(v auth.AccountAuthRepository) { a.accountAuthR = v }

// SetTenantRepo 设置租户仓库。
func (a *AppContext) SetTenantRepo(v tenant.TenantRepository) { a.tenantRepo = v }

// SetUserRepo 设置用户仓库。
func (a *AppContext) SetUserRepo(v pkgauth.UserRepository) { a.userRepo = v }

// SetRoleRepo 设置角色仓库。
func (a *AppContext) SetRoleRepo(v pkgauth.RoleRepository) { a.roleRepo = v }

// SetOrgRepo 设置组织仓库。
func (a *AppContext) SetOrgRepo(v pkgauth.OrganizationRepository) { a.orgRepo = v }

// SetPermRepo 设置角色-资源（权限）仓库。
func (a *AppContext) SetPermRepo(v pkgauth.RoleResourceRepository) { a.permRepo = v }
