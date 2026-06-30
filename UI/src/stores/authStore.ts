import { create } from "zustand"
import { persist } from "zustand/middleware"
import {
  authApi,
  setAuthTokens,
  clearAuthTokens,
  ApiError,
  type LoginScope,
  type TenantIdentity,
  type LoginPrecheckResponse,
  tenantApi,
} from "@/api"
import type { LoginResponse } from "@/api"

interface User {
  code: string
  id: number
  role: string
  /** scope=platform 时固定 0；scope=tenant 时为真实 tenant_id */
  tenant_id: number
  nickname?: string
  real_name?: string
  avatar?: string
  email?: string
  platform_roles?: string[]
  /** 资源权限码（"resource:action"），与后端 LoginResponse.user.permissions 对齐 */
  permissions?: string[]
}

interface AuthState {
  token: string | null
  refreshToken: string | null
  user: User | null
  /** 登录作用域：tenant / platform / null（未登录） */
  scope: LoginScope | null
  isAuthenticated: boolean
  isLoading: boolean
  error: string | null
  lastApiError: string | null

  // === 路径 B 多身份支持 ===
  /** 账号可用 tenant 身份缓存（precheck 后填，logout 清空） */
  availableIdentities: TenantIdentity[]
  /** 账号是否有 platform 角色（precheck 后填） */
  platformAvailable: boolean
  /** 账号的 platform 角色列表（precheck 后填） */
  availablePlatformRoles: string[]
  /** 当前账号的 account_id（precheck 后填） */
  accountId: number | null

  // === 平台管理员模拟登录租户（super_admin 专用） ===
  /**
   * 非空 = 当前正在模拟某个租户。保存原 platform tokens（用于"退出模拟"时
   * 调 /auth/refresh 恢复）+ 模拟 token + 目标租户信息。
   */
  impersonation: {
    originalTokens: { token: string; refreshToken: string }
    tenantId: number
    tenantName: string
    impersonatedUserId: number
    impersonatedBy: number
    startedAt: number
  } | null

  /** 租户域登录（业务用户登录）。跳转到 /app/dashboard。 */
  tenantLogin: (
    account: string,
    password: string,
    tenantId: number
  ) => Promise<boolean>
  /** 平台域登录（super_admin 登录）。跳转到 /platform/dashboard。 */
  platformLogin: (account: string, password: string) => Promise<boolean>
  /** 登录前置检查：列账号所有可用身份，不签 token。 */
  loginPrecheck: (
    account: string,
    password: string
  ) => Promise<LoginPrecheckResponse | null>
  /** 选择一个 tenant 身份签 token（多身份登录流的第二步）。 */
  selectTenant: (
    account: string,
    password: string,
    tenantId: number
  ) => Promise<boolean>
  /** 切租户（已登录后用 refresh_token 换新租户的 token，无需密码）。 */
  switchTenant: (tenantId: number) => Promise<boolean>
  /** 清空 identities 缓存（强制下次重新 precheck）。 */
  clearIdentities: () => void
  /**
   * 平台管理员模拟登录租户。
   * 流程：调 /platform/tenants/:id/impersonate → 保存原 tokens → 写入模拟 token → 跳转租户域。
   */
  startImpersonation: (tenantId: number, tenantName: string) => Promise<boolean>
  /**
   * 退出模拟：用原 platform refresh_token 调 /auth/refresh 恢复 platform token。
   * 不走 /auth/logout（会同时撤销原 session）。
   */
  stopImpersonation: () => Promise<boolean>
  logout: () => void
  clearError: () => void
  clearApiError: () => void
}

/**
 * 判断当前用户是否拥有指定权限码（"resource:action" 形式）。
 *
 * 通配符语义与后端 framework/pkg/permission.HasPermission 保持一致：
 *   - "menu:create"    精确匹配
 *   - "menu:*"         资源级通配（菜单所有操作）
 *   - "*:*"            全局通配（超级管理员等）
 *
 * 设计意图：0024+ 前后端统一按权限码（不是角色名）判定 UI 可见性。
 * super_admin 因为持有 "*:*" 通配而自然通过任何权限检查，
 * 不需要在组件里写 isSuperAdmin 这种硬编码。
 *
 * 参数 code 格式："resource:action"，与 permission.P(Res, Act) 输出对齐。
 */
