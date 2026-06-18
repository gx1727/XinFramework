const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:8087/api/v1"

interface ApiOptions extends RequestInit {
  params?: Record<string, string | number | boolean>
  retry?: number
  retryDelay?: number
}

export interface ApiResponse<T = unknown> {
  code: number
  msg: string
  data: T
}

export interface PageResponse<T> {
  list: T[]
  total: number
  page?: number
  size?: number
}

export interface LoginRequest {
  account: string
  password: string
  tenant_id?: number
}

export interface LoginResponse {
  token: string
  refresh_token: string
  user: {
    id: number
    tenant_id: number
    code: string
    role: string
  }
}

export interface RegisterRequest {
  account: string
  password: string
  tenant_id?: number
  real_name: string
}

export interface RefreshRequest {
  refresh_token: string
}

export interface RefreshResponse {
  token: string
  refresh_token?: string
}

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

export interface UserItem {
  id: number
  tenant_id?: number
  account_id?: number
  org_id?: number | null
  org_name?: string
  code: string
  status: number
  username?: string
  nickname?: string
  real_name: string
  avatar?: string
  phone?: string
  email?: string
  role: string
}

export interface RoleItem {
  id: number
  tenant_id?: number
  org_id?: number
  code: string
  name: string
  description?: string
  data_scope?: number
  extend?: string
  is_default?: boolean
  sort: number
  status: number
  created_at?: string
  updated_at?: string
}

export interface TenantItem {
  id: number
  code: string
  name: string
  status: number
  contact?: string
  phone?: string
  email?: string
  province?: string
  city?: string
  area?: string
  address?: string
  created_at?: string
  updated_at?: string
}

export interface OrganizationItem {
  id: number
  tenant_id?: number
  code: string
  name: string
  type: string
  description?: string
  admin_code?: string
  parent_id: number
  ancestors?: string
  sort: number
  status: number
  created_at?: string
  updated_at?: string
  children?: OrganizationItem[]
}

export interface DictItem {
  id: number
  tenant_id?: number
  code: string
  name: string
  sort: number
  status: number
  extend?: Record<string, unknown>
  item_count?: number
  created_at?: string
  updated_at?: string
  items?: DictValueItem[]
}

export interface DictValueItem {
  id: number
  tenant_id?: number
  dict_id: number
  code: string
  name: string
  sort: number
  status: number
  extend?: Record<string, unknown>
  created_at?: string
  updated_at?: string
}

// 配置项类型枚举
export type ConfigItemType =
  | "string"
  | "number"
  | "boolean"
  | "json"
  | "image"
  | "color"
  | "select"
  | "multiselect"
  | "text"
  | "password"

// 配置项可选项（select/multiselect 用）
export interface ConfigOption {
  label: string
  value: string | number | boolean
}

// 配置项校验规则
export interface ConfigValidation {
  min?: number
  max?: number
  required?: boolean
  regex?: string
  placeholder?: string
}

export interface ConfigGroup {
  id: number
  tenant_id?: number
  code: string
  name: string
  description?: string
  icon?: string
  sort: number
  is_system: boolean
  is_public: boolean
  status: number
  created_at?: string
  updated_at?: string
}

export interface ConfigItem {
  id: number
  tenant_id?: number
  group_id: number
  key: string
  value?: unknown
  default_value?: unknown
  type: ConfigItemType
  label?: string
  description?: string
  options?: ConfigOption[]
  validation?: ConfigValidation
  sort: number
  is_public: boolean
  is_readonly: boolean
  is_system: boolean
  status: number
  created_at?: string
  updated_at?: string
}

export interface PublicConfigResponse {
  group: string
  values: Record<string, unknown>
}

export interface ResourceItem {
  id: number
  tenant_id?: number
  menu_id: number
  code: string
  name: string
  action: string
  description?: string
  sort: number
  status: number
  created_at?: string
  updated_at?: string
}

class ApiError extends Error {
  status: number
  code: number
  data?: unknown

  constructor(
    status: number,
    code: number,
    message: string,
    data?: unknown
  ) {
    super(message)
    this.name = "ApiError"
    this.status = status
    this.code = code
    this.data = data
  }
}

