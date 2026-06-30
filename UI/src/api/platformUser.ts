// 平台用户管理（super_admin 域）
//
// 与 apps/platform/sys_user 后端对齐：
//   GET    /platform/sys-users
//   GET    /platform/sys-users/:id
//   POST   /platform/sys-users
//   PUT    /platform/sys-users/:id
//   PUT    /platform/sys-users/:id/status
//   DELETE /platform/sys-users/:id
//   PUT    /platform/sys-users/:id/roles
//
// 全部强制 super_admin（group 级 RequirePlatformRole）。
// 平台域无 tenant_id；一个 account 可以同时有 platform + tenant 身份。

import { api, type PageResponse } from "./common"

export interface PlatformUserRoleLite {
  id: number
  code: string
  name: string
}

export interface PlatformUserItem {
  id: number
  account_id: number
  org_id?: number | null
  code: string
  real_name: string
  nickname?: string
  avatar?: string
  status: number
  created_at?: string
  updated_at?: string
  roles?: PlatformUserRoleLite[]
}

export const platformUserApi = {
  list: (params?: { keyword?: string; page?: number; size?: number }) =>
    api<PageResponse<PlatformUserItem>>("/platform/sys-users", { params }),

  get: (id: number) => api<PlatformUserItem>(`/platform/sys-users/${id}`),

  create: (data: {
    // 模式 1：绑定已有账号（向后兼容）
    account_id?: number
    // 模式 2：AccountID 不传或为 0 时启用
    //   - phone + password 必填，username/email 可选
    //   - password 走 HTTPS，backend 用 Argon2id 哈希后入库
    username?: string
    phone?: string
    email?: string
    password?: string
    // 平台身份字段
    code?: string
    real_name: string
    nickname?: string
    avatar?: string
    status?: number
    role_ids?: number[]
  }) =>
    api<PlatformUserItem>("/platform/sys-users", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<PlatformUserItem>) =>
    api(`/platform/sys-users/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  updateStatus: (id: number, status: 0 | 1) =>
    api(`/platform/sys-users/${id}/status`, {
      method: "PUT",
      body: JSON.stringify({ status }),
    }),

  delete: (id: number) =>
    api(`/platform/sys-users/${id}`, {
      method: "DELETE",
    }),

  /** 全量替换用户的平台角色集合 */
  assignRoles: (id: number, roleIds: number[]) =>
    api(`/platform/sys-users/${id}/roles`, {
      method: "PUT",
      body: JSON.stringify({ role_ids: roleIds }),
    }),
}
