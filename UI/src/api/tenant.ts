// Sys 租户管理（super_admin 域）
//
// 与 apps/sys/tenants 后端对齐：
//   GET    /sys/tenants
//   GET    /sys/tenants/:id
//   POST   /sys/tenants
//   PUT    /sys/tenants/:id
//   PUT    /sys/tenants/:id/status
//   DELETE /sys/tenants/:id
//   POST   /sys/tenants/:id/purge
//
// 全部强制 super_admin（group 级 RequireAnySysRole）+ ResTenant.* 双层守卫。

import { api, type PageResponse } from "./common"

export interface TenantItem {
  id: number
  code: string
  name: string
  status: number
  contact?: string
  phone?: string
  email?: string
  province?: string
  city?: string
  area?: string
  address?: string
  admin_account_id?: number
  admin_username?: string
  user_count?: number
  created_at?: string
  updated_at?: string
}

export interface PurgeTenantResponse {
  tenant_id: number
  code: string
  tables_purged: number
  tables: Record<string, number>
}

export const tenantApi = {
  list: (params?: {
    page?: number
    size?: number
    keyword?: string
    status?: number
  }) => api<PageResponse<TenantItem>>("/sys/tenants", { params }),

  get: (id: number) => api<TenantItem>(`/sys/tenants/${id}`),

  create: (data: Partial<TenantItem> & { admin_account_id?: number }) =>
    api<TenantItem>("/sys/tenants", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<TenantItem>) =>
    api(`/sys/tenants/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  /** 更新状态（启用/停用） */
  updateStatus: (id: number, status: 0 | 1) =>
    api<TenantItem>(`/sys/tenants/${id}/status`, {
      method: "PUT",
      body: JSON.stringify({ status }),
    }),

  /** 软删（先软删后才能硬删） */
  delete: (id: number) =>
    api(`/sys/tenants/${id}`, {
      method: "DELETE",
    }),

  /** 硬删（不可逆，需先软删） */
  purge: (id: number) =>
    api<PurgeTenantResponse>(`/sys/tenants/${id}/purge`, { method: "POST" }),

  /**
   * 平台管理员（super_admin）模拟登录到指定租户。
   * 返回与 /auth/tenant-login 同构的 token 响应；前端应保存原 sys refresh_token
   * 用于"退出模拟"时调 /auth/refresh 恢复。
   *
   * 后端：POST /api/v1/sys/tenants/:id/impersonate
   */
  impersonate: (id: number) =>
    api<{
      scope: string
      token: string
      refresh_token: string
      expires_in: number
      tenant_id: number
      tenant_name: string
      impersonated_user_id: number
      impersonated_by: number
      impersonation_sid: string
    }>(`/sys/tenants/${id}/impersonate`, { method: "POST" }),
}
