// 平台租户管理（super_admin 域）
//
// 与 apps/admin/platform_tenant 后端对齐：
//   GET    /platform/tenants
//   GET    /platform/tenants/:id
//   POST   /platform/tenants
//   PUT    /platform/tenants/:id
//   PUT    /platform/tenants/:id/status
//   DELETE /platform/tenants/:id
//   POST   /platform/tenants/:id/purge
//
// 全部强制 super_admin（group 级 RequirePlatformRole）+ ResTenant.* 双层守卫。

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
  list: (params?: { page?: number; size?: number; keyword?: string; status?: number }) =>
    api<PageResponse<TenantItem>>("/platform/tenants", { params }),

  get: (id: number) =>
    api<TenantItem>(`/platform/tenants/${id}`),

  create: (data: Partial<TenantItem> & { admin_account_id?: number }) =>
    api<TenantItem>("/platform/tenants", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<TenantItem>) =>
    api(`/platform/tenants/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  /** 更新状态（启用/停用） */
  updateStatus: (id: number, status: 0 | 1) =>
    api<TenantItem>(`/platform/tenants/${id}/status`, {
      method: "PUT",
      body: JSON.stringify({ status }),
    }),

  /** 软删（先软删后才能硬删） */
  delete: (id: number) =>
    api(`/platform/tenants/${id}`, {
      method: "DELETE",
    }),

  /** 硬删（不可逆，需先软删） */
  purge: (id: number) =>
    api<PurgeTenantResponse>(
      `/platform/tenants/${id}/purge`,
      { method: "POST" }
    ),
}