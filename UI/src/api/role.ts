// 角色

import { api, type PageResponse } from "./common"
import type { ResourceItem } from "./resource"

export interface RoleItem {
  id: number
  tenant_id?: number
  code: string
  name: string
  description?: string
  sort: number
  is_system?: boolean
  is_default?: boolean
  is_public?: boolean
  status: number
  created_at?: string
  updated_at?: string
}

export const roleApi = {
  list: (params?: { keyword?: string; page?: number; size?: number }) =>
    api<PageResponse<RoleItem>>("/roles", { params }),

  get: (id: number) =>
    api<RoleItem>(`/roles/${id}`),

  create: (data: Partial<RoleItem>) =>
    api<RoleItem>("/roles", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<RoleItem>) =>
    api(`/roles/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  patch: (id: number, data: Partial<RoleItem>) =>
    api(`/roles/${id}`, {
      method: "PATCH",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/roles/${id}`, {
      method: "DELETE",
    }),

  getMenus: (id: number) =>
    api<{ menu_ids: number[] }>(`/roles/${id}/menus`),

  setMenus: (id: number, menuIds: number[]) =>
    api(`/roles/${id}/menus`, {
      method: "PUT",
      body: JSON.stringify({ menu_ids: menuIds }),
    }),

  getDataScopes: (id: number) =>
    api<{ org_ids: number[] }>(`/roles/${id}/data-scopes`),

  setDataScopes: (id: number, orgIds: number[]) =>
    api(`/roles/${id}/data-scopes`, {
      method: "PUT",
      body: JSON.stringify({ org_ids: orgIds }),
    }),

  getPermissions: (id: number) =>
    api<{ list: ResourceItem[] }>(`/roles/${id}/permissions`),

  setPermissions: (id: number, resourceIds: number[]) =>
    api(`/roles/${id}/permissions`, {
      method: "PUT",
      body: JSON.stringify({ resource_ids: resourceIds }),
    }),
}