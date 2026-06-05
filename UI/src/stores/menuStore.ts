import { create } from "zustand"
import { api, ApiError } from "@/api"

export interface MenuItem {
  id: number
  tenant_id?: number
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
  children?: MenuItem[]
}

interface MenuState {
  menus: MenuItem[]
  isLoading: boolean
  error: string | null
  fetchMenus: () => Promise<void>
  setMenus: (menus: MenuItem[]) => void
}

const mockMenus: MenuItem[] = [
  {
    id: 1,
    code: "dashboard",
    name: "仪表盘",
    path: "/dashboard",
    icon: "LayoutDashboardIcon",
    sort: 1,
    parent_id: 0,
  },
  {
    id: 2,
    code: "analytics",
    name: "数据分析",
    path: "/analytics",
    icon: "ChartBarIcon",
    sort: 2,
    parent_id: 0,
  },
  {
    id: 3,
    code: "projects",
    name: "项目管理",
    path: "/projects",
    icon: "FolderIcon",
    sort: 3,
    parent_id: 0,
  },
  {
    id: 4,
    code: "team",
    name: "团队管理",
    path: "/team",
    icon: "UsersIcon",
    sort: 4,
    parent_id: 0,
  },
  {
    id: 6,
    code: "frames",
    name: "相框管理",
    path: "/frames",
    icon: "FrameIcon",
    sort: 6,
    parent_id: 0,
    children: [
      {
        id: 61,
        code: "frame-list",
        name: "相框列表",
        path: "/frames",
        icon: "FileIcon",
        sort: 1,
        parent_id: 6,
      },
      {
        id: 62,
        code: "frame-categories",
        name: "相框分类",
        path: "/frame-categories",
        icon: "ListIcon",
        sort: 2,
        parent_id: 6,
      },
    ],
  },
  {
    id: 7,
    code: "avatars",
    name: "头像管理",
    path: "/avatars",
    icon: "ImageIcon",
    sort: 7,
    parent_id: 0,
    children: [
      {
        id: 71,
        code: "avatar-list",
        name: "头像列表",
        path: "/avatars",
        icon: "FileIcon",
        sort: 1,
        parent_id: 7,
      },
      {
        id: 72,
        code: "avatar-categories",
        name: "头像分类",
        path: "/avatar-categories",
        icon: "ListIcon",
        sort: 2,
        parent_id: 7,
      },
    ],
  },
  {
    id: 5,
    code: "system",
    name: "系统管理",
    path: "/system",
    icon: "SettingsIcon",
    sort: 5,
    parent_id: 0,
    children: [
      {
        id: 51,
        code: "users",
        name: "用户管理",
        path: "/users",
        icon: "FileIcon",
        sort: 1,
        parent_id: 5,
      },
      {
        id: 52,
        code: "roles",
        name: "角色管理",
        path: "/roles",
        icon: "ShieldIcon",
        sort: 2,
        parent_id: 5,
      },
      {
        id: 53,
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

export const useMenuStore = create<MenuState>((set) => ({
  menus: [],
  isLoading: false,
  error: null,

  fetchMenus: async () => {
    set({ isLoading: true, error: null })
    try {
      const menus = await api<MenuItem[]>("/menus/tree")
      set({ menus: (menus as MenuItem[])?.length ? menus as MenuItem[] : mockMenus, isLoading: false })
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        set({ menus: mockMenus, isLoading: false, error: "unauthorized" })
      } else {
        set({ menus: mockMenus, isLoading: false })
      }
    }
  },

  setMenus: (menus) => set({ menus }),
}))

export function buildMenuTree(menus: MenuItem[]): MenuItem[] {
  const menuMap = new Map<number, MenuItem>()
  const roots: MenuItem[] = []

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