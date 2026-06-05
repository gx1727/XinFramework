import { type ReactNode } from "react"
import { usePermissionStore } from "@/stores/permissionStore"

interface AuthProps {
  permission?: string | string[]
  permissions?: string[]
  requireAll?: boolean
  children: ReactNode
  fallback?: ReactNode
}

export function Auth({
  permission,
  permissions: permissionsProp,
  requireAll = false,
  children,
  fallback = null,
}: AuthProps) {
  const { hasPermission, hasPermissions } = usePermissionStore()

  const permissionList = permission
    ? Array.isArray(permission)
      ? permission
      : [permission]
    : permissionsProp || []

  if (permissionList.length === 0) {
    return children
  }

  const hasAccess = requireAll
    ? hasPermissions(permissionList)
    : hasPermission(...permissionList)

  if (!hasAccess) {
    return fallback as React.ReactElement
  }

  return children as React.ReactElement
}

interface UsePermissionReturn {
  hasPermission: (...permissions: string[]) => boolean
  hasPermissions: (permissions: string[]) => boolean
  hasAnyPermission: (...permissions: string[]) => boolean
}

export function usePermission(): UsePermissionReturn {
  const { hasPermission, hasPermissions } = usePermissionStore()

  return {
    hasPermission,
    hasPermissions,
    hasAnyPermission: (...permissions: string[]) => hasPermission(...permissions),
  }
}

interface PermissionDirectives {
  vIf: (permission: string) => boolean
  vShow: (permission: string) => boolean
}

export function usePermissionDirective(): PermissionDirectives {
  const { hasPermission } = usePermissionStore()

  return {
    vIf: (permission: string) => hasPermission(permission),
    vShow: (permission: string) => hasPermission(permission),
  }
}
