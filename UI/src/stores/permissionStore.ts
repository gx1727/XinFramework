import { create } from "zustand"
import { persist } from "zustand/middleware"
import type { MenuItem, PermissionContext } from "@/types/permission"

interface PermissionState {
  menus: MenuItem[]
  permissions: string[]
  
  setPermissions: (context: PermissionContext) => void
  clearPermissions: () => void
  
  hasPermission: (...permissions: string[]) => boolean
  hasPermissions: (permissions: string[]) => boolean
  hasAnyPermission: (...permissions: string[]) => boolean
  
  getAccessibleMenus: () => MenuItem[]
  isMenuAccessible: (path: string) => boolean
}

export const usePermissionStore = create<PermissionState>()(
  persist(
    (set, get) => ({
      menus: [],
      permissions: [],
      
      setPermissions: (context: PermissionContext) => {
        const permissionCodes = extractPermissionCodes(context.permissions)
        
        set({
          menus: context.menus,
          permissions: permissionCodes,
        })
      },
      
      clearPermissions: () => {
        set({
          menus: [],
          permissions: [],
        })
      },
      
      hasPermission: (...permissions: string[]) => {
        const { permissions: userPermissions } = get()
        return permissions.every((p) => userPermissions.includes(p))
      },
      
      hasPermissions: (permissions: string[]) => {
        const { permissions: userPermissions } = get()
        return permissions.every((p) => userPermissions.includes(p))
      },
      
      hasAnyPermission: (...permissions: string[]) => {
        const { permissions: userPermissions } = get()
        return permissions.some((p) => userPermissions.includes(p))
      },
      
      getAccessibleMenus: () => {
        return get().menus
      },
      
      isMenuAccessible: (path: string) => {
        const { menus } = get()
        return findMenuByPath(menus, path) !== null
      },
    }),
    {
      name: "permission-storage",
      partialize: (state) => ({
        menus: state.menus,
        permissions: state.permissions,
      }),
    }
  )
)

function extractPermissionCodes(permissions: string[]): string[] {
  return permissions
}

function findMenuByPath(menus: MenuItem[], path: string): MenuItem | null {
  for (const menu of menus) {
    if (menu.path === path) {
      return menu
    }
    if (menu.children) {
      const found = findMenuByPath(menu.children, path)
      if (found) return found
    }
  }
  return null
}

export function buildMenuTree(flatMenus: Array<{
  id: number
  parent_id?: number
  path: string
  name: string
  icon?: string
  component?: string
}>): MenuItem[] {
  const menuMap = new Map<number, MenuItem>()
  const rootMenus: MenuItem[] = []
  
  flatMenus.forEach((menu) => {
    menuMap.set(menu.id, { ...menu, children: [] })
  })
  
  flatMenus.forEach((menu) => {
    const menuItem = menuMap.get(menu.id)!
    if (menu.parent_id && menuMap.has(menu.parent_id)) {
      const parent = menuMap.get(menu.parent_id)!
      parent.children!.push(menuItem)
    } else {
      rootMenus.push(menuItem)
    }
  })
  
  return rootMenus
}
