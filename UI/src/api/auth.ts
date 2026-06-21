// 认证（登录/注册/刷新/登出）

import { api } from "./common"

export interface LoginRequest {
  account: string
  password: string
  tenant_id?: number
}

export interface LoginResponse {
  token: string
  refresh_token: string
  user: {
    id: number
    tenant_id: number
    code: string
    role: string
    // 展示资料（侧边栏 / NavUser 用）
    nickname?: string
    real_name?: string
    avatar?: string
    email?: string
    // 平台级角色（与后端 User.PlatformRoles 对齐）；空数组表示非平台角色
    platform_roles?: string[]
  }
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
  login: (data: LoginRequest) =>
    api<LoginResponse>("/auth/login", {
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
