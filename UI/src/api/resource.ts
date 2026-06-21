// 资源（菜单下的 action：list / get / create / ...）

import { api, type PageResponse } from "./common"

export interface ResourceItem {
  id: number
  tenant_id?: number
  menu_id: number
  code: string
  name: string
  action: string
  description?: string
  sort: number
  status: number
  created_at?: string
  updated_at?: string
}

export const resourceApi = {
  list: (params?: { menu_id?: number; action?: string; page?: number; size?: number }) =>
    api<PageResponse<ResourceItem>>("/resources", { params }),

  get: (id: number) =>
    api<ResourceItem>(`/resources/${id}`),

  create: (data: Partial<ResourceItem>) =>
    api<ResourceItem>("/resources", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<ResourceItem>) =>
    api(`/resources/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/resources/${id}`, {
      method: "DELETE",
    }),

  byMenu: (menuId: number) =>
    api<ResourceItem[]>(`/resources/by-menu/${menuId}`),
}