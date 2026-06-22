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

  /** 租户域登录（业务用户登录）。跳转到 /app/dashboard。 */
  tenantLogin: (account: string, password: string, tenantId: number) => Promise<boolean>
  /** 平台域登录（super_admin 登录）。跳转到 /platform/dashboard。 */
  platformLogin: (account: string, password: string) => Promise<boolean>
  /** 登录前置检查：列账号所有可用身份，不签 token。 */
  loginPrecheck: (account: string, password: string) => Promise<LoginPrecheckResponse | null>
  /** 选择一个 tenant 身份签 token（多身份登录流的第二步）。 */
  selectTenant: (account: string, password: string, tenantId: number) => Promise<boolean>
  /** 切租户（已登录后用 refresh_token 换新租户的 token，无需密码）。 */
  switchTenant: (tenantId: number) => Promise<boolean>
  /** 清空 identities 缓存（强制下次重新 precheck）。 */
  clearIdentities: () => void
  logout: () => void
  clearError: () => void
  clearApiError: () => void
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
        })
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
        }
      },
    },
  ),
)