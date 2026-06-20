// 租户

import { api, type PageResponse } from "./common"

export interface TenantItem {
  id: number
  code: string
  name: string
  status: number
  admin_account_id?: number
  admin_username?: string
  user_count?: number
  created_at?: string
  updated_at?: string
}

export const tenantApi = {
  list: (params?: { page?: number; size?: number; keyword?: string; status?: number }) =>
    api<PageResponse<TenantItem>>("/tenants", { params }),

  get: (id: number) =>
    api<TenantItem>(`/tenants/${id}`),

  create: (data: Partial<TenantItem> & { admin_account_id?: number }) =>
    api<TenantItem>("/tenants", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<TenantItem>) =>
    api(`/tenants/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  /** 更新状态（启用/停用） */
  updateStatus: (id: number, status: 0 | 1) =>
    api<TenantItem>(`/tenants/${id}/status`, {
      method: "PUT",
      body: JSON.stringify({ status }),
    }),

  /** 软删（先软删后才能硬删） */
  delete: (id: number) =>
    api(`/tenants/${id}`, {
      method: "DELETE",
    }),

  /** 硬删（不可逆，需先软删） */
  purge: (id: number) =>
    api<{ tenant_id: number; code: string; tables_purged: number; tables: Record<string, number> }>(
      `/tenants/${id}/purge`,
      { method: "POST" }
    ),
}
