// 头像

import { api, type PageResponse } from "./common"

export interface AvatarItem {
  id: number
  tenant_id?: number
  user_id: number
  category_id: number
  name: string
  source_url: string
  thumbnail_url?: string
  file_size?: number
  width?: number
  height?: number
  type: string
  is_public?: boolean
  like_count?: number
  view_count?: number
  sort?: number
  status: number
}

export const avatarApi = {
  list: (params?: { category_id?: number; user_id?: number; type?: string; page?: number; size?: number }) =>
    api<PageResponse<AvatarItem>>("/flag/avatars", { params }),

  get: (id: number) =>
    api<AvatarItem>(`/flag/avatars/${id}`),

  create: (data: { source_url: string; category_id?: number; name?: string; thumbnail_url?: string; file_size?: number; width?: number; height?: number; is_public?: boolean }) =>
    api<AvatarItem>("/flag/avatars", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: { name?: string; category_id?: number; source_url: string; thumbnail_url?: string; is_public?: boolean; status?: number }) =>
    api(`/flag/avatars/${id}`, {
      method: "PUT",
      body: JSON.stringify({ id, ...data }),
    }),

  delete: (id: number) =>
    api(`/flag/avatars/${id}`, {
      method: "DELETE",
    }),
}
