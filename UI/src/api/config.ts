// 配置中心（group + item）
//
// 路由约定（重构后）：
//
//   - public:   GET /api/v1/public/configs                  （无需 auth）
//   - tenant:   /api/v1/configs                            （Auth + RequireTenantContext，业务消费 + override）
//   - platform: /api/v1/platform/configs                   （RequirePlatformRole("super_admin")）
//
// 响应约定：resp.Success(c, x) 包成 {code:0,msg:"ok",data:x}，api<T>() 解出 data
//   - ListGroups/ListItems 返回裸数组，不是 {list,total}
//   - GetPublic 返回 {group,values}
//
// 路由约定（与 [server/framework/framework.go](../../server/framework/framework.go) 同步）：
//   - public   → /api/v1/public/configs        （无需 auth）
//   - tenant   → /api/v1/configs               （Auth + RequireTenantContext）
//   - platform → /api/v1/platform/configs      （Auth + RequirePlatformRole("super_admin")）

import { api } from "./common"

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
  scope?: string // 'platform' | 'tenant'
  visibility?: string // 'all' | 'whitelist' | 'blacklist'
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

// 创建 platform group 请求体（后端 createGroupRequest 字段）
export interface CreatePlatformGroupRequest {
  code: string
  name: string
  description?: string
  icon?: string
  sort?: number
  is_system?: boolean
  is_public?: boolean
}

// 更新 platform group 请求体（字段均为可选）
export interface UpdatePlatformGroupRequest {
  name?: string
  description?: string
  icon?: string
  sort?: number
  is_public?: boolean
  visibility?: "public" | "tenant_only" | "hidden"
  status?: number
}

// 创建 platform item 请求体
export interface CreatePlatformItemRequest {
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
  is_system?: boolean
}

// 更新 platform item 请求体
export interface UpdatePlatformItemRequest {
  value?: unknown
  label?: string
  description?: string
  sort?: number
  is_public?: boolean
  is_readonly?: boolean
  status?: number
}

export const configApi = {
  // =============== Public（无需 auth）===============
  //
  // 后端 /public/configs 忽略 ?group=，返回某租户的全部公开项；
  // 前端 store 按 group 参数作缓存 key 的命名空间。
  // 匿名场景下可传 tenantId 通过 X-Tenant-ID header 兜底。
  getPublic: (_group: string, tenantId?: number) =>
    api<PublicConfigResponse>("/public/configs", {
      headers: tenantId ? { "X-Tenant-ID": String(tenantId) } : undefined,
    }),

  // =============== Tenant 域（业务消费 / 租户自建读）===============
  //
  // 受 RLS + RequireTenantContext 约束，
  // 租户看到的是 platform 可见 group ∪ 自己 tenant scope group。
  listGroups: () => api<ConfigGroup[]>("/configs"),

  // 取某 group 下所有 item（含 platform + tenant override 合并）
  listItemsByGroup: (groupId: number) =>
    api<ConfigItem[]>(`/configs/${groupId}/items`),

  // 租户对某 platform item 的值覆盖（"重置" = 删 override，恢复平台默认值）
  deleteOverride: (groupId: number, itemId: number) =>
    api(`/configs/${groupId}/items/${itemId}/override`, {
      method: "DELETE",
    }),

  // =============== Platform 域（super_admin CRUD）===============
  //
  // 挂在 /api/v1/platform/configs 下，需 RequirePlatformRole("super_admin")。
  // 前端如果用租户 token 调用会 403——这是后端设计。
  createPlatformGroup: (data: CreatePlatformGroupRequest) =>
    api<ConfigGroup>("/platform/configs", {
      method: "POST",
      params: { scope: "platform" },
      body: JSON.stringify(data),
    }),

  updatePlatformGroup: (id: number, data: UpdatePlatformGroupRequest) =>
    api<ConfigGroup>(`/platform/configs/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  deletePlatformGroup: (id: number) =>
    api(`/platform/configs/${id}`, { method: "DELETE" }),

  listPlatformItems: (groupId: number) =>
    api<ConfigItem[]>(`/platform/configs/${groupId}/items`),

  createPlatformItem: (groupId: number, data: CreatePlatformItemRequest) =>
    api<ConfigItem>(`/platform/configs/${groupId}/items`, {
      method: "POST",
      body: JSON.stringify(data),
    }),

  updatePlatformItem: (
    groupId: number,
    itemId: number,
    data: UpdatePlatformItemRequest
  ) =>
    api<ConfigItem>(`/platform/configs/${groupId}/items/${itemId}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  deletePlatformItem: (groupId: number, itemId: number) =>
    api(`/platform/configs/${groupId}/items/${itemId}`, {
      method: "DELETE",
    }),
}