// Sys 菜单管理（super_admin 域）
//
// 与 apps/sys/menu 后端对齐：
//   GET    /sys/menus
//   GET    /sys/menus/tree
//   GET    /sys/menus/:id
//   POST   /sys/menus
//   PUT    /sys/menus/:id
//   DELETE /sys/menus/:id
//
// 全部强制 super_admin Sys 角色（中间件层短路），写操作走 db.RunInSysTx。

import { api, type PageResponse } from "./common"

export interface SysMenuItem {
  id: number
  tenant_id: number // Sys 菜单固定为 0
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
  children?: SysMenuItem[]
}

export const sysMenuApi = {
  list: (params?: { page?: number; size?: number; root?: boolean }) =>
    api<PageResponse<SysMenuItem>>("/sys/menus", { params }),

  tree: () => api<SysMenuItem[]>("/sys/menus/tree"),

  get: (id: number) => api<SysMenuItem>(`/sys/menus/${id}`),

  create: (data: Partial<SysMenuItem>) =>
    api<SysMenuItem>("/sys/menus", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<SysMenuItem>) =>
    api(`/sys/menus/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/sys/menus/${id}`, {
      method: "DELETE",
    }),
}

/** 工具：判断当前 user 是否拥有指定 sys 角色（super_admin 等）。 */
export function hasSysRole(
  user: { sys_role_codes?: string[] } | null | undefined,
  role: string
): boolean {
  if (!user?.sys_role_codes) return false
  return user.sys_role_codes.includes(role)
}
