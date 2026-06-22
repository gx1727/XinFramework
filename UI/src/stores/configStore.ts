import { create } from "zustand"
import {
  configApi,
  type ConfigCategory,
  type ConfigItem,
  type ConfigItemType,
  type CreatePlatformGroupRequest,
  type CreatePlatformItemRequest,
  type UpdatePlatformGroupRequest,
  type UpdatePlatformItemRequest,
} from "@/api"

interface ConfigState {
  // 已加载的公共配置：groupCode -> { key -> value }
  publicCache: Record<string, Record<string, unknown>>
  publicLoaded: Record<string, boolean>

  // 管理端状态
  groups: ConfigCategory[]
  categoryItems: Record<number, ConfigItem[]>
  isLoadingGroups: boolean
  isLoadingItems: boolean
  error: string | null

  // 公共读
  loadPublic: (group: string, force?: boolean) => Promise<Record<string, unknown>>
  getPublicValue: (group: string, key: string) => unknown
  refreshPublic: (group: string) => Promise<void>

  // 管理端
  loadGroups: () => Promise<ConfigCategory[]>
  loadItems: (categoryId: number) => Promise<ConfigItem[]>
  // 后端没有"全量 items"单端点；前端聚合 listGroups + 每 group listItemsByGroup
  loadAllItems: () => Promise<ConfigItem[]>

  createGroup: (data: CreatePlatformGroupRequest) => Promise<ConfigCategory>
  updateGroup: (id: number, data: UpdatePlatformGroupRequest) => Promise<ConfigCategory>
  deleteGroup: (id: number) => Promise<void>

  createItem: (
    categoryId: number,
    data: CreatePlatformItemRequest
  ) => Promise<ConfigItem>
  updateItem: (
    categoryId: number,
    id: number,
    data: UpdatePlatformItemRequest
  ) => Promise<ConfigItem>
  // "重置" 在新架构下 = 删除租户对 platform item 的 override
  resetItem: (categoryId: number, id: number) => Promise<void>
  deleteItem: (categoryId: number, id: number) => Promise<void>

  clear: () => void
}

// 工具：写/更新公共项后，按 groupCode 失效公共缓存
function refreshPublicByGroupCode(groupCode: string) {
  useConfigStore.setState((s) => ({
    publicLoaded: { ...s.publicLoaded, [groupCode]: false },
  }))
  // 后台异步重新加载
  void useConfigStore.getState().loadPublic(groupCode, true)
}

