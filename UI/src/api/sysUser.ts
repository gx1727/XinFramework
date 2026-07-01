// Sys 用户管理（super_admin 域）
//
// 与 apps/sys/user 后端对齐：
//   GET    /sys/sys-users
//   GET    /sys/sys-users/:id
//   POST   /sys/sys-users
//   PUT    /sys/sys-users/:id
//   PUT    /sys/sys-users/:id/status
//   DELETE /sys/sys-users/:id
//   PUT    /sys/sys-users/:id/roles
//
// 全部强制 super_admin（group 级 RequireAnySysRole）。
// sys 域无 tenant_id；一个 account 可以同时有 sys + tenant 身份。

import { api, type PageResponse } from "./common"

export interface SysUserRoleLite {
  id: number
  code: string
  name: string
}

export interface SysUserItem {
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
  roles?: SysUserRoleLite[]
}

export const sysUserApi = {
  list: (params?: { keyword?: string; page?: number; size?: number }) =>
    api<PageResponse<SysUserItem>>("/sys/sys-users", { params }),

  get: (id: number) => api<SysUserItem>(`/sys/sys-users/${id}`),

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
    // sys 身份字段
    code?: string
    real_name: string
    nickname?: string
    avatar?: string
    status?: number
    role_ids?: number[]
  }) =>
    api<SysUserItem>("/sys/sys-users", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<SysUserItem>) =>
    api(`/sys/sys-users/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  updateStatus: (id: number, status: 0 | 1) =>
    api(`/sys/sys-users/${id}/status`, {
      method: "PUT",
      body: JSON.stringify({ status }),
    }),

  delete: (id: number) =>
    api(`/sys/sys-users/${id}`, {
      method: "DELETE",
    }),

  /** 全量替换用户的 sys 角色集合 */
  assignRoles: (id: number, roleIds: number[]) =>
    api(`/sys/sys-users/${id}/roles`, {
      method: "PUT",
      body: JSON.stringify({ role_ids: roleIds }),
    }),
}
