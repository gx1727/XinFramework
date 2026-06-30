// 认证（登录/注册/刷新/登出）
//
// Phase 0022 拆分：登录入口按 scope 拆开
//   - tenantLogin    → POST /auth/tenant-login   （业务域，需 tenant_id）
//   - platformLogin  → POST /auth/platform-login （平台域，无需 tenant_id，要求 super_admin）
//
// Phase 0024 路径 B：多身份账号登录
//   - loginPrecheck  → POST /auth/login-precheck （账号+密码 → 列出所有可用身份）
//   - selectTenant   → POST /auth/select-tenant  （选 tenant 身份签 token）
//   - refresh        → POST /auth/refresh         （可选 tenant_id 切租户）

import { api } from "./common"

export interface LoginRequest {
  account: string
  password: string
  tenant_id?: number
}

/** 登录作用域：tenant（业务租户） / platform（平台管理员） */
export type LoginScope = "tenant" | "platform"

export interface LoginResponse {
  /** 登录作用域，与后端 LoginResult.Scope 对齐 */
  scope: LoginScope
  token: string
  refresh_token: string
  user: {
    id: number
    /** scope=platform 时固定为 0；scope=tenant 时为真实 tenant_id */
    tenant_id: number
    code: string
    role: string
    nickname?: string
    real_name?: string
    avatar?: string
    email?: string
    platform_roles?: string[]
    /**
     * 资源权限码列表（"resource:action" 形式，如 "menu:create"、"user:list"）。
     * 0024+：登录响应一次下发，前端用作按钮可见性与路由守门，
     * 与后端 Require(P(Res, Act)) 用同一份数据推导，避免 round-trip。
     * 不存在 = 零权限（不用纠结 [] vs undefined）。
     */
    permissions?: string[]
  }
}

export interface PlatformLoginRequest {
  account: string
  password: string
}

export interface RegisterRequest {
  account: string
  password: string
  tenant_id?: number
  real_name: string
}

export interface RefreshRequest {
  refresh_token: string
  /** 可选：切租户时传目标 tenant_id（路径 B 多身份支持） */
  tenant_id?: number
}

export interface RefreshResponse {
  token: string
  refresh_token?: string
}

/** 账号在某个租户内的身份记录（路径 B 多身份支持） */
export interface TenantIdentity {
  tenant_id: number
  tenant_code: string
  tenant_name: string
  user_id: number
  user_code: string
  role: string
  nickname?: string
  real_name?: string
  avatar?: string
  email?: string
}

/** 登录前置检查响应：不签 token，只列出账号可用身份 */
export interface LoginPrecheckResponse {
  account_id: number
  account_status: number
  real_name?: string
  email?: string
  /** 是否具备平台角色（如 super_admin），true 时前端可调 platform-login */
  platform_available: boolean
  platform_roles?: string[]
  /** 账号在所有租户的 users 身份列表；空数组 + platform_available=false → 无登录权限 */
  tenant_identities: TenantIdentity[]
}

export const authApi = {
  /** 租户域登录（业务用户）。必须传 tenant_id。 */
  tenantLogin: (data: LoginRequest) =>
    api<LoginResponse>("/auth/tenant-login", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  /** 平台域登录（super_admin）。无 tenant_id；后端校验账号有平台角色。 */
  platformLogin: (data: PlatformLoginRequest) =>
    api<LoginResponse>("/auth/platform-login", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  /**
   * 登录前置检查（路径 B 多身份支持）。
   * 输入账号密码，返回账号在所有租户的用户身份 + 平台角色列表。
   * 不签 token，前端根据返回结果决定下一步调用哪个登录入口。
   *
   * 后端错误码：
   *   - 401: 账号/密码错
   *   - 403 (1015): 账号无 tenant 身份且无 platform 角色
   */
  loginPrecheck: (data: { account: string; password: string }) =>
    api<LoginPrecheckResponse>("/auth/login-precheck", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  /**
   * 选择一个 tenant 身份签发 token（路径 B 多身份支持）。
   * 等价于 tenantLogin，区别在于语义化入口（"precheck 后选了某个身份"）。
   */
  selectTenant: (data: {
    account: string
    password: string
    tenant_id: number
  }) =>
    api<LoginResponse>("/auth/select-tenant", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  register: (data: RegisterRequest) =>
    api<LoginResponse>("/auth/register", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  refresh: (data: RefreshRequest) =>
    api<RefreshResponse>("/auth/refresh", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  logout: () =>
    api("/auth/logout", {
      method: "POST",
    }),
}
