// Sys 权限码管理（super_admin 域）
//
// 与 apps/sys/permission 后端对齐：
//   GET    /sys/sys-permissions
//   GET    /sys/sys-permissions/:id
//   POST   /sys/sys-permissions
//   PUT    /sys/sys-permissions/:id
//   DELETE /sys/sys-permissions/:id
//
// 全部强制 super_admin（group 级 RequireAnySysRole）。
// sys 域资源码：sys_permissions.code 格式必须为 `resource:action`（后端强校验）。

import { api, type PageResponse } from "./common"

export interface SysPermissionItem {
  id: number
  menu_id?: number | null
  menu_code?: string | null
  code: string
  name: string
  action: string
  description?: string
  sort: number
  status: number
  created_at?: string
  updated_at?: string
}

export const sysPermissionApi = {
  list: (params?: {
    menu_id?: number
    keyword?: string
    page?: number
    size?: number
  }) =>
    api<PageResponse<SysPermissionItem>>("/sys/sys-permissions", { params }),

  get: (id: number) => api<SysPermissionItem>(`/sys/sys-permissions/${id}`),

  create: (
    data: Partial<SysPermissionItem> & {
      code: string
      name: string
    }
  ) =>
    api<SysPermissionItem>("/sys/sys-permissions", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<SysPermissionItem>) =>
    api(`/sys/sys-permissions/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/sys/sys-permissions/${id}`, {
      method: "DELETE",
    }),
}