export function hasPermission(
  user: Pick<User, "permissions"> | null | undefined,
  code: string
): boolean {
  if (!user?.permissions?.length) return false
  const set = user.permissions
  if (set.includes(code)) return true
  const idx = code.indexOf(":")
  if (idx > 0) {
    if (set.includes(code.slice(0, idx) + ":*")) return true
  }
  return set.includes("*:*")
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      refreshToken: null,
      user: null,
      scope: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,
      lastApiError: null,

      // 路径 B 多身份默认值
      availableIdentities: [],
      platformAvailable: false,
      availablePlatformRoles: [],
      accountId: null,

      // 模拟登录默认 null
      impersonation: null,

      tenantLogin: async (account, password, tenantId) => {
        set({ isLoading: true, error: null })
        try {
          const data: LoginResponse = await authApi.tenantLogin({
            account,
            password,
            tenant_id: tenantId,
          })

          setAuthTokens(data.token, data.refresh_token)

          set({
            token: data.token,
            refreshToken: data.refresh_token,
            user: data.user,
            scope: data.scope,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          })

          return true
        } catch (err) {
          const errorMessage =
            err instanceof ApiError
              ? err.message
              : err instanceof Error
                ? err.message
                : "登录失败"
          set({
            isLoading: false,
            error: errorMessage,
            isAuthenticated: false,
            token: null,
            refreshToken: null,
            user: null,
            scope: null,
          })
          return false
        }
      },

      platformLogin: async (account, password) => {
        set({ isLoading: true, error: null })
        try {
          const data: LoginResponse = await authApi.platformLogin({
            account,
            password,
          })

          setAuthTokens(data.token, data.refresh_token)

          set({
            token: data.token,
            refreshToken: data.refresh_token,
            user: data.user,
            scope: data.scope,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          })

          return true
        } catch (err) {
          const errorMessage =
            err instanceof ApiError
              ? err.message
              : err instanceof Error
                ? err.message
                : "登录失败"
          set({
            isLoading: false,
            error: errorMessage,
            isAuthenticated: false,
            token: null,
            refreshToken: null,
            user: null,
            scope: null,
          })
          return false
        }
      },

      loginPrecheck: async (account, password) => {
        set({ isLoading: true, error: null })
        try {
          const data = await authApi.loginPrecheck({ account, password })

          set({
            accountId: data.account_id,
            availableIdentities: data.tenant_identities,
            platformAvailable: data.platform_available,
            availablePlatformRoles: data.platform_roles ?? [],
            isLoading: false,
            error: null,
          })

          return data
        } catch (err) {
          const errorMessage =
            err instanceof ApiError
              ? err.message
              : err instanceof Error
                ? err.message
                : "登录前置检查失败"
          set({
            isLoading: false,
            error: errorMessage,
            availableIdentities: [],
            platformAvailable: false,
            availablePlatformRoles: [],
            accountId: null,
          })
          return null
        }
      },

      selectTenant: async (account, password, tenantId) => {
        set({ isLoading: true, error: null })
        try {
          const data: LoginResponse = await authApi.selectTenant({
            account,
            password,
            tenant_id: tenantId,
          })

          setAuthTokens(data.token, data.refresh_token)

          set({
            token: data.token,
            refreshToken: data.refresh_token,
            user: data.user,
            scope: data.scope,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          })

          return true
        } catch (err) {
          const errorMessage =
            err instanceof ApiError
              ? err.message
              : err instanceof Error
                ? err.message
                : "登录失败"
          set({
            isLoading: false,
            error: errorMessage,
            isAuthenticated: false,
          })
          return false
        }
      },

      switchTenant: async (tenantId) => {
        const { refreshToken, availableIdentities, user } = get()
        if (!refreshToken) {
          set({ error: "未登录" })
          return false
        }

        // 在缓存里找目标 tenant 的 identity（用于更新 user.code / role / id）
        const target = availableIdentities.find((i) => i.tenant_id === tenantId)
        if (!target) {
          set({ error: "目标租户不在账号可用身份列表中，请重新登录" })
          return false
        }

        set({ isLoading: true, error: null })
        try {
          const data = await authApi.refresh({
            refresh_token: refreshToken,
            tenant_id: tenantId,
          })

          setAuthTokens(data.token, data.refresh_token)

          // 更新 user：tenant_id / role / code / id 都从目标 identity 取
          set({
            token: data.token,
            refreshToken: data.refresh_token ?? refreshToken,
            user: user
              ? {
                  ...user,
                  id: target.user_id,
                  tenant_id: target.tenant_id,
                  code: target.user_code,
                  role: target.role,
                }
              : user,
            isLoading: false,
            error: null,
          })

          return true
        } catch (err) {
          const errorMessage =
            err instanceof ApiError
              ? err.message
              : err instanceof Error
                ? err.message
                : "切换租户失败"
          set({
            isLoading: false,
            error: errorMessage,
          })
          return false
        }
      },

      clearIdentities: () => {
        set({
          availableIdentities: [],
          platformAvailable: false,
          availablePlatformRoles: [],
          accountId: null,
        })
      },

      logout: () => {
        clearAuthTokens()
        set({
          token: null,
          refreshToken: null,
          user: null,
          scope: null,
          isAuthenticated: false,
          error: null,
          lastApiError: null,
          availableIdentities: [],
          platformAvailable: false,
          availablePlatformRoles: [],
          accountId: null,
          impersonation: null,
        })
      },

      startImpersonation: async (tenantId, tenantName) => {
        const {
          token: currentToken,
          refreshToken: currentRefresh,
          user,
        } = get()
        if (!currentToken || !currentRefresh) {
          set({ error: "未登录" })
          return false
        }
        set({ isLoading: true, error: null })
        try {
          // 用当前 platform token 调 impersonate 端点（后端会保留原 platform session）
          const data = await tenantApi.impersonate(tenantId)

          // 1. 把当前 platform tokens 保存到 impersonation.originalTokens
          //    （退出模拟时用 refresh_token 调 /auth/refresh 即可恢复 platform token）
          // 2. 把模拟 token 写到 localStorage + authStore 主 token
          setAuthTokens(data.token, data.refresh_token)
          set({
            token: data.token,
            refreshToken: data.refresh_token,
            user: user
              ? {
                  ...user,
                  id: data.impersonated_user_id,
                  tenant_id: data.tenant_id,
                  role: "admin",
                  // 模拟期间 PlatformRoles 留空（不走 super_admin 短路）
                  platform_roles: [],
                }
              : user,
            scope: "tenant",
            isAuthenticated: true,
            isLoading: false,
            error: null,
            impersonation: {
              originalTokens: {
                token: currentToken,
                refreshToken: currentRefresh,
              },
              tenantId: data.tenant_id,
              tenantName: data.tenant_name ?? tenantName,
              impersonatedUserId: data.impersonated_user_id,
              impersonatedBy: data.impersonated_by,
              startedAt: Date.now(),
            },
          })

          return true
        } catch (err) {
          const errorMessage =
            err instanceof ApiError
              ? err.message
              : err instanceof Error
                ? err.message
                : "模拟登录失败"
          set({
            isLoading: false,
            error: errorMessage,
          })
          return false
        }
      },

      stopImpersonation: async () => {
        const { impersonation } = get()
        if (!impersonation) {
          return false
        }
        set({ isLoading: true, error: null })
        try {
          // 用原 platform refresh_token 调 /auth/refresh（不传 tenant_id）恢复 platform token
          const data = await authApi.refresh({
            refresh_token: impersonation.originalTokens.refreshToken,
          })

          setAuthTokens(data.token, data.refresh_token)

          set({
            token: data.token,
            refreshToken:
              data.refresh_token ?? impersonation.originalTokens.refreshToken,
            scope: "platform",
            isLoading: false,
            error: null,
            impersonation: null,
          })

          return true
        } catch (err) {
          const errorMessage =
            err instanceof ApiError
              ? err.message
              : err instanceof Error
                ? err.message
                : "退出模拟失败"
          set({
            isLoading: false,
            error: errorMessage,
          })
          return false
        }
      },

      clearError: () => {
        set({ error: null })
      },

      clearApiError: () => {
        set({ lastApiError: null })
      },
    }),
    {
      name: "auth-storage",
      partialize: (state) => ({
        token: state.token,
        refreshToken: state.refreshToken,
        user: state.user,
        scope: state.scope,
        isAuthenticated: state.isAuthenticated,
        availableIdentities: state.availableIdentities,
        platformAvailable: state.platformAvailable,
        availablePlatformRoles: state.availablePlatformRoles,
        accountId: state.accountId,
        // 模拟登录状态必须持久化：刷新页面后 stopImpersonation 仍能拿到原 platform refresh_token
        impersonation: state.impersonation,
      }),
      onRehydrateStorage: () => (state) => {
        const token = localStorage.getItem("token")
        if (!token && state) {
          state.isAuthenticated = false
          state.token = null
          state.refreshToken = null
          state.user = null
          state.scope = null
          state.availableIdentities = []
          state.platformAvailable = false
          state.availablePlatformRoles = []
          state.accountId = null
          state.impersonation = null
        }
      },
    }
  )
)
