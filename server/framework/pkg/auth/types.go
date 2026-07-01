// Package auth 暴露业务模块依赖的公开鉴权契约。
//
// 具体的 AccountRepository / AccountAuthRepository 实现在 apps/boot/auth
// （原来在 framework/internal/module/auth，Phase 2 重构后迁移到 apps）。
// 业务模块（如 user）不能直接 import apps/*，因此只能看到 framework/pkg/auth
// 中的接口。实现通过 plugin.AppContext.SetAccountRepo 在启动期注入。
//
// 为什么要公开接口、却把实现放在 apps？
//   - 业务模块能合法 import framework 包、不能 import apps 包
//   - 接口公开在 framework，实现具体放在 apps，两边都能访问
//   - 实现可插拔：未来想替换实现（比如迁到 LDAP），只需重写 apps/boot/auth
//   - 字段重复：Account / AccountAuth 结构体在两边都有，字段类型必须一致
//     以保证接口隐式满足
package auth

import (
	"context"
	"time"
)

// Account 是跨租户的全局账号记录。结构与 apps/boot/auth 中的同名结构一致，
// 这里重复定义是为了让 framework 包不依赖 apps 包。
type Account struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	RealName  string    `json:"real_name"`
	Status    int8      `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AccountAuth 是第三方授权绑定记录（微信/QQ/微博）。
type AccountAuth struct {
	ID         uint      `json:"id"`
	TenantID   uint      `json:"tenant_id"`
	AccountID  uint      `json:"account_id"`
	Type       string    `json:"type"`
	OpenID     string    `json:"openid"`
	UnionID    string    `json:"unionid"`
	Nickname   string    `json:"nickname"`
	Avatar     string    `json:"avatar"`
	SessionKey string    `json:"session_key"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TenantIdentity 账号在某个租户内的身份记录。
//
// 路径 B 下，一个账号（accounts.id）可以在多个租户各有一条 users 记录，
// 每个 (account_id, tenant_id) 对应一个独立身份。LoginPrecheck 用它
// 列出账号可用身份供前端选择。
type TenantIdentity struct {
	TenantID   uint   `json:"tenant_id"`
	TenantCode string `json:"tenant_code"`
	TenantName string `json:"tenant_name"`
	UserID     uint   `json:"user_id"`
	UserCode   string `json:"user_code"`
	Role       string `json:"role"`
	Nickname   string `json:"nickname,omitempty"`
	RealName   string `json:"real_name,omitempty"`
	Avatar     string `json:"avatar,omitempty"`
	Email      string `json:"email,omitempty"`
}

// AccountRepository 是业务模块可依赖的账号数据访问接口子集。
// 具体实现在 apps/boot/auth，依靠字段类型一致隐式满足本接口。
type AccountRepository interface {
	GetByID(ctx context.Context, id uint) (*Account, error)
	GetByUsername(ctx context.Context, username string) (*Account, error)
	GetByPhone(ctx context.Context, phone string) (*Account, error)
	GetByEmail(ctx context.Context, email string) (*Account, error)
	Create(ctx context.Context, username, phone, email, realName, passwordHash string) (*Account, error)
	Exists(ctx context.Context, account string) (bool, error)
	// GetPasswordAndStatus 取账号的 password_hash + id + status，用于 sys-login。
	// 注意：account 在 DB 中可能是 username / phone / email 任一，按顺序匹配。
	GetPasswordAndStatus(ctx context.Context, account string) (passwordHash string, accountID uint, status int8, err error)
	// GetAccountIDByUserID 通过 user_id 反查 account_id。
	// 用于 Refresh 切租户流程：先查 account_id，再用 ListTenantIdentities 跨租户列身份。
	//
	// RLS 说明：调用方需保证 ctx 处在能查到该 user 行的租户上下文里
	// （即 user_id 来自的租户事务）；不在事务里调用会被 RLS 拒绝。
	GetAccountIDByUserID(ctx context.Context, userID uint) (uint, error)
	// ListTenantIdentities 列出账号在所有租户的用户身份。
	//
	// 路径 B 下，accounts 不绑 tenant，但 users 表的 (account_id, tenant_id)
	// 可以有多条记录 —— 一个账号可以是多个租户的用户。
	// 该方法返回账号在所有未删除租户里的 users 身份，用于 LoginPrecheck。
	//
	// RLS 说明：本方法跨租户查 users / tenants 表（两表都启用了 RLS），
	// 实现必须走 db.RunInSysTx（设 app.bypass_rls='on'），
	// 或等效的绕过机制。
	ListTenantIdentities(ctx context.Context, accountID uint) ([]TenantIdentity, error)
}

// AccountAuthRepository 是第三方授权绑定的数据访问接口。
type AccountAuthRepository interface {
	GetByOpenID(ctx context.Context, tenantID uint, authType, openID string) (*AccountAuth, error)
	GetByAccountID(ctx context.Context, accountID uint) ([]AccountAuth, error)
	Create(ctx context.Context, tenantID, accountID uint, authType, openID, unionID, sessionKey string) (*AccountAuth, error)
	UpdateSessionKey(ctx context.Context, id uint, sessionKey string) error
	Delete(ctx context.Context, id uint) error
}
