// 菜单（业务域：/menus/*）

import { api, type PageResponse } from "./common"

export interface MenuItem {
  id: number
  tenant_id?: number
  code: string
  name: string
  subtitle?: string
  url?: string
  path: string
  icon?: string
  sort: number
  parent_id: number
  ancestors?: string
  visible?: boolean
  enabled?: boolean
  created_at?: string
  updated_at?: string
  children?: MenuItem[]
}

export const menuApi = {
  list: (params?: { page?: number; size?: number; root?: boolean }) =>
    api<PageResponse<MenuItem>>("/menus", { params }),

  tree: () =>
    api<MenuItem[]>("/menus/tree"),

  get: (id: number) =>
    api<MenuItem>(`/menus/${id}`),

  create: (data: Partial<MenuItem>) =>
    api<MenuItem>("/menus", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<MenuItem>) =>
    api(`/menus/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/menus/${id}`, {
      method: "DELETE",
    }),
}