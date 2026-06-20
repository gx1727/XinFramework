// 配置中心（group + item）

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
