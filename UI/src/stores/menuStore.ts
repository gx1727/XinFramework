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
  /** 数据源："api" 实时；"mock" 来自前端 mock；null 加载中。 */
  dataSource: "api" | "mock" | null
  isLoading: boolean
  /** 顶部错误条消息；null 时不显示。 */
  error: string | null
  /** 用户主动勾选的 mock 兜底（持久化到 localStorage）。 */
  useMockFallback: boolean

  fetchMenus: () => Promise<void>
  setMenus: (menus: UnifiedMenuItem[]) => void
  setUseMockFallback: (v: boolean) => void
  clearError: () => void
}

const LS_KEY_USE_MOCK = "menuStore.useMockFallback"

// ---------- mock 数据（仅在 useMockFallback=true 时使用） ----------

const mockPlatformMenus: UnifiedMenuItem[] = [
  {
    id: 100,
    scope: "platform",
    code: "admin",
    name: "平台管理",
    path: "/admin",
    icon: "ShieldIcon",
    sort: 999,
    parent_id: 0,
    children: [
      {
        id: 101,
        scope: "platform",
        code: "platform-tenants",
        name: "平台租户",
        path: "/tenants",
        icon: "Building2Icon",
        sort: 1,
        parent_id: 100,
      },
      {
        id: 102,
        scope: "platform",
        code: "platform-menus",
        name: "平台菜单",
        path: "/menus",
        icon: "MenuIcon",
        sort: 2,
        parent_id: 100,
      },
    ],
  },
]

const mockTenantMenus: UnifiedMenuItem[] = [
  {
    id: 1,
    scope: "tenant",
    code: "dashboard",
    name: "仪表盘",
    path: "/dashboard",
    icon: "LayoutDashboardIcon",
    sort: 1,
    parent_id: 0,
  },
  {
    id: 2,
    scope: "tenant",
    code: "analytics",
    name: "数据分析",
    path: "/analytics",
    icon: "ChartBarIcon",
    sort: 2,
    parent_id: 0,
  },
  {
    id: 5,
    scope: "tenant",
    code: "system",
    name: "系统管理",
    path: "/system",
    icon: "SettingsIcon",
    sort: 5,
    parent_id: 0,
    children: [
      {
        id: 51,
        scope: "tenant",
        code: "users",
        name: "用户管理",
        path: "/users",
        icon: "UsersIcon",
        sort: 1,
        parent_id: 5,
      },
      {
        id: 52,
        scope: "tenant",
        code: "roles",
        name: "角色管理",
        path: "/roles",
        icon: "ShieldIcon",
        sort: 2,
        parent_id: 5,
      },
      {
        id: 53,
        scope: "tenant",
        code: "menus",
        name: "菜单管理",
        path: "/menus",
        icon: "MenuIcon",
        sort: 3,
        parent_id: 5,
      },
    ],
  },
]

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
  tenant: UnifiedMenuItem[],
): UnifiedMenuItem[] {
  const seen = new Set<string>()
  const out: UnifiedMenuItem[] = []

  const addTree = (items: UnifiedMenuItem[]) => {
    items.forEach((it) => {
      const key = `${it.scope}:${it.code}`
      if (seen.has(key)) return
      seen.add(key)
      const cloned: UnifiedMenuItem = { ...it, children: [] }
      if (it.children?.length) cloned.children = mergeMenus(it.children as UnifiedMenuItem[], [])
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

export const useMenuStore = create<MenuState>((set, get) => ({
  menus: [],
  dataSource: null,
  isLoading: false,
  error: null,
  useMockFallback:
    typeof window !== "undefined" &&
    window.localStorage.getItem(LS_KEY_USE_MOCK) === "1",

  setMenus: (menus) => set({ menus }),
  setUseMockFallback: (v) => {
    if (typeof window !== "undefined") {
      window.localStorage.setItem(LS_KEY_USE_MOCK, v ? "1" : "0")
    }
    set({ useMockFallback: v })
  },
  clearError: () => set({ error: null }),

  fetchMenus: async () => {
    const useMock = get().useMockFallback
    const user = useAuthStore.getState().user
    const isSuperAdmin = (user?.platform_roles ?? []).includes("super_admin")

    set({ isLoading: true, error: null })

    // ---------- mock 分支（用户主动勾选） ----------
    if (useMock) {
      const merged = mergeMenus(mockPlatformMenus, mockTenantMenus)
      set({ menus: merged, dataSource: "mock", isLoading: false })
      return
    }

    // ---------- api 分支 ----------
    try {
      let platformMenus: UnifiedMenuItem[] = []
      let tenantMenus: UnifiedMenuItem[] = []

      // super_admin 才并发请求平台菜单；普通用户只请求租户菜单
      const promises: Array<Promise<void>> = [
        menuApi
          .tree()
          .then((res) => {
            tenantMenus = ((res as MenuItem[]) ?? []).map(fromTenantMenu)
          }),
      ]
      if (isSuperAdmin) {
        promises.push(
          platformMenuApi
            .tree()
            .then((res) => {
              platformMenus = ((res as PlatformMenuItem[]) ?? []).map(
                fromPlatformMenu,
              )
            })
            // 平台菜单接口不可用时降级（不影响租户菜单渲染）
            .catch((e) => {
              console.warn("[menuStore] platform menus unavailable:", e)
            }),
        )
      }

      await Promise.all(promises)

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