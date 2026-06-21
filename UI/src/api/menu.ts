// 菜单（业务域：/t/menus/*）

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
    api<PageResponse<MenuItem>>("/t/menus", { params }),

  tree: () =>
    api<MenuItem[]>("/t/menus/tree"),

  get: (id: number) =>
    api<MenuItem>(`/t/menus/${id}`),

  create: (data: Partial<MenuItem>) =>
    api<MenuItem>("/t/menus", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<MenuItem>) =>
    api(`/t/menus/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/t/menus/${id}`, {
      method: "DELETE",
    }),
}