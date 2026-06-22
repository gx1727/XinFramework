// 平台菜单管理（super_admin 域）
//
// 与 apps/platform/menu 后端对齐：
//   GET    /platform/menus
//   GET    /platform/menus/tree
//   GET    /platform/menus/:id
//   POST   /platform/menus
//   PUT    /platform/menus/:id
//   DELETE /platform/menus/:id
//
// 全部强制 super_admin 平台角色（中间件层短路），写操作走 db.RunInPlatformTx。

import { api, type PageResponse } from "./common"

export interface PlatformMenuItem {
  id: number
  tenant_id: number // 平台菜单固定为 0
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
  children?: PlatformMenuItem[]
}

export const platformMenuApi = {
  list: (params?: { page?: number; size?: number; root?: boolean }) =>
    api<PageResponse<PlatformMenuItem>>("/platform/menus", { params }),

  tree: () => api<PlatformMenuItem[]>("/platform/menus/tree"),

  get: (id: number) => api<PlatformMenuItem>(`/platform/menus/${id}`),

  create: (data: Partial<PlatformMenuItem>) =>
    api<PlatformMenuItem>("/platform/menus", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<PlatformMenuItem>) =>
    api(`/platform/menus/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/platform/menus/${id}`, {
      method: "DELETE",
    }),
}

/** 工具：判断当前 user 是否拥有指定平台角色（super_admin 等）。 */
export function hasPlatformRole(
  user: { platform_roles?: string[] } | null | undefined,
  role: string,
): boolean {
  if (!user?.platform_roles) return false
  return user.platform_roles.includes(role)
}