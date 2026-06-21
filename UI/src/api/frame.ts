// 头像框

import { api, type PageResponse } from "./common"

export interface FrameItem {
  id: number
  tenant_id?: number
  category_id: number
  name: string
  description?: string
  preview_url?: string
  template_url?: string
  template_config?: {
    avatar_x: number
    avatar_y: number
    avatar_width: number
    avatar_height: number
  }
  type: string
  sort: number
  status: number
}

export const frameApi = {
  list: (params?: { category_id?: number; page?: number; size?: number }) =>
    api<PageResponse<FrameItem>>("/flag/frames", { params }),

  get: (id: number) =>
    api<FrameItem>(`/flag/frames/${id}`),

  create: (data: { name: string; category_id?: number; description?: string; preview_url?: string; template_url?: string; type?: string; sort?: number }) =>
    api<FrameItem>("/flag/frames", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: { id?: number; category_id?: number; name?: string; description?: string; preview_url?: string; template_url?: string; type?: string; sort?: number; status?: number }) =>
    api(`/flag/frames/${id}`, {
      method: "PUT",
      body: JSON.stringify({ id, ...data }),
    }),

  delete: (id: number) =>
    api(`/flag/frames/${id}`, {
      method: "DELETE",
    }),
}