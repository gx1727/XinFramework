// Sys 角色管理（super_admin 域）
//
// 与 apps/sys/role 后端对齐：
//   GET    /sys/sys-roles
//   GET    /sys/sys-roles/:id
//   POST   /sys/sys-roles
//   PUT    /sys/sys-roles/:id
//   DELETE /sys/sys-roles/:id
//   PUT    /sys/sys-roles/:id/menus
//   PUT    /sys/sys-roles/:id/permissions
//
// 全部强制 super_admin（group 级 RequireAnySysRole）。
// sys 域 RBAC：sys_role + sys_role_menus + sys_role_permissions，不带 tenant_id。

import { api, type PageResponse } from "./common"

export interface SysRoleMenuLite {
  id: number
  code: string
  name: string
}

export interface SysRolePermissionLite {
  id: number
  code: string
  name: string
  menu_id?: number | null
}

export interface SysRoleItem {
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
  menus?: SysRoleMenuLite[]
  permissions?: SysRolePermissionLite[]
}

export const sysRoleApi = {
  list: (params?: { keyword?: string; page?: number; size?: number }) =>
    api<PageResponse<SysRoleItem>>("/sys/sys-roles", { params }),

  get: (id: number) => api<SysRoleItem>(`/sys/sys-roles/${id}`),

  create: (
    data: Partial<SysRoleItem> & {
      code: string
      name: string
      menu_ids?: number[]
      permission_ids?: number[]
    }
  ) =>
    api<SysRoleItem>("/sys/sys-roles", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<SysRoleItem>) =>
    api(`/sys/sys-roles/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/sys/sys-roles/${id}`, {
      method: "DELETE",
    }),

  /** 全量替换角色的 sys 菜单集合 */
  assignMenus: (id: number, menuIds: number[]) =>
    api(`/sys/sys-roles/${id}/menus`, {
      method: "PUT",
      body: JSON.stringify({ menu_ids: menuIds }),
    }),

  /** 全量替换角色的 sys 权限码集合 */
  assignPermissions: (id: number, permissionIds: number[]) =>
    api(`/sys/sys-roles/${id}/permissions`, {
      method: "PUT",
      body: JSON.stringify({ permission_ids: permissionIds }),
    }),
}
