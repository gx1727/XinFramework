// 头像框分类

import { api } from "./common"

export interface FrameCategoryItem {
  id: number
  tenant_id?: number
  code: string
  name: string
  type: string
  sort: number
  status: number
}

export const frameCategoryApi = {
  list: () =>
    api<FrameCategoryItem[]>("/t/flag/frames-categories"),

  create: (data: { code: string; name: string; type?: string; sort?: number }) =>
    api<FrameCategoryItem>("/t/flag/frames-categories", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<FrameCategoryItem>) =>
    api(`/t/flag/frames-categories/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/t/flag/frames-categories/${id}`, {
      method: "DELETE",
    }),
}
