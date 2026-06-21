// 头像空间（活动 / 邀请）

import { api } from "./common"

export interface SpaceItem {
  id: number
  tenant_id?: number
  name: string
  description?: string
  frame_id?: number
  space_config?: {
    fields: Array<{
      key: string
      label: string
      required: boolean
      show: boolean
    }>
  }
  access_type: string
  invite_code?: string
  max_usage?: number
  usage_count?: number
  status: number
  start_at?: string
  end_at?: string
}

export interface GenerateAvatarResponse {
  id: number
  result_url: string
  share_text?: string
}

export const spaceApi = {
  list: () =>
    api<SpaceItem[]>("/flag/spaces"),

  get: (code: string) =>
    api<SpaceItem>(`/flag/spaces/${code}`),

  create: (data: { name: string; description?: string; frame_id?: number; access_type?: string; start_at?: string; end_at?: string }) =>
    api<SpaceItem>("/flag/spaces", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<SpaceItem>) =>
    api(`/flag/spaces/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/flag/spaces/${id}`, {
      method: "DELETE",
    }),
}