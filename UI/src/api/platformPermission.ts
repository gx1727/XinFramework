// 平台权限码管理（super_admin 域）
//
// 与 apps/platform/sys_permission 后端对齐：
//   GET    /platform/sys-permissions
//   GET    /platform/sys-permissions/:id
//   POST   /platform/sys-permissions
//   PUT    /platform/sys-permissions/:id
//   DELETE /platform/sys-permissions/:id
//
// 全部强制 super_admin（group 级 RequirePlatformRole）。
// 平台域资源码：sys_permissions.code 格式必须为 `resource:action`（后端强校验）。

import { api, type PageResponse } from "./common"

export interface PlatformPermissionItem {
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

export const platformPermissionApi = {
  list: (params?: { menu_id?: number; keyword?: string; page?: number; size?: number }) =>
    api<PageResponse<PlatformPermissionItem>>("/platform/sys-permissions", { params }),

  get: (id: number) =>
    api<PlatformPermissionItem>(`/platform/sys-permissions/${id}`),

  create: (
    data: Partial<PlatformPermissionItem> & {
      code: string
      name: string
    },
  ) =>
    api<PlatformPermissionItem>("/platform/sys-permissions", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<PlatformPermissionItem>) =>
    api(`/platform/sys-permissions/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/platform/sys-permissions/${id}`, {
      method: "DELETE",
    }),
}