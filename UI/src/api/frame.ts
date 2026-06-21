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
    api<PageResponse<FrameItem>>("/t/flag/frames", { params }),

  get: (id: number) =>
    api<FrameItem>(`/t/flag/frames/${id}`),

  create: (data: { name: string; category_id?: number; description?: string; preview_url?: string; template_url?: string; type?: string; sort?: number }) =>
    api<FrameItem>("/t/flag/frames", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: { id?: number; category_id?: number; name?: string; description?: string; preview_url?: string; template_url?: string; type?: string; sort?: number; status?: number }) =>
    api(`/t/flag/frames/${id}`, {
      method: "PUT",
      body: JSON.stringify({ id, ...data }),
    }),

  delete: (id: number) =>
    api(`/t/flag/frames/${id}`, {
      method: "DELETE",
    }),
}