let isRefreshing = false
let refreshSubscribers: Array<(token: string) => void> = []

function subscribeTokenRefresh(callback: (token: string) => void) {
  refreshSubscribers.push(callback)
}

function onTokenRefreshed(token: string) {
  refreshSubscribers.forEach((callback) => callback(token))
  refreshSubscribers = []
}

async function buildUrl(endpoint: string, params?: Record<string, string | number | boolean>): Promise<string> {
  const url = new URL(`${API_BASE_URL}${endpoint}`, window.location.origin)
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value === undefined || value === null) return
      url.searchParams.append(key, String(value))
    })
  }
  return url.toString()
}

function getToken(): string | null {
  if (typeof window !== "undefined") {
    return localStorage.getItem("token")
  }
  return null
}

function getRefreshToken(): string | null {
  if (typeof window !== "undefined") {
    return localStorage.getItem("refresh_token")
  }
  return null
}

function setTokens(token: string, refreshToken?: string) {
  localStorage.setItem("token", token)
  if (refreshToken) {
    localStorage.setItem("refresh_token", refreshToken)
  }
}

function clearTokens() {
  localStorage.removeItem("token")
  localStorage.removeItem("refresh_token")
}

function redirectToLogin() {
  clearTokens()
  if (typeof window !== "undefined") {
    window.location.href = "/login"
  }
}

async function refreshAccessToken(): Promise<string | null> {
  const refreshToken = getRefreshToken()
  if (!refreshToken) {
    redirectToLogin()
    return null
  }

  try {
    const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })

    if (!response.ok) {
      redirectToLogin()
      return null
    }

    const data = await response.json()

    if (data.code === 0 && data.data) {
      setTokens(data.data.token, data.data.refresh_token)
      return data.data.token
    }

    redirectToLogin()
    return null
  } catch {
    redirectToLogin()
    return null
  }
}

async function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

export async function api<T = unknown>(
  endpoint: string,
  options: ApiOptions = {}
): Promise<T> {
  const {
    params,
    retry = 0,
    retryDelay = 1000,
    ...fetchOptions
  } = options

  const url = await buildUrl(endpoint, params)

  const headers: Record<string, string> = {
    ...(options.headers as Record<string, string> | undefined || {}),
  }
  if (!(options.body instanceof FormData) && !headers["Content-Type"]) {
    headers["Content-Type"] = "application/json"
  }

  const token = getToken()
  if (token) {
    headers["Authorization"] = `Bearer ${token}`
  }

  let lastError: Error | null = null

  for (let attempt = 0; attempt <= retry; attempt++) {
    try {
      const response = await fetch(url, {
        ...fetchOptions,
        headers,
      })

      if (response.status === 401) {
        if (!isRefreshing) {
          isRefreshing = true

          const newToken = await refreshAccessToken()

          isRefreshing = false

          if (newToken) {
            onTokenRefreshed(newToken)
            ;(headers as Record<string, string>)["Authorization"] = `Bearer ${newToken}`

            const retryResponse = await fetch(url, {
              ...fetchOptions,
              headers,
            })

            const data = await retryResponse.json().catch(() => null)

            if (!retryResponse.ok) {
              redirectToLogin()
              throw new ApiError(
                retryResponse.status,
                data?.code || retryResponse.status,
                data?.msg || `HTTP error! status: ${retryResponse.status}`,
                data
              )
            }

            const apiResponse = data as ApiResponse<T>

            if (apiResponse.code !== 0) {
              throw new ApiError(
                200,
                apiResponse.code,
                apiResponse.msg,
                apiResponse.data
              )
            }

            return apiResponse.data as T
          }

          redirectToLogin()
          throw new ApiError(401, 401, "Token refresh failed")
        }

        return new Promise((resolve, reject) => {
          subscribeTokenRefresh(async (newToken) => {
            try {
              ;(headers as Record<string, string>)["Authorization"] = `Bearer ${newToken}`
              const retryResponse = await fetch(url, {
                ...fetchOptions,
                headers,
              })
              const data = await retryResponse.json().catch(() => null)
              
              if (!retryResponse.ok) {
                redirectToLogin()
                reject(new ApiError(
                  retryResponse.status,
                  (data as ApiResponse<unknown>)?.code || retryResponse.status,
                  (data as ApiResponse<unknown>)?.msg || "Request failed"
                ))
                return
              }
              
              const apiResponse = data as ApiResponse<T>

              if (apiResponse?.code !== 0) {
                reject(new ApiError(200, apiResponse?.code || 0, apiResponse?.msg || "Request failed", apiResponse?.data))
              } else {
                resolve(apiResponse?.data as T)
              }
            } catch (err) {
              reject(err)
            }
          })
        })
      }

      const data = await response.json().catch(() => null)

      if (!response.ok) {
        throw new ApiError(
          response.status,
          data?.code || response.status,
          data?.msg || `HTTP error! status: ${response.status}`,
          data
        )
      }

      const apiResponse = data as ApiResponse<T>

      if (apiResponse.code !== 0) {
        throw new ApiError(
          200,
          apiResponse.code,
          apiResponse.msg,
          apiResponse.data
        )
      }

      return apiResponse.data as T
    } catch (err) {
      lastError = err as Error

      if (attempt < retry && !(err instanceof ApiError && err.status === 401)) {
        await delay(retryDelay * Math.pow(2, attempt))
        continue
      }

      throw lastError
    }
  }

  throw lastError
}

