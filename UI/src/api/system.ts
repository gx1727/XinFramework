// 系统运维（health / cache）
// cache 相关端点在 platform 域（仅 super_admin 可访问）。

import { api, type PageResponse } from "./common"

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
  getCacheInfo: () => api<CacheInfo>("/platform/system/cache/info"),

  getCacheKeys: (
    pattern: string = "*",
    page: number = 1,
    size: number = 50,
    excludePrefixes?: string[]
  ) =>
    api<PageResponse<string>>("/platform/system/cache/keys", {
      params: {
        pattern,
        page,
        size,
        exclude_prefixes:
          excludePrefixes && excludePrefixes.length > 0
            ? excludePrefixes.join(",")
            : undefined,
      },
    }),

  getCacheValue: (key: string) =>
    api<CacheValue>(`/platform/system/cache/value/${encodeURIComponent(key)}`),

  deleteCacheKey: (key: string) =>
    api(`/platform/system/cache/keys/${encodeURIComponent(key)}`, {
      method: "DELETE",
    }),
}