export const useConfigStore = create<ConfigState>()((set, get) => ({
  publicCache: {},
  publicLoaded: {},
  groups: [],
  categoryItems: {},
  isLoadingGroups: false,
  isLoadingItems: false,
  error: null,

  // =============== 公共读 ===============

  loadPublic: async (group, force = false) => {
    const { publicCache, publicLoaded } = get()
    if (!force && publicLoaded[group]) {
      return publicCache[group] || {}
    }
    try {
      // 公共读在匿名场景下也能调（login 页面加载 logo / 站点信息等），
      // 用 tenantId=0 表示平台范围公开配置。
      // 已登录场景如果需要按租户筛选，可以传 useAuthStore().user?.tenant_id。
      const res = await configApi.getPublic(group, 0)
      set((s) => ({
        publicCache: { ...s.publicCache, [group]: res.values || {} },
        publicLoaded: { ...s.publicLoaded, [group]: true },
      }))
      return res.values || {}
    } catch (e) {
      // 公共读失败不阻塞 UI，记 warn 即可
      console.warn(`[config] load public group=${group} failed`, e)
      set((s) => ({
        publicCache: { ...s.publicCache, [group]: s.publicCache[group] || {} },
        publicLoaded: { ...s.publicLoaded, [group]: true },
      }))
      return get().publicCache[group] || {}
    }
  },

  getPublicValue: (group, key) => {
    return get().publicCache[group]?.[key]
  },

  refreshPublic: async (group) => {
    await get().loadPublic(group, true)
  },

  // =============== 管理端 ===============

  loadGroups: async () => {
    set({ isLoadingGroups: true, error: null })
    try {
      // 后端返回裸数组，不是 {list,total}
      const list = await configApi.listGroups()
      set({ groups: list, isLoadingGroups: false })
      return list
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "加载配置分组失败"
      set({ error: msg, isLoadingGroups: false })
      throw e
    }
  },

  loadItems: async (categoryId) => {
    set({ isLoadingItems: true, error: null })
    try {
      const list = await configApi.listItemsByGroup(categoryId)
      set((s) => ({
        categoryItems: { ...s.categoryItems, [categoryId]: list },
        isLoadingItems: false,
      }))
      return list
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "加载配置项失败"
      set({ error: msg, isLoadingItems: false })
      throw e
    }
  },

  loadAllItems: async () => {
    // 后端没有"一次取全量 items"的端点；前端聚合
    set({ isLoadingItems: true, error: null })
    try {
      const groups = await configApi.listGroups()
      const lists = await Promise.all(
        groups.map((g) => configApi.listItemsByGroup(g.id).catch(() => []))
      )
      const categoryItems: Record<number, ConfigItem[]> = {}
      const flat: ConfigItem[] = []
      groups.forEach((g, i) => {
        categoryItems[g.id] = lists[i]
        flat.push(...lists[i])
      })
      set({ groups, categoryItems, isLoadingItems: false })
      return flat
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "加载配置项失败"
      set({ error: msg, isLoadingItems: false })
      throw e
    }
  },

  createGroup: async (data) => {
    const g = await configApi.createPlatformGroup(data)
    set((s) => ({ groups: [...s.groups, g] }))
    if (g.is_public) refreshPublicByGroupCode(g.code)
    return g
  },

  updateGroup: async (id, data) => {
    const old = get().groups.find((g) => g.id === id)
    const g = await configApi.updatePlatformGroup(id, data)
    set((s) => ({ groups: s.groups.map((it) => (it.id === id ? g : it)) }))
    if (old?.is_public || g.is_public) {
      refreshPublicByGroupCode(old?.code || g.code)
    }
    return g
  },

  deleteGroup: async (id) => {
    const old = get().groups.find((g) => g.id === id)
    await configApi.deletePlatformGroup(id)
    set((s) => {
      const next = { ...s.categoryItems }
      delete next[id]
      return {
        groups: s.groups.filter((g) => g.id !== id),
        categoryItems: next,
      }
    })
    if (old?.is_public) refreshPublicByGroupCode(old.code)
  },

  createItem: async (categoryId, data) => {
    const it = await configApi.createPlatformItem(categoryId, data)
    set((s) => ({
      categoryItems: {
        ...s.categoryItems,
        [categoryId]: [...(s.categoryItems[categoryId] || []), it].sort((a, b) => a.sort - b.sort),
      },
    }))
    if (it.is_public) invalidateItemPublicCache(it.category_id)
    return it
  },

  updateItem: async (categoryId, id, data) => {
    const old = findItemInState(get().categoryItems, id)
    const it = await configApi.updatePlatformItem(categoryId, id, data)
    set((s) => {
      const next = { ...s.categoryItems }
      for (const gid in next) {
        next[gid] = next[gid].map((x) => (x.id === id ? it : x))
      }
      return { categoryItems: next }
    })
    if (it.is_public) invalidateItemPublicCache(it.category_id)
    return it
  },

  resetItem: async (categoryId, id) => {
    // 新架构下"重置"语义 = 删除租户对 platform item 的 override
    const old = findItemInState(get().categoryItems, id)
    await configApi.deleteOverride(categoryId, id)
    set((s) => {
      const next = { ...s.categoryItems }
      for (const gid in next) {
        next[gid] = next[gid].map((x) =>
          x.id === id ? { ...x, value: x.default_value } : x
        )
      }
      return { categoryItems: next }
    })
    if (old?.is_public) invalidateItemPublicCache(old.category_id)
  },

  deleteItem: async (categoryId, id) => {
    const old = findItemInState(get().categoryItems, id)
    await configApi.deletePlatformItem(categoryId, id)
    set((s) => {
      const next = { ...s.categoryItems }
      for (const gid in next) {
        next[gid] = next[gid].filter((x) => x.id !== id)
      }
      return { categoryItems: next }
    })
    if (old?.is_public) invalidateItemPublicCache(old.category_id)
  },

  clear: () => {
    set({
      publicCache: {},
      publicLoaded: {},
      groups: [],
      categoryItems: {},
      isLoadingGroups: false,
      isLoadingItems: false,
      error: null,
    })
  },
}))

// 工具：在 categoryItems 中找 item
function findItemInState(categoryItems: Record<number, ConfigItem[]>, id: number): ConfigItem | undefined {
  for (const gid in categoryItems) {
    const it = categoryItems[gid].find((x) => x.id === id)
    if (it) return it
  }
  return undefined
}

// 工具：通过 category_id 找到 group code，失效 public 缓存
function invalidateItemPublicCache(categoryId: number) {
  const group = useConfigStore.getState().groups.find((g) => g.id === categoryId)
  if (group?.is_public) {
    refreshPublicByGroupCode(group.code)
  }
}

// 便捷 hook：取单个公共值（无则返回 undefined）
export function useConfigItem(group: string, key: string): unknown {
  return useConfigStore((s) => s.publicCache[group]?.[key])
}

// 便捷 hook：取整组公共配置
export function useConfigGroup(group: string): Record<string, unknown> {
  return useConfigStore((s) => s.publicCache[group] || {})
}

// 仅占位——避免 type-only import 在某些打包器下被消除
export type _ConfigItemTypeAlias = ConfigItemType