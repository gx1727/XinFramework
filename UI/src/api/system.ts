// 系统运维（health / cache）
// cache 相关端点在 platform 域（仅 super_admin 可访问）。

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
    api<CacheInfo>("/platform/system/cache/info"),

  getCacheKeys: (pattern: string = "*") =>
    api<string[]>("/platform/system/cache/keys", { params: { pattern } }),

  getCacheValue: (key: string) =>
    api<CacheValue>(`/platform/system/cache/value/${encodeURIComponent(key)}`),

  deleteCacheKey: (key: string) =>
    api(`/platform/system/cache/keys/${encodeURIComponent(key)}`, {
      method: "DELETE",
    }),
}
