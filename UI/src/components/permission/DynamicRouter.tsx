import { useEffect } from "react"
import { useNavigate } from "react-router-dom"
import type { RouteObject } from "react-router-dom"
import { usePermissionStore } from "@/stores/permissionStore"
import { useAuthStore } from "@/stores/authStore"
import React from "react"

interface RouteMeta {
  title?: string
  icon?: string
  requiresAuth?: boolean
  permission?: string[]
}

type RouteObjectWithMeta = RouteObject & { meta?: RouteMeta }

interface DynamicRouterProps {
  routeConfig: RouteObject[]
  children?: React.ReactNode
}

export function DynamicRouter({ routeConfig, children }: DynamicRouterProps) {
  const navigate = useNavigate()
  const { menus, isMenuAccessible } = usePermissionStore()
  const { isAuthenticated, isLoading } = useAuthStore()

  useEffect(() => {
    if (isLoading) return
    
    if (!isAuthenticated) {
      navigate("/login")
      return
    }
    
    const accessibleRoutes = filterRoutesByPermission(routeConfig, menus, isMenuAccessible)
    console.log("Accessible routes:", accessibleRoutes)
  }, [isAuthenticated, isLoading, menus, isMenuAccessible, navigate, routeConfig])

  return <>{children}</>
}

function filterRoutesByPermission(
  routes: RouteObject[],
  menus: { path: string; children?: { path: string }[] }[],
  isMenuAccessible: (path: string) => boolean
): RouteObject[] {
  const accessiblePaths = new Set<string>()
  
  function collectPaths(menuItems: { path: string; children?: { path: string }[] }[]) {
    menuItems.forEach((menu) => {
      accessiblePaths.add(menu.path)
      if (menu.children) {
        collectPaths(menu.children)
      }
    })
  }
  
  collectPaths(menus)
  
  return routes.filter((route) => {
    const path = (route as { path?: string }).path
    
    if (!path) return true
    
    if (accessiblePaths.size === 0) return true
    
    return accessiblePaths.has(path) || isMenuAccessible(path)
  })
}

export function generateRoutesFromMenus(
  menus: { path: string; name: string; component?: string; children?: unknown[] }[],
  routeComponents: Record<string, () => Promise<{ default: React.ComponentType }>>
): RouteObjectWithMeta[] {
  return menus.map((menu) => {
    const route: RouteObjectWithMeta = {
      path: menu.path,
      element: menu.component && routeComponents[menu.component]
        ? createLazyComponent(routeComponents[menu.component])
        : undefined,
      meta: {
        title: menu.name,
      },
    }
    
    if (menu.children && menu.children.length > 0) {
      route.children = generateRoutesFromMenus(
        menu.children as { path: string; name: string; component?: string; children?: unknown[] }[],
        routeComponents
      )
    }
    
    return route
  })
}

function createLazyComponent(
  loader: () => Promise<{ default: React.ComponentType }>
): React.ReactElement {
  const Component = React.lazy(loader)
  return (
    <React.Suspense fallback={<div>Loading...</div>}>
      <Component />
    </React.Suspense>
  )
}
