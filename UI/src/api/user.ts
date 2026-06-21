// 用户

import { api, type PageResponse } from "./common"

export interface UserItem {
  id: number
  tenant_id?: number
  account_id?: number
  org_id?: number | null
  org_name?: string
  code: string
  status: number
  username?: string
  nickname?: string
  real_name: string
  avatar?: string
  phone?: string
  email?: string
  created_at?: string
  updated_at?: string
}

export const userApi = {
  list: (params?: { keyword?: string; org_id?: number; page?: number; size?: number }) =>
    api<PageResponse<UserItem>>("/t/users", { params }),

  get: (id: number) =>
    api<UserItem>(`/t/users/${id}`),

  create: (data: Partial<UserItem> & { username: string, password?: string }) =>
    api<UserItem>("/t/users", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  updateStatus: (id: number, status: number) =>
    api(`/t/users/${id}/status`, {
      method: "PUT",
      body: JSON.stringify({ id, status }),
    }),

  patch: (id: number, data: Partial<UserItem>) =>
    api(`/t/users/${id}`, {
      method: "PATCH",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/t/users/${id}`, {
      method: "DELETE",
    }),

  getProfile: () =>
    api<UserItem>("/t/user/profile"),

  updateProfile: (data: { nickName: string; avatarUrl?: string }) =>
    api("/t/user/profile", {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  uploadAvatar: (file: File) => {
    const formData = new FormData()
    formData.append("file", file)
    return api<{ url: string }>("/t/user/avatar", {
      method: "POST",
      body: formData,
      headers: {
        // 不设置 Content-Type，让浏览器自动设置，包括 boundary
      },
    })
  },
}
