// 头像分类

import { api } from "./common"

export interface AvatarCategoryItem {
  id: number
  tenant_id?: number
  code: string
  name: string
  icon?: string
  type: string
  sort: number
  status: number
}

export const avatarCategoryApi = {
  list: (params?: { type?: string }) =>
    api<AvatarCategoryItem[]>("/t/flag/avatar-categories", { params }),

  create: (data: { code: string; name: string; icon?: string; type?: string; sort?: number }) =>
    api<AvatarCategoryItem>("/t/flag/avatar-categories", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<AvatarCategoryItem>) =>
    api(`/t/flag/avatar-categories/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/t/flag/avatar-categories/${id}`, {
      method: "DELETE",
    }),
}
