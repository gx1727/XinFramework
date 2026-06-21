// 字典

import { api, type PageResponse } from "./common"

export interface DictItem {
  id: number
  tenant_id?: number
  code: string
  name: string
  sort?: number
  extend?: Record<string, unknown>
  item_count?: number
  status: number
  created_at?: string
  updated_at?: string
}

export interface DictValueItem {
  id: number
  tenant_id?: number
  dict_id: number
  code: string
  name: string
  sort?: number
  extend?: Record<string, unknown>
  status: number
  created_at?: string
  updated_at?: string
}

export const dictApi = {
  list: (params?: { keyword?: string; page?: number; size?: number }) =>
    api<PageResponse<DictItem>>("/t/dicts", { params }),

  get: (id: number) =>
    api<DictItem>(`/t/dicts/${id}`),

  create: (data: { code: string; name: string; sort?: number; extend?: Record<string, unknown> }) =>
    api<DictItem>("/t/dicts", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: { name: string; sort?: number; status?: number; extend?: Record<string, unknown> }) =>
    api(`/t/dicts/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/t/dicts/${id}`, {
      method: "DELETE",
    }),

  // ???
  listItems: (dictId: number) =>
    api<PageResponse<DictValueItem>>(`/t/dicts/${dictId}/items`),

  createItem: (dictId: number, data: { code: string; name: string; sort?: number; extend?: Record<string, unknown> }) =>
    api<DictValueItem>(`/t/dicts/${dictId}/items`, {
      method: "POST",
      body: JSON.stringify(data),
    }),

  updateItem: (dictId: number, itemId: number, data: { name: string; sort?: number; status?: number; extend?: Record<string, unknown> }) =>
    api(`/t/dicts/${dictId}/items/${itemId}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  deleteItem: (dictId: number, itemId: number) =>
    api(`/t/dicts/${dictId}/items/${itemId}`, {
      method: "DELETE",
    }),
}
