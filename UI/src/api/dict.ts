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
    api<PageResponse<DictItem>>("/dicts", { params }),

  get: (id: number) =>
    api<DictItem>(`/dicts/${id}`),

  create: (data: { code: string; name: string; sort?: number; extend?: Record<string, unknown> }) =>
    api<DictItem>("/dicts", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: { name: string; sort?: number; status?: number; extend?: Record<string, unknown> }) =>
    api(`/dicts/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/dicts/${id}`, {
      method: "DELETE",
    }),

  // 字典项
  listItems: (dictId: number) =>
    api<PageResponse<DictValueItem>>(`/dicts/${dictId}/items`),

  createItem: (dictId: number, data: { code: string; name: string; sort?: number; extend?: Record<string, unknown> }) =>
    api<DictValueItem>(`/dicts/${dictId}/items`, {
      method: "POST",
      body: JSON.stringify(data),
    }),

  updateItem: (dictId: number, itemId: number, data: { name: string; sort?: number; status?: number; extend?: Record<string, unknown> }) =>
    api(`/dicts/${dictId}/items/${itemId}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  deleteItem: (dictId: number, itemId: number) =>
    api(`/dicts/${dictId}/items/${itemId}`, {
      method: "DELETE",
    }),

  // =============== Platform 域（super_admin CRUD）===============
  // 挂在 /api/v1/platform/dicts 下，需 RequirePlatformRole("super_admin")。
  // 注意：响应约定与业务端略有不同
  //   - 列表：`ListPlatformDicts` 返回 `{list, total, page, size}`
  //   - 字典项：`ListPlatformItems` 返回 `{list, total}`（无 page/size，与业务端 items 一致）
  listPlatformDicts: (params?: { keyword?: string; page?: number; size?: number }) =>
    api<PageResponse<DictItem>>("/platform/dicts", { params }),

  getPlatformDict: (id: number) =>
    api<DictItem>(`/platform/dicts/${id}`),

  createPlatformDict: (data: {
    code: string
    name: string
    sort?: number
    visibility?: "all" | "whitelist" | "blacklist"
    extend?: Record<string, unknown>
  }) =>
    api<DictItem>("/platform/dicts", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  updatePlatformDict: (
    id: number,
    data: {
      name: string
      sort?: number
      status?: number
      visibility?: "all" | "whitelist" | "blacklist"
      extend?: Record<string, unknown>
    }
  ) =>
    api<DictItem>(`/platform/dicts/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  deletePlatformDict: (id: number) =>
    api(`/platform/dicts/${id}`, { method: "DELETE" }),

  listPlatformItems: (dictId: number) =>
    api<PageResponse<DictValueItem>>(`/platform/dicts/${dictId}/items`),

  createPlatformItem: (
    dictId: number,
    data: {
      code: string
      name: string
      sort?: number
      extend?: Record<string, unknown>
    }
  ) =>
    api<DictValueItem>(`/platform/dicts/${dictId}/items`, {
      method: "POST",
      body: JSON.stringify(data),
    }),

  updatePlatformItem: (
    dictId: number,
    itemId: number,
    data: {
      name: string
      sort?: number
      status?: number
      extend?: Record<string, unknown>
    }
  ) =>
    api(`/platform/dicts/${dictId}/items/${itemId}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  deletePlatformItem: (dictId: number, itemId: number) =>
    api(`/platform/dicts/${dictId}/items/${itemId}`, {
      method: "DELETE",
    }),
}