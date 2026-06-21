// 认证（登录/注册/刷新/登出）
//
// Phase 0022 拆分：登录入口按 scope 拆开
//   - tenantLogin    → POST /auth/tenant-login   （业务域，需 tenant_id）
//   - platformLogin  → POST /auth/platform-login （平台域，无需 tenant_id，要求 super_admin）

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
}

export interface RefreshResponse {
  token: string
  refresh_token?: string
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