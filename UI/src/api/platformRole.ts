// 平台角色管理（super_admin 域）
//
// 与 apps/platform/sys_role 后端对齐：
//   GET    /platform/sys-roles
//   GET    /platform/sys-roles/:id
//   POST   /platform/sys-roles
//   PUT    /platform/sys-roles/:id
//   DELETE /platform/sys-roles/:id
//   PUT    /platform/sys-roles/:id/menus
//   PUT    /platform/sys-roles/:id/permissions
//
// 全部强制 super_admin（group 级 RequirePlatformRole）。
// 平台域 RBAC：sys_role + sys_role_menus + sys_role_permissions，不带 tenant_id。

import { api, type PageResponse } from "./common"

export interface PlatformRoleMenuLite {
  id: number
  code: string
  name: string
}

export interface PlatformRolePermissionLite {
  id: number
  code: string
  name: string
  menu_id?: number | null
}

export interface PlatformRoleItem {
  id: number
  org_id?: number | null
  code: string
  name: string
  description?: string
  data_scope: number
  is_default: boolean
  sort: number
  status: number
  created_at?: string
  updated_at?: string
  menus?: PlatformRoleMenuLite[]
  permissions?: PlatformRolePermissionLite[]
}

export const platformRoleApi = {
  list: (params?: { keyword?: string; page?: number; size?: number }) =>
    api<PageResponse<PlatformRoleItem>>("/platform/sys-roles", { params }),

  get: (id: number) =>
    api<PlatformRoleItem>(`/platform/sys-roles/${id}`),

  create: (
    data: Partial<PlatformRoleItem> & {
      code: string
      name: string
      menu_ids?: number[]
      permission_ids?: number[]
    },
  ) =>
    api<PlatformRoleItem>("/platform/sys-roles", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<PlatformRoleItem>) =>
    api(`/platform/sys-roles/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/platform/sys-roles/${id}`, {
      method: "DELETE",
    }),

  /** 全量替换角色的平台菜单集合 */
  assignMenus: (id: number, menuIds: number[]) =>
    api(`/platform/sys-roles/${id}/menus`, {
      method: "PUT",
      body: JSON.stringify({ menu_ids: menuIds }),
    }),

  /** 全量替换角色的平台权限码集合 */
  assignPermissions: (id: number, permissionIds: number[]) =>
    api(`/platform/sys-roles/${id}/permissions`, {
      method: "PUT",
      body: JSON.stringify({ permission_ids: permissionIds }),
    }),
}