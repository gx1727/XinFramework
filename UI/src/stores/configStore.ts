import { create } from "zustand"
import { configApi, type ConfigGroup, type ConfigItem, type ConfigItemType } from "@/api"

interface ConfigState {
  // 已加载的公共配置：groupCode -> { key -> value }
  publicCache: Record<string, Record<string, unknown>>
  publicLoaded: Record<string, boolean>

  // 管理端状态
  groups: ConfigGroup[]
  groupItems: Record<number, ConfigItem[]>
  isLoadingGroups: boolean
  isLoadingItems: boolean
  error: string | null

  // 公共读
  loadPublic: (group: string, force?: boolean) => Promise<Record<string, unknown>>
  getPublicValue: (group: string, key: string) => unknown
  refreshPublic: (group: string) => Promise<void>

  // 管理端
  loadGroups: () => Promise<ConfigGroup[]>
  loadItems: (groupId: number) => Promise<ConfigItem[]>
  loadAllItems: () => Promise<ConfigItem[]>

  createGroup: (data: {
    code: string
    name: string
    description?: string
    icon?: string
    sort?: number
    is_public?: boolean
  }) => Promise<ConfigGroup>
  updateGroup: (
    id: number,
    data: {
      name?: string
      description?: string
      icon?: string
      sort?: number
      is_public?: boolean
      status?: number
    }
  ) => Promise<ConfigGroup>
  deleteGroup: (id: number) => Promise<void>

  createItem: (
    groupId: number,
    data: {
      key: string
      value?: unknown
      default_value?: unknown
      type: ConfigItemType
      label?: string
      description?: string
      options?: { label: string; value: string | number | boolean }[]
      validation?: Record<string, unknown>
      sort?: number
      is_public?: boolean
      is_readonly?: boolean
    }
  ) => Promise<ConfigItem>
  updateItem: (id: number, data: { value?: unknown; is_public?: boolean; is_readonly?: boolean; label?: string; description?: string; sort?: number; status?: number }) => Promise<ConfigItem>
  resetItem: (id: number) => Promise<ConfigItem>
  deleteItem: (id: number) => Promise<void>

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
  groupItems: {},
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
      const res = await configApi.getPublic(group)
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
      const res = await configApi.listGroups()
      set({ groups: res.list, isLoadingGroups: false })
      return res.list
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "加载配置分组失败"
      set({ error: msg, isLoadingGroups: false })
      throw e
    }
  },

  loadItems: async (groupId) => {
    set({ isLoadingItems: true, error: null })
    try {
      const res = await configApi.listItemsByGroup(groupId)
      set((s) => ({
        groupItems: { ...s.groupItems, [groupId]: res.list },
        isLoadingItems: false,
      }))
      return res.list
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "加载配置项失败"
      set({ error: msg, isLoadingItems: false })
      throw e
    }
  },

  loadAllItems: async () => {
    set({ isLoadingItems: true, error: null })
    try {
      const res = await configApi.listAllItems()
      const grouped: Record<number, ConfigItem[]> = {}
      for (const it of res.list) {
        if (!grouped[it.group_id]) grouped[it.group_id] = []
        grouped[it.group_id].push(it)
      }
      set({ groupItems: grouped, isLoadingItems: false })
      return res.list
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "加载配置项失败"
      set({ error: msg, isLoadingItems: false })
      throw e
    }
  },

  createGroup: async (data) => {
    const g = await configApi.createGroup(data)
    set((s) => ({ groups: [...s.groups, g] }))
    if (g.is_public) refreshPublicByGroupCode(g.code)
    return g
  },

  updateGroup: async (id, data) => {
    const old = get().groups.find((g) => g.id === id)
    const g = await configApi.updateGroup(id, data)
    set((s) => ({ groups: s.groups.map((it) => (it.id === id ? g : it)) }))
    if (old?.is_public || g.is_public) {
      refreshPublicByGroupCode(old?.code || g.code)
    }
    return g
  },

  deleteGroup: async (id) => {
    const old = get().groups.find((g) => g.id === id)
    await configApi.deleteGroup(id)
    set((s) => {
      const next = { ...s.groupItems }
      delete next[id]
      return {
        groups: s.groups.filter((g) => g.id !== id),
        groupItems: next,
      }
    })
    if (old?.is_public) refreshPublicByGroupCode(old.code)
  },

  createItem: async (groupId, data) => {
    const it = await configApi.createItem(groupId, data)
    set((s) => ({
      groupItems: {
        ...s.groupItems,
        [groupId]: [...(s.groupItems[groupId] || []), it].sort((a, b) => a.sort - b.sort),
      },
    }))
    if (it.is_public) invalidateItemPublicCache(it.group_id)
    return it
  },

  updateItem: async (id, data) => {
    const old = findItemInState(get().groupItems, id)
    const it = await configApi.updateItem(id, data)
    set((s) => {
      const next = { ...s.groupItems }
      for (const gid in next) {
        next[gid] = next[gid].map((x) => (x.id === id ? it : x))
      }
      return { groupItems: next }
    })
    if (it.is_public) invalidateItemPublicCache(it.group_id)
    return it
  },

  resetItem: async (id) => {
    const old = findItemInState(get().groupItems, id)
    const it = await configApi.resetItem(id)
    set((s) => {
      const next = { ...s.groupItems }
      for (const gid in next) {
        next[gid] = next[gid].map((x) => (x.id === id ? it : x))
      }
      return { groupItems: next }
    })
    if (it.is_public) invalidateItemPublicCache(it.group_id)
    return it
  },

  deleteItem: async (id) => {
    const old = findItemInState(get().groupItems, id)
    await configApi.deleteItem(id)
    set((s) => {
      const next = { ...s.groupItems }
      for (const gid in next) {
        next[gid] = next[gid].filter((x) => x.id !== id)
      }
      return { groupItems: next }
    })
    if (old?.is_public) invalidateItemPublicCache(old.group_id)
  },

  clear: () => {
    set({
      publicCache: {},
      publicLoaded: {},
      groups: [],
      groupItems: {},
      isLoadingGroups: false,
      isLoadingItems: false,
      error: null,
    })
  },
}))

// 工具：在 groupItems 中找 item
function findItemInState(groupItems: Record<number, ConfigItem[]>, id: number): ConfigItem | undefined {
  for (const gid in groupItems) {
    const it = groupItems[gid].find((x) => x.id === id)
    if (it) return it
  }
  return undefined
}

// 工具：通过 group_id 找到 group code，失效 public 缓存
function invalidateItemPublicCache(groupId: number) {
  const group = useConfigStore.getState().groups.find((g) => g.id === groupId)
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
