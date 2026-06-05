export interface Permission {
  id: number
  code: string
  name: string
  type: "menu" | "button" | "api"
  path?: string
  icon?: string
  parent_id?: number
  children?: Permission[]
}

export interface Role {
  id: number
  code: string
  name: string
  description?: string
  permissions: Permission[]
}

export interface User {
  id: number
  code: string
  name?: string
  email?: string
  avatar?: string
  role: Role
  tenant_id: number
}

export interface MenuItem {
  id: number
  path: string
  name: string
  icon?: string
  component?: string
  children?: MenuItem[]
}

export interface PermissionContext {
  menus: MenuItem[]
  permissions: string[]
}

export interface RouteConfig {
  path: string
  name: string
  component?: () => Promise<unknown>
  children?: RouteConfig[]
  meta?: {
    title?: string
    icon?: string
    requiresAuth?: boolean
    permission?: string[]
  }
}

export interface LoginResponse {
  token: string
  refresh_token?: string
  user: User
  permissions: PermissionContext
}

export interface LoginRequest {
  account: string
  password: string
}
