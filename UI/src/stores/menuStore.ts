import { create } from "zustand"
import { menuApi, platformMenuApi, ApiError } from "@/api"
import type { MenuItem, PlatformMenuItem } from "@/api"
import { useAuthStore } from "@/stores/authStore"

// 统一的菜单 shape（侧边栏 / 树视图都用），兼容 platform menu 与 tenant menu
export type UnifiedMenuItem = {
  id: number
  /** tenant menu: 真实租户 id；platform menu: 0。用来去重时区分来源。 */
  scope: "platform" | "tenant"
  code: string
  name: string
  subtitle?: string
  url?: string
  path: string
  icon?: string
  sort: number
  parent_id: number
  ancestors?: string
  visible?: boolean
  enabled?: boolean
  created_at?: string
  updated_at?: string
  children?: UnifiedMenuItem[]
}

interface MenuState {
  /** 当前用户能看到的所有菜单（已合并 platform + tenant）。 */
  menus: UnifiedMenuItem[]
  /** 数据源："api" 实时；null 加载中。 */
  dataSource: "api" | null
  isLoading: boolean
  /** 顶部错误条消息；null 时不显示。 */
  error: string | null

  fetchMenus: () => Promise<void>
  setMenus: (menus: UnifiedMenuItem[]) => void
  clearError: () => void
}

// ---------- helpers ----------

function fromPlatformMenu(m: PlatformMenuItem): UnifiedMenuItem {
  return {
    ...m,
    scope: "platform",
    children: m.children?.map(fromPlatformMenu),
  }
}

function fromTenantMenu(m: MenuItem): UnifiedMenuItem {
  return {
    ...m,
    scope: "tenant",
    children: m.children?.map(fromTenantMenu),
  }
}

/**
 * 合并 platform + tenant 菜单并去重：
 * - 顶层按 (scope, code) 去重，platform 优先
 * - 同 code 的子项保留 platform 版本
 *
 * 为什么需要合并：super_admin 登录后既要看到"平台管理"（来自 platform menus），
 * 也要看到本租户的"系统管理 → 用户管理"等业务菜单。
 */
export function mergeMenus(
  platform: UnifiedMenuItem[],
  tenant: UnifiedMenuItem[]
): UnifiedMenuItem[] {
  const seen = new Set<string>()
  const out: UnifiedMenuItem[] = []

  const addTree = (items: UnifiedMenuItem[]) => {
    items.forEach((it) => {
      const key = `${it.scope}:${it.code}`
      if (seen.has(key)) return
      seen.add(key)
      const cloned: UnifiedMenuItem = { ...it, children: [] }
      if (it.children?.length)
        cloned.children = mergeMenus(it.children as UnifiedMenuItem[], [])
      out.push(cloned)
    })
  }

  // platform 优先
  addTree(platform)
  addTree(tenant)

  // 顶层按 sort 排序
  out.sort((a, b) => a.sort - b.sort)
  return out
}

// ---------- store ----------

export const useMenuStore = create<MenuState>((set) => ({
  menus: [],
  dataSource: null,
  isLoading: false,
  error: null,

  setMenus: (menus) => set({ menus }),
  clearError: () => set({ error: null }),

  fetchMenus: async () => {
    const scope = useAuthStore.getState().scope

    set({ isLoading: true, error: null })

    // 路径 B 单身份登录：scope 决定走哪组菜单接口。
    //   - tenant   : /menus/tree          (受 RequireTenantContext 约束)
    //   - platform : /platform/menus/tree (受 RequirePlatformRole 约束)
    //   - null     : 未登录或 token 异常，不调任何接口
    try {
      if (scope === null) {
        set({ menus: [], dataSource: "api", isLoading: false, error: null })
        return
      }

      let platformMenus: UnifiedMenuItem[] = []
      let tenantMenus: UnifiedMenuItem[] = []

      if (scope === "platform") {
        const res = await platformMenuApi.tree()
        platformMenus = ((res as PlatformMenuItem[]) ?? []).map(
          fromPlatformMenu
        )
      } else {
        const res = await menuApi.tree()
        tenantMenus = ((res as MenuItem[]) ?? []).map(fromTenantMenu)
      }

      const merged = mergeMenus(platformMenus, tenantMenus)
      set({ menus: merged, dataSource: "api", isLoading: false, error: null })
    } catch (err) {
      const msg =
        err instanceof ApiError
          ? `${err.status} ${err.message}`
          : err instanceof Error
            ? err.message
            : "菜单加载失败"
      console.error("[menuStore] fetch failed:", err)
      set({
        menus: [],
        dataSource: null,
        isLoading: false,
        error: msg,
      })
    }
  },
}))

/**
 * 工具：从扁平的 menus（任意 scope）构造前端树形菜单。
 * 已被 app-sidebar 替代；保留以兼容旧调用方。
 */
export function buildMenuTree(menus: UnifiedMenuItem[]): UnifiedMenuItem[] {
  const menuMap = new Map<number, UnifiedMenuItem>()
  const roots: UnifiedMenuItem[] = []

  menus.forEach((menu) => {
    menuMap.set(menu.id, { ...menu, children: [] })
  })

  menuMap.forEach((menu) => {
    if (menu.parent_id === 0) {
      roots.push(menu)
    } else {
      const parent = menuMap.get(menu.parent_id)
      if (parent) {
        parent.children = parent.children || []
        parent.children.push(menu)
      }
    }
  })

  roots.sort((a, b) => a.sort - b.sort)
  return roots
}
