package xincontext

// 本文件定义业务领域强类型 ID（alias 形式，渐进式落地）。
//
// # 设计原则：alias vs named type
//
// 本阶段使用 **type alias**（`type TenantID = uint`）而非 named type（`type TenantID uint`）：
//   - alias：与底层类型完全等价，可直接传给接收 uint 的函数，**零迁移成本**
//   - named：必须显式 .Uint() 转换，50+ 调用点级联修改
//
// alias 的代价是"不防误用"——但通过下面两层防御补充：
//   1. 关键函数签名（auth.Service.Login / permission.BuildDataScopeFilter）改用 alias，
//      文档 / lint 工具 / code review 共同防止 TenantID 误传给 userID 参数
//   2. V2 阶段：把 alias 升级为 named（单点替换 types.go），编译器立刻指出所有误用点
//
// # 转换约定
//   - 强类型 ↔ uint：直接赋值（同为 uint 别名）
//   - JSON 序列化：wire format 仍是 number，无破坏
//   - 数据库 scan：保持 uint，service 层做 alias 转换
//
// # 约束
//   - alias 类型不能定义 method（Go 语法限制）
//   - 校验函数用 `IsValidXxxID(uint) bool` 形式提供
//
// # 类型清单
//
//   TenantID     - 租户 ID（对应 tenants / tenant_users.tenant_id）
//   UserID       - 用户 ID（对应 tenant_users.id 或 sys_users.id）
//   AccountID    - 账号 ID（对应 accounts.id，跨 tenant 唯一）
//   OrgID        - 组织 ID（对应 tenant_organizations.id）
//   RoleID       - 角色 ID（对应 tenant_roles.id 或 sys_roles.id）
//   SessionID    - 会话 ID（对应 auth_sessions.id / Redis key），string 强类型

// TenantID 租户 ID（对应 accounts / tenants / tenant_users.tenant_id）。
//
// 业务边界：所有租户域数据查询必须带 TenantID 上下文。
// 文档约定：仅在租户域上下文中使用，禁止与 UserID 互换。
type TenantID = uint

// NewTenantID 从 uint 构造（用于 HTTP / DB 层）。
func NewTenantID(v uint) TenantID { return v }

// IsValidTenantID 报告 TenantID 是否非零（0 = 平台域 / 未指定）。
func IsValidTenantID(t TenantID) bool { return t != 0 }

// UserID 用户 ID（对应 tenant_users.id 或 sys_users.id）。
//
// 文档约定：platform admin 登录时 UserID 字段实际是 account_id（语义复用）。
type UserID = uint

// NewUserID 从 uint 构造。
func NewUserID(v uint) UserID { return v }

// IsValidUserID 报告 UserID 是否非零。
func IsValidUserID(u UserID) bool { return u != 0 }

// AccountID 账号 ID（对应 accounts.id，跨 tenant 唯一）。
//
// 业务场景：登录 subject / login_history.account_id / RecipientResolver 反查手机/邮箱。
type AccountID = uint

// NewAccountID 从 uint 构造。
func NewAccountID(v uint) AccountID { return v }

// IsValidAccountID 报告 AccountID 是否非零。
func IsValidAccountID(a AccountID) bool { return a != 0 }

// OrgID 组织 ID（对应 tenant_organizations.id）。
//
// 文档约定：仅用于组织树相关 API，禁止与 UserID 互换。
type OrgID = int64

// NewOrgID 从 int64 构造。
func NewOrgID(v int64) OrgID { return v }

// IsValidOrgID 报告 OrgID 是否非零。
func IsValidOrgID(o OrgID) bool { return o != 0 }

// RoleID 角色 ID（对应 tenant_roles.id 或 sys_roles.id）。
type RoleID = uint

// NewRoleID 从 uint 构造。
func NewRoleID(v uint) RoleID { return v }

// IsValidRoleID 报告 RoleID 是否非零。
func IsValidRoleID(r RoleID) bool { return r != 0 }

// SessionID 会话 ID（对应 auth_sessions.id / Redis key / JWT claims.sid）。
//
// 强类型 string：编译期防"任意 string 误传为 session id"。
type SessionID = string

// NewSessionID 从 string 构造（用于 HTTP / Redis）。
func NewSessionID(s string) SessionID { return s }

// IsValidSessionID 报告 SessionID 是否非空。
func IsValidSessionID(s SessionID) bool { return s != "" }