export function setAuthTokens(token: string, refreshToken?: string) {
  setTokens(token, refreshToken)
}

export function clearAuthTokens() {
  clearTokens()
}

export { getToken, getRefreshToken }

export const authApi = {
  login: (data: LoginRequest) =>
    api<LoginResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  register: (data: RegisterRequest) =>
    api<LoginResponse>("/auth/register", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  refresh: (data: RefreshRequest) =>
    api<RefreshResponse>("/auth/refresh", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  logout: () =>
    api("/auth/logout", {
      method: "POST",
    }),
}

export const userApi = {
  list: (params?: { keyword?: string; org_id?: number; page?: number; size?: number }) =>
    api<PageResponse<UserItem>>("/users", { params }),

  get: (id: number) =>
    api<UserItem>(`/users/${id}`),

  create: (data: Partial<UserItem> & { username: string, password?: string }) =>
    api<UserItem>("/users", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  updateStatus: (id: number, status: number) =>
    api(`/users/${id}/status`, {
      method: "PUT",
      body: JSON.stringify({ id, status }),
    }),

  patch: (id: number, data: Partial<UserItem>) =>
    api(`/users/${id}`, {
      method: "PATCH",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/users/${id}`, {
      method: "DELETE",
    }),

  getProfile: () =>
    api<UserItem>("/user/profile"),

  updateProfile: (data: { nickName: string; avatarUrl?: string }) =>
    api("/user/profile", {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  uploadAvatar: (file: File) => {
    const formData = new FormData()
    formData.append("file", file)
    return api<{ url: string }>("/user/avatar", {
      method: "POST",
      body: formData,
      headers: {
        // 不设置 Content-Type，让浏览器自动设置，包括 boundary
      },
    })
  },
}

export const menuApi = {
  list: (params?: { page?: number; size?: number; root?: boolean }) =>
    api<PageResponse<MenuItem>>("/menus", { params }),

  tree: () =>
    api<MenuItem[]>("/menus/tree"),

  get: (id: number) =>
    api<MenuItem>(`/menus/${id}`),

  create: (data: Partial<MenuItem>) =>
    api<MenuItem>("/menus", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<MenuItem>) =>
    api(`/menus/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/menus/${id}`, {
      method: "DELETE",
    }),
}

export const roleApi = {
  list: (params?: { keyword?: string; page?: number; size?: number }) =>
    api<PageResponse<RoleItem>>("/roles", { params }),

  get: (id: number) =>
    api<RoleItem>(`/roles/${id}`),

  create: (data: Partial<RoleItem>) =>
    api<RoleItem>("/roles", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<RoleItem>) =>
    api(`/roles/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  patch: (id: number, data: Partial<RoleItem>) =>
    api(`/roles/${id}`, {
      method: "PATCH",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/roles/${id}`, {
      method: "DELETE",
    }),

  getMenus: (id: number) =>
    api<{ menu_ids: number[] }>(`/roles/${id}/menus`),

  setMenus: (id: number, menuIds: number[]) =>
    api(`/roles/${id}/menus`, {
      method: "PUT",
      body: JSON.stringify({ menu_ids: menuIds }),
    }),

  getDataScopes: (id: number) =>
    api<{ org_ids: number[] }>(`/roles/${id}/data-scopes`),

  setDataScopes: (id: number, orgIds: number[]) =>
    api(`/roles/${id}/data-scopes`, {
      method: "PUT",
      body: JSON.stringify({ org_ids: orgIds }),
    }),

  getPermissions: (id: number) =>
    api<{ list: ResourceItem[] }>(`/roles/${id}/permissions`),

  setPermissions: (id: number, resourceIds: number[]) =>
    api(`/roles/${id}/permissions`, {
      method: "PUT",
      body: JSON.stringify({ resource_ids: resourceIds }),
    }),
}

export const organizationApi = {
  list: (params?: { keyword?: string; parent_id?: number; page?: number; size?: number }) =>
    api<PageResponse<OrganizationItem>>("/organizations", { params }),

  tree: () =>
    api<{ tree: OrganizationItem[] }>("/organizations/tree"),

  get: (id: number) =>
    api<OrganizationItem>(`/organizations/${id}`),

  create: (data: Partial<OrganizationItem>) =>
    api<OrganizationItem>("/organizations", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<OrganizationItem>) =>
    api(`/organizations/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/organizations/${id}`, {
      method: "DELETE",
    }),
}
export const dictApi = {
  list: (params?: { keyword?: string; page?: number; size?: number }) =>
    api<PageResponse<DictItem>>("/dicts", { params }),

  get: (id: number) =>
    api<DictItem>(`/dicts/${id}`),

  create: (data: { code: string; name: string; sort?: number; extend?: Record<string, unknown> }) =>
    api<DictItem>("/dicts", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: { name: string; sort?: number; status?: number; extend?: Record<string, unknown> }) =>
    api(`/dicts/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/dicts/${id}`, {
      method: "DELETE",
    }),

  // ???
  listItems: (dictId: number) =>
    api<PageResponse<DictValueItem>>(`/dicts/${dictId}/items`),

  createItem: (dictId: number, data: { code: string; name: string; sort?: number; extend?: Record<string, unknown> }) =>
    api<DictValueItem>(`/dicts/${dictId}/items`, {
      method: "POST",
      body: JSON.stringify(data),
    }),

  updateItem: (dictId: number, itemId: number, data: { name: string; sort?: number; status?: number; extend?: Record<string, unknown> }) =>
    api(`/dicts/${dictId}/items/${itemId}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  deleteItem: (dictId: number, itemId: number) =>
    api(`/dicts/${dictId}/items/${itemId}`, {
      method: "DELETE",
    }),
}

// ============================================
// 通用配置 (config)
// ============================================
export const configApi = {
  // 公共读（未登录可用）：按 group code 取所有 is_public 项
  getPublic: (group: string, tenantId?: number) => {
    const params: Record<string, string | number> = { group }
    if (tenantId) params.tenant_id = tenantId
    return api<PublicConfigResponse>("/config", { params })
  },

  // 管理端
  listGroups: () =>
    api<{ list: ConfigGroup[]; total: number }>("/config/groups"),

  createGroup: (data: {
    code: string
    name: string
    description?: string
    icon?: string
    sort?: number
    is_public?: boolean
  }) =>
    api<ConfigGroup>("/config/groups", {
      method: "POST",
      body: JSON.stringify(data),
    }),

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
  ) =>
    api<ConfigGroup>(`/config/groups/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  deleteGroup: (id: number) =>
    api(`/config/groups/${id}`, { method: "DELETE" }),

  listItemsByGroup: (groupId: number) =>
    api<{ list: ConfigItem[]; total: number }>(`/config/groups/${groupId}/items`),

  createItem: (
    groupId: number,
    data: {
      key: string
      value?: unknown
      default_value?: unknown
      type: ConfigItemType
      label?: string
      description?: string
      options?: ConfigOption[]
      validation?: ConfigValidation
      sort?: number
      is_public?: boolean
      is_readonly?: boolean
    }
  ) =>
    api<ConfigItem>(`/config/groups/${groupId}/items`, {
      method: "POST",
      body: JSON.stringify(data),
    }),

  listAllItems: () =>
    api<{ list: ConfigItem[]; total: number }>("/config/items"),

  updateItem: (
    id: number,
    data: {
      value?: unknown
      label?: string
      description?: string
      sort?: number
      is_public?: boolean
      is_readonly?: boolean
      status?: number
    }
  ) =>
    api<ConfigItem>(`/config/items/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  resetItem: (id: number) =>
    api<ConfigItem>(`/config/items/${id}/reset`, { method: "POST" }),

  deleteItem: (id: number) =>
    api(`/config/items/${id}`, { method: "DELETE" }),
}



export const tenantApi = {
  list: (params?: { page?: number; size?: number; keyword?: string; status?: number }) =>
    api<PageResponse<TenantItem>>("/tenants", { params }),

  get: (id: number) =>
    api<TenantItem>(`/tenants/${id}`),

  create: (data: Partial<TenantItem>) =>
    api<TenantItem>("/tenants", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<TenantItem>) =>
    api(`/tenants/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/tenants/${id}`, {
      method: "DELETE",
    }),
}



export const resourceApi = {
  list: (params?: { menu_id?: number; action?: string; page?: number; size?: number }) =>
    api<PageResponse<ResourceItem>>("/resources", { params }),

  get: (id: number) =>
    api<ResourceItem>(`/resources/${id}`),

  create: (data: Partial<ResourceItem>) =>
    api<ResourceItem>("/resources", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<ResourceItem>) =>
    api(`/resources/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/resources/${id}`, {
      method: "DELETE",
    }),

  byMenu: (menuId: number) =>
    api<ResourceItem[]>(`/resources/by-menu/${menuId}`),
}

export interface FrameItem {
  id: number
  tenant_id?: number
  category_id: number
  name: string
  description?: string
  preview_url?: string
  template_url?: string
  template_config?: {
    avatar_x: number
    avatar_y: number
    avatar_width: number
    avatar_height: number
  }
  type: string
  sort: number
  status: number
}

export interface FrameCategoryItem {
  id: number
  tenant_id?: number
  code: string
  name: string
  type: string
  sort: number
  status: number
}

export interface AvatarItem {
  id: number
  tenant_id?: number
  user_id: number
  category_id: number
  name: string
  source_url: string
  thumbnail_url?: string
  file_size?: number
  width?: number
  height?: number
  type: string
  is_public?: boolean
  like_count?: number
  view_count?: number
  sort?: number
  status: number
}

export interface AvatarCategoryItem {
  id: number
  tenant_id?: number
  code: string
  name: string
  icon?: string
  type: string
  sort: number
  status: number
}

export interface SpaceItem {
  id: number
  tenant_id?: number
  name: string
  description?: string
  frame_id?: number
  space_config?: {
    fields: Array<{
      key: string
      label: string
      required: boolean
      show: boolean
    }>
  }
  access_type: string
  invite_code?: string
  max_usage?: number
  usage_count?: number
  status: number
  start_at?: string
  end_at?: string
}

export interface GenerateAvatarResponse {
  id: number
  result_url: string
  share_text?: string
}

export const frameApi = {
  list: (params?: { category_id?: number; page?: number; size?: number }) =>
    api<PageResponse<FrameItem>>("/flag/frames", { params }),

  get: (id: number) =>
    api<FrameItem>(`/flag/frames/${id}`),

  create: (data: { name: string; category_id?: number; description?: string; preview_url?: string; template_url?: string; type?: string; sort?: number }) =>
    api<FrameItem>("/flag/frames", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: { id?: number; category_id?: number; name?: string; description?: string; preview_url?: string; template_url?: string; type?: string; sort?: number; status?: number }) =>
    api(`/flag/frames/${id}`, {
      method: "PUT",
      body: JSON.stringify({ id, ...data }),
    }),

  delete: (id: number) =>
    api(`/flag/frames/${id}`, {
      method: "DELETE",
    }),
}

export const frameCategoryApi = {
  list: () =>
    api<FrameCategoryItem[]>("/flag/frames-categories"),

  create: (data: { code: string; name: string; type?: string; sort?: number }) =>
    api<FrameCategoryItem>("/flag/frames-categories", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<FrameCategoryItem>) =>
    api(`/flag/frames-categories/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/flag/frames-categories/${id}`, {
      method: "DELETE",
    }),
}

export const avatarApi = {
  list: (params?: { category_id?: number; user_id?: number; type?: string; page?: number; size?: number }) =>
    api<PageResponse<AvatarItem>>("/flag/avatars", { params }),

  get: (id: number) =>
    api<AvatarItem>(`/flag/avatars/${id}`),

  create: (data: { source_url: string; category_id?: number; name?: string; thumbnail_url?: string; file_size?: number; width?: number; height?: number; is_public?: boolean }) =>
    api<AvatarItem>("/flag/avatars", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: { name?: string; category_id?: number; source_url: string; thumbnail_url?: string; is_public?: boolean; status?: number }) =>
    api(`/flag/avatars/${id}`, {
      method: "PUT",
      body: JSON.stringify({ id, ...data }),
    }),

  delete: (id: number) =>
    api(`/flag/avatars/${id}`, {
      method: "DELETE",
    }),
}

export const avatarCategoryApi = {
  list: (params?: { type?: string }) =>
    api<AvatarCategoryItem[]>("/flag/avatar-categories", { params }),

  create: (data: { code: string; name: string; icon?: string; type?: string; sort?: number }) =>
    api<AvatarCategoryItem>("/flag/avatar-categories", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<AvatarCategoryItem>) =>
    api(`/flag/avatar-categories/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/flag/avatar-categories/${id}`, {
      method: "DELETE",
    }),
}

export const spaceApi = {
  list: () =>
    api<SpaceItem[]>("/flag/spaces"),

  get: (code: string) =>
    api<SpaceItem>(`/flag/spaces/${code}`),

  create: (data: { name: string; description?: string; frame_id?: number; access_type?: string; start_at?: string; end_at?: string }) =>
    api<SpaceItem>("/flag/spaces", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<SpaceItem>) =>
    api(`/flag/spaces/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api(`/flag/spaces/${id}`, {
      method: "DELETE",
    }),
}

export interface AssetUploadResponse {
  id: number
  url: string
}

export const assetApi = {
  upload: (file: File) => {
    const formData = new FormData()
    formData.append("file", file)
    const token = getToken()

    const headers: HeadersInit = {}
    if (token) {
      (headers as Record<string, string>)["Authorization"] = `Bearer ${token}`
    }

    return fetch(`${API_BASE_URL}/asset/upload`, {
      method: "POST",
      headers,
      body: formData,
    }).then(async (response) => {
      const data = await response.json()
      if (!response.ok) {
        throw new ApiError(
          response.status,
          data?.code || response.status,
          data?.msg || `Upload failed: ${response.status}`,
          data
        )
      }
      const apiResponse = data as ApiResponse<AssetUploadResponse>
      if (apiResponse.code !== 0) {
        throw new ApiError(200, apiResponse.code, apiResponse.msg, apiResponse.data)
      }
      return apiResponse.data as AssetUploadResponse
    })
  },
}

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
  getCacheInfo: () =>
    api<CacheInfo>("/system/cache/info"),

  getCacheKeys: (pattern: string = "*") =>
    api<string[]>("/system/cache/keys", { params: { pattern } }),

  getCacheValue: (key: string) =>
    api<CacheValue>(`/system/cache/value/${encodeURIComponent(key)}`),

  deleteCacheKey: (key: string) =>
    api(`/system/cache/keys/${encodeURIComponent(key)}`, {
      method: "DELETE",
    }),
}

export { ApiError, type MenuItem as MenuItemType }