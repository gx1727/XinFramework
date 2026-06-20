// 系统运维（health / cache）

import { api } from "./common"

export interface CacheInfo {
  info: string
  dbSize: number
  commandStats: Record<string, any>
}

export interface CacheKeyItem {
  key: string
}

export interface CacheValue {
  key: string
  value: any
  type: string
  ttl: number
}

export const systemApi = {
  getCacheInfo: () =>
    api<CacheInfo>("/system/cache/info"),

  getCacheKeys: (pattern: string = "*") =>
    api<string[]>("/system/cache/keys", { params: { pattern } }),

  getCacheValue: (key: string) =>
    api<CacheValue>(`/system/cache/value/${encodeURIComponent(key)}`),

  deleteCacheKey: (key: string) =>
    api(`/system/cache/keys/${encodeURIComponent(key)}`, {
      method: "DELETE",
    }),
}
