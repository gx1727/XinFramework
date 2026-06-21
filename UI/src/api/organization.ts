// 组织

import { api, type PageResponse } from "./common"

export interface OrganizationItem {
  id: number
  tenant_id?: number
  parent_id: number
  code: string
  name: string
  type?: string
  sort: number
  status: number
  path?: string
  ancestors?: string
  created_at?: string
  updated_at?: string
  children?: OrganizationItem[]
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